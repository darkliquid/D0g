package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func commandHandler(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if message.Author.ID == me.ID {
		return
	}

	// Find command and arguments
	if strings.HasPrefix(message.Content, *cmdpfx) {
		commandParts := strings.SplitN(strings.TrimPrefix(message.Content, *cmdpfx), " ", 2)
		command, args := strings.ToLower(commandParts[0]), commandParts[1:]

		switch command {
		case "ping":
			pingHandler(session, message, args...)
		case "roll":
			rollHandler(session, message, args...)
		case "uptime":
			uptimeHandler(session, message, args...)
		case "+", "-", "score", "top":
			scoreHandler(command, session, message, args...)
		}
	}
}
