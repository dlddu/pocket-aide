// Package repository provides data access layer implementations.
package repository

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrSentenceCategoryNotFound is returned when no sentence category with the given criteria exists.
var ErrSentenceCategoryNotFound = errors.New("sentence category not found")

// SentenceCategory represents a sentence_categories record from the database.
type SentenceCategory struct {
	ID        int64
	UserID    int64
	Name      string
	CreatedAt string
	UpdatedAt string
}

// SentenceCategoryRepository provides database access for sentence_categories records.
type SentenceCategoryRepository struct {
	db *sql.DB
}

// NewSentenceCategoryRepository creates a new SentenceCategoryRepository backed by the given database.
func NewSentenceCategoryRepository(db *sql.DB) *SentenceCategoryRepository {
	return &SentenceCategoryRepository{db: db}
}

// Create inserts a new sentence category for the given user and returns the persisted record.
// Returns an error if name is empty.
func (r *SentenceCategoryRepository) Create(userID int64, name string) (*SentenceCategory, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	result, err := r.db.Exec(
		`INSERT INTO sentence_categories (user_id, name) VALUES (?, ?)`,
		userID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert sentence category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return r.FindByID(id, userID)
}

// ListByUserID returns all sentence categories belonging to the given user ordered by id ASC.
// Returns a non-nil empty slice when the user has no categories.
func (r *SentenceCategoryRepository) ListByUserID(userID int64) ([]*SentenceCategory, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, created_at, updated_at
		 FROM sentence_categories WHERE user_id = ? ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query sentence categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*SentenceCategory, 0)
	for rows.Next() {
		c, err := scanSentenceCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return categories, nil
}

// FindByID returns the sentence category with the given ID owned by userID.
// Returns ErrSentenceCategoryNotFound when the ID does not exist or belongs to another user.
func (r *SentenceCategoryRepository) FindByID(id int64, userID int64) (*SentenceCategory, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, name, created_at, updated_at
		 FROM sentence_categories WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	c, err := scanSentenceCategoryRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w", ErrSentenceCategoryNotFound)
		}
		return nil, fmt.Errorf("failed to query sentence category: %w", err)
	}
	return c, nil
}

// Update updates the name of the sentence category with the given ID owned by userID.
// Returns ErrSentenceCategoryNotFound when the ID does not exist or belongs to another user.
func (r *SentenceCategoryRepository) Update(id int64, userID int64, name string) (*SentenceCategory, error) {
	_, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(
		`UPDATE sentence_categories SET name = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND user_id = ?`,
		name, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update sentence category: %w", err)
	}

	return r.FindByID(id, userID)
}

// Delete removes the sentence category with the given ID owned by userID.
// Returns ErrSentenceCategoryNotFound when 0 rows are affected.
func (r *SentenceCategoryRepository) Delete(id int64, userID int64) error {
	result, err := r.db.Exec(
		`DELETE FROM sentence_categories WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete sentence category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w", ErrSentenceCategoryNotFound)
	}
	return nil
}

// scanSentenceCategory scans a row from *sql.Rows into a SentenceCategory struct.
func scanSentenceCategory(rows *sql.Rows) (*SentenceCategory, error) {
	var c SentenceCategory
	if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan sentence category: %w", err)
	}
	return &c, nil
}

// scanSentenceCategoryRow scans a *sql.Row into a SentenceCategory struct.
func scanSentenceCategoryRow(row *sql.Row) (*SentenceCategory, error) {
	var c SentenceCategory
	if err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	return &c, nil
}
