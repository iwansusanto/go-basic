package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Connect initializes the database connection
func Connect(connStr string) (*sql.DB, error) {
	if connStr == "" {
		return nil, fmt.Errorf("connection string is empty")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %v", err)
	}

	// Verify the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return db, nil
}
