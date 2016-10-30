package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/paked/configure"
	werr "github.com/pkg/errors"
)

var (
	conf   = configure.New()
	token  = conf.String("token", "", "Discord bot token")
	cmdpfx = conf.String("command-prefix", "!", "The prefix symbol used to indicate commands")
	me     *discordgo.User
)

func main() {
	// Pull in configuration
	conf.Use(configure.NewFlag())
	conf.Use(configure.NewEnvironment())

	if len(os.Args) > 1 {
		if _, err := os.Stat(os.Args[1]); err == nil {
			conf.Use(configure.NewJSONFromFile(os.Args[1]))
		} else {
			log.Println("can't load specified config file", err)
		}
	}
	conf.Parse()

	if token == nil {
		log.Fatalln("A token MUST be set to authenticate")
	}

	log.Println("Creating session...")
	discord, err := discordgo.New(fmt.Sprint("Bot ", *token))
	if err != nil {
		err = werr.Wrap(err, "failed to create discord session")
		log.Fatalln(err)
	}

	log.Println("Registering command handler...")
	discord.AddHandler(commandHandler)

	// Open the websocket and begin listening.
	log.Println("Opening connection...")
	if err = discord.Open(); err != nil {
		err = werr.Wrap(err, "failed to open discord connection")
		log.Fatalln(err)
	}

	// Get our own user details
	me, err = discord.User("@me")
	if err != nil {
		log.Fatalln("error obtaining account details,", err)
	}

	// Store the account ID for later use.

	// Block forever
	log.Println("Running...")
	lock := make(chan struct{})
	<-lock
}
