package config

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func NewDB() (*sqlx.DB, error) {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	dbToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbURL == "" || dbToken == "" {
		return nil, fmt.Errorf("TURSO_DATABASE_URL and TURSO_AUTH_TOKEN must be set")
	}

	// Add auth token to the URL
	dbURL = fmt.Sprintf("%s?authToken=%s", dbURL, dbToken)

	db, err := sql.Open("libsql", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Wrap with sqlx
	return sqlx.NewDb(db, "libsql"), nil
}
