package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

func TestHealthOK(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "h.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	rec := httptest.NewRecorder()
	handlers.Health(conn).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
}

func TestAuthConfigEcho(t *testing.T) {
	cfg := handlers.AuthConfig{
		Issuer:      "https://issuer.example",
		ClientID:    "client-x",
		RedirectURI: "pocketaide://callback",
		Audience:    "aud-x",
	}
	rec := httptest.NewRecorder()
	handlers.AuthConfigHandler(cfg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/config", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}
	var got handlers.AuthConfig
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got != cfg {
		t.Errorf("body mismatch: got %+v want %+v", got, cfg)
	}
}

func TestMeReadsContext(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	ctx := auth.WithUser(req.Context(), auth.User{ID: 7, Sub: "abc"})
	handlers.Me().ServeHTTP(rec, req.WithContext(ctx))
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
	var got map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if got["sub"] != "abc" {
		t.Errorf("sub: %v", got["sub"])
	}
}
