package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/notificationsettings"
)

func chiRouterForNotificationSettings(t *testing.T) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "ns.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	store := notificationsettings.New(conn)
	r := chi.NewRouter()
	r.Get("/api/notification-settings", handlers.GetNotificationSettings(store))
	r.Patch("/api/notification-settings", handlers.UpdateNotificationSettings(store))
	return r
}

func decodeSettings(t *testing.T, rec *httptest.ResponseRecorder) notificationsettings.Settings {
	t.Helper()
	var s notificationsettings.Settings
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return s
}

func TestNotificationSettings_GetDefaultThenPatch(t *testing.T) {
	router := chiRouterForNotificationSettings(t)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodGet, "/api/notification-settings", nil, 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("get: got %d want 200", rec.Code)
	}
	if got := decodeSettings(t, rec); got != notificationsettings.Default() {
		t.Errorf("default: got %+v", got)
	}

	body, _ := json.Marshal(map[string]any{"outcomes": "success"})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPatch, "/api/notification-settings", body, 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("patch: got %d want 200 (body=%s)", rec.Code, rec.Body.String())
	}
	want := notificationsettings.Settings{Enabled: true, Outcomes: notificationsettings.OutcomesSuccess}
	if got := decodeSettings(t, rec); got != want {
		t.Errorf("patch response: got %+v want %+v", got, want)
	}

	body, _ = json.Marshal(map[string]any{"enabled": false})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPatch, "/api/notification-settings", body, 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("patch enabled: got %d want 200", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodGet, "/api/notification-settings", nil, 1))
	want = notificationsettings.Settings{Enabled: false, Outcomes: notificationsettings.OutcomesSuccess}
	if got := decodeSettings(t, rec); got != want {
		t.Errorf("reload: got %+v want %+v", got, want)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodGet, "/api/notification-settings", nil, 2))
	if got := decodeSettings(t, rec); got != notificationsettings.Default() {
		t.Errorf("other user must keep defaults, got %+v", got)
	}
}

func TestNotificationSettings_PatchRejectsBadInput(t *testing.T) {
	router := chiRouterForNotificationSettings(t)

	for _, raw := range []string{
		`{"outcomes":"cancelled"}`,
		`{"enabled":"yes"}`,
		`{"unknown":true}`,
		`not json`,
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, authedRequest(http.MethodPatch, "/api/notification-settings", []byte(raw), 1))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("patch %s: got %d want 400", raw, rec.Code)
		}
	}
}
