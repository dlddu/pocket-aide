// Package devicetokens stores APNs device tokens registered by iOS clients.
// Per-user fan-out: ListByUserID returns tokens for a single user so the
// PR-monitor pipeline can push only to that user's devices after blacklist
// matching.
package devicetokens

import (
	"context"
	"database/sql"
	"fmt"
)

// Store is the user-scoped upsert facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

// Upsert inserts a new (user, token) row or, when the token already exists,
// rebinds it to the current user and bumps updated_at. Tokens are globally
// unique so the same physical device cannot be claimed by multiple accounts.
func (s *Store) Upsert(ctx context.Context, userID int64, token string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO device_tokens (user_id, token)
		VALUES (?, ?)
		ON CONFLICT(token) DO UPDATE SET
			user_id = excluded.user_id,
			updated_at = unixepoch()
	`, userID, token)
	if err != nil {
		return fmt.Errorf("upsert device token: %w", err)
	}
	return nil
}

// ListAll returns every registered token across all users. Kept for tests
// that pre-date per-user fan-out; production dispatch should use
// ListByUserID for the blacklist-matched push path.
func (s *Store) ListAll(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT token FROM device_tokens`)
	if err != nil {
		return nil, fmt.Errorf("list device tokens: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]string, 0)
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan device token: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// ListByUserID returns every token registered for a single user. A user with
// no devices returns an empty slice (not an error).
func (s *Store) ListByUserID(ctx context.Context, userID int64) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT token FROM device_tokens WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("list device tokens for user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]string, 0)
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan device token: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}
