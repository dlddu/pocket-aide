// Package repository provides data access layer implementations.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrSentenceNotFound is returned when no sentence with the given criteria exists.
var ErrSentenceNotFound = errors.New("sentence not found")

// Sentence represents a sentences record from the database.
type Sentence struct {
	ID         int64
	UserID     int64
	CategoryID int64
	Content    string
	CreatedAt  string
	UpdatedAt  string
}

// SentenceRepository provides database access for sentences records.
type SentenceRepository struct {
	db *sql.DB
}

// NewSentenceRepository creates a new SentenceRepository backed by the given database.
func NewSentenceRepository(db *sql.DB) *SentenceRepository {
	return &SentenceRepository{db: db}
}

// Create inserts a new sentence for the given user and category and returns the persisted record.
// Returns an error if content is empty.
func (r *SentenceRepository) Create(userID int64, categoryID int64, content string) (*Sentence, error) {
	if content == "" {
		return nil, fmt.Errorf("content must not be empty")
	}

	result, err := r.db.Exec(
		`INSERT INTO sentences (user_id, category_id, content) VALUES (?, ?, ?)`,
		userID, categoryID, content,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert sentence: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.FindByID(id, userID)
}

// ListByUserID returns all sentences belonging to the given user ordered by id ASC.
// Returns a non-nil empty slice when the user has no sentences.
func (r *SentenceRepository) ListByUserID(userID int64) ([]*Sentence, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, category_id, content, created_at, updated_at
		 FROM sentences WHERE user_id = ? ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query sentences: %w", err)
	}
	defer rows.Close()

	sentences := make([]*Sentence, 0)
	for rows.Next() {
		s, err := scanSentence(rows)
		if err != nil {
			return nil, err
		}
		sentences = append(sentences, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return sentences, nil
}

// ListByCategoryID returns all sentences belonging to the given user and category ordered by id ASC.
// Returns a non-nil empty slice when there are no matching sentences.
func (r *SentenceRepository) ListByCategoryID(userID int64, categoryID int64) ([]*Sentence, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, category_id, content, created_at, updated_at
		 FROM sentences WHERE user_id = ? AND category_id = ? ORDER BY id ASC`,
		userID, categoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query sentences by category: %w", err)
	}
	defer rows.Close()

	sentences := make([]*Sentence, 0)
	for rows.Next() {
		s, err := scanSentence(rows)
		if err != nil {
			return nil, err
		}
		sentences = append(sentences, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return sentences, nil
}

// FindByID returns the sentence with the given ID owned by userID.
// Returns ErrSentenceNotFound when the ID does not exist or belongs to another user.
func (r *SentenceRepository) FindByID(id int64, userID int64) (*Sentence, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, category_id, content, created_at, updated_at
		 FROM sentences WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	s, err := scanSentenceRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrSentenceNotFound)
		}
		return nil, fmt.Errorf("failed to query sentence: %w", err)
	}
	return s, nil
}

// Update updates the content of the sentence with the given ID owned by userID.
// Returns ErrSentenceNotFound when the ID does not exist or belongs to another user.
func (r *SentenceRepository) Update(id int64, userID int64, content string) (*Sentence, error) {
	_, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(
		`UPDATE sentences SET content = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		content, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update sentence: %w", err)
	}

	return r.FindByID(id, userID)
}

// Delete removes the sentence with the given ID owned by userID.
// Returns ErrSentenceNotFound when 0 rows are affected.
func (r *SentenceRepository) Delete(id int64, userID int64) error {
	result, err := r.db.Exec(
		`DELETE FROM sentences WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete sentence: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", ErrSentenceNotFound)
	}
	return nil
}

// scanSentence scans a row from *sql.Rows into a Sentence struct.
func scanSentence(rows *sql.Rows) (*Sentence, error) {
	var s Sentence
	if err := rows.Scan(&s.ID, &s.UserID, &s.CategoryID, &s.Content, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan sentence: %w", err)
	}
	return &s, nil
}

// scanSentenceRow scans a *sql.Row into a Sentence struct.
func scanSentenceRow(row *sql.Row) (*Sentence, error) {
	var s Sentence
	if err := row.Scan(&s.ID, &s.UserID, &s.CategoryID, &s.Content, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	return &s, nil
}
