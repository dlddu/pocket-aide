package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/affirmations"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

func newStore(t *testing.T) *affirmations.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "aff.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	// Seed two distinct users so the user-scoping checks have something to
	// confuse handlers with if they forget to filter on user_id.
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	return affirmations.New(conn)
}

func authedRequest(method, target string, body []byte, userID int64) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, target, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	ctx := auth.WithUser(r.Context(), auth.User{ID: userID, Sub: "test-sub"})
	return r.WithContext(ctx)
}

func chiRouterForAffirmations(store *affirmations.Store) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/affirmations", handlers.ListAffirmations(store))
	r.Post("/api/affirmations", handlers.CreateAffirmation(store))
	r.Patch("/api/affirmations/{id}", handlers.UpdateAffirmation(store))
	r.Delete("/api/affirmations/{id}", handlers.DeleteAffirmation(store))

	// Wrap so the auth.User context survives chi's routing — chi normally
	// preserves it, but we wrap once defensively for our own context plumbing.
	return r
}

func TestCreateThenList(t *testing.T) {
	store := newStore(t)
	router := chiRouterForAffirmations(store)

	body, _ := json.Marshal(map[string]any{"text": "작게 시작해서 매일 1%씩", "priority": "high"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/affirmations", body, 1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status: got %d want 201 (body=%s)", rec.Code, rec.Body.String())
	}
	var created affirmations.Affirmation
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID == 0 || created.Text == "" || created.Priority != affirmations.PriorityHigh {
		t.Errorf("unexpected created: %+v", created)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodGet, "/api/affirmations", nil, 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status: got %d want 200", rec.Code)
	}
	var listed struct {
		Items []affirmations.Affirmation `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Items) != 1 || listed.Items[0].ID != created.ID {
		t.Errorf("unexpected list: %+v", listed.Items)
	}
}

func TestCreateRejectsInvalidPriority(t *testing.T) {
	store := newStore(t)
	router := chiRouterForAffirmations(store)

	body, _ := json.Marshal(map[string]any{"text": "x", "priority": "URGENT"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/affirmations", body, 1))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestCreateRejectsEmptyText(t *testing.T) {
	store := newStore(t)
	router := chiRouterForAffirmations(store)

	body, _ := json.Marshal(map[string]any{"text": "   ", "priority": "normal"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/affirmations", body, 1))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestUpdateOtherUserReturns404(t *testing.T) {
	store := newStore(t)
	router := chiRouterForAffirmations(store)

	body, _ := json.Marshal(map[string]any{"text": "u1 sentence", "priority": "high"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/affirmations", body, 1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed create: %d", rec.Code)
	}
	var created affirmations.Affirmation
	_ = json.NewDecoder(rec.Body).Decode(&created)

	updateBody, _ := json.Marshal(map[string]any{"text": "tampered", "priority": "low"})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodPatch,
		"/api/affirmations/"+strconv.FormatInt(created.ID, 10),
		updateBody,
		2, // different user
	))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user update should 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestDeleteOtherUserReturns404(t *testing.T) {
	store := newStore(t)
	router := chiRouterForAffirmations(store)

	body, _ := json.Marshal(map[string]any{"text": "u1 sentence", "priority": "normal"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/affirmations", body, 1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed create: %d", rec.Code)
	}
	var created affirmations.Affirmation
	_ = json.NewDecoder(rec.Body).Decode(&created)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodDelete,
		"/api/affirmations/"+strconv.FormatInt(created.ID, 10),
		nil,
		2,
	))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user delete should 404, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestUpdateHappyPath(t *testing.T) {
	store := newStore(t)
	router := chiRouterForAffirmations(store)

	body, _ := json.Marshal(map[string]any{"text": "before", "priority": "high"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/affirmations", body, 1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed create: %d", rec.Code)
	}
	var created affirmations.Affirmation
	_ = json.NewDecoder(rec.Body).Decode(&created)

	patch, _ := json.Marshal(map[string]any{"text": "after", "priority": "low"})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodPatch,
		"/api/affirmations/"+strconv.FormatInt(created.ID, 10),
		patch,
		1,
	))
	if rec.Code != http.StatusOK {
		t.Fatalf("update status: %d (body=%s)", rec.Code, rec.Body.String())
	}
	var updated affirmations.Affirmation
	_ = json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Text != "after" || updated.Priority != affirmations.PriorityLow {
		t.Errorf("unexpected update result: %+v", updated)
	}
}
