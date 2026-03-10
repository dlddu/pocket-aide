// Package db provides SQLite database connection and migration utilities.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// New opens a new SQLite database connection using the given DSN.
// For in-memory databases, use ":memory:" as the DSN.
func New(dsn string) (*sql.DB, error) {
	database, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return database, nil
}

// RunMigrations runs all pending database migrations from the given directory path.
// It tracks applied migrations in a schema_migrations table to ensure idempotency.
func RunMigrations(db *sql.DB, migrationsPath string) error {
	// Create schema_migrations tracking table if it doesn't exist
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read migration files from the directory
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Collect and sort .up.sql files
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	// Apply each migration that hasn't been applied yet
	for _, filename := range files {
		var existing string
		row := db.QueryRow("SELECT filename FROM schema_migrations WHERE filename = ?", filename)
		if err := row.Scan(&existing); err == nil {
			// Already applied, skip
			continue
		}

		// Read and execute the migration file
		content, err := os.ReadFile(filepath.Join(migrationsPath, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record this migration as applied
		if _, err := db.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", filename); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}
	}

	return nil
}
