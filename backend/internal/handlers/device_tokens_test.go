package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

func newDeviceStore(t *testing.T) *devicetokens.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "dt.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return devicetokens.New(conn)
}

func TestRegisterDeviceToken_Created(t *testing.T) {
	store := newDeviceStore(t)
	hex64 := strings.Repeat("a", 64)
	body := []byte(`{"token":"` + hex64 + `"}`)

	rec := httptest.NewRecorder()
	handlers.RegisterDeviceToken(store).ServeHTTP(
		rec,
		authedRequest(http.MethodPost, "/api/device-tokens", body, 1),
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d want 201 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestRegisterDeviceToken_RejectsShortToken(t *testing.T) {
	store := newDeviceStore(t)
	rec := httptest.NewRecorder()
	handlers.RegisterDeviceToken(store).ServeHTTP(
		rec,
		authedRequest(http.MethodPost, "/api/device-tokens", []byte(`{"token":"abcd"}`), 1),
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestRegisterDeviceToken_RejectsNonHex(t *testing.T) {
	store := newDeviceStore(t)
	rec := httptest.NewRecorder()
	body := []byte(`{"token":"` + strings.Repeat("z", 64) + `"}`)
	handlers.RegisterDeviceToken(store).ServeHTTP(
		rec,
		authedRequest(http.MethodPost, "/api/device-tokens", body, 1),
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestRegisterDeviceToken_UpsertRebindsToNewUser(t *testing.T) {
	store := newDeviceStore(t)
	hex64 := strings.Repeat("b", 64)
	body := []byte(`{"token":"` + hex64 + `"}`)

	for _, uid := range []int64{1, 2} {
		rec := httptest.NewRecorder()
		handlers.RegisterDeviceToken(store).ServeHTTP(
			rec,
			authedRequest(http.MethodPost, "/api/device-tokens", body, uid),
		)
		if rec.Code != http.StatusCreated {
			t.Fatalf("uid=%d status: got %d want 201", uid, rec.Code)
		}
	}

	tokens, err := store.ListAll(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tokens) != 1 {
		t.Errorf("expected 1 token (upsert rebind), got %d", len(tokens))
	}
}
