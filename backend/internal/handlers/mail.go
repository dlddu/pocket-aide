package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/mail"
)

type mailAccountPayload struct {
	Email       string `json:"email"`
	AppPassword string `json:"app_password"`
	IMAPHost    string `json:"imap_host"`
}

type mailPasswordPayload struct {
	AppPassword string `json:"app_password"`
}

// MailDisabled answers every mail route with 503 when no encryption key is
// configured: the feature is off, the rest of the server is unaffected.
func MailDisabled() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "mail is not configured", http.StatusServiceUnavailable)
	}
}

// ListMailAccounts handles GET /api/mail/accounts.
func ListMailAccounts(svc *mail.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		items, err := svc.Store.List(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "list failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

// ConnectMailAccount handles POST /api/mail/accounts.
func ConnectMailAccount(svc *mail.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var body mailAccountPayload
		if err := decodeStrict(r, &body); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		acct, err := svc.Connect(r.Context(), u.ID, body.Email, body.AppPassword, body.IMAPHost)
		if writeMailError(w, err) {
			return
		}
		writeJSON(w, http.StatusCreated, acct)
	}
}

// ReconnectMailAccount handles PUT /api/mail/accounts/{id}/password.
func ReconnectMailAccount(svc *mail.Service) http.HandlerFunc {
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
		var body mailPasswordPayload
		if err := decodeStrict(r, &body); err != nil {
			http.Error(w, "invalid json body", http.StatusBadRequest)
			return
		}
		if writeMailError(w, svc.Reconnect(r.Context(), u.ID, id, body.AppPassword)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// DisconnectMailAccount handles DELETE /api/mail/accounts/{id}.
func DisconnectMailAccount(svc *mail.Service) http.HandlerFunc {
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
		if writeMailError(w, svc.Store.Delete(r.Context(), u.ID, id)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// ListMailMessages handles GET /api/mail/messages[?account_id=N].
func ListMailMessages(svc *mail.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var accountID int64
		if raw := r.URL.Query().Get("account_id"); raw != "" {
			id, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || id <= 0 {
				http.Error(w, "invalid account_id", http.StatusBadRequest)
				return
			}
			accountID = id
		}
		items, statuses, err := svc.Messages(r.Context(), u.ID, accountID)
		if writeMailError(w, err) {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "accounts": statuses})
	}
}

func writeMailError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, mail.ErrInvalidAccount), errors.Is(err, mail.ErrHostRequired):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, mail.ErrAuth):
		http.Error(w, "mail login failed", http.StatusUnprocessableEntity)
	case errors.Is(err, mail.ErrNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	case errors.Is(err, mail.ErrAlreadyConnected):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "mail request failed", http.StatusBadGateway)
	}
	return true
}
