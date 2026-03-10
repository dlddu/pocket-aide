// Package testutil provides helpers for writing isolated database tests
// using an in-memory SQLite database.
package testutil

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/dlddu/pocket-aide/db"
)

// TestDB wraps an in-memory SQLite database for use in tests.
type TestDB struct {
	DB *sql.DB
}

// NewTestDB creates a new in-memory SQLite database, runs all migrations,
// and registers a cleanup function to close the DB when the test ends.
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}

	if err := db.RunMigrations(database, "../db/migrations"); err != nil {
		database.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	tdb := &TestDB{DB: database}

	t.Cleanup(func() {
		database.Close()
	})

	return tdb
}

// Seed executes a SQL statement with the given arguments to insert test data.
// It fails the test immediately if the statement returns an error.
func (tdb *TestDB) Seed(t *testing.T, query string, args ...interface{}) {
	t.Helper()

	if _, err := tdb.DB.Exec(query, args...); err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}
}

// Truncate deletes all rows from the specified table.
// It fails the test immediately if the operation returns an error.
func (tdb *TestDB) Truncate(t *testing.T, table string) {
	t.Helper()

	if _, err := tdb.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
		t.Fatalf("failed to truncate table %s: %v", table, err)
	}
}
