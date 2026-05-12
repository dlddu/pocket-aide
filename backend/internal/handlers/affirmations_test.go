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

func newAffirmationsRouter(t *testing.T, userID int64) (http.Handler, *affirmations.Store) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "h.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO users (oidc_sub) VALUES ('u1'), ('u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	store := affirmations.NewStore(conn)

	withAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.WithUser(r.Context(), auth.User{ID: userID, Sub: "u"})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	r := chi.NewRouter()
	r.With(withAuth).Get("/api/affirmations", handlers.ListAffirmations(store))
	r.With(withAuth).Post("/api/affirmations", handlers.CreateAffirmation(store))
	r.With(withAuth).Patch("/api/affirmations/{id}", handlers.UpdateAffirmation(store))
	r.With(withAuth).Delete("/api/affirmations/{id}", handlers.DeleteAffirmation(store))
	return r, store
}

func TestCreateListUpdateDelete(t *testing.T) {
	r, _ := newAffirmationsRouter(t, 1)

	// Create
	body := bytes.NewBufferString(`{"text":"one step","priority":"high"}`)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/affirmations", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status: %d body=%s", rec.Code, rec.Body.String())
	}
	var created affirmations.Affirmation
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID == 0 || created.Text != "one step" || created.Priority != affirmations.PriorityHigh {
		t.Fatalf("created mismatch: %+v", created)
	}

	// List
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/affirmations", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status: %d", rec.Code)
	}
	var list []affirmations.Affirmation
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}

	// Update
	body = bytes.NewBufferString(`{"text":"two steps","priority":"normal"}`)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/affirmations/"+strconv.FormatInt(created.ID, 10), body))
	if rec.Code != http.StatusOK {
		t.Fatalf("update status: %d body=%s", rec.Code, rec.Body.String())
	}
	var updated affirmations.Affirmation
	_ = json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Text != "two steps" || updated.Priority != affirmations.PriorityNormal {
		t.Fatalf("update did not apply: %+v", updated)
	}

	// Delete
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/affirmations/"+strconv.FormatInt(created.ID, 10), nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status: %d", rec.Code)
	}
}

func TestCreateRejectsInvalidPriority(t *testing.T) {
	r, _ := newAffirmationsRouter(t, 1)

	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"text":"x","priority":"urgent"}`)
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/affirmations", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateRejectsEmptyText(t *testing.T) {
	r, _ := newAffirmationsRouter(t, 1)

	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"text":"   ","priority":"low"}`)
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/affirmations", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateOtherUserReturns404(t *testing.T) {
	// Create as user 1
	r1, store := newAffirmationsRouter(t, 1)
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"text":"u1 row","priority":"low"}`)
	r1.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/affirmations", body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed create: %d body=%s", rec.Code, rec.Body.String())
	}
	var created affirmations.Affirmation
	_ = json.NewDecoder(rec.Body).Decode(&created)
	_ = store

	// User 2 tries to update user 1's row
	withAuthUser2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.WithUser(r.Context(), auth.User{ID: 2, Sub: "u2"})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	r2 := chi.NewRouter()
	r2.With(withAuthUser2).Patch("/api/affirmations/{id}", handlers.UpdateAffirmation(store))
	r2.With(withAuthUser2).Delete("/api/affirmations/{id}", handlers.DeleteAffirmation(store))

	rec = httptest.NewRecorder()
	body = bytes.NewBufferString(`{"text":"hijack","priority":"high"}`)
	r2.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/api/affirmations/"+strconv.FormatInt(created.ID, 10), body))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on cross-user PATCH, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r2.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/affirmations/"+strconv.FormatInt(created.ID, 10), nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on cross-user DELETE, got %d", rec.Code)
	}
}
