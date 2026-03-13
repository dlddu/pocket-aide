// Package repository_test contains unit tests for the UserRepository.
// These tests use an in-memory SQLite database via testutil.NewTestDB and
// verify that UserRepository correctly reads and writes user records.
//
// TDD Red Phase: the UserRepository type and its methods do not yet exist.
// All tests are expected to fail until repository/user_repository.go is implemented.
package repository_test

import (
	"errors"
	"testing"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// CreateUser
// ---------------------------------------------------------------------------

// TestUserRepository_CreateUser_ReturnsNewUser verifies that CreateUser inserts
// a record and returns the persisted user with a non-zero ID and the correct email.
//
// Scenario:
//
//	CreateUser(email="new@example.com", passwordHash="hashed")
//	→ User{ID: >0, Email: "new@example.com"}
func TestUserRepository_CreateUser_ReturnsNewUser(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	user, err := repo.CreateUser("new@example.com", "hashed_password")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user, got nil")
	}
	if user.ID == 0 {
		t.Error("expected user ID to be non-zero after creation")
	}
	if user.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", user.Email)
	}
}

// TestUserRepository_CreateUser_PersistsToDatabase verifies that CreateUser
// actually writes the record so that a subsequent query finds it.
//
// Scenario:
//
//	CreateUser(email="persisted@example.com", passwordHash="hashed")
//	→ row exists in the users table
func TestUserRepository_CreateUser_PersistsToDatabase(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	_, err := repo.CreateUser("persisted@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("CreateUser returned unexpected error: %v", err)
	}

	// Assert: verify the row is actually in the DB
	var count int
	row := tdb.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "persisted@example.com")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query users table: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row in users table, got %d", count)
	}
}

// TestUserRepository_CreateUser_DuplicateEmail_ReturnsError verifies that
// attempting to create a second user with an already-registered email returns
// an error (the users table has a UNIQUE constraint on email).
//
// Scenario:
//
//	Seed: user with email "exists@example.com" already present.
//	CreateUser(email="exists@example.com", ...)
//	→ non-nil error
func TestUserRepository_CreateUser_DuplicateEmail_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"exists@example.com", "existing_hash",
	)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	user, err := repo.CreateUser("exists@example.com", "another_hash")

	// Assert
	if err == nil {
		t.Error("expected error for duplicate email, got nil")
	}
	if user != nil {
		t.Errorf("expected nil user on error, got %+v", user)
	}
}

// TestUserRepository_CreateUser_DuplicateEmail_ReturnsErrDuplicateEmail
// verifies that the specific sentinel error repository.ErrDuplicateEmail is
// returned when the email already exists, so callers can map it to HTTP 409.
//
// Scenario:
//
//	Seed: user with email "dup@example.com".
//	CreateUser(email="dup@example.com", ...)
//	→ errors.Is(err, repository.ErrDuplicateEmail) == true
func TestUserRepository_CreateUser_DuplicateEmail_ReturnsErrDuplicateEmail(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"dup@example.com", "existing_hash",
	)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	_, err := repo.CreateUser("dup@example.com", "another_hash")

	// Assert
	if !errors.Is(err, repository.ErrDuplicateEmail) {
		t.Errorf("expected ErrDuplicateEmail, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// FindByEmail
// ---------------------------------------------------------------------------

// TestUserRepository_FindByEmail_ReturnsUser verifies that FindByEmail returns
// the matching user record when the email exists in the database.
//
// Scenario:
//
//	Seed: user{email:"find@example.com", password_hash:"stored_hash"}.
//	FindByEmail("find@example.com")
//	→ User{Email: "find@example.com", PasswordHash: "stored_hash"}
func TestUserRepository_FindByEmail_ReturnsUser(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"find@example.com", "stored_hash",
	)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	user, err := repo.FindByEmail("find@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user, got nil")
	}
	if user.Email != "find@example.com" {
		t.Errorf("expected email 'find@example.com', got '%s'", user.Email)
	}
	if user.PasswordHash != "stored_hash" {
		t.Errorf("expected password_hash 'stored_hash', got '%s'", user.PasswordHash)
	}
}

// TestUserRepository_FindByEmail_ReturnsNonZeroID verifies that FindByEmail
// populates the ID field from the database row.
//
// Scenario:
//
//	Seed: user with email "idcheck@example.com".
//	FindByEmail("idcheck@example.com")
//	→ User.ID > 0
func TestUserRepository_FindByEmail_ReturnsNonZeroID(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"idcheck@example.com", "hash",
	)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	user, err := repo.FindByEmail("idcheck@example.com")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.ID == 0 {
		t.Error("expected non-zero ID from FindByEmail")
	}
}

// TestUserRepository_FindByEmail_NotFound_ReturnsError verifies that
// FindByEmail returns an error when no user with the given email exists.
//
// Scenario:
//
//	Empty database.
//	FindByEmail("ghost@example.com")
//	→ non-nil error
func TestUserRepository_FindByEmail_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	user, err := repo.FindByEmail("ghost@example.com")

	// Assert
	if err == nil {
		t.Error("expected error for non-existent email, got nil")
	}
	if user != nil {
		t.Errorf("expected nil user when not found, got %+v", user)
	}
}

// TestUserRepository_FindByEmail_NotFound_ReturnsErrNotFound verifies that
// the specific sentinel error repository.ErrUserNotFound is returned when the
// email does not exist, so callers can map it to HTTP 401/404.
//
// Scenario:
//
//	Empty database.
//	FindByEmail("nobody@example.com")
//	→ errors.Is(err, repository.ErrUserNotFound) == true
func TestUserRepository_FindByEmail_NotFound_ReturnsErrNotFound(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(tdb.DB)

	// Act
	_, err := repo.FindByEmail("nobody@example.com")

	// Assert
	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}
