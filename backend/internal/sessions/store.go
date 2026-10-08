package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Store keeps the agent_sessions table: which session-platform sessions each user owns.
type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Add(ctx context.Context, userID int64, sessionID string) error {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO agent_sessions (session_id, user_id) VALUES (?, ?)`, sessionID, userID); err != nil {
		return fmt.Errorf("insert agent session: %w", err)
	}
	return nil
}

func (s *Store) Owns(ctx context.Context, userID int64, sessionID string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM agent_sessions WHERE session_id = ? AND user_id = ?`, sessionID, userID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("select agent session: %w", err)
	}
	return true, nil
}

func (s *Store) Owned(ctx context.Context, userID int64) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT session_id FROM agent_sessions WHERE user_id = ?`, userID)
	if err != nil {
		return nil, fmt.Errorf("list agent sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()
	owned := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan agent session: %w", err)
		}
		owned[id] = true
	}
	return owned, rows.Err()
}

func (s *Store) Remove(ctx context.Context, userID int64, sessionID string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM agent_sessions WHERE session_id = ? AND user_id = ?`, sessionID, userID); err != nil {
		return fmt.Errorf("delete agent session: %w", err)
	}
	return nil
}
