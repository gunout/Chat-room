package main

import (
	"math/rand"
	"os"
	"time"

	"git.kirsle.net/apps/barertc/cmd/BareBot/commands"
	"github.com/urfave/cli/v2"
)

const Version = "0.0.1"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	app := cli.NewApp()
	app.Name = "BareBot"
	app.Usage = "chatbot client for the BareRTC chat server"
	app.Version = Version
	app.Commands = []*cli.Command{
		commands.Init,
		commands.Run,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
