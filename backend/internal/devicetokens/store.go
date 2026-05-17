// Package devicetokens stores APNs device tokens registered by iOS clients.
// Broadcast-style fan-out: ListAll returns every token across all users so the
// PR-monitor pipeline can push to each registered device.
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

// ListAll returns every registered token across all users. The PR-monitor
// pipeline broadcasts to all of them — single-user assumption per the draft.
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
