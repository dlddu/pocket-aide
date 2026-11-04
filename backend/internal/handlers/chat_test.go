package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/llm"
)

type fakeStreamer struct {
	chunks []string
	err    error
	got    llm.ChatRequest
}

func (f *fakeStreamer) ChatStream(_ context.Context, req llm.ChatRequest, onDelta func(llm.StreamDelta) error) error {
	f.got = req
	for _, c := range f.chunks {
		var d llm.StreamDelta
		d.Choices = make([]struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		}, 1)
		d.Choices[0].Delta.Content = c
		if err := onDelta(d); err != nil {
			return err
		}
	}
	return f.err
}

func newChatRouter(t *testing.T, streamer handlers.ChatStreamer) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := chat.New(conn)
	r := chi.NewRouter()
	r.Get("/api/chat/sessions", handlers.ListChatSessions(store))
	r.Post("/api/chat/sessions", handlers.CreateChatSession(store))
	r.Get("/api/chat/sessions/{id}", handlers.GetChatSession(store))
	r.Post("/api/chat/sessions/{id}/messages", handlers.SendChatMessage(store, streamer, "test-model"))
	return r
}

func chatDo(t *testing.T, router http.Handler, method, target string, body any, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		if raw, err = json.Marshal(body); err != nil {
			t.Fatalf("marshal: %v", err)
		}
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, authedRequest(method, target, raw, userID))
	return rec
}

func newSession(t *testing.T, router http.Handler, userID int64) int64 {
	t.Helper()
	rec := chatDo(t, router, http.MethodPost, "/api/chat/sessions", nil, userID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create session: %d %s", rec.Code, rec.Body.String())
	}
	var s chat.Session
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return s.ID
}

func TestChat_SendStreamsAndPersists(t *testing.T) {
	fs := &fakeStreamer{chunks: []string{"안녕", "하세요"}}
	router := newChatRouter(t, fs)
	id := newSession(t, router, 1)
	target := "/api/chat/sessions/" + strconv.FormatInt(id, 10)

	rec := chatDo(t, router, http.MethodPost, target+"/messages", map[string]string{"content": "  hello  "}, 1)
	if rec.Code != http.StatusOK {
		t.Fatalf("send: %d %s", rec.Code, rec.Body.String())
	}
	out := rec.Body.String()
	for _, want := range []string{`data: {"delta":"안녕"}`, `data: {"delta":"하세요"}`, "event: done"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stream missing %q:\n%s", want, out)
		}
	}
	if fs.got.Model != "test-model" || len(fs.got.Messages) != 1 || fs.got.Messages[0].Content != "hello" {
		t.Fatalf("unexpected LLM request: %+v", fs.got)
	}

	// Re-entry restores the history (AC1) and the title comes from the first message.
	rec = chatDo(t, router, http.MethodGet, target, nil, 1)
	var got struct {
		Session  chat.Session   `json:"session"`
		Messages []chat.Message `json:"messages"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Session.Title != "hello" || len(got.Messages) != 2 ||
		got.Messages[0].Role != chat.RoleUser || got.Messages[1].Role != chat.RoleAssistant ||
		got.Messages[1].Content != "안녕하세요" {
		t.Fatalf("unexpected history: %+v", got)
	}

	// The next turn carries the earlier context to the LLM (AC4 resume).
	fs.chunks = []string{"ok"}
	chatDo(t, router, http.MethodPost, target+"/messages", map[string]string{"content": "again"}, 1)
	if len(fs.got.Messages) != 3 {
		t.Fatalf("want 3 context messages, got %+v", fs.got.Messages)
	}
}

func TestChat_SessionsListAndIsolation(t *testing.T) {
	router := newChatRouter(t, &fakeStreamer{})
	a := newSession(t, router, 1)
	newSession(t, router, 1)
	newSession(t, router, 2)

	rec := chatDo(t, router, http.MethodGet, "/api/chat/sessions", nil, 1)
	var list struct {
		Items []chat.Session `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("user 1 should see 2 sessions, got %d", len(list.Items))
	}

	target := "/api/chat/sessions/" + strconv.FormatInt(a, 10)
	if rec := chatDo(t, router, http.MethodGet, target, nil, 2); rec.Code != http.StatusNotFound {
		t.Fatalf("other user's session read: want 404, got %d", rec.Code)
	}
	if rec := chatDo(t, router, http.MethodPost, target+"/messages", map[string]string{"content": "x"}, 2); rec.Code != http.StatusNotFound {
		t.Fatalf("other user's session write: want 404, got %d", rec.Code)
	}
}

func TestChat_NotConfiguredIs503(t *testing.T) {
	router := newChatRouter(t, nil)
	id := newSession(t, router, 1)
	target := "/api/chat/sessions/" + strconv.FormatInt(id, 10)
	if rec := chatDo(t, router, http.MethodPost, target+"/messages", map[string]string{"content": "hi"}, 1); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}
	rec := chatDo(t, router, http.MethodGet, target, nil, 1)
	if !strings.Contains(rec.Body.String(), `"messages":[]`) {
		t.Fatalf("503 must not store the message: %s", rec.Body.String())
	}
}

func TestChat_RejectsBadInput(t *testing.T) {
	router := newChatRouter(t, &fakeStreamer{})
	id := newSession(t, router, 1)
	target := "/api/chat/sessions/" + strconv.FormatInt(id, 10) + "/messages"
	if rec := chatDo(t, router, http.MethodPost, target, map[string]string{"content": "   "}, 1); rec.Code != http.StatusBadRequest {
		t.Fatalf("blank content: want 400, got %d", rec.Code)
	}
	long := strings.Repeat("a", 16*1024+1)
	if rec := chatDo(t, router, http.MethodPost, target, map[string]string{"content": long}, 1); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("long content: want 413, got %d", rec.Code)
	}
	if rec := chatDo(t, router, http.MethodPost, "/api/chat/sessions/abc/messages", map[string]string{"content": "x"}, 1); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id: want 400, got %d", rec.Code)
	}
}

func TestChat_StreamErrorKeepsPartialReply(t *testing.T) {
	fs := &fakeStreamer{chunks: []string{"부분"}, err: errors.New("upstream down")}
	router := newChatRouter(t, fs)
	id := newSession(t, router, 1)
	target := "/api/chat/sessions/" + strconv.FormatInt(id, 10)
	rec := chatDo(t, router, http.MethodPost, target+"/messages", map[string]string{"content": "hi"}, 1)
	if !strings.Contains(rec.Body.String(), "event: error") {
		t.Fatalf("want error event: %s", rec.Body.String())
	}
	rec = chatDo(t, router, http.MethodGet, target, nil, 1)
	if !strings.Contains(rec.Body.String(), "부분") {
		t.Fatalf("partial reply should be kept: %s", rec.Body.String())
	}
}
