package externalnotify

import (
	"context"
	"database/sql"
	"fmt"
)

// Notification is what a caller asked to send, with the requester the caller reported (PRD-12 AC17).
type Notification struct {
	Title     string
	Body      string
	Requester string
}

// Item is one stored notification with the registered caller that sent it (PRD-12 AC17).
type Item struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	CallerName string `json:"caller_name"`
	CallerKind string `json:"caller_kind"`
	Requester  string `json:"requester"`
	Devices    int    `json:"devices"`
	Sent       int    `json:"sent"`
	ReadAt     *int64 `json:"read_at,omitempty"`
	CreatedAt  int64  `json:"created_at"`
}

// Store persists external notifications per user (PRD-12 AC17).
type Store struct {
	db *sql.DB
}

// NewStore wires a Store onto an open *sql.DB.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Insert records a notification before it is sent.
func (s *Store) Insert(ctx context.Context, userID, callerKeyID int64, n Notification) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO external_notifications (user_id, caller_key_id, requester, title, body)
		VALUES (?, ?, ?, ?, ?)
	`, userID, callerKeyID, n.Requester, n.Title, n.Body)
	if err != nil {
		return 0, fmt.Errorf("insert external notification: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// SetDelivery stores how many devices the send reached.
func (s *Store) SetDelivery(ctx context.Context, id int64, r Result) error {
	if _, err := s.db.ExecContext(ctx, `
		UPDATE external_notifications SET devices = ?, sent = ? WHERE id = ?
	`, r.Devices, r.Sent, id); err != nil {
		return fmt.Errorf("set external notification delivery: %w", err)
	}
	return nil
}

// Delete drops a row whose send never started.
func (s *Store) Delete(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM external_notifications WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete external notification: %w", err)
	}
	return nil
}

const itemColumns = `n.id, n.title, n.body, k.name, k.kind, n.requester, n.devices, n.sent, n.read_at, n.created_at`

// List returns up to limit of userID's notifications, newest first, older than beforeID when it is positive.
func (s *Store) List(ctx context.Context, userID int64, limit int, beforeID int64) ([]Item, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	var (
		rows *sql.Rows
		err  error
	)
	if beforeID > 0 {
		rows, err = s.db.QueryContext(ctx, `SELECT `+itemColumns+`
			FROM external_notifications n
			JOIN approval_caller_keys k ON k.id = n.caller_key_id
			WHERE n.user_id = ? AND n.id < ?
			ORDER BY n.id DESC
			LIMIT ?`, userID, beforeID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `SELECT `+itemColumns+`
			FROM external_notifications n
			JOIN approval_caller_keys k ON k.id = n.caller_key_id
			WHERE n.user_id = ?
			ORDER BY n.id DESC
			LIMIT ?`, userID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list external notifications: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]Item, 0)
	for rows.Next() {
		var (
			it     Item
			readAt sql.NullInt64
		)
		if err := rows.Scan(&it.ID, &it.Title, &it.Body, &it.CallerName, &it.CallerKind,
			&it.Requester, &it.Devices, &it.Sent, &readAt, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan external notification: %w", err)
		}
		if readAt.Valid {
			t := readAt.Int64
			it.ReadAt = &t
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// UnreadCount is how many of userID's notifications have not been seen in the history screen.
func (s *Store) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM external_notifications WHERE user_id = ? AND read_at IS NULL
	`, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("count unread external notifications: %w", err)
	}
	return n, nil
}

// MarkRead stamps userID's unread notifications up to and including upToID, the newest one the screen showed.
func (s *Store) MarkRead(ctx context.Context, userID, upToID int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE external_notifications SET read_at = unixepoch()
		WHERE user_id = ? AND read_at IS NULL AND id <= ?
	`, userID, upToID)
	if err != nil {
		return 0, fmt.Errorf("mark external notifications read: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return n, nil
}
