package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// AppDatabase is the interface that defines the methods your application
// needs for interacting with the database. Add methods as needed.
type AppDatabase interface {
	Ping() error
	// You will also implement methods defined in user.go, conversation.go, message.go, etc.
}

// appdbimpl is the internal implementation of AppDatabase.
// It holds a reference to the SQL database connection.
type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the provided *sql.DB.
// It expects that the schema is already initialized (InitSchema called beforehand).
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}
	return &appdbimpl{c: db}, nil
}

// Ping checks the database connection.
func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

// OpenDB opens a connection to the given SQLite database file.
func OpenDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("could not open SQLite database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("could not ping SQLite database: %w", err)
	}
	return db, nil
}

// InitSchema reads and executes the SQL schema from the given file.
func InitSchema(db *sql.DB, schemaPath string) error {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Println("Database schema initialized successfully.")
	return nil
}
