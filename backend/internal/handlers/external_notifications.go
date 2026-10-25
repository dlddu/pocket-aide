package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dlddu/pocket-aide/backend/internal/externalnotify"
)

const maxNotificationTitleRunes = 120

const externalNotificationTimeout = 15 * time.Second

type externalNotificationPayload struct {
	Title *string `json:"title"`
	Body  *string `json:"body"`
}

type NotificationSender interface {
	Send(ctx context.Context, userID int64, title, body string) (externalnotify.Result, error)
}

// CreateExternalNotification sends a plain push to the devices of the user who registered the caller (PRD-12 AC16).
func CreateExternalNotification(push NotificationSender) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := callerFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no caller in context")
			return
		}
		var p externalNotificationPayload
		if err := decodeApprovalBody(w, r, &p); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if p.Title == nil || strings.TrimSpace(*p.Title) == "" {
			writeError(w, http.StatusBadRequest, "title is required")
			return
		}
		title := strings.TrimSpace(*p.Title)
		if utf8.RuneCountInString(title) > maxNotificationTitleRunes {
			writeError(w, http.StatusBadRequest, "title is too long")
			return
		}
		var body string
		if p.Body != nil {
			body = strings.TrimSpace(*p.Body)
		}
		if push == nil {
			writeError(w, http.StatusServiceUnavailable, "push notifications unavailable")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), externalNotificationTimeout)
		defer cancel()
		res, err := push.Send(ctx, caller.UserID, title, body)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "send failed")
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}
