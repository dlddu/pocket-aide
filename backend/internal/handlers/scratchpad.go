package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/affirmations"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/routines"
	"github.com/dlddu/pocket-aide/backend/internal/scratchpad"
	"github.com/dlddu/pocket-aide/backend/internal/todos"
)

type scratchpadPayload struct {
	Text       string             `json:"text"`
	Source     *scratchpad.Source `json:"source"`
	CapturedAt *int64             `json:"captured_at"`
}

type movePayload struct {
	Target scratchpad.Target `json:"target"`
}

// moveResponse carries the row created in the destination area so the app can
// show it without a second round trip — the moved affirmation, for instance,
// goes straight into the priority sheet.
type moveResponse struct {
	Target      scratchpad.Target         `json:"target"`
	Todo        *todos.Todo               `json:"todo,omitempty"`
	Affirmation *affirmations.Affirmation `json:"affirmation,omitempty"`
	Routine     *routines.Routine         `json:"routine,omitempty"`
}

// ListScratchpad handles GET /api/scratchpad.
func ListScratchpad(store *scratchpad.Store) http.HandlerFunc {
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

// CreateScratchpadItem handles POST /api/scratchpad.
func CreateScratchpadItem(store *scratchpad.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var p scratchpadPayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		text := strings.TrimSpace(p.Text)
		if text == "" {
			http.Error(w, "text is required", http.StatusBadRequest)
			return
		}
		source := scratchpad.SourceText
		if p.Source != nil {
			source = *p.Source
		}
		if !source.Valid() {
			http.Error(w, "source must be one of text|voice|shortcut", http.StatusBadRequest)
			return
		}
		created, err := store.Create(r.Context(), u.ID, text, source, p.CapturedAt)
		if err != nil {
			http.Error(w, "create failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

// DeleteScratchpadItem handles DELETE /api/scratchpad/{id}.
func DeleteScratchpadItem(store *scratchpad.Store) http.HandlerFunc {
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
		if errors.Is(err, scratchpad.ErrNotFound) {
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

// MoveScratchpadItem handles POST /api/scratchpad/{id}/move.
func MoveScratchpadItem(store *scratchpad.Store, todoStore *todos.Store, affStore *affirmations.Store, routineStore *routines.Store) http.HandlerFunc {
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
		var p movePayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		if !p.Target.Valid() {
			http.Error(w, "target must be one of personal|work|affirmation|routine", http.StatusBadRequest)
			return
		}
		newID, err := store.Move(r.Context(), u.ID, id, p.Target)
		if errors.Is(err, scratchpad.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "move failed", http.StatusInternalServerError)
			return
		}
		resp := moveResponse{Target: p.Target}
		switch p.Target {
		case scratchpad.TargetAffirmation:
			a, err := affStore.Get(r.Context(), u.ID, newID)
			if err != nil {
				http.Error(w, "read moved item failed", http.StatusInternalServerError)
				return
			}
			resp.Affirmation = &a
		case scratchpad.TargetRoutine:
			rt, err := routineStore.Get(r.Context(), u.ID, newID)
			if err != nil {
				http.Error(w, "read moved item failed", http.StatusInternalServerError)
				return
			}
			resp.Routine = &rt
		default:
			t, err := todoStore.Get(r.Context(), todos.Area(p.Target), u.ID, newID)
			if err != nil {
				http.Error(w, "read moved item failed", http.StatusInternalServerError)
				return
			}
			resp.Todo = &t
		}
		writeJSON(w, http.StatusOK, resp)
	}
}
