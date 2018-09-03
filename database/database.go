// Package database manages storage for bitly-lambda
package database

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/olekukonko/tablewriter"

	// The database is sqlite3
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database and implements methods to perform operations on the database.
type Database struct {
	File string
	DB   *sqlx.DB
}

// QueryOptions represents the options you can have for a query and how the result will be rendered
type QueryOptions struct {
	Writer     io.Writer
	Query      string
	MergeCells bool
	RowLine    bool
	Caption    string
	Render     bool
}

// QueryResponse represents the response from a query
type QueryResponse struct {
	Rows        [][]string
	ColumnNames []string
	Table       *tablewriter.Table
}

// New creates a connection to the database.
func New(filename string) (*Database, error) {
	// Making sure the file exists
	_, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("file %s does not exist", filename)
	}

	// Connect to the database
	dbase, err := sqlx.Open("sqlite3", filename)
	if err != nil {
		return nil, fmt.Errorf("error while opening connection to database: %s", err.Error())
	}

	// Return a new struct
	return &Database{File: filename, DB: dbase}, nil
}

// Close closes all handles to the database
func (db *Database) Close() error {
	err := db.DB.Close()
	if err != nil {
		return fmt.Errorf("error while closing database: %s", err.Error())
	}
	return nil
}

// ExecWithTransaction executes a query and wraps the execution in a transaction
func (db *Database) ExecWithTransaction(query string) error {
	// Start a transaction to add everything into the database
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("error while starting database transaction: %s", err.Error())
	}

	// Execute the query
	_, err = db.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error while executing query: %s", err.Error())
	}

	// Commit the transaction
	tx.Commit()

	return nil
}

// Exec executes a query without any transaction support
func (db *Database) Exec(query string) error {
	// Execute the query
	_, err := db.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error while executing query: %s", err.Error())
	}

	return nil
}

// RunQuery run a query on the database and prints the result in a table
func (db *Database) RunQuery(opts QueryOptions) (QueryResponse, error) {
	queryResponse := QueryResponse{}

	// Execute the query
	rows, err := db.DB.Queryx(opts.Query)
	if err != nil {
		return queryResponse, fmt.Errorf("error while executing query: %s", err.Error())
	}
	defer rows.Close()

	// Get the column names
	colnames, _ := rows.Columns()

	// Prepare the output table
	table := tablewriter.NewWriter(opts.Writer)
	table.SetHeader(colnames)
	table.SetAutoMergeCells(opts.MergeCells)
	table.SetRowLine(opts.RowLine)
	if len(opts.Caption) > 0 {
		table.SetCaption(true, opts.Caption)
	}

	// Prepare a result array
	var resultArray [][]string

	// Loop over the result
	for rows.Next() {
		cols, _ := rows.SliceScan()
		tempStringArray := make([]string, len(cols))
		for idx := range cols {
			switch v := cols[idx].(type) {
			case int64:
				tempStringArray[idx] = strconv.Itoa(int(v))
			case string:
				tempStringArray[idx] = v
			case nil:
				tempStringArray[idx] = ""
			case []uint8:
				tempStringArray[idx] = string(v)
			default:
				tempStringArray[idx] = fmt.Sprintf("%v", v)
			}
		}
		table.Append(tempStringArray)
		resultArray = append(resultArray, tempStringArray)
	}

	// Print the table
	if opts.Render {
		table.Render()
	}

	queryResponse.ColumnNames = colnames
	queryResponse.Rows = resultArray
	queryResponse.Table = table

	return queryResponse, nil
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
