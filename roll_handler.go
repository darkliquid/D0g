package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/darkliquid/roll"
)

func rollHandler(session *discordgo.Session, message *discordgo.MessageCreate, args ...string) {
	user := message.Author.Username
	rollArg := strings.Join(args, " ")
	rand.Seed(time.Now().UnixNano())
	parser := roll.NewParser(strings.NewReader(rollArg))

	// Parse dice roll and whine if it's a bad roll string
	stmt, err := parser.Parse()
	if err != nil {
		_, _ = session.ChannelMessageSend(message.ChannelID, fmt.Sprintf(
			":poop: **%s** rolled `%s` but that is incomprehensible gibberish, so they critical fail at life",
			user,
			rollArg,
		))
		log.Println(err)
		return
	}

	// Build up result string
	results := stmt.Roll()
	rolls := make([]string, len(results.Results))
	for i, result := range results.Results {
		rolls[i] = result.Symbol
	}

	// TODO: improve results output by taking into account whether it's a roll
	// for number of successes instead of just a total.
	_, _ = session.ChannelMessageSend(message.ChannelID, fmt.Sprintf(
		":game_die: **%s** rolled `%s` and got the following results: `%s`\n\nFinal result: **%v**",
		user,
		strings.TrimPrefix(stmt.String(), "+"),
		strings.Join(rolls, ", "),
		results.Total,
	))
}
