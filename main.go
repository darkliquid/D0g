package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"github.com/paked/configure"
	werr "github.com/pkg/errors"
)

var (
	conf   = configure.New()
	token  = conf.String("token", "", "Discord bot token")
	cmdpfx = conf.String("command-prefix", "!", "The prefix symbol used to indicate commands")
	dbfile = conf.String("dbfile", "D0g.db", "The file that stores persistent info")
	me     *discordgo.User
	db     *bolt.DB
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

	if token == nil || *token == "" {
		log.Fatalln("A token MUST be set to authenticate")
	}

	if dbfile == nil || *dbfile == "" {
		log.Fatalln("A dbfile MUST be specified")
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

	// Initialise boltdb
	db, err = bolt.Open(*dbfile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Block forever
	log.Println("Running...")
	lock := make(chan struct{})
	<-lock
}
