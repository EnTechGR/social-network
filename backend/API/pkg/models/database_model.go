package models

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	dbmigrate "social-network/pkg/db/sqlite"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the database and returns a connection.
// It applies migrations via golang-migrate and seeds default categories.
func InitDB() (*sql.DB, error) {
	dbPath := filepath.Join("./database", "social-network.db")

	// Ensure database directory exists before opening the SQLite file
	if err := os.MkdirAll("./database", 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Open database connection with foreign keys enabled
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Set connection pool limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Apply SQL migrations using golang-migrate
	if err := dbmigrate.Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to apply migrations: %v", err)
	}
	fmt.Println("✅ Database migrations applied via golang-migrate.")

	// Seed default categories idempotently
	// Note: Categories are seeded by application code, not migrations
	// This allows easy customization without creating new migration files

	return db, nil
}

// populateCategories inserts default categories into the database.
// Uses INSERT OR IGNORE to make the operation idempotent.
// This is called after migrations to seed initial data.
func populateCategories(db *sql.DB, categories []string) error {
	if len(categories) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO categories (name) VALUES (?)`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %v", err)
	}
	defer stmt.Close()

	for _, c := range categories {
		if _, err := stmt.Exec(c); err != nil {
			return fmt.Errorf("insert category '%s': %v", c, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %v", err)
	}

	fmt.Println("✅ Categories populated (duplicates ignored).")
	return nil
}

// createBackup creates a timestamped backup of the database.
// Backups are stored in ./database/backups/ directory.
// Returns the path to the backup file.
func createBackup(dbPath string) (string, error) {
	backupDir := filepath.Join("./database", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("social_network_backup_%s.db", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	sourceFile, err := os.Open(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source database: %v", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return "", fmt.Errorf("failed to copy database: %v", err)
	}

	if err := destFile.Sync(); err != nil {
		return "", fmt.Errorf("failed to sync backup file: %v", err)
	}

	return backupPath, nil
}

// cleanupOldBackups removes backup files older than maxAgeDays.
// Useful for automatic cleanup via cron job or periodic task.
func cleanupOldBackups(maxAgeDays int) error {
	backupDir := filepath.Join("./database", "backups")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %v", err)
	}

	cutoff := time.Now().AddDate(0, 0, -maxAgeDays)
	deleted := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Only process files that match backup naming pattern
		if filepath.Ext(entry.Name()) != ".db" || len(entry.Name()) < 12 || entry.Name()[:12] != "social_network_backup" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			path := filepath.Join(backupDir, entry.Name())
			if err := os.Remove(path); err == nil {
				deleted++
			}
		}
	}

	if deleted > 0 {
		fmt.Printf("🧹 Cleaned up %d old backup(s)\n", deleted)
	}
	return nil
}

// RestoreFromBackup restores the database from a backup file.
// Creates a backup of the current database before restoring.
// Use this for disaster recovery or rolling back to a previous state.
func RestoreFromBackup(backupPath string) error {
	dbPath := filepath.Join("./database", "social-network.db")

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup does not exist: %s", backupPath)
	}

	// Create backup of current database before restoring
	currentBackup, err := createBackup(dbPath)
	if err == nil {
		fmt.Printf("📦 Current DB backed up to: %s\n", currentBackup)
	}

	// Open and copy backup file
	src, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("open backup: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(dbPath)
	if err != nil {
		return fmt.Errorf("create DB file: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("restore copy: %v", err)
	}

	if err := dst.Sync(); err != nil {
		return fmt.Errorf("sync restored DB: %v", err)
	}

	fmt.Printf("✅ Database restored from: %s\n", backupPath)
	return nil
}

// ListBackups returns a list of available backup files with metadata.
// Each entry includes file path, size, and modification time.
func ListBackups() ([]string, error) {
	backupDir := filepath.Join("./database", "backups")

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("read backup dir: %v", err)
	}

	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Only list files that match backup naming pattern
		if filepath.Ext(entry.Name()) != ".db" || len(entry.Name()) < 12 || entry.Name()[:12] != "social_network_backup" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(backupDir, entry.Name())
		backups = append(backups, fmt.Sprintf(
			"%s (size: %d bytes, modified: %s)",
			path,
			info.Size(),
			info.ModTime().Format("2006-01-02 15:04:05"),
		))
	}

	return backups, nil
}

// CleanupExpiredOAuthStates removes expired OAuth state records.
// Should be called periodically (e.g., via cron job) to prevent
// the oauth_states table from growing indefinitely.
// Recommended: Run daily or hourly depending on OAuth usage.
func CleanupExpiredOAuthStates(db *sql.DB) error {
	result, err := db.Exec("DELETE FROM oauth_states WHERE expires_at < ?", time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired OAuth states: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("🧹 Cleaned up %d expired OAuth state(s)\n", rowsAffected)
	}

	return nil
}
