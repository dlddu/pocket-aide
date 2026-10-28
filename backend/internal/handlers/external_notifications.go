package handlers

import (
	"context"
	"errors"
	stdlog "log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/externalnotify"
)

const maxNotificationTitleRunes = 120

const maxNotificationRequesterRunes = 200

const externalNotificationTimeout = 15 * time.Second

type externalNotificationPayload struct {
	Title     *string `json:"title"`
	Body      *string `json:"body"`
	Requester *string `json:"requester"`
}

type NotificationSender interface {
	Send(ctx context.Context, userID, notificationID int64, title, body string) (externalnotify.Result, error)
}

// NotificationLog records every external notification with its caller (PRD-12 AC17).
type NotificationLog interface {
	Insert(ctx context.Context, userID, callerKeyID int64, n externalnotify.Notification) (int64, error)
	SetDelivery(ctx context.Context, id int64, r externalnotify.Result) error
	Delete(ctx context.Context, id int64) error
}

// CreateExternalNotification records, then sends, a plain push to the devices of the user who registered the caller (PRD-12 AC16·AC17).
func CreateExternalNotification(push NotificationSender, log NotificationLog) http.HandlerFunc {
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
		requester := trimmedOptional(p.Requester)
		if requester == "" {
			writeError(w, http.StatusBadRequest, "requester is required")
			return
		}
		if utf8.RuneCountInString(requester) > maxNotificationRequesterRunes {
			writeError(w, http.StatusBadRequest, "requester is too long")
			return
		}
		if push == nil {
			writeError(w, http.StatusServiceUnavailable, "push notifications unavailable")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), externalNotificationTimeout)
		defer cancel()
		id, err := log.Insert(ctx, caller.UserID, caller.KeyID, externalnotify.Notification{Title: title, Body: body, Requester: requester})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "record failed")
			return
		}
		res, err := push.Send(ctx, caller.UserID, id, title, body)
		if err != nil {
			if err := log.Delete(context.WithoutCancel(ctx), id); err != nil {
				stdlog.Printf("external notification %d: drop unsent record: %v", id, err)
			}
			writeError(w, http.StatusInternalServerError, "send failed")
			return
		}
		if err := log.SetDelivery(context.WithoutCancel(ctx), id, res); err != nil {
			stdlog.Printf("external notification %d: record delivery: %v", id, err)
		}
		writeJSON(w, http.StatusOK, res)
	}
}

// ExternalNotificationReader is the user-facing side of the notification log.
type ExternalNotificationReader interface {
	List(ctx context.Context, userID int64, limit int, beforeID int64) ([]externalnotify.Item, error)
	UnreadCount(ctx context.Context, userID int64) (int, error)
	Get(ctx context.Context, userID, id int64) (externalnotify.Item, error)
	MarkRead(ctx context.Context, userID, id int64) error
}

// ListExternalNotifications handles GET /api/external-notifications (PRD-12 AC17).
func ListExternalNotifications(store ExternalNotificationReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
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
			writeError(w, http.StatusInternalServerError, "list failed")
			return
		}
		unread, err := store.UnreadCount(r.Context(), u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "count failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "unread_count": unread})
	}
}

// GetExternalNotification handles GET /api/external-notifications/{id}: the detail a push tap opens (PRD-12 AC16·AC17).
func GetExternalNotification(store ExternalNotificationReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		id, err := pathID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		item, err := store.Get(r.Context(), u.ID, id)
		if errors.Is(err, externalnotify.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

// MarkExternalNotificationRead handles POST /api/external-notifications/{id}/read: opening the detail marks it seen (PRD-12 AC17).
func MarkExternalNotificationRead(store ExternalNotificationReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		id, err := pathID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		err = store.MarkRead(r.Context(), u.ID, id)
		if errors.Is(err, externalnotify.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "mark read failed")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
