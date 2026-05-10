package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/oidcmock"
)

func TestOIDCMiddlewareWithMockIdP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mock, err := oidcmock.New(oidcmock.Options{Subject: "integration-user"})
	if err != nil {
		t.Fatalf("oidcmock new: %v", err)
	}
	ts := httptest.NewServer(mock.Handler())
	t.Cleanup(ts.Close)
	mock.SetIssuer(ts.URL)

	verifier, err := auth.NewVerifier(ctx, ts.URL, mock.Audience())
	if err != nil {
		t.Fatalf("new verifier: %v", err)
	}

	conn := openTempDB(t)

	mw := auth.Middleware(verifier, conn)
	called := false
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		u, ok := auth.FromContext(r.Context())
		if !ok {
			t.Errorf("expected user in context")
			return
		}
		if u.Sub != "integration-user" {
			t.Errorf("sub mismatch: got %q want %q", u.Sub, "integration-user")
		}
		if u.ID == 0 {
			t.Errorf("expected non-zero user id")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	tok, err := mock.SignAccessToken("integration-user", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !called {
		t.Fatal("inner handler never ran")
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/me", nil)
	req2.Header.Set("Authorization", "Bearer "+tok)
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("second call expected 204, got %d", rec2.Code)
	}

	var count int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE oidc_sub = ?`, "integration-user").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user row, got %d", count)
	}
}

func TestRejectsMissingAuth(t *testing.T) {
	conn := openTempDB(t)
	mw := auth.Middleware(stubVerifier{}, conn)
	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("inner should not run")
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

type stubVerifier struct{}

func (stubVerifier) Verify(context.Context, string) (*oidc.IDToken, error) {
	return nil, errors.New("should not be called")
}

func openTempDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}
