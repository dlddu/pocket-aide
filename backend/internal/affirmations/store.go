// Package affirmations is the storage layer for PRD-5 (다짐) — user-scoped
// CRUD over the affirmations table with a string priority of high|normal|low.
package affirmations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Priority is the exposure-frequency tier. The DB CHECK constraint mirrors
// these three string values exactly so the API and DB stay in sync.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// Valid reports whether p is one of the three allowed values.
func (p Priority) Valid() bool {
	switch p {
	case PriorityHigh, PriorityNormal, PriorityLow:
		return true
	default:
		return false
	}
}

// Affirmation is one row of the affirmations table.
type Affirmation struct {
	ID        int64    `json:"id"`
	Text      string   `json:"text"`
	Priority  Priority `json:"priority"`
	CreatedAt int64    `json:"created_at"`
	UpdatedAt int64    `json:"updated_at"`
}

// ErrNotFound is returned when a query targets a row that does not exist or
// is not owned by the caller.
var ErrNotFound = errors.New("affirmation not found")

// Store is the user-scoped CRUD facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

// List returns every affirmation that belongs to userID, ordered newest first.
func (s *Store) List(ctx context.Context, userID int64) ([]Affirmation, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, text, priority, created_at, updated_at
		FROM affirmations
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list affirmations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]Affirmation, 0)
	for rows.Next() {
		var a Affirmation
		if err := rows.Scan(&a.ID, &a.Text, &a.Priority, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan affirmation: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// Get returns a single affirmation that belongs to userID.
func (s *Store) Get(ctx context.Context, userID, id int64) (Affirmation, error) {
	var a Affirmation
	err := s.db.QueryRowContext(ctx, `
		SELECT id, text, priority, created_at, updated_at
		FROM affirmations
		WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&a.ID, &a.Text, &a.Priority, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Affirmation{}, ErrNotFound
	}
	if err != nil {
		return Affirmation{}, fmt.Errorf("get affirmation: %w", err)
	}
	return a, nil
}

// Create inserts a new affirmation for userID and returns the row.
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

// Update changes text and/or priority on an existing affirmation. Both args
// are required (full overwrite) — partial updates are intentionally not
// supported to keep the contract small.
func (s *Store) Update(ctx context.Context, userID, id int64, text string, priority Priority) (Affirmation, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE affirmations
		SET text = ?, priority = ?, updated_at = unixepoch()
		WHERE id = ? AND user_id = ?
	`, text, priority, id, userID)
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

// Delete removes an affirmation that belongs to userID.
func (s *Store) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM affirmations WHERE id = ? AND user_id = ?`, id, userID)
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
