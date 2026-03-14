// Package repository_test contains unit tests for the ChatRepository.
// These tests use an in-memory SQLite database via testutil.NewTestDB and
// verify that ChatRepository correctly reads and writes message records.
//
// TDD Red Phase: the ChatRepository type and its methods do not yet exist.
// All tests are expected to fail until repository/chat_repository.go is implemented.
package repository_test

import (
	"testing"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// SaveMessage
// ---------------------------------------------------------------------------

// TestChatRepository_SaveMessage_Success verifies that SaveMessage inserts a
// message record and returns the persisted message with a non-zero ID.
//
// Scenario:
//
//	Seed: user with id=1.
//	SaveMessage(userID=1, role="user", content="Hello", model="")
//	→ Message{ID: >0, UserID: 1, Role: "user", Content: "Hello"}, nil
func TestChatRepository_SaveMessage_Success(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		1, "user@example.com", "hashed_password",
	)
	repo := repository.NewChatRepository(tdb.DB)

	// Act
	msg, err := repo.SaveMessage(1, "user", "Hello", "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message, got nil")
	}
	if msg.ID == 0 {
		t.Error("expected message ID to be non-zero after creation")
	}
	if msg.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", msg.UserID)
	}
	if msg.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", msg.Role)
	}
	if msg.Content != "Hello" {
		t.Errorf("expected content 'Hello', got '%s'", msg.Content)
	}
}

// TestChatRepository_SaveMessage_WithModel verifies that SaveMessage correctly
// stores the model field when a non-empty model name is provided.
//
// Scenario:
//
//	Seed: user with id=1.
//	SaveMessage(userID=1, role="assistant", content="Hi there", model="gpt-4o")
//	→ Message{Model: "gpt-4o"}, nil
func TestChatRepository_SaveMessage_WithModel(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		1, "user@example.com", "hashed_password",
	)
	repo := repository.NewChatRepository(tdb.DB)

	// Act
	msg, err := repo.SaveMessage(1, "assistant", "Hi there", "gpt-4o")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message, got nil")
	}
	if msg.Model != "gpt-4o" {
		t.Errorf("expected model 'gpt-4o', got '%s'", msg.Model)
	}
	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got '%s'", msg.Role)
	}

	// Verify persistence: query directly from the DB
	var storedModel string
	row := tdb.DB.QueryRow("SELECT model FROM messages WHERE id = ?", msg.ID)
	if err := row.Scan(&storedModel); err != nil {
		t.Fatalf("failed to query messages table: %v", err)
	}
	if storedModel != "gpt-4o" {
		t.Errorf("expected stored model 'gpt-4o', got '%s'", storedModel)
	}
}

// ---------------------------------------------------------------------------
// FindByUserID
// ---------------------------------------------------------------------------

// TestChatRepository_FindByUserID_ReturnsMessages verifies that FindByUserID
// returns all messages belonging to the specified user in chronological order.
//
// Scenario:
//
//	Seed: user id=1 with two messages.
//	FindByUserID(1)
//	→ []Message with len == 2, ordered by created_at asc
func TestChatRepository_FindByUserID_ReturnsMessages(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		1, "user@example.com", "hashed_password",
	)
	tdb.Seed(t,
		"INSERT INTO messages (user_id, role, content, model) VALUES (?, ?, ?, ?)",
		1, "user", "First message", "",
	)
	tdb.Seed(t,
		"INSERT INTO messages (user_id, role, content, model) VALUES (?, ?, ?, ?)",
		1, "assistant", "First response", "gpt-4o",
	)
	repo := repository.NewChatRepository(tdb.DB)

	// Act
	messages, err := repo.FindByUserID(1)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != "user" {
		t.Errorf("expected first message role 'user', got '%s'", messages[0].Role)
	}
	if messages[0].Content != "First message" {
		t.Errorf("expected first message content 'First message', got '%s'", messages[0].Content)
	}
	if messages[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got '%s'", messages[1].Role)
	}
}

// TestChatRepository_FindByUserID_EmptyHistory verifies that FindByUserID
// returns an empty (non-nil) slice when the user has no messages.
//
// Scenario:
//
//	Seed: user id=1 with no messages.
//	FindByUserID(1)
//	→ []Message{} (empty slice, not nil), nil
func TestChatRepository_FindByUserID_EmptyHistory(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		1, "user@example.com", "hashed_password",
	)
	repo := repository.NewChatRepository(tdb.DB)

	// Act
	messages, err := repo.FindByUserID(1)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if messages == nil {
		t.Error("expected non-nil slice, got nil")
	}
	if len(messages) != 0 {
		t.Errorf("expected empty slice, got %d messages", len(messages))
	}
}

// TestChatRepository_FindByUserID_OnlyOwnerMessages verifies that FindByUserID
// returns only messages belonging to the requested user and excludes messages
// from other users.
//
// Scenario:
//
//	Seed: user id=1 with 1 message, user id=2 with 1 message.
//	FindByUserID(1)
//	→ []Message with len == 1, all UserID == 1
func TestChatRepository_FindByUserID_OnlyOwnerMessages(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		1, "owner@example.com", "hashed_password",
	)
	tdb.Seed(t,
		"INSERT INTO users (id, email, password_hash) VALUES (?, ?, ?)",
		2, "other@example.com", "hashed_password",
	)
	tdb.Seed(t,
		"INSERT INTO messages (user_id, role, content, model) VALUES (?, ?, ?, ?)",
		1, "user", "Owner message", "",
	)
	tdb.Seed(t,
		"INSERT INTO messages (user_id, role, content, model) VALUES (?, ?, ?, ?)",
		2, "user", "Other user message", "",
	)
	repo := repository.NewChatRepository(tdb.DB)

	// Act
	messages, err := repo.FindByUserID(1)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message for user 1, got %d", len(messages))
	}
	if messages[0].UserID != 1 {
		t.Errorf("expected message UserID 1, got %d", messages[0].UserID)
	}
	if messages[0].Content != "Owner message" {
		t.Errorf("expected content 'Owner message', got '%s'", messages[0].Content)
	}
}
