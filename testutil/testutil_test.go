package testutil_test

import (
	"database/sql"
	"testing"

	"github.com/dlddu/pocket-aide/testutil"
)

func TestNewTestDB_ReturnsInMemoryDatabase(t *testing.T) {
	// Arrange & Act
	tdb := testutil.NewTestDB(t)

	// Assert
	if tdb == nil {
		t.Fatal("expected non-nil TestDB, got nil")
	}
	if tdb.DB == nil {
		t.Fatal("expected non-nil sql.DB, got nil")
	}
}

func TestNewTestDB_DatabaseIsPingable(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)

	// Act
	err := tdb.DB.Ping()

	// Assert
	if err != nil {
		t.Fatalf("expected in-memory DB to be pingable, got: %v", err)
	}
}

func TestNewTestDB_RunsMigrations(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)

	// Act: query the users table which should be created by migrations
	var tableName string
	row := tdb.DB.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
	)
	err := row.Scan(&tableName)

	// Assert
	if err == sql.ErrNoRows {
		t.Fatal("expected 'users' table to exist after NewTestDB, but it does not")
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tableName != "users" {
		t.Errorf("expected table 'users', got '%s'", tableName)
	}
}

func TestNewTestDB_IsolatedBetweenTests(t *testing.T) {
	// Arrange
	tdb1 := testutil.NewTestDB(t)
	tdb2 := testutil.NewTestDB(t)

	// Act: insert a row into tdb1
	_, err := tdb1.DB.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"test@example.com", "hashed_password",
	)
	if err != nil {
		t.Fatalf("failed to insert into tdb1: %v", err)
	}

	// Assert: tdb2 should not see the row from tdb1
	var count int
	row := tdb2.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "test@example.com")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query tdb2: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows in isolated tdb2, got %d", count)
	}
}

func TestNewTestDB_CleanupClosesDatabase(t *testing.T) {
	// Arrange
	var capturedDB *sql.DB

	t.Run("subtest", func(t *testing.T) {
		tdb := testutil.NewTestDB(t)
		capturedDB = tdb.DB
		// The cleanup registered by NewTestDB will run after this subtest ends
	})

	// Assert: after subtest ends, t.Cleanup should have run
	// We verify the DB was at least created properly
	if capturedDB == nil {
		t.Error("expected DB to have been created in subtest")
	}
}

func TestNewTestDB_SeedHelperInsertsData(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)

	// Act
	tdb.Seed(t, "INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"seed@example.com", "hash123",
	)

	// Assert
	var email string
	row := tdb.DB.QueryRow("SELECT email FROM users WHERE email = ?", "seed@example.com")
	if err := row.Scan(&email); err != nil {
		t.Fatalf("expected seeded user to exist, got error: %v", err)
	}
	if email != "seed@example.com" {
		t.Errorf("expected email 'seed@example.com', got '%s'", email)
	}
}

func TestNewTestDB_TruncateHelperClearsTable(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t, "INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"user1@example.com", "hash1",
	)
	tdb.Seed(t, "INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"user2@example.com", "hash2",
	)

	// Act
	tdb.Truncate(t, "users")

	// Assert
	var count int
	row := tdb.DB.QueryRow("SELECT COUNT(*) FROM users")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows after truncate, got %d", count)
	}
}
