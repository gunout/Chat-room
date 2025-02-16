package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	barertc "git.kirsle.net/apps/barertc/pkg"
	"git.kirsle.net/apps/barertc/pkg/config"
	"git.kirsle.net/apps/barertc/pkg/log"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Command line flags.
	var (
		debug   bool
		address string
	)
	flag.BoolVar(&debug, "debug", false, "Enable debug-level logging in the app.")
	flag.StringVar(&address, "address", ":9000", "Address to listen on, like localhost:5000 or :8080")
	flag.Parse()

	if debug {
		log.SetDebug(true)
	}

	// Load configuration.
	if err := config.LoadSettings(); err != nil {
		panic(fmt.Sprintf("Error loading settings.toml: %s", err))
	}

	app := barertc.NewServer()
	app.Setup()

	log.Info("Listening at %s", address)
	panic(app.ListenAndServe(address))
}
