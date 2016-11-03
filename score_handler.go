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

// Scores are kept in the `scores:{uid}` bucket, with the key being `{reason}`
// where reason is the string representing the reason for a score and uid is
// the uid of the user with the scores.

func scoreHandler(cmd string, session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	adjust := int64(1)
	adjSym := ":+1:"
	switch cmd {
	case "score":
		retrieveScores(session, message, args...)
	case "-":
		adjust = -1
		adjSym = ":-1:"
		fallthrough
	case "+":
		if err := adjustScore(adjust, session, message, args...); err != nil {
			log.Print(err)
			_, _ = session.ChannelMessageSend(message.ChannelID, ":trophy: I couldn't log this score due to an error. :sob:")
			return
		}

		_, _ = session.ChannelMessageSend(message.ChannelID, ":trophy: score logged! "+adjSym)
	case "top":
		retrieveTopScores(session, message)
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
func retrieveScores(session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	user := message.Author
	if len(args) == 1 {
		// If the first mention isn't the first arg, its an invalid command
		if args[0] != fmt.Sprintf("<@%s>", message.Mentions[0].ID) {
			_, _ = session.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":trophy: no idea who %v is", args[0]))
			return
		}

		user = message.Mentions[0]
	}

	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		log.Print(err)
		_, _ = session.ChannelMessageSend(message.ChannelID, ":trophy: I can't get the scores due to an error. :sob:")
		return
	}
	scores := getKeyScoreList([]byte("guild:" + channel.GuildID + ":scores:" + user.ID))

	if len(scores) == 0 {
		_, _ = session.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":trophy: %v has not been rated", user.Username))
		return
	}

	output := make([]string, len(scores)+1)
	var total int64
	for i, score := range scores {
		output[1+i] = fmt.Sprintf("**%v** for `%v`", score.Score, score.Key)
		total += score.Score
	}

	output[0] = fmt.Sprintf(":trophy: %v has been rated %v for the following:\n", user.Username, total)

	_, _ = session.ChannelMessageSend(message.ChannelID, strings.Join(output, "\n"))
	return
}

// retrieveTopScores lists the users by total score
func retrieveTopScores(session *discordgo.Session, message *discordgo.MessageCreate) {
	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		log.Print(err)
		_, _ = session.ChannelMessageSend(message.ChannelID, ":trophy: I can't get the top scores due to an error. :sob:")
		return
	}

	scores := getKeyScoreList([]byte("guild:" + channel.GuildID + ":scorestotals"))

	if len(scores) == 0 {
		_, _ = session.ChannelMessageSend(message.ChannelID, ":trophy: No-one has been rated")
		return
	}

	output := make([]string, len(scores)+1)
	output[0] = ":trophy: Everyone has been rated for the following:"
	for i, score := range scores {
		user, err := session.User(score.Key)
		if err == nil {
			score.Key = user.Username
		}

		output[1+i] = fmt.Sprintf("**%v** has a score of **%v**", score.Key, score.Score)
	}

	_, _ = session.ChannelMessageSend(message.ChannelID, strings.Join(output, "\n"))
	return
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
	if parts[0] != fmt.Sprintf("<@%s>", message.Mentions[0].ID) || parts[1] != "for" {
		return fmt.Errorf("invalid arguments: %#v (%#v)", parts, message.Mentions[0].ID)
	}

	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		return err
	}

	bucketname := []byte("guild:" + channel.GuildID + ":scores:" + message.Mentions[0].ID)
	key := []byte(strings.TrimSpace(parts[2]))
	totalsbucketname := []byte("guild:" + channel.GuildID + ":scorestotals")
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
