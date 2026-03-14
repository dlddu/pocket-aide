// Package chat provides business logic for AI chat functionality.
package chat

import (
	"context"
	"fmt"

	"github.com/dlddu/pocket-aide/repository"
)

// ChatRepository is the interface that ChatService uses to persist and retrieve
// chat messages. Defined here so that the service layer does not depend
// directly on the concrete repository type.
type ChatRepository interface {
	SaveMessage(userID int64, role, content, model string) (*repository.Message, error)
	FindByUserID(userID int64) ([]*repository.Message, error)
}

// LLMRouter is the interface that ChatService uses to route prompts to the
// appropriate LLM provider.
type LLMRouter interface {
	Route(ctx context.Context, prompt, model string) (string, error)
}

// ChatService orchestrates LLM routing and message history persistence.
type ChatService struct {
	repo   ChatRepository
	router LLMRouter
}

// NewChatService creates a new ChatService with the provided repository and
// LLM router.
func NewChatService(repo ChatRepository, router LLMRouter) *ChatService {
	return &ChatService{
		repo:   repo,
		router: router,
	}
}

// SendMessage saves the user message, routes it to the LLM, saves the
// assistant response, and returns the LLM response.
func (s *ChatService) SendMessage(ctx context.Context, userID int64, message, model string) (string, error) {
	// Save user message
	if _, err := s.repo.SaveMessage(userID, "user", message, model); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	// Route to LLM
	response, err := s.router.Route(ctx, message, model)
	if err != nil {
		return "", err
	}

	// Save assistant response
	if _, err := s.repo.SaveMessage(userID, "assistant", response, model); err != nil {
		return "", fmt.Errorf("failed to save assistant message: %w", err)
	}

	return response, nil
}

// GetHistory returns all chat messages for the given user.
func (s *ChatService) GetHistory(ctx context.Context, userID int64) ([]*repository.Message, error) {
	messages, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
