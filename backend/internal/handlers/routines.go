package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/routines"
)

const routineHistoryDays = 30

type routinePayload struct {
	Name     string   `json:"name"`
	Cadence  string   `json:"cadence"`
	Weekdays int      `json:"weekdays"`
	MonthDay int      `json:"month_day"`
	StartDay *string  `json:"start_day"`
	Steps    []string `json:"steps"`
}

type routineStepPayload struct {
	Title string `json:"title"`
}

type routineCheckPayload struct {
	Checked *bool `json:"checked"`
}

func decodeStrict(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func int64Param(r *http.Request, name string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, name), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid " + name)
	}
	return id, nil
}

func writeRoutineError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, routines.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, routines.ErrInvalidDay):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, routines.ErrInvalidSchedule):
		http.Error(w, "cadence must be daily, weekdays (weekdays mask 1..127), weekly (one weekday bit) or monthly (month_day 1..31)", http.StatusBadRequest)
	case errors.Is(err, routines.ErrNotScheduled):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}

func ListRoutines(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		items, err := store.List(r.Context(), u.ID)
		if err != nil {
			writeRoutineError(w, err, "list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func CreateRoutine(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var p routinePayload
		if err := decodeStrict(r, &p); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		steps := make([]string, 0, len(p.Steps))
		for _, s := range p.Steps {
			if t := strings.TrimSpace(s); t != "" {
				steps = append(steps, t)
			}
		}
		startDay := time.Now().UTC().Format("2006-01-02")
		if p.StartDay != nil {
			startDay = *p.StartDay
		}
		schedule := routines.Schedule{Cadence: routines.Cadence(p.Cadence), Weekdays: p.Weekdays, MonthDay: p.MonthDay}
		created, err := store.Create(r.Context(), u.ID, name, schedule, startDay, steps)
		if err != nil {
			writeRoutineError(w, err, "create failed")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func DeleteRoutine(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := int64Param(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := store.Delete(r.Context(), u.ID, id); err != nil {
			writeRoutineError(w, err, "delete failed")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func AddRoutineStep(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := int64Param(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var p routineStepPayload
		if err := decodeStrict(r, &p); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		title := strings.TrimSpace(p.Title)
		if title == "" {
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}
		updated, err := store.AddStep(r.Context(), u.ID, id, title)
		if err != nil {
			writeRoutineError(w, err, "add step failed")
			return
		}
		writeJSON(w, http.StatusCreated, updated)
	}
}

func DeleteRoutineStep(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := int64Param(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stepID, err := int64Param(r, "stepID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := store.DeleteStep(r.Context(), u.ID, id, stepID); err != nil {
			writeRoutineError(w, err, "delete step failed")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func ListRoutineDay(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		day := chi.URLParam(r, "day")
		items, err := store.Day(r.Context(), u.ID, day)
		if err != nil {
			writeRoutineError(w, err, "list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"day": day, "items": items})
	}
}

func SetRoutineStepCheck(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := int64Param(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		stepID, err := int64Param(r, "stepID")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var p routineCheckPayload
		if err := decodeStrict(r, &p); err != nil || p.Checked == nil {
			http.Error(w, "checked is required", http.StatusBadRequest)
			return
		}
		updated, err := store.SetCheck(r.Context(), u.ID, id, stepID, chi.URLParam(r, "day"), *p.Checked)
		if err != nil {
			writeRoutineError(w, err, "check failed")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func RoutineHistory(store *routines.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id, err := int64Param(r, "id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		items, err := store.History(r.Context(), u.ID, id, chi.URLParam(r, "day"), routineHistoryDays)
		if err != nil {
			writeRoutineError(w, err, "history failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}
