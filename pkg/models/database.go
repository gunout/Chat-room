package models

import (
	"database/sql"
	"errors"

	_ "github.com/glebarez/go-sqlite"
)

var (
	DB                *sql.DB
	ErrNotInitialized = errors.New("database is not initialized")
)

func Initialize(connString string) error {
	db, err := sql.Open("sqlite", connString)
	if err != nil {
		return err
	}

	DB = db

	// Run table migrations
	if err := (DirectMessage{}).CreateTable(); err != nil {
		return err
	}

	return nil
}
