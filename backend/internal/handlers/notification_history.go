package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/notificationhistory"
)

// ListNotificationHistory handles GET /api/notification-history with optional
// `?limit=&before=` query parameters (limit clamped to 100, before is a row
// id used for keyset pagination — pass the last id of the previous page).
func ListNotificationHistory(store *notificationhistory.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		limit := 50
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if n, err := strconv.Atoi(raw); err == nil && n > 0 {
				limit = n
			}
		}
		var before int64
		if raw := r.URL.Query().Get("before"); raw != "" {
			if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
				before = n
			}
		}
		items, err := store.List(r.Context(), u.ID, limit, before)
		if err != nil {
			http.Error(w, "list failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

// AcknowledgeNotification handles POST /api/notification-history/{id}/ack.
// AC12: only the explicit button click reaches here. Push taps (AC7) and
// external-link taps must NOT call this endpoint.
func AcknowledgeNotification(store *notificationhistory.Store) http.HandlerFunc {
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
		err = store.Acknowledge(r.Context(), u.ID, id)
		if errors.Is(err, notificationhistory.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "ack failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
