// Package llm_test contains unit tests for the LLM Router.
// These tests verify that the Router correctly selects and delegates to
// the appropriate Provider based on the requested model name.
//
// TDD Red Phase: the Router type and its methods do not yet exist.
// All tests are expected to fail until service/llm/router.go is implemented.
package llm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dlddu/pocket-aide/service/llm"
)

// ---------------------------------------------------------------------------
// Router.Route
// ---------------------------------------------------------------------------

// TestRouter_Route_DefaultModel verifies that when no model is specified
// (empty string), the Router uses the registered default model's Provider.
//
// Scenario:
//
//	Router with default model "default-model" registered.
//	Route(ctx, prompt="Hello", model="")
//	→ defaultProvider.CompleteFunc is called, response returned
func TestRouter_Route_DefaultModel(t *testing.T) {
	// Arrange
	defaultProvider := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "default response", nil
		},
	}
	router := llm.NewRouter("default-model")
	router.RegisterProvider("default-model", defaultProvider)

	// Act
	result, err := router.Route(context.Background(), "Hello", "")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "default response" {
		t.Errorf("expected 'default response', got '%s'", result)
	}
	if defaultProvider.CallCount != 1 {
		t.Errorf("expected default provider to be called once, got %d", defaultProvider.CallCount)
	}
}

// TestRouter_Route_SpecifiedModel verifies that when a model name is
// explicitly specified, the Router delegates to that model's Provider.
//
// Scenario:
//
//	Router with "gpt-4o" Provider registered.
//	Route(ctx, prompt="Hello", model="gpt-4o")
//	→ gpt4oProvider.CompleteFunc is called, response returned
func TestRouter_Route_SpecifiedModel(t *testing.T) {
	// Arrange
	defaultProvider := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "default response", nil
		},
	}
	gpt4oProvider := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "gpt-4o response", nil
		},
	}
	router := llm.NewRouter("default-model")
	router.RegisterProvider("default-model", defaultProvider)
	router.RegisterProvider("gpt-4o", gpt4oProvider)

	// Act
	result, err := router.Route(context.Background(), "Hello", "gpt-4o")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "gpt-4o response" {
		t.Errorf("expected 'gpt-4o response', got '%s'", result)
	}
	if gpt4oProvider.CallCount != 1 {
		t.Errorf("expected gpt-4o provider to be called once, got %d", gpt4oProvider.CallCount)
	}
	if defaultProvider.CallCount != 0 {
		t.Errorf("expected default provider not to be called, got %d calls", defaultProvider.CallCount)
	}
}

// TestRouter_Route_UnknownModel_UsesFallback verifies that when an unregistered
// model name is requested, the Router falls back to the default model's Provider
// instead of returning an error.
//
// Scenario:
//
//	Router with default model "default-model" registered (no "unknown-model").
//	Route(ctx, prompt="Hello", model="unknown-model")
//	→ defaultProvider.CompleteFunc is called (fallback behavior)
func TestRouter_Route_UnknownModel_UsesFallback(t *testing.T) {
	// Arrange
	defaultProvider := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "fallback response", nil
		},
	}
	router := llm.NewRouter("default-model")
	router.RegisterProvider("default-model", defaultProvider)

	// Act
	result, err := router.Route(context.Background(), "Hello", "unknown-model")

	// Assert
	if err != nil {
		t.Fatalf("expected no error for unknown model (should fallback), got: %v", err)
	}
	if result != "fallback response" {
		t.Errorf("expected 'fallback response' from fallback, got '%s'", result)
	}
	if defaultProvider.CallCount != 1 {
		t.Errorf("expected default provider to be called once as fallback, got %d", defaultProvider.CallCount)
	}
}

// TestRouter_RegisterProvider verifies that a Provider registered under a
// given model name is discoverable and callable by the Router.
//
// Scenario:
//
//	Register two providers: "model-a" and "model-b".
//	Route with "model-a" → model-a Provider is called.
//	Route with "model-b" → model-b Provider is called.
func TestRouter_RegisterProvider(t *testing.T) {
	// Arrange
	providerA := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "response from A", nil
		},
	}
	providerB := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "response from B", nil
		},
	}
	router := llm.NewRouter("model-a")
	router.RegisterProvider("model-a", providerA)
	router.RegisterProvider("model-b", providerB)

	// Act: route to model-a
	resultA, errA := router.Route(context.Background(), "prompt", "model-a")

	// Assert model-a
	if errA != nil {
		t.Fatalf("expected no error for model-a, got: %v", errA)
	}
	if resultA != "response from A" {
		t.Errorf("expected 'response from A', got '%s'", resultA)
	}
	if providerA.CallCount != 1 {
		t.Errorf("expected providerA call count 1, got %d", providerA.CallCount)
	}

	// Act: route to model-b
	resultB, errB := router.Route(context.Background(), "prompt", "model-b")

	// Assert model-b
	if errB != nil {
		t.Fatalf("expected no error for model-b, got: %v", errB)
	}
	if resultB != "response from B" {
		t.Errorf("expected 'response from B', got '%s'", resultB)
	}
	if providerB.CallCount != 1 {
		t.Errorf("expected providerB call count 1, got %d", providerB.CallCount)
	}

	// Verify providerA was not called again
	if providerA.CallCount != 1 {
		t.Errorf("expected providerA to be called exactly once total, got %d", providerA.CallCount)
	}
}

// TestRouter_Route_PropagatesProviderError verifies that when the selected
// Provider returns an error, the Router propagates it to the caller.
//
// Scenario:
//
//	Provider is configured to return an error.
//	Route(ctx, prompt="Hello", model="")
//	→ non-nil error matching the provider's error
func TestRouter_Route_PropagatesProviderError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("provider unavailable")
	failingProvider := &llm.MockProvider{
		CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
			return "", expectedErr
		},
	}
	router := llm.NewRouter("default-model")
	router.RegisterProvider("default-model", failingProvider)

	// Act
	result, err := router.Route(context.Background(), "Hello", "")

	// Assert
	if err == nil {
		t.Error("expected error from provider, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
	if result != "" {
		t.Errorf("expected empty result on error, got '%s'", result)
	}
}
