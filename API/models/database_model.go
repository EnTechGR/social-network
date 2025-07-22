package models

import (
	"database/sql"
	"fmt"
	"forum/config"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Database version constants
const (
	CURRENT_DB_VERSION = 4 // Updated to version 4 for image uploads
	INITIAL_VERSION    = 1
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         []string
}

// GetMigrations returns all available migrations
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     2,
			Description: "Add OAuth support",
			SQL: []string{
				config.CreateOAuthTable,
				`CREATE INDEX IF NOT EXISTS idx_oauth_provider_user ON oauth_accounts(provider, provider_user_id)`,
				`CREATE INDEX IF NOT EXISTS idx_oauth_user_id ON oauth_accounts(user_id)`,
			},
		},
		{
			Version:     3,
			Description: "Add OAuth state management",
			SQL: []string{
				`CREATE TABLE IF NOT EXISTS oauth_states (
					state TEXT PRIMARY KEY,
					provider TEXT NOT NULL,
					ip_address TEXT,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					expires_at TIMESTAMP NOT NULL
				)`,
				`CREATE INDEX IF NOT EXISTS idx_oauth_states_expires ON oauth_states(expires_at)`,
				`CREATE INDEX IF NOT EXISTS idx_oauth_states_provider ON oauth_states(provider)`,
			},
		},
		{
			Version:     4,
			Description: "Add images table",
			SQL: []string{
				config.CreateImagesTable,
				config.IdxImagesPostID,
			},
		},
		// Add future migrations here
	}
}

// InitDB initializes the database and returns a connection
func InitDB() (*sql.DB, error) {
	dbPath := filepath.Join("./database", "forum.db")

	firstTime := false
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		firstTime = true
		if err := os.MkdirAll("./database", 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %v", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Always ensure the database_version table exists before doing anything with versions
	if err := createDatabaseVersionTable(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create database version table: %v", err)
	}

	// Determine if it's truly a first-time setup or just missing version info
	currentVersion, err := getDatabaseVersion(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to get database version: %v", err)
	}

	if firstTime || currentVersion == 0 { // If file didn't exist, or version is 0 (meaning no version recorded yet)
		fmt.Println("Performing initial database setup...")
		if err := createTables(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create tables: %v", err)
		}
		if err := createIndexes(db); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create indexes: %v", err)
		}
		if err := populateCategories(db, config.Categories); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to populate categories: %v", err)
		}
		// After creating initial tables, set the version to INITIAL_VERSION (1)
		// and then run any pending migrations from there to CURRENT_DB_VERSION.
		if err := setDatabaseVersion(db, INITIAL_VERSION); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set initial database version: %v", err)
		}
		fmt.Println("Initial database setup completed.")
	}

	// Always run migrations to catch up to the CURRENT_DB_VERSION
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}
	fmt.Println("Database initialization and migrations completed successfully.")

	return db, nil
}

func createDatabaseVersionTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS database_version (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getDatabaseVersion(db *sql.DB) (int, error) {

	var version int
	err := db.QueryRow("SELECT version FROM database_version ORDER BY version DESC LIMIT 1").Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			// If there are no rows in database_version, check if other tables exist.
			// This handles cases where a very old database might exist without a version table,
			// or a newly created one before any version is set.
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count) // Changed to 'users' for consistency
			if err != nil {
				return 0, fmt.Errorf("failed to check existing tables: %v", err)
			}
			if count > 0 {
				// If 'users' table exists but no database_version, assume INITIAL_VERSION
				// This might be redundant with the new InitDB logic, but provides a fallback.
				if err := setDatabaseVersion(db, INITIAL_VERSION); err != nil {
					return 0, fmt.Errorf("failed to set initial version: %v", err)
				}
				return INITIAL_VERSION, nil
			}
			return 0, nil  // No version and no existing user table, implies brand new DB
		}
		return 0, fmt.Errorf("failed to get database version: %v", err)
	}
	return version, nil
}

func setDatabaseVersion(db *sql.DB, version int) error {
	_, err := db.Exec("INSERT INTO database_version (version) VALUES (?)", version)
	return err
}

