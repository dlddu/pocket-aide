package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dlddu/pocket-aide/service/llm"
)

func TestLLMProviderInterface_MockImplementsInterface(t *testing.T) {
	// Arrange & Act: verify Mock satisfies the LLMProvider interface at compile time
	var _ llm.Provider = (*llm.MockProvider)(nil)

	// Assert: if this compiles, the interface is satisfied
	t.Log("MockProvider correctly implements LLMProvider interface")
}

func TestMockProvider_CompleteReturnsConfiguredResponse(t *testing.T) {
	// Arrange
	expectedResponse := "Hello, I am a mock LLM response."
	mock := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return expectedResponse, nil
		},
	}

	// Act
	result, err := mock.Complete(context.Background(), "Say hello")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != expectedResponse {
		t.Errorf("expected '%s', got '%s'", expectedResponse, result)
	}
}

func TestMockProvider_CompleteReturnsConfiguredError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("LLM service unavailable")
	mock := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "", expectedErr
		},
	}

	// Act
	_, err := mock.Complete(context.Background(), "Say hello")

	// Assert
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestMockProvider_CompleteReceivesCorrectPrompt(t *testing.T) {
	// Arrange
	var capturedPrompt string
	mock := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			capturedPrompt = prompt
			return "response", nil
		},
	}
	expectedPrompt := "Translate this to Korean: Hello"

	// Act
	_, err := mock.Complete(context.Background(), expectedPrompt)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if capturedPrompt != expectedPrompt {
		t.Errorf("expected prompt '%s', got '%s'", expectedPrompt, capturedPrompt)
	}
}

func TestMockProvider_CompleteRespectsContextCancellation(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mock := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			if ctx.Err() != nil {
				return "", ctx.Err()
			}
			return "response", nil
		},
	}

	// Act
	_, err := mock.Complete(ctx, "some prompt")

	// Assert
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestMockProvider_CallCountTracking(t *testing.T) {
	// Arrange
	mock := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "response", nil
		},
	}

	// Act
	_, _ = mock.Complete(context.Background(), "prompt 1")
	_, _ = mock.Complete(context.Background(), "prompt 2")
	_, _ = mock.Complete(context.Background(), "prompt 3")

	// Assert
	if mock.CallCount != 3 {
		t.Errorf("expected call count of 3, got %d", mock.CallCount)
	}
}

func TestMockProvider_LastPromptTracking(t *testing.T) {
	// Arrange
	mock := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "response", nil
		},
	}

	// Act
	_, _ = mock.Complete(context.Background(), "first prompt")
	_, _ = mock.Complete(context.Background(), "last prompt")

	// Assert
	if mock.LastPrompt != "last prompt" {
		t.Errorf("expected last prompt 'last prompt', got '%s'", mock.LastPrompt)
	}
}
