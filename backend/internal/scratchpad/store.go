// Package scratchpad is the storage layer for PRD-4 (임시 공간) — unclassified
// captures that stay here until the user moves them into another area.
package scratchpad

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Source is how an item was captured (PRD-4 AC3). The DB CHECK constraint
// mirrors these values.
type Source string

const (
	SourceText     Source = "text"
	SourceVoice    Source = "voice"
	SourceShortcut Source = "shortcut"
)

// Valid reports whether s is one of the three capture paths.
func (s Source) Valid() bool {
	switch s {
	case SourceText, SourceVoice, SourceShortcut:
		return true
	default:
		return false
	}
}

// Target is an area an item can be moved into (PRD-4 AC4).
type Target string

const (
	TargetPersonal    Target = "personal"
	TargetWork        Target = "work"
	TargetAffirmation Target = "affirmation"
	TargetRoutine     Target = "routine"
)

var moveInserts = map[Target]string{
	TargetPersonal:    `INSERT INTO personal_todos (user_id, title) VALUES (?, ?)`,
	TargetWork:        `INSERT INTO work_todos (user_id, title) VALUES (?, ?)`,
	TargetAffirmation: `INSERT INTO affirmations (user_id, text, priority) VALUES (?, ?, 'normal')`,
	TargetRoutine:     `INSERT INTO routines (user_id, name, start_day) VALUES (?, ?, date('now'))`,
}

// Valid reports whether t is a supported move target.
func (t Target) Valid() bool {
	_, ok := moveInserts[t]
	return ok
}

// Item is one row of the scratchpad_items table.
type Item struct {
	ID         int64  `json:"id"`
	Text       string `json:"text"`
	Source     Source `json:"source"`
	CapturedAt int64  `json:"captured_at"`
	CreatedAt  int64  `json:"created_at"`
}

// ErrNotFound is returned when a query targets a row that does not exist or
// is not owned by the caller.
var ErrNotFound = errors.New("scratchpad item not found")

// ErrInvalidTarget is returned for a move target outside the supported set.
var ErrInvalidTarget = errors.New("invalid move target")

// Store is the user-scoped facade.
type Store struct {
	db *sql.DB
}

// New wires a Store onto an open *sql.DB.
func New(db *sql.DB) *Store { return &Store{db: db} }

const columns = `id, text, source, captured_at, created_at`

type scanner interface{ Scan(dest ...any) error }

func scanItem(s scanner) (Item, error) {
	var it Item
	err := s.Scan(&it.ID, &it.Text, &it.Source, &it.CapturedAt, &it.CreatedAt)
	return it, err
}

// List returns every item that belongs to userID, most recent capture first.
func (s *Store) List(ctx context.Context, userID int64) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT `+columns+`
		FROM scratchpad_items
		WHERE user_id = ?
		ORDER BY captured_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list scratchpad: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]Item, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan scratchpad item: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}
	return out, nil
}

// Get returns a single item that belongs to userID.
func (s *Store) Get(ctx context.Context, userID, id int64) (Item, error) {
	it, err := scanItem(s.db.QueryRowContext(ctx, `
		SELECT `+columns+`
		FROM scratchpad_items
		WHERE id = ? AND user_id = ?
	`, id, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Item{}, ErrNotFound
	}
	if err != nil {
		return Item{}, fmt.Errorf("get scratchpad item: %w", err)
	}
	return it, nil
}

// Create inserts a new item for userID. A nil capturedAt means "now".
func (s *Store) Create(ctx context.Context, userID int64, text string, source Source, capturedAt *int64) (Item, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO scratchpad_items (user_id, text, source, captured_at)
		VALUES (?, ?, ?, COALESCE(?, unixepoch()))
	`, userID, text, source, capturedAt)
	if err != nil {
		return Item{}, fmt.Errorf("insert scratchpad item: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Item{}, fmt.Errorf("last insert id: %w", err)
	}
	return s.Get(ctx, userID, id)
}

// Delete removes an item that belongs to userID.
func (s *Store) Delete(ctx context.Context, userID, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM scratchpad_items WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete scratchpad item: %w", err)
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

// Move creates a row in the target area from the item's text and removes the
// item, in one transaction, so an item is never in both places or neither
// (PRD-4 AC4). It returns the id of the row created in the target table.
func (s *Store) Move(ctx context.Context, userID, id int64, target Target) (int64, error) {
	insert, ok := moveInserts[target]
	if !ok {
		return 0, ErrInvalidTarget
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin move: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var text string
	err = tx.QueryRowContext(ctx, `SELECT text FROM scratchpad_items WHERE id = ? AND user_id = ?`, id, userID).Scan(&text)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("read moved item: %w", err)
	}
	res, err := tx.ExecContext(ctx, insert, userID, text)
	if err != nil {
		return 0, fmt.Errorf("insert into %s: %w", target, err)
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM scratchpad_items WHERE id = ? AND user_id = ?`, id, userID); err != nil {
		return 0, fmt.Errorf("remove moved item: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit move: %w", err)
	}
	return newID, nil
}
