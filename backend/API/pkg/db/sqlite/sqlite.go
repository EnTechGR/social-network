package sqlite

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrate applies SQL migrations located in pkg/db/migrations/sqlite using golang-migrate.
// It is idempotent; migrate.ErrNoChange is treated as success.
func Migrate(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("sqlite driver: %w", err)
	}

	migrationsPath, err := migrationsDir()
	if err != nil {
		return err
	}

	sourceURL := "file://" + migrationsPath
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// migrationsDir resolves the absolute path to the sqlite migrations folder.
// It uses the location of this file to remain robust to working directory changes.
func migrationsDir() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("cannot determine caller for migrations path")
	}
	base := filepath.Dir(filepath.Dir(thisFile)) // pkg/db
	path := filepath.Join(base, "migrations", "sqlite")
	return path, nil
}
