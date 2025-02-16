package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"git.kirsle.net/apps/barertc/client"
	"git.kirsle.net/apps/barertc/client/config"
	"git.kirsle.net/apps/barertc/pkg/log"
	"github.com/urfave/cli/v2"
)

// Init implements `BareBot init`
var Init *cli.Command

// Default folder structure
var defaultFolders = []string{
	"brain",
	"logs",
	"userdata",
}

func init() {
	Init = &cli.Command{
		Name:      "init",
		Usage:     "initialize a new BareBot robot at the given directory",
		ArgsUsage: "<chatbot directory>",
		Flags:     []cli.Flag{},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.Exit("Usage: BareBot init <directory>\n"+
					"    Example: BareBot init ./chatbot\n"+
					"    Example: BareBot init .",
					1,
				)
			}

			// If they named an existing directory, ensure it is empty.
			var botdir = c.Args().First()
			stat, err := os.Stat(botdir)
			if os.IsNotExist(err) {
				log.Info("Creating chatbot directory: %s", botdir)
				if err := os.MkdirAll(botdir, 0755); err != nil {
					log.Error("Error creating chatbot directory: %s: %s", botdir, err)
					return cli.Exit(err, 1)
				}
			} else if stat.IsDir() {
				// They named an existing directory: is it empty?
				fh, err := os.Open(botdir)
				if err != nil {
					log.Error("Checking if %s is empty: couldn't open: %s", botdir, err)
					return cli.Exit("Exited", 1)
				}
				defer fh.Close()

				_, err = fh.Readdirnames(1)
				if err != io.EOF {
					return cli.Exit(fmt.Sprintf(
						"%s: not an empty directory, will not initialize the chatbot into it", botdir),
						1,
					)
				}
			}

			// Enter the directory.
			if err := os.Chdir(botdir); err != nil {
				log.Error("Couldn't enter directory %s: %s", botdir, err)
				return cli.Exit("Exited", 1)
			}

			// Initialize the folders.
			for _, folder := range defaultFolders {
				log.Info("Creating: %s", folder)
				if err := os.MkdirAll(folder, 0755); err != nil {
					return cli.Exit(fmt.Sprintf(
						"Couldn't create %s: %s", folder, err,
					), 1)
				}
			}

			// Extract the default RiveScript brain.
			if files, err := client.Embedded.ReadDir("brain"); err == nil {
				for _, file := range files {
					log.Info("Extracting: brain/%s", file.Name())
					var (
						filename = filepath.Join("brain", file.Name())
					)
					data, err := client.Embedded.ReadFile(filename)
					if err != nil {
						log.Error("Reading built-in brain file %s: %s", filename, err)
						continue
					}
					ioutil.WriteFile(filename, data, 0644)
				}
			} else {
				log.Error("Couldn't read default brain: %s", err)
			}

			// Initialize the settings file.
			if err := config.WriteSettings(); err != nil {
				log.Error("Writing chatbot.toml: %s", err)
				return cli.Exit("Exited", 1)
			}

			return nil
		},
	}
}
