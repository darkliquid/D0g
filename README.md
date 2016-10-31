# D0g [![Go Report Card](https://goreportcard.com/badge/github.com/darkliquid/D0g)](https://goreportcard.com/report/github.com/darkliquid/D0g) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/darkliquid/D0g/blob/master/LICENSE) [![GoDoc](https://godoc.org/github.com/darkliquid/D0g?status.svg)](https://godoc.org/github.com/darkliquid/D0g) [![Build Status](https://travis-ci.org/darkliquid/D0g.svg?branch=master)](https://travis-ci.org/darkliquid/D0g)

A discord bot written in Go for City17.

## Setup

To setup the bot, you need first create an application via the discord dev area
here: https://discordapp.com/developers/applications/me

You can safely ignore virtually everything it says, just give it a name and
then click create. After that, you'll have a "Create a Bot User" button for
you app, click it to get the actual bot user.

Now you have a bot user, you can get the token for the bot. You'll need this
for the bot config file. You'll also see a client ID, you'l want this too for
the next step.

The final step to add the bot to your server is to go to the following url,
with YOUR_CLIENT_ID substituted for the client ID of your bot: https://discordapp.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&scope=bot&permissions=0

That will give you the option to select a server you own to join the bot to.
