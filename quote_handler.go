package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	cmdGet  = "get"
	cmdAdd  = "add"
	cmdDel  = "del"
	cmdList = "list"
)

// Quotes are kept in the `quotes:{uid}` bucket, with the key being `{reason}`
// where reason is the string representing the reason for a score and uid is
// the uid of the user with the scores.

func quoteHandler(session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	var err error

	// Parse the quote command, which will be one of the following forms:
	// * [verb] [user] [args]
	// * [user]
	// * [verb] [args]
	// * [verb]
	// * nothing

	cmd := cmdGet
	var uid string
	user := message.Author
	if len(args) > 0 {
		cmd, args = args[0], args[1:]
	}

	switch cmd {
	case cmdAdd, cmdDel, cmdGet, cmdList:
		// The next argument must be a user (or nothing)
		if len(args) > 0 {
			uid = getUIDFromMention(args[0])
			if uid == message.Mentions[0].ID {
				user = message.Mentions[0]
				args = args[1:]
			}
		}
	default:
		// In the default case, the only valid argument is a user (get is implied)
		uid := getUIDFromMention(cmd)
		if uid != message.Mentions[0].ID {
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

func addUserQuote(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User, quote string) error {
	return nil
}

func delUserQuote(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User, quoteID string) error {
	return nil
}

func listUserQuotes(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User) error {
	return nil
}

func getUserQuote(s *discordgo.Session, m *discordgo.MessageCreate, u *discordgo.User) error {
	return nil
}
