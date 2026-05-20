package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/excludedrepos"
)

type excludedRepoPayload struct {
	RepoFullName string `json:"repo_full_name"`
}

// ListExcludedRepos handles GET /api/excluded-repos.
func ListExcludedRepos(store *excludedrepos.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		items, err := store.List(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "list failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

// AddExcludedRepo handles POST /api/excluded-repos.
func AddExcludedRepo(store *excludedrepos.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		body, err := decodeExcludedRepoPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := store.Add(r.Context(), u.ID, body.RepoFullName)
		if errors.Is(err, excludedrepos.ErrInvalidRepo) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, excludedrepos.ErrAlreadyExcluded) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, "add failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

// DeleteExcludedRepo handles DELETE /api/excluded-repos/{id}.
func DeleteExcludedRepo(store *excludedrepos.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := pathID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = store.Delete(r.Context(), u.ID, id)
		if errors.Is(err, excludedrepos.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "delete failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func decodeExcludedRepoPayload(r *http.Request) (excludedRepoPayload, error) {
	var p excludedRepoPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return p, errors.New("invalid json body")
	}
	p.RepoFullName = strings.TrimSpace(p.RepoFullName)
	if p.RepoFullName == "" {
		return p, errors.New("repo_full_name is required")
	}
	return p, nil
}
