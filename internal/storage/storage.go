package storage

import (
	"database/sql"
	"fmt"
	"os"

	// postgres driver
	_ "github.com/lib/pq"
)

// Datastore ...
type Datastore interface{}

// Database ...
type Database struct {
	*sql.DB
}

// NewDatabase ...
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
