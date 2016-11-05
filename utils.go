package main

import (
	"bytes"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func getUIDFromMention(mention string) string {
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(mention, "<@"), "!"), ">")
}

func getGuildIDFromMessage(session *discordgo.Session, message *discordgo.MessageCreate) (string, error) {
	channel, err := session.Channel(message.ChannelID)
	if err != nil {
		return "", err
	}
	return channel.GuildID, nil
}

func cleanDiscordString(input string) string {
	// Strip all backticks
	input = strings.Replace(input, "`", "", -1)

	// Replace all new lines with spaces
	input = strings.Replace(input, "\r", "", -1)
	input = strings.Replace(input, "\n", " ", -1)

	// Arbitrarily limit number of runes in a string
	runes := bytes.Runes([]byte(input))
	if len(runes) > 100 {
		input = string(runes[:100])
	}

	// Get rid of superfluous whitespace
	input = strings.TrimSpace(input)

	// Trim trailing backslashes to avoid accidentally escaping stuff
	input = strings.TrimRight(input, "\\")

	// Final trim incase the is whitespace before some backslashes
	input = strings.TrimSpace(input)

	if len(runes) > 100 {
		input += "..."
	}

	return input
}
