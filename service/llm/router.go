// Package llm defines the LLM provider interface and a mock implementation for testing.
package llm

import (
	"context"
	"fmt"
)

// Router selects and delegates to the appropriate Provider based on the
// requested model name.
type Router struct {
	providers    map[string]Provider
	defaultModel string
}

// NewRouter creates a new Router with the given default model name.
func NewRouter(defaultModel string) *Router {
	return &Router{
		providers:    make(map[string]Provider),
		defaultModel: defaultModel,
	}
}

// RegisterProvider registers a Provider under the given model name.
func (r *Router) RegisterProvider(model string, p Provider) {
	r.providers[model] = p
}

// Route delegates the prompt to the Provider registered for the given model.
// If model is empty or not registered, it falls back to the default model's Provider.
// It returns the provider's response or propagates any error.
func (r *Router) Route(ctx context.Context, prompt, model string) (string, error) {
	selected := model
	if selected == "" {
		selected = r.defaultModel
	}

	provider, ok := r.providers[selected]
	if !ok {
		// Fall back to the default model's provider
		provider, ok = r.providers[r.defaultModel]
		if !ok {
			return "", fmt.Errorf("no provider registered for model %q and no default provider available", selected)
		}
	}

	return provider.Complete(ctx, prompt)
}
