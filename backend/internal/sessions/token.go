package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/k8sauth"
)

// PlatformAudience is the audience session-platform accepts tenant tokens for.
const PlatformAudience = "session-platform"

const (
	tokenLifetime = time.Hour
	// tokenRenewBefore renews well before expiry so a long chat turn never
	// starts on a token about to lapse.
	tokenRenewBefore = 15 * time.Minute
)

// TokenSource supplies the bearer token a Client sends.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

// ServiceAccountTokens issues tokens for another ServiceAccount through the
// TokenRequest API, so session-platform sees that account's tenant
// (flux-cd-apps apps/pocket-aide/session-platform-tenant-accounts.yaml).
type ServiceAccountTokens struct {
	api       k8sauth.APIServer
	namespace string
	name      string
	audience  string
	now       func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func NewServiceAccountTokens(api k8sauth.APIServer, namespace, name, audience string) *ServiceAccountTokens {
	return &ServiceAccountTokens{api: api, namespace: namespace, name: name, audience: audience, now: time.Now}
}

func (s *ServiceAccountTokens) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token != "" && s.now().Before(s.expires.Add(-tokenRenewBefore)) {
		return s.token, nil
	}
	own, err := s.api.Token()
	if err != nil {
		return "", fmt.Errorf("read own service account token: %w", err)
	}
	body, _ := json.Marshal(map[string]any{
		"apiVersion": "authentication.k8s.io/v1",
		"kind":       "TokenRequest",
		"spec": map[string]any{
			"audiences":         []string{s.audience},
			"expirationSeconds": int64(tokenLifetime / time.Second),
		},
	})
	target := s.api.URL + "/api/v1/namespaces/" + url.PathEscape(s.namespace) +
		"/serviceaccounts/" + url.PathEscape(s.name) + "/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+own)
	resp, err := s.api.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request for %s/%s: status %d: %s", s.namespace, s.name, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var out struct {
		Status struct {
			Token               string    `json:"token"`
			ExpirationTimestamp time.Time `json:"expirationTimestamp"`
		} `json:"status"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.Status.Token == "" {
		return "", fmt.Errorf("token request for %s/%s: no token in response", s.namespace, s.name)
	}
	s.token = out.Status.Token
	s.expires = out.Status.ExpirationTimestamp
	if s.expires.IsZero() {
		s.expires = s.now().Add(tokenLifetime)
	}
	return s.token, nil
}
