// Package database manages storage for bitly-lambda
package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	// The database is sqlite3
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database and implements methods to perform operations on the database.
type Database struct {
	File string
}

// New creates a connection to the database. filename represents the exact file location of the database
// file.
func New(filename string) (*Database, error) {
	// Making sure the file exists
	_, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("file %s does not exist", filename)
	}

	db := &Database{File: filename}

	return db, nil
}

// InsertItem inserts bitlink data into the database.
func (db *Database) InsertItem(item map[string]interface{}) error {
	// Open a connection to the database
	dbase, err := sqlx.Open("sqlite3", db.File)
	if err != nil {
		return fmt.Errorf("error while opening connection to database: %s", err.Error())
	}
	defer dbase.Close()

	// Start a transaction to add everything into the database
	tx, err := dbase.Begin()
	if err != nil {
		return fmt.Errorf("error while starting database transaction: %s", err.Error())
	}

	// Create a prepared statement
	stmt, err := tx.Prepare("insert into links(host, path, date, link, url, clicks, utm_source, utm_medium, utm_campaign, utm_term, utm_content) values(?,?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error while creating prepared statement: %s", err.Error())
	}
	defer stmt.Close()

	// Insert items into database
	host := item["host"]
	path := item["path"]
	date := item["date"]
	link := item["link"]
	url := item["url"]
	clicks := item["clicks"]
	utmSource := item["utm_source"]
	utmMedium := item["utm_medium"]
	utmCampaign := item["utm_campaign"]
	utmTerm := item["utm_term"]
	utmContent := item["utm_content"]

	// Execute the prepared statement
	_, err = stmt.Exec(host, path, date, link, url, clicks, utmSource, utmMedium, utmCampaign, utmTerm, utmContent)
	if err != nil {
		return fmt.Errorf("error while inserting data into database: %s", err.Error())
	}

	// Commit the transaction
	tx.Commit()

	return nil
}
