package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/todos"
)

// todoPayload has no area field: the area comes only from the URL, so a
// request body cannot move an item into the other collection (PRD-3 AC4).
type todoPayload struct {
	Title    string          `json:"title"`
	Memo     string          `json:"memo"`
	DueDate  *string         `json:"due_date"`
	Priority *todos.Priority `json:"priority"`
	Done     bool            `json:"done"`
}

// ListTodos handles GET /api/todos/{area}.
func ListTodos(store *todos.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		area, ok := todoArea(w, r)
		if !ok {
			return
		}
		items, err := store.List(r.Context(), area, u.ID)
		if err != nil {
			http.Error(w, "list failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

// CreateTodo handles POST /api/todos/{area}.
func CreateTodo(store *todos.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		area, ok := todoArea(w, r)
		if !ok {
			return
		}
		fields, err := decodeTodoPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := store.Create(r.Context(), area, u.ID, fields)
		if err != nil {
			http.Error(w, "create failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

// UpdateTodo handles PATCH /api/todos/{area}/{id}.
func UpdateTodo(store *todos.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		area, ok := todoArea(w, r)
		if !ok {
			return
		}
		id, err := pathID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fields, err := decodeTodoPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		updated, err := store.Update(r.Context(), area, u.ID, id, fields)
		if errors.Is(err, todos.ErrNotFound) {
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

// DeleteTodo handles DELETE /api/todos/{area}/{id}.
func DeleteTodo(store *todos.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		area, ok := todoArea(w, r)
		if !ok {
			return
		}
		id, err := pathID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = store.Delete(r.Context(), area, u.ID, id)
		if errors.Is(err, todos.ErrNotFound) {
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

func todoArea(w http.ResponseWriter, r *http.Request) (todos.Area, bool) {
	area := todos.Area(chi.URLParam(r, "area"))
	if !area.Valid() {
		http.Error(w, "not found", http.StatusNotFound)
		return "", false
	}
	return area, true
}

func decodeTodoPayload(r *http.Request) (todos.Fields, error) {
	var p todoPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return todos.Fields{}, errors.New("invalid json body")
	}
	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		return todos.Fields{}, errors.New("title is required")
	}
	if p.DueDate != nil {
		if _, err := time.Parse(time.DateOnly, *p.DueDate); err != nil {
			return todos.Fields{}, errors.New("due_date must be YYYY-MM-DD")
		}
	}
	if p.Priority != nil && !p.Priority.Valid() {
		return todos.Fields{}, errors.New("priority must be one of high|normal|low")
	}
	return todos.Fields{
		Title:    p.Title,
		Memo:     strings.TrimSpace(p.Memo),
		DueDate:  p.DueDate,
		Priority: p.Priority,
		Done:     p.Done,
	}, nil
}
