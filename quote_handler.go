package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
)

const (
	cmdGet  = "get"
	cmdAdd  = "add"
	cmdDel  = "del"
	cmdList = "list"
)

// Quotes are kept in the `quotes:{uid}` bucket, with the key being `{id}` - a
// new id

func quoteHandler(session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	var err error

	// Parse the quote command, which will be one of the following forms:
	// * [verb] [user] [args]
	// * [user]
	// * [verb] [args]
	// * [verb]
	// * nothing
	cmd := cmdGet
	if len(args) > 0 {
		args = strings.Split(args[0], " ")
		cmd, args = args[0], args[1:]
	}

	var uid string
	user := message.Author

	switch cmd {
	case cmdAdd, cmdDel, cmdGet, cmdList:
		// The next argument must be a user (or nothing)
		if len(args) > 0 {
			uid = getUIDFromMention(args[0])
			if len(message.Mentions) != 0 && uid == message.Mentions[0].ID {
				user = message.Mentions[0]
				args = args[1:]
			}
		}
	default:
		// In the default case, the only valid argument is a user (get is implied)
		uid := getUIDFromMention(cmd)
		if len(message.Mentions) == 0 || uid != message.Mentions[0].ID {
			log.Printf("invalid quote command %q", cmd)
			return
		}
		user = message.Mentions[0]
		cmd = cmdGet
	}

	switch cmd {
	case cmdAdd:
		err = addUserQuote(session, message, user, strings.Join(args, " "))
	case cmdDel:
		err = delUserQuote(session, message, user, strings.Join(args, " "))
	case cmdList:
		err = listUserQuotes(session, message, user)
	case cmdGet:
		err = getUserQuote(session, message, user)
	}

	if err != nil {
		_, err = session.ChannelMessageSend(message.ChannelID, ":poop: I couldn't perform the requested quote command")
		log.Print(err)
	}
}

func quotesBucket(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User) ([]byte, error) {
	guildid, err := getGuildIDFromMessage(s, m)
	if err != nil {
		return nil, err
	}
	return []byte("guild:" + guildid + ":quotes:" + u.ID), nil
}

func addUserQuote(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User, quote string) error {
	bucket, err := quotesBucket(s, m, u)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}

		id, _ := b.NextSequence()

		// Persist bytes to users bucket.
		log.Printf("putting %q into %v--%v", quote, string(bucket), id)
		return b.Put(itob(int(id)), []byte(quote))
	})
}

func delUserQuote(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User, quoteID string) error {
	bucket, err := quotesBucket(s, m, u)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}

		// Delete entry
		log.Printf("delete %v--%v", string(bucket), quoteID)
		return b.Delete([]byte(quoteID))
	})
}

func listUserQuotes(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User) error {
	bucket, err := quotesBucket(s, m, u)
	if err != nil {
		return err
	}

	return db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}

		c := b.Cursor()

		quotes := fmt.Sprintf(":speech_balloon: All quotes by %v\n", u.Username)
		for k, v := c.First(); k != nil; k, v = c.Next() {
			i, _ := binary.ReadVarint(bytes.NewBuffer(k))
			quotes += fmt.Sprintf("  %v: %q\n", i, string(v))
		}

		log.Printf("got quotes: %v", quotes)
		_, err = s.ChannelMessageSend(m.ChannelID, quotes)

		return err
	})
}

func getUserQuote(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User) error {
	bucket, err := quotesBucket(s, m, u)
	if err != nil {
		return err
	}

	return db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}

		c := b.Cursor()

		var quotes []string
		for k, v := c.First(); k != nil; k, v = c.Next() {
			quotes = append(quotes, string(v))
		}

		source := rand.NewSource(time.Now().Unix())
		r := rand.New(source)
		_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(":speech_balloon: %v said %q", u.Username, quotes[r.Intn(len(quotes))]))
		log.Printf("got an entry from quotes %v", quotes)

		return err
	})
}
