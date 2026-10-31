// Package devicetokens stores push device tokens registered by iOS (APNs) and Android (FCM) clients.
package devicetokens

import (
	"context"
	"database/sql"
	"fmt"
)

// Platform values stored in device_tokens.platform.
const (
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
)

// Device is one registered push token and the platform whose push service accepts it.
type Device struct {
	Token    string
	Platform string
}

// Store is the user-scoped upsert facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

// Upsert inserts a new (user, token) row or, when the token already exists,
// rebinds it to the current user and bumps updated_at. Tokens are globally
// unique so the same physical device cannot be claimed by multiple accounts.
func (s *Store) Upsert(ctx context.Context, userID int64, token, platform string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO device_tokens (user_id, token, platform)
		VALUES (?, ?, ?)
		ON CONFLICT(token) DO UPDATE SET
			user_id = excluded.user_id,
			platform = excluded.platform,
			updated_at = unixepoch()
	`, userID, token, platform)
	if err != nil {
		return fmt.Errorf("upsert device token: %w", err)
	}
	return nil
}

// ListByUserID returns every device registered for a single user.
func (s *Store) ListByUserID(ctx context.Context, userID int64) ([]Device, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT token, platform FROM device_tokens WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("list device tokens for user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]Device, 0)
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.Token, &d.Platform); err != nil {
			return nil, fmt.Errorf("scan device token: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

func (s *Store) Delete(ctx context.Context, userID int64, token string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM device_tokens WHERE user_id = ? AND token = ?`, userID, token); err != nil {
		return fmt.Errorf("delete device token: %w", err)
	}
	return nil
}
