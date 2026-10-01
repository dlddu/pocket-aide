package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/notificationsettings"
)

func GetNotificationSettings(store *notificationsettings.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		settings, err := store.Get(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, settings)
	}
}

func UpdateNotificationSettings(store *notificationsettings.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var patch notificationsettings.Patch
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&patch); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		settings, err := store.Update(r.Context(), u.ID, patch)
		if errors.Is(err, notificationsettings.ErrInvalidOutcomes) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "update failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, settings)
	}
}
