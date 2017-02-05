package main

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
)

// ErrSelfRating is returned when trying to rate yourself
var ErrSelfRating = errors.New("self-rating is not allowed")

// Scores are kept in the `scores:{uid}` bucket, with the key being `{reason}`
// where reason is the string representing the reason for a score and uid is
// the uid of the user with the scores.

func scoreHandler(cmd string, session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	adjust := int64(1)
	adjSym := ":+1:"
	var err error

	switch cmd {
	case "score":
		err = retrieveScores(session, message, args...)
	case "-":
		adjust = -1
		adjSym = ":-1:"
		fallthrough
	case "+":
		err = adjustScore(adjust, session, message, args...)
		if err == nil {
			_, err = session.ChannelMessageSend(message.ChannelID, ":trophy: score logged! "+adjSym)
		} else if err == ErrSelfRating {
			_, err = session.ChannelMessageSend(message.ChannelID, ":poop: You can't rate yourself, scumbag!")
		} else {
			log.Print(err)
			_, err = session.ChannelMessageSend(message.ChannelID, ":trophy: I couldn't log this score due to an error. :sob:")
		}
	case "top":
		err = retrieveTopScores(session, message)
	}

	if err != nil {
		log.Print(err)
	}
}

// KeyScore is a Key and total scores logged against it
type KeyScore struct {
	Key   string
	Score int64
}

// KeyScoreList is a list of KeyScore's
type KeyScoreList []KeyScore

func (rsl KeyScoreList) Len() int           { return len(rsl) }
func (rsl KeyScoreList) Less(i, j int) bool { return rsl[i].Score < rsl[j].Score }
func (rsl KeyScoreList) Swap(i, j int)      { rsl[i], rsl[j] = rsl[j], rsl[i] }

func getKeyScoreList(bucketname []byte) (scores KeyScoreList) {
	db.View(func(tx *bolt.Tx) (err error) {
		// Assume bucket exists and has keys
		b := tx.Bucket(bucketname)
		if b == nil {
			return
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var intV int64
			intV, err = strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				return err
			}
			scores = append(scores, KeyScore{Key: string(k), Score: intV})
		}

		return nil
	})

	// Sort it correctly
	sort.Sort(sort.Reverse(scores))

	return
}

// retrieveScores gets all the scores for a given user
func retrieveScores(session *discordgo.Session, message *discordgo.MessageCreate, args ...string) error {
	user := message.Author
	if len(args) == 1 {
		// If the first mention isn't the first arg, its an invalid command
		uid := getUIDFromMention(args[0])
		if uid != message.Mentions[0].ID {
			_, err := session.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":trophy: no idea who %v is", uid))
			return err
		}

		user = message.Mentions[0]
	}

	guildid, err := getGuildIDFromMessage(session, message)
	if err != nil {
		return err
	}
	scores := getKeyScoreList([]byte("guild:" + guildid + ":scores:" + user.ID))

	if len(scores) == 0 {
		_, err = session.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":trophy: %v has not been rated", user.Username))
		return err
	}

	output := make([]string, len(scores)+1)
	var total int64
	for i, score := range scores {
		output[1+i] = fmt.Sprintf("**%v** for `%v`", score.Score, score.Key)
		total += score.Score
	}

	output[0] = fmt.Sprintf(":trophy: %v has been rated %v for the following:\n", user.Username, total)

	_, err = session.ChannelMessageSend(message.ChannelID, strings.Join(output, "\n"))
	return err
}

// retrieveTopScores lists the users by total score
func retrieveTopScores(session *discordgo.Session, message *discordgo.MessageCreate) error {
	guildid, err := getGuildIDFromMessage(session, message)
	if err != nil {
		return err
	}

	scores := getKeyScoreList([]byte("guild:" + guildid + ":scorestotals"))

	if len(scores) == 0 {
		_, err = session.ChannelMessageSend(message.ChannelID, ":trophy: No-one has been rated")
		return err
	}

	output := make([]string, len(scores)+1)
	output[0] = ":trophy: Everyone has been rated for the following:"
	for i, score := range scores {
		user, uerr := session.User(score.Key)
		if uerr == nil {
			score.Key = user.Username
		}

		output[1+i] = fmt.Sprintf("**%v** has a score of **%v**", score.Key, score.Score)
	}

	_, err = session.ChannelMessageSend(message.ChannelID, strings.Join(output, "\n"))
	return err
}

func adjustScore(adjust int64, session *discordgo.Session, message *discordgo.MessageCreate, args ...string) error {
	// First, we check to see there has at least been one mention
	if len(message.Mentions) == 0 {
		return errors.New("no-one was mentioned in this score adjustment")
	}

	// Next, we check the first two args are that mention, and "for"
	parts := strings.SplitN(args[0], " ", 3)
	if len(parts) < 3 {
		return errors.New("missing reason")
	}

	// If the first mention isn't the first arg, its an invalid command
	user := getUIDFromMention(parts[0])
	if user != message.Mentions[0].ID || parts[1] != "for" {
		return fmt.Errorf("invalid arguments: %#v (%#v)", parts, message.Mentions[0].ID)
	}

	// No self-rating
	if user == message.Author.ID {
		return ErrSelfRating
	}

	guildid, err := getGuildIDFromMessage(session, message)
	if err != nil {
		return err
	}

	bucketname := []byte("guild:" + guildid + ":scores:" + message.Mentions[0].ID)
	key := []byte(cleanDiscordString(parts[2]))
	totalsbucketname := []byte("guild:" + guildid + ":scorestotals")
	totalskey := []byte(message.Mentions[0].ID)

	return db.Update(func(tx *bolt.Tx) error {
		// Individual reasons
		bucket, err := tx.CreateBucketIfNotExists(bucketname)
		if err != nil {
			return err
		}

		value := bucket.Get(key)
		if value != nil {
			intVal, serr := strconv.ParseInt(string(value), 10, 64)
			if serr != nil {
				return serr
			}

			if err = bucket.Put(key, []byte(strconv.FormatInt(intVal+adjust, 10))); err != nil {
				return err
			}
		} else {
			if err = bucket.Put(key, []byte(strconv.FormatInt(adjust, 10))); err != nil {
				return err
			}
		}

		// User totals
		bucket, err = tx.CreateBucketIfNotExists(totalsbucketname)
		if err != nil {
			return err
		}

		value = bucket.Get(totalskey)
		if value != nil {
			intVal, serr := strconv.ParseInt(string(value), 10, 64)
			if serr != nil {
				return serr
			}

			if err = bucket.Put(totalskey, []byte(strconv.FormatInt(intVal+adjust, 10))); err != nil {
				return err
			}
		} else {
			if err = bucket.Put(totalskey, []byte(strconv.FormatInt(adjust, 10))); err != nil {
				return err
			}
		}

		return nil
	})
}
