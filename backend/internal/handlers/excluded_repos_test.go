package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/excludedrepos"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

func newExcludedStore(t *testing.T) *excludedrepos.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "ex.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return excludedrepos.New(conn)
}

func chiRouterForExcludedRepos(store *excludedrepos.Store) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/excluded-repos", handlers.ListExcludedRepos(store))
	r.Post("/api/excluded-repos", handlers.AddExcludedRepo(store))
	r.Delete("/api/excluded-repos/{id}", handlers.DeleteExcludedRepo(store))
	return r
}

func TestExcludedRepos_AddListDelete(t *testing.T) {
	store := newExcludedStore(t)
	router := chiRouterForExcludedRepos(store)

	body, _ := json.Marshal(map[string]any{"repo_full_name": "dlddu/pocket-aide"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/excluded-repos", body, 1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("add: got %d want 201 (body=%s)", rec.Code, rec.Body.String())
	}
	var created excludedrepos.ExcludedRepo
	_ = json.NewDecoder(rec.Body).Decode(&created)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodGet, "/api/excluded-repos", nil, 1))
	if rec.Code != http.StatusOK {
		t.Fatalf("list: got %d want 200", rec.Code)
	}
	var listed struct {
		Items []excludedrepos.ExcludedRepo `json:"items"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&listed)
	if len(listed.Items) != 1 || listed.Items[0].ID != created.ID {
		t.Errorf("list: got %+v", listed.Items)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodDelete,
		"/api/excluded-repos/"+strconv.FormatInt(created.ID, 10),
		nil, 1,
	))
	if rec.Code != http.StatusNoContent {
		t.Errorf("delete: got %d want 204", rec.Code)
	}
}

func TestExcludedRepos_AddRejectsBadShape(t *testing.T) {
	store := newExcludedStore(t)
	router := chiRouterForExcludedRepos(store)

	body, _ := json.Marshal(map[string]any{"repo_full_name": "not-a-repo"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/excluded-repos", body, 1))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestExcludedRepos_AddDuplicateReturns409(t *testing.T) {
	store := newExcludedStore(t)
	router := chiRouterForExcludedRepos(store)

	body, _ := json.Marshal(map[string]any{"repo_full_name": "dlddu/x"})

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/excluded-repos", body, 1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed add: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/excluded-repos", body, 1))
	if rec.Code != http.StatusConflict {
		t.Errorf("dup add: got %d want 409", rec.Code)
	}
}

func TestExcludedRepos_DeleteOtherUserReturns404(t *testing.T) {
	store := newExcludedStore(t)
	router := chiRouterForExcludedRepos(store)

	body, _ := json.Marshal(map[string]any{"repo_full_name": "dlddu/y"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(http.MethodPost, "/api/excluded-repos", body, 1))
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed: %d", rec.Code)
	}
	var created excludedrepos.ExcludedRepo
	_ = json.NewDecoder(rec.Body).Decode(&created)

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(
		http.MethodDelete,
		"/api/excluded-repos/"+strconv.FormatInt(created.ID, 10),
		nil, 2, // different user
	))
	if rec.Code != http.StatusNotFound {
		t.Errorf("cross-user delete: got %d want 404", rec.Code)
	}
}
