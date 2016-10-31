package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

var start = time.Now()

func uptimeHandler(session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	_, _ = session.ChannelMessageSend(message.ChannelID, fmt.Sprint("Uptime: ", time.Since(start)))
}
