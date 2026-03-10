// Package llm defines the LLM provider interface and a mock implementation for testing.
package llm

import "context"

// Provider defines the interface for interacting with an LLM service.
type Provider interface {
	// Complete sends a prompt to the LLM and returns the generated completion.
	Complete(ctx context.Context, prompt string) (string, error)
}

// MockProvider is a test double for the Provider interface.
// It records calls and delegates to a user-supplied function.
type MockProvider struct {
	// CompleteFunc is called when Complete is invoked. Must be set before use.
	CompleteFunc func(ctx context.Context, prompt string) (string, error)

	// CallCount tracks how many times Complete has been called.
	CallCount int

	// LastPrompt stores the most recent prompt passed to Complete.
	LastPrompt string
}

// Complete implements Provider by delegating to CompleteFunc and tracking call metadata.
func (m *MockProvider) Complete(ctx context.Context, prompt string) (string, error) {
	m.CallCount++
	m.LastPrompt = prompt
	return m.CompleteFunc(ctx, prompt)
}
