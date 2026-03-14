// Package chat_test contains unit tests for ChatService.
// These tests verify that ChatService correctly orchestrates LLM routing
// and message history persistence.
//
// ChatService depends on a ChatRepository and an LLM Router. These tests use
// mock implementations so they run without a real database or LLM, keeping
// them fast and isolated.
//
// TDD Red Phase: the ChatService type and its methods do not yet exist.
// All tests are expected to fail until service/chat/chat_service.go is implemented.
package chat_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/service/chat"
)

// ---------------------------------------------------------------------------
// Mock ChatRepository
// ---------------------------------------------------------------------------

// mockChatRepository is a test double for the ChatRepository interface that
// ChatService will depend on.
// Each field holds an optional stub function; if nil the method returns an error.
type mockChatRepository struct {
	saveMessageFunc  func(userID int64, role, content, model string) (*repository.Message, error)
	findByUserIDFunc func(userID int64) ([]*repository.Message, error)
}

func (m *mockChatRepository) SaveMessage(userID int64, role, content, model string) (*repository.Message, error) {
	if m.saveMessageFunc != nil {
		return m.saveMessageFunc(userID, role, content, model)
	}
	return nil, errors.New("mockChatRepository.SaveMessage not configured")
}

func (m *mockChatRepository) FindByUserID(userID int64) ([]*repository.Message, error) {
	if m.findByUserIDFunc != nil {
		return m.findByUserIDFunc(userID)
	}
	return nil, errors.New("mockChatRepository.FindByUserID not configured")
}

// ---------------------------------------------------------------------------
// Mock LLMRouter
// ---------------------------------------------------------------------------

// mockLLMRouter is a test double for the LLMRouter interface that ChatService
// will depend on for model routing decisions.
type mockLLMRouter struct {
	routeFunc func(ctx context.Context, prompt, model string) (string, error)
}

func (m *mockLLMRouter) Route(ctx context.Context, prompt, model string) (string, error) {
	if m.routeFunc != nil {
		return m.routeFunc(ctx, prompt, model)
	}
	return "", errors.New("mockLLMRouter.Route not configured")
}

// ---------------------------------------------------------------------------
// SendMessage
// ---------------------------------------------------------------------------

// TestChatService_SendMessage_Success verifies that SendMessage calls the LLM
// router with the provided prompt and returns the LLM response.
//
// Scenario:
//
//	LLM router returns "AI response".
//	SendMessage(ctx, userID=1, message="Hello", model="")
//	→ "AI response", nil
func TestChatService_SendMessage_Success(t *testing.T) {
	// Arrange
	router := &mockLLMRouter{
		routeFunc: func(ctx context.Context, prompt, model string) (string, error) {
			return "AI response", nil
		},
	}
	repo := &mockChatRepository{
		saveMessageFunc: func(userID int64, role, content, model string) (*repository.Message, error) {
			return &repository.Message{ID: 1, UserID: userID, Role: role, Content: content, Model: model}, nil
		},
	}
	svc := chat.NewChatService(repo, router)

	// Act
	response, err := svc.SendMessage(context.Background(), 1, "Hello", "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if response != "AI response" {
		t.Errorf("expected 'AI response', got '%s'", response)
	}
}

// TestChatService_SendMessage_SavesHistory verifies that SendMessage saves
// both the user message and the assistant response to the repository.
//
// Scenario:
//
//	SendMessage(ctx, userID=1, message="Hello", model="")
//	→ repository.SaveMessage called twice:
//	   1st call: role="user", content="Hello"
//	   2nd call: role="assistant", content=<LLM response>
func TestChatService_SendMessage_SavesHistory(t *testing.T) {
	// Arrange
	var savedMessages []struct {
		role    string
		content string
		model   string
	}

	router := &mockLLMRouter{
		routeFunc: func(ctx context.Context, prompt, model string) (string, error) {
			return "Assistant reply", nil
		},
	}
	repo := &mockChatRepository{
		saveMessageFunc: func(userID int64, role, content, model string) (*repository.Message, error) {
			savedMessages = append(savedMessages, struct {
				role    string
				content string
				model   string
			}{role, content, model})
			return &repository.Message{ID: int64(len(savedMessages)), UserID: userID, Role: role, Content: content, Model: model}, nil
		},
	}
	svc := chat.NewChatService(repo, router)

	// Act
	_, err := svc.SendMessage(context.Background(), 1, "Hello", "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(savedMessages) != 2 {
		t.Fatalf("expected 2 saved messages (user + assistant), got %d", len(savedMessages))
	}

	// First saved message must be the user message
	if savedMessages[0].role != "user" {
		t.Errorf("expected first saved message role 'user', got '%s'", savedMessages[0].role)
	}
	if savedMessages[0].content != "Hello" {
		t.Errorf("expected first saved message content 'Hello', got '%s'", savedMessages[0].content)
	}

	// Second saved message must be the assistant response
	if savedMessages[1].role != "assistant" {
		t.Errorf("expected second saved message role 'assistant', got '%s'", savedMessages[1].role)
	}
	if savedMessages[1].content != "Assistant reply" {
		t.Errorf("expected second saved message content 'Assistant reply', got '%s'", savedMessages[1].content)
	}
}

