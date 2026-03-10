package db_test

import (
	"database/sql"
	"testing"

	"github.com/dlddu/pocket-aide/db"
)

func TestNew_OpensSQLiteConnection(t *testing.T) {
	// Arrange
	dsn := ":memory:"

	// Act
	database, err := db.New(dsn)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if database == nil {
		t.Fatal("expected database instance, got nil")
	}
	defer database.Close()
}

func TestNew_ReturnsErrorOnInvalidDSN(t *testing.T) {
	// Arrange
	dsn := "/nonexistent/path/to/db.sqlite"

	// Act
	database, err := db.New(dsn)

	// Assert
	if err == nil {
		t.Error("expected error for invalid DSN, got nil")
		if database != nil {
			database.Close()
		}
	}
}

func TestNew_DatabaseIsPingable(t *testing.T) {
	// Arrange
	dsn := ":memory:"

	// Act
	database, err := db.New(dsn)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	// Assert
	if err := database.Ping(); err != nil {
		t.Fatalf("expected database to be pingable, got: %v", err)
	}
}

func TestRunMigrations_CreatesTables(t *testing.T) {
	// Arrange
	dsn := ":memory:"
	database, err := db.New(dsn)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	// Act
	err = db.RunMigrations(database, "migrations")

	// Assert
	if err != nil {
		t.Fatalf("expected migrations to run successfully, got: %v", err)
	}
}

func TestRunMigrations_UsersTableExists(t *testing.T) {
	// Arrange
	dsn := ":memory:"
	database, err := db.New(dsn)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	err = db.RunMigrations(database, "migrations")
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Act
	var tableName string
	row := database.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='users'",
	)
	err = row.Scan(&tableName)

	// Assert
	if err == sql.ErrNoRows {
		t.Fatal("expected 'users' table to exist after migration, but it does not")
	}
	if err != nil {
		t.Fatalf("unexpected error querying sqlite_master: %v", err)
	}
	if tableName != "users" {
		t.Errorf("expected table name 'users', got '%s'", tableName)
	}
}

func TestRunMigrations_IsIdempotent(t *testing.T) {
	// Arrange
	dsn := ":memory:"
	database, err := db.New(dsn)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer database.Close()

	// Act: run migrations twice
	err = db.RunMigrations(database, "migrations")
	if err != nil {
		t.Fatalf("first migration failed: %v", err)
	}

	err = db.RunMigrations(database, "migrations")

	// Assert: second run should not return an error
	if err != nil {
		t.Fatalf("expected idempotent migrations, got error on second run: %v", err)
	}
}
