package sessions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/k8sauth"
)

// The chat tab calls session-platform as its own tenant: a TokenRequest
// token for its ServiceAccount, audience session-platform, reused until it
// is close to expiring.
func TestServiceAccountTokensAndBearer(t *testing.T) {
	var issued atomic.Int32
	now := time.Date(2026, 11, 6, 0, 0, 0, 0, time.UTC)
	apiserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/namespaces/pocket-aide/serviceaccounts/pocket-aide-chat/token" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer own-token" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		var req struct {
			Spec struct {
				Audiences         []string `json:"audiences"`
				ExpirationSeconds int64    `json:"expirationSeconds"`
			} `json:"spec"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Spec.Audiences) != 1 || req.Spec.Audiences[0] != PlatformAudience || req.Spec.ExpirationSeconds != 3600 {
			t.Errorf("spec = %+v", req.Spec)
		}
		n := issued.Add(1)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": map[string]any{
			"token":               "chat-token-" + string(rune('0'+n)),
			"expirationTimestamp": now.Add(time.Hour).Format(time.RFC3339),
		}})
	}))
	defer apiserver.Close()

	tokens := NewServiceAccountTokens(k8sauth.APIServer{
		URL:        apiserver.URL,
		Token:      func() (string, error) { return "own-token", nil },
		HTTPClient: apiserver.Client(),
	}, "pocket-aide", "pocket-aide-chat", PlatformAudience)
	tokens.now = func() time.Time { return now }

	var seen []string
	platform := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{}`))
	}))
	defer platform.Close()
	client := NewClient(platform.URL).WithTokens(tokens)

	for i := 0; i < 2; i++ {
		if _, _, err := client.Do(context.Background(), http.MethodGet, "/sessions", nil); err != nil {
			t.Fatal(err)
		}
	}
	now = now.Add(50 * time.Minute)
	if _, _, err := client.Do(context.Background(), http.MethodGet, "/sessions", nil); err != nil {
		t.Fatal(err)
	}
	if got := issued.Load(); got != 2 {
		t.Fatalf("tokens issued = %d, want 2 (reuse, then renew near expiry)", got)
	}
	if seen[0] != "Bearer chat-token-1" || seen[1] != "Bearer chat-token-1" || seen[2] != "Bearer chat-token-2" {
		t.Fatalf("platform saw %v", seen)
	}

	bare := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, has := r.Header["Authorization"]; has {
			t.Error("a client without tokens must send no Authorization header")
		}
	}))
	defer bare.Close()
	if _, _, err := NewClient(bare.URL).Do(context.Background(), http.MethodGet, "/sessions", nil); err != nil {
		t.Fatal(err)
	}
}
