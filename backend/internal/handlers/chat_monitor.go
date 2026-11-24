package handlers

import (
	"errors"
	"net/http"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// The separated monitor's opt-in, from the settings screen or the conversation.
//
// Both routes act on the caller's own conversation: the user comes from the
// authenticated context and the parent session from their chat row, so no
// request names a session and nobody can bind someone else's chat session.
func monitorError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, chat.ErrNoConversation):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": "채팅 세션이 아직 없어 모니터를 붙일 수 없습니다.",
			"code":  "no_conversation",
		})
	case errors.Is(err, chat.ErrPairRefused):
		// An authorization failure, kept apart from "no monitor" on purpose.
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": err.Error(), "code": "monitor_pair_refused",
		})
	case errors.Is(err, chat.ErrPairUnknown):
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": err.Error(), "code": "monitor_pair_unknown",
		})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": err.Error(), "code": "monitor_setup_failed",
		})
	}
}

// PostChatMonitor turns the monitor on. It is idempotent: a conversation that
// already has one gets it back with `created: false`, and no second session is
// made.
func PostChatMonitor(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		setup, err := engine.EnableMonitor(r.Context(), u.ID)
		if err != nil {
			monitorError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, setup)
	}
}

// GetChatMonitor reads whether this conversation has a monitor. It is a read:
// it creates nothing and does not wake the monitor.
func GetChatMonitor(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		setup, err := engine.MonitorStatus(r.Context(), u.ID)
		if err != nil {
			monitorError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, setup)
	}
}

// DeleteChatMonitor turns the monitor off. It is the rollback the platform runs
// in order — the monitor is closed before the work records are handed back — so
// this route only reports which of the two happened. Repeating it is safe, and
// a half-done rollback answers 202 with `retry` rather than claiming success.
func DeleteChatMonitor(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		off, err := engine.DisableMonitor(r.Context(), u.ID)
		if err != nil {
			monitorError(w, err)
			return
		}
		if off.Retry {
			writeJSON(w, http.StatusAccepted, off)
			return
		}
		writeJSON(w, http.StatusOK, off)
	}
}
