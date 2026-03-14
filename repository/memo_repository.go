// Package repository provides data access layer implementations.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrMemoNotFound is returned when no memo with the given criteria exists.
var ErrMemoNotFound = errors.New("memo not found")

// Memo represents a memo record from the database.
type Memo struct {
	ID        int64
	UserID    int64
	Content   string
	Source    string
	CreatedAt string
	UpdatedAt string
}

// MemoRepository provides database access for memo records.
type MemoRepository struct {
	db *sql.DB
}

// NewMemoRepository creates a new MemoRepository backed by the given database.
func NewMemoRepository(db *sql.DB) *MemoRepository {
	return &MemoRepository{db: db}
}

// Create inserts a new memo for the given user and returns the persisted record.
// Returns an error if content is empty. The source defaults to "text" if empty.
func (r *MemoRepository) Create(userID int64, content string, source string) (*Memo, error) {
	if content == "" {
		return nil, fmt.Errorf("content must not be empty")
	}
	if source == "" {
		source = "text"
	}

	result, err := r.db.Exec(
		`INSERT INTO memos (user_id, content, source) VALUES (?, ?, ?)`,
		userID, content, source,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert memo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.FindByID(id, userID)
}

// ListByUserID returns all memos belonging to the given user ordered by id ASC.
// Returns a non-nil empty slice when the user has no memos.
func (r *MemoRepository) ListByUserID(userID int64) ([]*Memo, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, content, source, created_at, updated_at
		 FROM memos WHERE user_id = ? ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query memos: %w", err)
	}
	defer rows.Close()

	memos := make([]*Memo, 0)
	for rows.Next() {
		m, err := scanMemo(rows)
		if err != nil {
			return nil, err
		}
		memos = append(memos, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return memos, nil
}

// FindByID returns the memo with the given ID owned by userID.
// Returns ErrMemoNotFound when the ID does not exist or belongs to another user.
func (r *MemoRepository) FindByID(id int64, userID int64) (*Memo, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, content, source, created_at, updated_at
		 FROM memos WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	m, err := scanMemoRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrMemoNotFound)
		}
		return nil, fmt.Errorf("failed to query memo: %w", err)
	}
	return m, nil
}

// Update updates the content of the memo with the given ID owned by userID.
// Returns ErrMemoNotFound when the ID does not exist or belongs to another user.
func (r *MemoRepository) Update(id int64, userID int64, content string) (*Memo, error) {
	_, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(
		`UPDATE memos SET content = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		content, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update memo: %w", err)
	}

	return r.FindByID(id, userID)
}

// Delete removes the memo with the given ID owned by userID.
// Returns ErrMemoNotFound when 0 rows are affected.
func (r *MemoRepository) Delete(id int64, userID int64) error {
	result, err := r.db.Exec(
		`DELETE FROM memos WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete memo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", ErrMemoNotFound)
	}
	return nil
}

// scanMemo scans a row from *sql.Rows into a Memo struct.
func scanMemo(rows *sql.Rows) (*Memo, error) {
	var m Memo
	if err := rows.Scan(&m.ID, &m.UserID, &m.Content, &m.Source, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan memo: %w", err)
	}
	return &m, nil
}

// scanMemoRow scans a *sql.Row into a Memo struct.
func scanMemoRow(row *sql.Row) (*Memo, error) {
	var m Memo
	if err := row.Scan(&m.ID, &m.UserID, &m.Content, &m.Source, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}
