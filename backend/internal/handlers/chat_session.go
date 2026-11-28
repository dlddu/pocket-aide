package handlers

import (
	"errors"
	"net/http"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// GetChatSession handles GET /api/chat/session: whether the codex session
// answering the caller's own conversation is on, and the detail the tab shows
// beside that (PRD-1 AC12).
//
// Both the session and whose conversation it is come from the signed in user's
// row, so this route takes no parameters and cannot be pointed at another
// user's session. It asks the platform about that one session only and never
// reads its transcript, so the chat tab may poll it without promoting an idle
// session or restoring an archived one (PRD-11 AC5).
func GetChatSession(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		power, err := engine.SessionPowerOf(r.Context(), u.ID)
		if err != nil {
			writeChatPowerError(w, err)
			return
		}
		// The answer is about a session whose state changes outside this
		// server, so a stored copy must never stand in for asking again.
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, power)
	}
}

// PowerOnChatSession handles POST /api/chat/session/power: turns the session
// answering the caller's own conversation back on (PRD-1 AC12).
//
// It restores the session the conversation row already names — the same
// explicit restore the sessions tab offers (PRD-11 AC9) — so the conversation
// carries on in it (AC4) and no second session appears, however many screens
// tap the button. The body is empty on purpose: there is nothing to choose,
// and "turn on the one answering me" cannot name somebody else's session.
//
// A session that is already on is answered with its state and no platform
// write. The refusals stay apart: no session yet (the first message opens
// one), a session the platform no longer has (nothing is lost — the history is
// this server's), the platform busy with that session, and no login slot free
// for the restore (PRD-11 AC11 · PRD-1 AC7).
func PowerOnChatSession(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		power, err := engine.PowerOnSession(r.Context(), u.ID)
		if err != nil {
			writeChatPowerError(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, power)
	}
}

func writeChatPowerError(w http.ResponseWriter, err error) {
	var platformErr *chat.PlatformError
	switch {
	case errors.Is(err, chat.ErrNoConversation):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": "send a message first", "code": "chat_power_no_session",
		})
	case errors.Is(err, chat.ErrPowerSessionGone):
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(), "code": "chat_power_session_gone",
		})
	case errors.Is(err, chat.ErrPowerBusy):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": err.Error(), "code": "chat_power_busy",
		})
	case errors.As(err, &platformErr) && platformErr.Status == http.StatusServiceUnavailable && !platformErr.AuthPoolExhausted:
		// Nothing here is behind the archive gate, and PRD-11 AC11 reads every
		// 503 other than the exhausted login slots as the platform not being
		// reachable — so this must not borrow the sessions tab's "이 환경에서는
		// 비활성" wording, which would tell the person a feature is off.
		writeUnreachable(w)
	default:
		writeChatError(w, err)
	}
}
