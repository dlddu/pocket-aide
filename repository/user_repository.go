// Package repository provides data access layer implementations.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrDuplicateEmail is returned when attempting to create a user with an email
// that already exists in the database.
var ErrDuplicateEmail = errors.New("duplicate email")

// ErrUserNotFound is returned when no user with the given criteria exists.
var ErrUserNotFound = errors.New("user not found")

// User represents a user record from the database.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

// UserRepository provides database access for user records.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository backed by the given database.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser inserts a new user with the given email and password hash.
// It returns the newly created user with its assigned ID.
// If the email already exists, ErrDuplicateEmail is returned.
func (r *UserRepository) CreateUser(email, passwordHash string) (*User, error) {
	result, err := r.db.Exec(
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, passwordHash,
	)
	if err != nil {
		if isSQLiteUniqueConstraintError(err) {
			return nil, fmt.Errorf("%w", ErrDuplicateEmail)
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
	}, nil
}

// FindByEmail looks up a user by their email address.
// If no user is found, ErrUserNotFound is returned.
func (r *UserRepository) FindByEmail(email string) (*User, error) {
	row := r.db.QueryRow(
		"SELECT id, email, password_hash FROM users WHERE email = ?",
		email,
	)

	var user User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// FindByID looks up a user by their numeric ID.
// If no user is found, ErrUserNotFound is returned.
func (r *UserRepository) FindByID(id int64) (*User, error) {
	row := r.db.QueryRow(
		"SELECT id, email, password_hash FROM users WHERE id = ?",
		id,
	)

	var user User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}

	return &user, nil
}

// isSQLiteUniqueConstraintError returns true when the error is a SQLite
// UNIQUE constraint violation.
func isSQLiteUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "constraint failed")
}
