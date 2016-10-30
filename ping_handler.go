package main

import "github.com/bwmarrin/discordgo"

func pingHandler(session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	_, _ = session.ChannelMessageSend(message.ChannelID, "Pong!")
}
