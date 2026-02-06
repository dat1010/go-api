package config

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() error {
	timeout := migrationTimeout()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dsn, err := buildPostgresDSN(ctx)
	if err != nil {
		return err
	}

	dsn, err = withStatementTimeout(dsn, timeout)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open postgres for migrations: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping postgres for migrations: %w", err)
	}

	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to init migration driver: %w", err)
	}

	sourceURL, err := migrationSourceURL()
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to init migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func migrationSourceURL() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	migrationsPath := filepath.Join(wd, "migrations")
	return "file://" + filepath.ToSlash(migrationsPath), nil
}

func migrationTimeout() time.Duration {
	const defaultTimeout = 30 * time.Second
	raw := os.Getenv("MIGRATION_TIMEOUT_MS")
	if raw == "" {
		return defaultTimeout
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultTimeout
	}

	return time.Duration(value) * time.Millisecond
}

func withStatementTimeout(dsn string, timeout time.Duration) (string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("failed to parse db dsn: %w", err)
	}

	query := parsed.Query()
	query.Set("options", fmt.Sprintf("-c statement_timeout=%d", timeout.Milliseconds()))
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}