func createBackup(dbPath string) (string, error) {
	backupDir := filepath.Join("./database", "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("forum_backup_%s.db", timestamp)
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
		if filepath.Ext(entry.Name()) != ".db" || len(entry.Name()) < 12 || entry.Name()[:12] != "forum_backup" {
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
		fmt.Printf("Cleaned up %d old backup(s)\n", deleted)
	}
	return nil
}

func runMigrations(db *sql.DB) error {
	currentVersion, err := getDatabaseVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current database version: %v", err)
	}

	migrations := GetMigrations()
	var pending []Migration
	for _, m := range migrations {
		if m.Version > currentVersion {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 && currentVersion == CURRENT_DB_VERSION {
		fmt.Printf("Database is up to date (version %d)\n", currentVersion)
		return nil
		} else if len(pending) == 0 && currentVersion < CURRENT_DB_VERSION {
			// This case means there are no migrations defined beyond the current version
			// but the current version is not yet the latest expected version.
			// This can happen if CURRENT_DB_VERSION constant is updated, but no
			// corresponding migration is added to GetMigrations().
			fmt.Printf("Warning: Database version (%d) is not at the latest expected version (%d), but no pending migrations found.\n", currentVersion, CURRENT_DB_VERSION)
			return nil
	}

	fmt.Printf("Running %d migration(s)...\n", len(pending))

	dbPath := filepath.Join("./database", "forum.db")
	backupPath, err := createBackup(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}
	fmt.Printf("✅ Database backup created: %s\n", backupPath)

	if err := cleanupOldBackups(30); err != nil {
		fmt.Printf("Warning: failed to clean backups: %v\n", err)
	}

	for _, m := range pending {
		fmt.Printf("Applying migration %d: %s\n", m.Version, m.Description)
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %v\nBackup: %s", m.Version, err, backupPath)
		}
		for i, stmt := range m.SQL {
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d stmt %d failed: %v\nSQL: %s\nBackup: %s", m.Version, i+1, err, stmt, backupPath)
			}
		}
		if _, err := tx.Exec("INSERT INTO database_version (version) VALUES (?)", m.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update version %d: %v\nBackup: %s", m.Version, err, backupPath)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d failed: %v\nBackup: %s", m.Version, err, backupPath)
		}
		fmt.Printf("Migration %d completed\n", m.Version)
	}

	fmt.Println("🎉 All migrations completed successfully!")
	fmt.Printf("📁 Backup stored at: %s\n", backupPath)
	return nil
}

func createTables(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %v", err)
	}
	defer tx.Rollback()

	statements := []string{
		config.CreateUserTable,
		config.CreateUserAuthTable,
		config.CreateSessionsTable,
		config.CreateCategoriesTable,
		config.CreatePostsTable,
		config.CreateCommentsTable,
		config.CreateReactionsTable,
		config.CreateImagesTable,
		config.CreatePostCategoriesTable,
		config.CreateOAuthTable,
		// Add OAuth state table for new installations
		`CREATE TABLE IF NOT EXISTS oauth_states (
			state TEXT PRIMARY KEY,
			provider TEXT NOT NULL,
			ip_address TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL
		)`,
	}

	for i, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("statement %d failed: %v\nSQL: %s", i+1, err, stmt)
		}
	}

	return tx.Commit()
}

func createIndexes(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %v", err)
	}
	defer tx.Rollback()

	indexes := []string{
		config.IdxPostsUserID,
		config.IdxPostCategoriesPostID,
		config.IdxPostCategoriesCategoryID,
		config.IdxCommentsPostID,
		config.IdxCommentsUserID,
		config.IdxReactionsUserID,
		config.IdxReactionsPostID,
		config.IdxReactionsCommentID,
		config.IdxImagesPostID,
		// OAuth indexes
		`CREATE INDEX IF NOT EXISTS idx_oauth_provider_user ON oauth_accounts(provider, provider_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_oauth_user_id ON oauth_accounts(user_id)`,
		// OAuth state indexes
		`CREATE INDEX IF NOT EXISTS idx_oauth_states_expires ON oauth_states(expires_at)`,
		`CREATE INDEX IF NOT EXISTS idx_oauth_states_provider ON oauth_states(provider)`,
	}

	for _, stmt := range indexes {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to create index: %v\nSQL: %s", err, stmt)
		}
	}

	return tx.Commit()
}

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

	fmt.Println("Categories populated (duplicates ignored).")
	return nil
}

// CleanupExpiredOAuthStates removes expired OAuth state records
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
		fmt.Printf("Cleaned up %d expired OAuth state(s)\n", rowsAffected)
	}

	return nil
}

func RestoreFromBackup(backupPath string) error {
	dbPath := filepath.Join("./database", "forum.db")

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup does not exist: %s", backupPath)
	}

	currentBackup, err := createBackup(dbPath)
	if err == nil {
		fmt.Printf("Current DB backed up to: %s\n", currentBackup)
	}

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
		if filepath.Ext(entry.Name()) != ".db" || len(entry.Name()) < 12 || entry.Name()[:12] != "forum_backup" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(backupDir, entry.Name())
		backups = append(backups, fmt.Sprintf("%s (size: %d bytes, modified: %s)", path, info.Size(), info.ModTime().Format("2006-01-02 15:04:05")))
	}

	return backups, nil
}
