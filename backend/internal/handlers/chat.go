package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

const maxChatMessageBytes = 16 * 1024

// GetChat handles GET /api/chat: the user's one conversation, its messages,
// and the reply still being written, if any — the apps poll this while a
// reply is pending (PRD-1 AC1·AC4).
func GetChat(store *chat.Store, engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		conv, err := store.Conversation(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		messages, err := store.Messages(r.Context(), u.ID, conv.ID, 0)
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		queued, err := store.GetQueued(r.Context(), u.ID, conv.ID)
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"conversation": conv,
			"messages":     messages,
			"pending":      engine.Pending(u.ID),
			"queued":       queued,
		})
	}
}

type chatMessagePayload struct {
	Content string `json:"content"`
}

// SendChatMessage handles POST /api/chat/messages. It stores the message and
// hands it to the user's codex session at once, or — if a reply is already
// in progress — holds it to send right after (PRD-1 AC1, 2026-11-09 개정),
// answering 202 either way; the reply arrives through GET /api/chat. The
// first message of a codex session asks for a ChatGPT sign-in, and that
// request is the reply.
func SendChatMessage(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var body chatMessagePayload
		r.Body = http.MaxBytesReader(w, r.Body, 2*maxChatMessageBytes)
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		content := strings.TrimSpace(body.Content)
		if content == "" {
			http.Error(w, "content is required", http.StatusBadRequest)
			return
		}
		if len(content) > maxChatMessageBytes {
			http.Error(w, "content too long", http.StatusRequestEntityTooLarge)
			return
		}
		msg, queued, err := engine.Queue(r.Context(), u.ID, content)
		if err != nil {
			writeChatError(w, err)
			return
		}
		if queued != nil {
			writeJSON(w, http.StatusAccepted, map[string]any{"queued": queued})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"message": msg})
	}
}

// CancelQueuedChatMessage handles DELETE /api/chat/queue: drops the message
// held for after the reply in progress, if it has not already been sent.
func CancelQueuedChatMessage(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		if err := engine.CancelQueued(r.Context(), u.ID); err != nil {
			writeChatError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// RefreshChatPlugins handles POST /api/chat/refresh-plugins: the codex
// session fetches its plugin skills again without losing the conversation.
func RefreshChatPlugins(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		result, err := engine.RefreshPlugins(r.Context(), u.ID)
		if err != nil {
			writeChatError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func writeChatError(w http.ResponseWriter, err error) {
	var platformErr *chat.PlatformError
	switch {
	case errors.Is(err, chat.ErrQueueFull):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "a message is already queued"})
	case errors.Is(err, chat.ErrNoQueuedMessage):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no message queued"})
	case errors.Is(err, chat.ErrBusy):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "a reply is still in progress"})
	case errors.Is(err, chat.ErrNoConversation):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "send a message first"})
	case errors.As(err, &platformErr) && (platformErr.Status == http.StatusUnauthorized || platformErr.Status == http.StatusForbidden):
		// The platform refused the chat tab's tenant; that is not the app user's credential.
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "session platform refused the chat tenant"})
	case errors.As(err, &platformErr) && platformErr.AuthPoolExhausted:
		writeAuthPoolExhausted(w)
	case errors.As(err, &platformErr):
		writeSessionError(w, platformErr.Status)
	default:
		writeUnreachable(w)
	}
}
