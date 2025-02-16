package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git.kirsle.net/apps/barertc/client"
	"git.kirsle.net/apps/barertc/client/config"
	"git.kirsle.net/apps/barertc/pkg/jwt"
	"git.kirsle.net/apps/barertc/pkg/log"
	xjwt "github.com/golang-jwt/jwt/v4"
	"github.com/urfave/cli/v2"
)

// Run implements `BareBot run`
var Run *cli.Command

func init() {
	Run = &cli.Command{
		Name:      "run",
		Usage:     "run the BareBot client program and connect to your chat room",
		ArgsUsage: "<chatbot directory>",
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			// Chatbot directory
			var botdir = c.Args().First()
			if botdir == "" {
				botdir = "."
			}

			// Check for the chatbot.toml file.
			if _, err := os.Stat(filepath.Join(botdir, "chatbot.toml")); os.IsNotExist(err) {
				return cli.Exit(fmt.Errorf(
					"Did not find chatbot.toml in your chatbot directory (%s): did you run `BareBot init`?",
					botdir,
				), 1)
			}

			// Enter the directory.
			if err := os.Chdir(botdir); err != nil {
				log.Error("Couldn't enter directory %s: %s", botdir, err)
				return cli.Exit("Exited", 1)
			}

			// Load the settings.
			if err := config.LoadSettings(); err != nil {
				return cli.Exit(fmt.Sprintf(
					"Couldn't load chatbot.toml: %s", err,
				), 1)
			}

			log.Info("Initializing BareBot")

			// Get the JWT auth token.
			log.Info("Authenticating with BareRTC (getting JWT token)")
			client, err := client.NewClient(config.Current.BareRTC.URL, jwt.Claims{
				IsAdmin:    config.Current.Profile.IsAdmin,
				Avatar:     config.Current.Profile.AvatarURL,
				ProfileURL: config.Current.Profile.ProfileURL,
				Nick:       config.Current.Profile.Nickname,
				Emoji:      config.Current.Profile.Emoji,
				Gender:     config.Current.Profile.Gender,
				RegisteredClaims: xjwt.RegisteredClaims{
					Subject: config.Current.Profile.Username,
				},
			})
			if err != nil {
				return cli.Exit(err, 1)
			}

			// Register handler funcs for the chatbot.
			client.SetupChatbot()

			// Run!
			log.Info("Connecting to ChatServer")
			err = client.Run()
			if err != nil {
				log.Error("Error: %s (and sleeping 5 seconds before exit)", err)
				time.Sleep(5 * time.Second)
			}
			return cli.Exit(client.Run(), 1)
		},
	}
}
