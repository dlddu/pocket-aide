package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/affirmations"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
)

type affirmationPayload struct {
	Text     string                `json:"text"`
	Priority affirmations.Priority `json:"priority"`
}

// ListAffirmations handles GET /api/affirmations.
func ListAffirmations(store *affirmations.Store) http.HandlerFunc {
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

// CreateAffirmation handles POST /api/affirmations.
func CreateAffirmation(store *affirmations.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		body, err := decodePayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := store.Create(r.Context(), u.ID, body.Text, body.Priority)
		if err != nil {
			http.Error(w, "create failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

// UpdateAffirmation handles PATCH /api/affirmations/{id}.
func UpdateAffirmation(store *affirmations.Store) http.HandlerFunc {
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
		body, err := decodePayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updated, err := store.Update(r.Context(), u.ID, id, body.Text, body.Priority)
		if errors.Is(err, affirmations.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "update failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

// DeleteAffirmation handles DELETE /api/affirmations/{id}.
func DeleteAffirmation(store *affirmations.Store) http.HandlerFunc {
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
		if errors.Is(err, affirmations.ErrNotFound) {
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

func decodePayload(r *http.Request) (affirmationPayload, error) {
	var p affirmationPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return p, errors.New("invalid json body")
	}
	p.Text = strings.TrimSpace(p.Text)
	if p.Text == "" {
		return p, errors.New("text is required")
	}
	if !p.Priority.Valid() {
		return p, errors.New("priority must be one of high|normal|low")
	}
	return p, nil
}

func pathID(r *http.Request) (int64, error) {
	raw := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