// TestChatService_SendMessage_WithModel verifies that when a specific model is
// requested, SendMessage passes that model name to the LLM router and stores
// it in the saved assistant message.
//
// Scenario:
//
//	SendMessage(ctx, userID=1, message="Hello", model="gpt-4o")
//	→ router.Route called with model="gpt-4o"
//	→ assistant message saved with model="gpt-4o"
func TestChatService_SendMessage_WithModel(t *testing.T) {
	// Arrange
	var capturedModel string
	var savedAssistantModel string

	router := &mockLLMRouter{
		routeFunc: func(ctx context.Context, prompt, model string) (string, error) {
			capturedModel = model
			return "gpt-4o response", nil
		},
	}
	repo := &mockChatRepository{
		saveMessageFunc: func(userID int64, role, content, model string) (*repository.Message, error) {
			if role == "assistant" {
				savedAssistantModel = model
			}
			return &repository.Message{ID: 1, UserID: userID, Role: role, Content: content, Model: model}, nil
		},
	}
	svc := chat.NewChatService(repo, router)

	// Act
	_, err := svc.SendMessage(context.Background(), 1, "Hello", "gpt-4o")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if capturedModel != "gpt-4o" {
		t.Errorf("expected router to receive model 'gpt-4o', got '%s'", capturedModel)
	}
	if savedAssistantModel != "gpt-4o" {
		t.Errorf("expected assistant message saved with model 'gpt-4o', got '%s'", savedAssistantModel)
	}
}

// TestChatService_SendMessage_LLMError_ReturnsError verifies that when the
// LLM router returns an error, SendMessage propagates it.
//
// Scenario:
//
//	LLM router returns an error.
//	SendMessage(ctx, userID=1, message="Hello", model="")
//	→ non-nil error, empty response
func TestChatService_SendMessage_LLMError_ReturnsError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("LLM service unavailable")
	router := &mockLLMRouter{
		routeFunc: func(ctx context.Context, prompt, model string) (string, error) {
			return "", expectedErr
		},
	}
	repo := &mockChatRepository{
		saveMessageFunc: func(userID int64, role, content, model string) (*repository.Message, error) {
			return &repository.Message{ID: 1, UserID: userID, Role: role, Content: content, Model: model}, nil
		},
	}
	svc := chat.NewChatService(repo, router)

	// Act
	response, err := svc.SendMessage(context.Background(), 1, "Hello", "")

	// Assert
	if err == nil {
		t.Error("expected error from LLM failure, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
	if response != "" {
		t.Errorf("expected empty response on error, got '%s'", response)
	}
}

// ---------------------------------------------------------------------------
// GetHistory
// ---------------------------------------------------------------------------

// TestChatService_GetHistory_Success verifies that GetHistory returns all
// messages for the given user from the repository.
//
// Scenario:
//
//	Repository returns 2 messages for userID=1.
//	GetHistory(ctx, userID=1)
//	→ []*repository.Message with len == 2, nil
func TestChatService_GetHistory_Success(t *testing.T) {
	// Arrange
	repoMessages := []*repository.Message{
		{ID: 1, UserID: 1, Role: "user", Content: "Hello", Model: ""},
		{ID: 2, UserID: 1, Role: "assistant", Content: "Hi there", Model: "gpt-4o"},
	}
	repo := &mockChatRepository{
		findByUserIDFunc: func(userID int64) ([]*repository.Message, error) {
			return repoMessages, nil
		},
	}
	router := &mockLLMRouter{}
	svc := chat.NewChatService(repo, router)

	// Act
	messages, err := svc.GetHistory(context.Background(), 1)

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
	if messages[1].Role != "assistant" {
		t.Errorf("expected second message role 'assistant', got '%s'", messages[1].Role)
	}
}

// TestChatService_GetHistory_Empty verifies that GetHistory returns an empty
// (non-nil) slice when the user has no chat history.
//
// Scenario:
//
//	Repository returns empty slice for userID=1.
//	GetHistory(ctx, userID=1)
//	→ []*repository.Message{} (empty slice, not nil), nil
func TestChatService_GetHistory_Empty(t *testing.T) {
	// Arrange
	repo := &mockChatRepository{
		findByUserIDFunc: func(userID int64) ([]*repository.Message, error) {
			return []*repository.Message{}, nil
		},
	}
	router := &mockLLMRouter{}
	svc := chat.NewChatService(repo, router)

	// Act
	messages, err := svc.GetHistory(context.Background(), 1)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if messages == nil {
		t.Error("expected non-nil slice for empty history, got nil")
	}
	if len(messages) != 0 {
		t.Errorf("expected empty slice, got %d messages", len(messages))
	}
}

// TestChatService_GetHistory_RepositoryError_ReturnsError verifies that when
// the repository returns an error, GetHistory propagates it.
//
// Scenario:
//
//	Repository returns an error for FindByUserID.
//	GetHistory(ctx, userID=1)
//	→ non-nil error, nil messages
func TestChatService_GetHistory_RepositoryError_ReturnsError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("database error")
	repo := &mockChatRepository{
		findByUserIDFunc: func(userID int64) ([]*repository.Message, error) {
			return nil, expectedErr
		},
	}
	router := &mockLLMRouter{}
	svc := chat.NewChatService(repo, router)

	// Act
	messages, err := svc.GetHistory(context.Background(), 1)

	// Assert
	if err == nil {
		t.Error("expected error from repository failure, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
	if messages != nil {
		t.Errorf("expected nil messages on error, got %v", messages)
	}
}
