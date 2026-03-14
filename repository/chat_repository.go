// Package repository provides data access layer implementations.
package repository

import (
	"database/sql"
	"fmt"
)

// Message represents a chat message record from the database.
type Message struct {
	ID      int64
	UserID  int64
	Role    string
	Content string
	Model   string
}

// ChatRepository provides database access for chat message records.
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository creates a new ChatRepository backed by the given database.
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// SaveMessage inserts a new message record for the given user and returns the
// persisted Message with its assigned ID.
func (r *ChatRepository) SaveMessage(userID int64, role, content, model string) (*Message, error) {
	result, err := r.db.Exec(
		"INSERT INTO messages (user_id, role, content, model) VALUES (?, ?, ?, ?)",
		userID, role, content, model,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &Message{
		ID:      id,
		UserID:  userID,
		Role:    role,
		Content: content,
		Model:   model,
	}, nil
}

// FindByUserID returns all messages belonging to the specified user ordered by
// created_at ascending. If the user has no messages, an empty (non-nil) slice
// is returned.
func (r *ChatRepository) FindByUserID(userID int64) ([]*Message, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, role, content, model FROM messages WHERE user_id = ? ORDER BY created_at ASC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	messages := make([]*Message, 0)
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Role, &msg.Content, &msg.Model); err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}
		messages = append(messages, &msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	return messages, nil
}
