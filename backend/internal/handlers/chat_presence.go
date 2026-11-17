package handlers

import (
	"net/http"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// ChatPresenceReader reports whether a user is on the chat screen.
type ChatPresenceReader interface {
	State(userID int64) chat.State
}

// GetExternalChatPresence handles GET /api/external/chat-presence: whether the
// user who registered this caller is looking at the chat screen right now
// (PRD-12 AC22).
//
// The subject is that registration, not a parameter — the same attribution as
// every other caller route (AC2) — and the answer carries no conversation
// content, only whether and when the screen last read it.
func GetExternalChatPresence(presence ChatPresenceReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := callerFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no caller in context")
			return
		}
		writeJSON(w, http.StatusOK, presence.State(caller.UserID))
	}
}
