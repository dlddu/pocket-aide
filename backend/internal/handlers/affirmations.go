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

const maxAffirmationTextBytes = 4096

type affirmationPayload struct {
	Text     string `json:"text"`
	Priority string `json:"priority"`
}

// ListAffirmations returns all of the authenticated user's affirmations.
func ListAffirmations(store *affirmations.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		items, err := store.List(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if items == nil {
			items = []affirmations.Affirmation{}
		}
		writeJSON(w, http.StatusOK, items)
	}
}

// CreateAffirmation persists a new affirmation under the authenticated user.
func CreateAffirmation(store *affirmations.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		text, priority, err := decodeAffirmationPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a, err := store.Create(r.Context(), user.ID, text, priority)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, a)
	}
}

// UpdateAffirmation applies text + priority changes. 404 if the row does not
// belong to the authenticated user.
func UpdateAffirmation(store *affirmations.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := pathInt64(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		text, priority, err := decodeAffirmationPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		a, err := store.Update(r.Context(), user.ID, id, text, priority)
		if errors.Is(err, affirmations.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, a)
	}
}

// DeleteAffirmation removes the row. 404 if not owned by the user.
func DeleteAffirmation(store *affirmations.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := pathInt64(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = store.Delete(r.Context(), user.ID, id)
		if errors.Is(err, affirmations.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func decodeAffirmationPayload(r *http.Request) (string, affirmations.Priority, error) {
	var p affirmationPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return "", "", errors.New("invalid JSON body")
	}
	text := strings.TrimSpace(p.Text)
	if text == "" {
		return "", "", errors.New("text is required")
	}
	if len(text) > maxAffirmationTextBytes {
		return "", "", errors.New("text too long")
	}
	priority := affirmations.Priority(p.Priority)
	if !priority.Valid() {
		return "", "", errors.New("priority must be high, normal, or low")
	}
	return text, priority, nil
}

func pathInt64(r *http.Request, key string) (int64, error) {
	raw := chi.URLParam(r, key)
	if raw == "" {
		return 0, errors.New("missing path parameter: " + key)
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	return v, nil
}
