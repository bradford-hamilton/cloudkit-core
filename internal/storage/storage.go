package storage

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/bradford-hamilton/cloudkit-core/internal/cloudkit"

	// postgres driver
	_ "github.com/lib/pq"
)

// Datastore descirbes all the behaviors the persistance layer must implement.
type Datastore interface {
	CreateVM(vm cloudkit.VM) (int, error)
	RecordVMMemory(domainID int, usage float64) error
	GetVMIDFromDomainID(domainID int) (int, error)
}

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

// CreateVM inserts a cloud kit VM into our datastore.
func (db *Database) CreateVM(vm cloudkit.VM) (int, error) {
	var id int
	query := `INSERT INTO vms (name, domain_id) VALUES ($1, $2) RETURNING id;`

	row := db.QueryRow(query, vm.Name, vm.DomainID)
	if err := row.Err(); err != nil {
		return 0, err
	}
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// GetVMIDFromDomainID ....
func (db *Database) GetVMIDFromDomainID(domainID int) (int, error) {
	var id int
	query := `SELECT id FROM vms WHERE domain_id = $1;`

	row := db.QueryRow(query, domainID)
	if err := row.Err(); err != nil {
		return 0, err
	}
	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// RecordVMMemory inserts a snapshot of a VMs memory into storage.
func (db *Database) RecordVMMemory(domainID int, usage float64) error {
	vmID, err := db.GetVMIDFromDomainID(domainID)
	if err != nil {
		return err
	}

	query := `INSERT INTO measurements (time, vm_id, mem_usage) VALUES ($1, $2, $3);`
	row := db.QueryRow(query, time.Now(), vmID, usage)
	if err := row.Err(); err != nil {
		return err
	}

	return nil
}
