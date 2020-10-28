package storage

import (
	"database/sql"
	"fmt"
	"os"

	// postgres driver
	_ "github.com/lib/pq"
)

// Datastore descirbes all the behaviors the persistance layer must implement.
type Datastore interface{}

// Database implements our Datastore interface.
type Database struct {
	*sql.DB
}

// NewDatabase aquires a connection to Postgres, embeds it in a Database, and pings
// the db before returning it.
func NewDatabase() (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("CLOUDKIT_DB_HOST"),
		os.Getenv("CLOUDKIT_DB_PORT"),
		os.Getenv("CLOUDKIT_DB_USER"),
		os.Getenv("CLOUDKIT_DB_PASSWORD"),
		os.Getenv("CLOUDKIT_DB_NAME"),
		os.Getenv("CLOUDKIT_SSL_MODE"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db}, nil
}
