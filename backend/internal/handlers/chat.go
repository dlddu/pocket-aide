package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

const maxChatMessageBytes = 16 * 1024

// chatArchiveBodyLimit bounds the archive request's body. It carries one
// session id, so anything larger is not a body this route has to read.
const chatArchiveBodyLimit = 4 << 10

// GetChat handles GET /api/chat: the user's one conversation, its messages,
// and the reply still being written, if any — the apps poll this while a
// reply is pending (PRD-1 AC1·AC4).
//
// Only the chat screen polls this route, and only while it is visible, so the
// read doubles as the presence signal (PRD-1 AC8). Recorded before the store
// is read: a DB failure does not mean the user stopped looking.
//
// ?after=<message id> returns only the messages after it, and If-None-Match
// with the last ETag gets an empty 304 — one conversation grows without end
// and the apps poll it every 1.5 s during a reply, so a full read per poll
// cost gigabytes a day of mobile data (PRD-1 AC1, 2026-11-27 개정).
func GetChat(store *chat.Store, engine *chat.Engine, presence *chat.Presence) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		presence.Touch(u.ID)
		after, ok := chatAfterParam(r)
		if !ok {
			http.Error(w, "after must be a non-negative message id", http.StatusBadRequest)
			return
		}
		conv, err := store.Conversation(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		var messages []chat.Message
		if after > 0 {
			messages, err = store.MessagesAfter(r.Context(), u.ID, conv.ID, after)
		} else {
			messages, err = store.Messages(r.Context(), u.ID, conv.ID, 0)
		}
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		queued, err := store.GetQueued(r.Context(), u.ID, conv.ID)
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		// A restart kills the turn that would have sent a waiting message, so
		// this poll is what gets it moving again (PRD-1 AC1 2026-11-09 개정).
		if queued != nil {
			engine.ResumeQueued(r.Context(), u.ID, conv.ID)
		}
		writeJSONWithETag(w, r, map[string]any{
			"conversation": conv,
			"messages":     messages,
			"pending":      engine.Pending(u.ID),
			"queued":       queued,
		})
	}
}

func chatAfterParam(r *http.Request) (int64, bool) {
	raw := r.URL.Query().Get("after")
	if raw == "" {
		return 0, true
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// writeJSONWithETag tags the body with a hash of itself, so the tag changes
// exactly when the response would, and answers a matching If-None-Match with
// an empty 304. no-cache keeps clients from reusing a stored copy without
// asking — the conversation can change on every poll.
func writeJSONWithETag(w http.ResponseWriter, r *http.Request, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "encode failed", http.StatusInternalServerError)
		return
	}
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "private, no-cache")
	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(append(body, '\n'))
}

func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimPrefix(strings.TrimSpace(candidate), "W/")
		if candidate == etag || candidate == "*" {
			return true
		}
	}
	return false
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

// CancelQueuedChatMessage handles DELETE /api/chat/queue?id=<id>: drops the
// message held for after the reply in progress, if that message is still the
// one waiting. The id is required so a cancel that arrives just after its
// message went out answers 404 instead of dropping the next one.
func CancelQueuedChatMessage(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		id := strings.TrimSpace(r.URL.Query().Get("id"))
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := engine.CancelQueued(r.Context(), u.ID, id); err != nil {
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

// chatArchivePayload names the session the person confirmed. There is no user
// field: the server derives whose conversation it is from the signed in user,
// and the session from that conversation's own row.
type chatArchivePayload struct {
	SessionID string `json:"session_id"`
}

// ArchiveChatSession handles POST /api/chat/archive: hands back the
// environment of the codex session answering the caller's own conversation and
// has the platform keep its work (PRD-1 AC4).
//
// The chat history is not touched — it is this server's, not the session's —
// and the conversation keeps pointing at the same session, so the next message
// restores it and carries on. The body names the session so an archive cannot
// land on one that took over between the confirmation and the tap; that
// refusal is its own code, told apart from the others the app explains
// differently: no session yet, a reply in progress, a message already waiting,
// a session already gone, and archiving turned off on the platform — which is
// not a failure and loses nothing.
func ArchiveChatSession(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var body chatArchivePayload
		if err := json.NewDecoder(io.LimitReader(r.Body, chatArchiveBodyLimit)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "body must be {session_id}", "code": "chat_archive_session_unnamed",
			})
			return
		}
		result, err := engine.ArchiveSession(r.Context(), u.ID, body.SessionID)
		if err != nil {
			writeChatArchiveError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func writeChatArchiveError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, chat.ErrArchiveSessionUnnamed):
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(), "code": "chat_archive_session_unnamed",
		})
	case errors.Is(err, chat.ErrNoConversation):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": "send a message first", "code": "chat_archive_no_session",
		})
	case errors.Is(err, chat.ErrArchiveSessionChanged):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": err.Error(), "code": "chat_archive_session_changed",
		})
	case errors.Is(err, chat.ErrArchiveQueued):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": err.Error(), "code": "chat_archive_queued",
		})
	case errors.Is(err, chat.ErrBusy):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error": err.Error(), "code": "chat_archive_busy",
		})
	case errors.Is(err, chat.ErrArchiveSessionGone):
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": err.Error(), "code": "chat_archive_session_gone",
		})
	case errors.Is(err, chat.ErrArchiveDisabled):
		// Its own 503 code, so the app reads it as "this environment does not
		// archive" and not as the platform being out of login slots.
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": err.Error(), "code": "chat_archive_disabled",
		})
	default:
		writeChatError(w, err)
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
