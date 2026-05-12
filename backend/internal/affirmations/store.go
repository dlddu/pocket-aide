// Package affirmations persists user affirmation rows and exposes the CRUD
// surface used by the HTTP handlers. All operations are scoped to a userID so
// callers cannot accidentally read or mutate another user's data.
package affirmations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Priority is the rotation weight tier for an affirmation. The values match the
// strings persisted in SQLite (see migration 0002).
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// Valid reports whether p is one of the three accepted tiers.
func (p Priority) Valid() bool {
	switch p {
	case PriorityHigh, PriorityNormal, PriorityLow:
		return true
	}
	return false
}

// Affirmation is one row in the affirmations table.
type Affirmation struct {
	ID        int64    `json:"id"`
	UserID    int64    `json:"-"`
	Text      string   `json:"text"`
	Priority  Priority `json:"priority"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
}

// ErrNotFound is returned when no row matches the (userID, id) pair.
var ErrNotFound = errors.New("affirmation not found")

// Store wraps a *sql.DB with CRUD helpers.
type Store struct {
	db *sql.DB
}

// NewStore returns a Store backed by db.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// List returns all affirmations for the user, newest first.
func (s *Store) List(ctx context.Context, userID int64) ([]Affirmation, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, user_id, text, priority, created_at, updated_at
        FROM affirmations
        WHERE user_id = ?
        ORDER BY created_at DESC, id DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("list affirmations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Affirmation
	for rows.Next() {
		var a Affirmation
		if err := rows.Scan(&a.ID, &a.UserID, &a.Text, &a.Priority, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan affirmation: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate affirmations: %w", err)
	}
	return out, nil
}

// Get returns the single affirmation matching (userID, id) or ErrNotFound.
func (s *Store) Get(ctx context.Context, userID, id int64) (Affirmation, error) {
	var a Affirmation
	err := s.db.QueryRowContext(ctx, `
        SELECT id, user_id, text, priority, created_at, updated_at
        FROM affirmations
        WHERE user_id = ? AND id = ?
    `, userID, id).Scan(&a.ID, &a.UserID, &a.Text, &a.Priority, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Affirmation{}, ErrNotFound
	}
	if err != nil {
		return Affirmation{}, fmt.Errorf("get affirmation: %w", err)
	}
	return a, nil
}

// Create inserts a new affirmation for the user.
func (s *Store) Create(ctx context.Context, userID int64, text string, priority Priority) (Affirmation, error) {
	res, err := s.db.ExecContext(ctx, `
        INSERT INTO affirmations (user_id, text, priority)
        VALUES (?, ?, ?)
    `, userID, text, priority)
	if err != nil {
		return Affirmation{}, fmt.Errorf("insert affirmation: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Affirmation{}, fmt.Errorf("last insert id: %w", err)
	}
	return s.Get(ctx, userID, id)
}

// Update applies new text + priority. Returns ErrNotFound if the row does not
// belong to the user.
func (s *Store) Update(ctx context.Context, userID, id int64, text string, priority Priority) (Affirmation, error) {
	res, err := s.db.ExecContext(ctx, `
        UPDATE affirmations
        SET text = ?, priority = ?, updated_at = unixepoch()
        WHERE user_id = ? AND id = ?
    `, text, priority, userID, id)
	if err != nil {
		return Affirmation{}, fmt.Errorf("update affirmation: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return Affirmation{}, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return Affirmation{}, ErrNotFound
	}
	return s.Get(ctx, userID, id)
}

// Delete removes the row. ErrNotFound if it does not belong to the user.
func (s *Store) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `
        DELETE FROM affirmations
        WHERE user_id = ? AND id = ?
    `, userID, id)
	if err != nil {
		return fmt.Errorf("delete affirmation: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
