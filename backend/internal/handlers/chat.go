package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/llm"
)

const chatHistoryLimit = 50

const maxChatMessageBytes = 16 * 1024

// ChatStreamer is the slice of *llm.Client the chat handler needs.
type ChatStreamer interface {
	ChatStream(ctx context.Context, req llm.ChatRequest, onDelta func(llm.StreamDelta) error) error
}

// ListChatSessions handles GET /api/chat/sessions.
func ListChatSessions(store *chat.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		items, err := store.ListSessions(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "list failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

// CreateChatSession handles POST /api/chat/sessions.
func CreateChatSession(store *chat.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		created, err := store.CreateSession(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "create failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

// GetChatSession handles GET /api/chat/sessions/{id}: the session and its
// full message history, so re-entry restores the conversation (PRD-1 AC1·AC4).
func GetChatSession(store *chat.Store) http.HandlerFunc {
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
		session, err := store.GetSession(r.Context(), u.ID, id)
		if errors.Is(err, chat.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		messages, err := store.Messages(r.Context(), u.ID, id, 0)
		if err != nil {
			http.Error(w, "get failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"session": session, "messages": messages})
	}
}

type chatMessagePayload struct {
	Content string `json:"content"`
}

// SendChatMessage handles POST /api/chat/sessions/{id}/messages. It stores the
// user's message, streams the LLM reply as server-sent events
// (`data: {"delta":"…"}` per chunk, then `event: done` with the stored
// assistant message), and stores the reply once the stream ends. A nil
// streamer (no OPENROUTER_API_KEY) answers 503 before anything is stored.
func SendChatMessage(store *chat.Store, llmClient ChatStreamer, model string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		if llmClient == nil {
			http.Error(w, "chat is not configured", http.StatusServiceUnavailable)
			return
		}
		id, err := pathID(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		if _, err := store.AppendMessage(r.Context(), u.ID, id, chat.RoleUser, content); err != nil {
			if errors.Is(err, chat.ErrNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "save failed", http.StatusInternalServerError)
			return
		}
		history, err := store.Messages(r.Context(), u.ID, id, chatHistoryLimit)
		if err != nil {
			http.Error(w, "history failed", http.StatusInternalServerError)
			return
		}
		msgs := make([]llm.ChatMessage, 0, len(history))
		for _, m := range history {
			msgs = append(msgs, llm.ChatMessage{Role: string(m.Role), Content: m.Content})
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()

		var reply strings.Builder
		streamErr := llmClient.ChatStream(r.Context(), llm.ChatRequest{Model: model, Messages: msgs}, func(d llm.StreamDelta) error {
			for _, c := range d.Choices {
				if c.Delta.Content == "" {
					continue
				}
				reply.WriteString(c.Delta.Content)
				if err := writeSSE(w, flusher, "", map[string]string{"delta": c.Delta.Content}); err != nil {
					return err
				}
			}
			return nil
		})
		// Keep whatever arrived: the user saw it, and it is context for the
		// next turn. The save uses a fresh context so a client disconnect
		// does not drop the partial reply.
		var saved *chat.Message
		if reply.Len() > 0 {
			m, err := store.AppendMessage(context.WithoutCancel(r.Context()), u.ID, id, chat.RoleAssistant, reply.String())
			if err == nil {
				saved = &m
			}
		}
		if streamErr != nil {
			_ = writeSSE(w, flusher, "error", map[string]string{"message": "reply failed"})
			return
		}
		_ = writeSSE(w, flusher, "done", map[string]any{"message": saved})
	}
}

func writeSSE(w http.ResponseWriter, f http.Flusher, event string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if event != "" {
		if _, err := w.Write([]byte("event: " + event + "\n")); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte("data: " + string(raw) + "\n\n")); err != nil {
		return err
	}
	f.Flush()
	return nil
}
