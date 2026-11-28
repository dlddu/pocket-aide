package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

// chatPlatform is a session-platform stand-in with one codex session per
// create, answering each prompt as soon as it is written.
type chatPlatform struct {
	mu         sync.Mutex
	transcript map[string]string
	status     int    // forced status for every call when non-zero
	statusBody string // body sent with the forced status
	refresh    int
	// workStatus forces the read-only work task route's status; work is the
	// page it answers otherwise.
	workStatus int
	work       string
	// snapshot forces the archive route's status (0 = 200) and snapshotBody is
	// the body it sends with it — the 503s are told apart by body alone.
	snapshot     int
	snapshotBody string
	// archived is which sessions read as archived now, which is what the
	// engine re-reads rather than assuming from its own request.
	archived map[string]bool
}

func (p *chatPlatform) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status != 0 {
		w.WriteHeader(p.status)
		_, _ = w.Write([]byte(p.statusBody))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	switch {
	case path == "/sessions" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"sessions":[]}`))
	case path == "/sessions":
		id := fmt.Sprintf("cx%d", len(p.transcript)+1)
		p.transcript[id] = ""
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": id, "workloadType": body["workloadType"].(string)})
	case strings.HasSuffix(path, "/write"):
		id := strings.Split(path, "/")[2]
		seq := strings.Count(p.transcript[id], "input seq=") + 1
		p.transcript[id] += fmt.Sprintf("\x1esession-platform input seq=%d\n%s\n\x1esession-platform output seq=%d\nre: %s\n\x1esession-platform end seq=%d status=ok\n",
			seq, body["payload"], seq, body["payload"], seq)
		_, _ = w.Write([]byte(`{}`))
	case strings.HasSuffix(path, "/read"):
		id := strings.Split(path, "/")[2]
		t := p.transcript[id]
		offset := int(body["offset"].(float64))
		_ = json.NewEncoder(w).Encode(map[string]any{"payload": t[min(offset, len(t)):], "nextOffset": len(t)})
	case strings.Contains(path, "/work/tasks"):
		if p.workStatus != 0 {
			w.WriteHeader(p.workStatus)
			_, _ = w.Write([]byte(`{"code":"x","error":"no"}`))
			return
		}
		_, _ = w.Write([]byte(p.work))
	case strings.HasSuffix(path, "/snapshot"):
		if p.snapshot != 0 {
			w.WriteHeader(p.snapshot)
			_, _ = w.Write([]byte(p.snapshotBody))
			return
		}
		p.archived[strings.Split(path, "/")[2]] = true
		_, _ = w.Write([]byte(`{}`))
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/sessions/") && strings.Count(path, "/") == 2:
		id := strings.Split(path, "/")[2]
		out := map[string]any{"id": id, "state": "active"}
		if p.archived[id] {
			out["state"] = "snapshot"
			out["checkpoint"] = map[string]any{"createdAt": "2026-11-28T06:00:00Z", "sizeBytes": 2048}
		}
		_ = json.NewEncoder(w).Encode(out)
	case strings.HasSuffix(path, "/plugins/refresh"):
		if p.refresh != 0 {
			w.WriteHeader(p.refresh)
			return
		}
		_, _ = w.Write([]byte(`{"state":"ready","added":[],"updated":["p--s"],"removed":[]}`))
	default:
		w.WriteHeader(http.StatusTeapot)
	}
}

func newChatRouter(t *testing.T) (http.Handler, *chatPlatform, *chat.Engine) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	platform := &chatPlatform{transcript: map[string]string{}, work: `{"tasks":[],"next_cursor":""}`, archived: map[string]bool{}}
	srv := httptest.NewServer(platform)
	t.Cleanup(srv.Close)
	store := chat.New(conn)
	engine := chat.NewEngine(store, sessions.NewClient(srv.URL), nil).WithPoll(5 * time.Millisecond)
	t.Cleanup(engine.Wait)
	r := chi.NewRouter()
	r.Get("/api/chat", handlers.GetChat(store, engine, chat.NewPresence()))
	r.Get("/api/chat/tasks", handlers.GetChatTasks(engine))
	r.Post("/api/chat/messages", handlers.SendChatMessage(engine))
	r.Delete("/api/chat/queue", handlers.CancelQueuedChatMessage(engine))
	r.Post("/api/chat/refresh-plugins", handlers.RefreshChatPlugins(engine))
	r.Post("/api/chat/archive", handlers.ArchiveChatSession(engine))
	return r, platform, engine
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

type chatView struct {
	Conversation struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
		// The session answering this conversation, which is what an action on
		// it names so it cannot land on one that took over since (PRD-1 AC11).
		SessionID string `json:"session_id"`
	} `json:"conversation"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Pending *struct {
		Content string `json:"content"`
	} `json:"pending"`
	Queued *struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	} `json:"queued"`
}

func getChat(t *testing.T, router http.Handler, userID int64) chatView {
	t.Helper()
	rec := chatDo(t, router, http.MethodGet, "/api/chat", nil, userID)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/chat = %d %s", rec.Code, rec.Body)
	}
	var v chatView
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

// PRD-1 AC1·AC4: a message is accepted at once (202) and its reply appears
// in the one conversation; each user sees only their own.
func TestChat_SendThenPoll(t *testing.T) {
	router, _, engine := newChatRouter(t)
	if v := getChat(t, router, 1); len(v.Messages) != 0 || v.Pending != nil || v.Conversation.ID == 0 {
		t.Fatalf("fresh chat = %+v", v)
	}
	rec := chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "  안녕  "}, 1)
	if rec.Code != http.StatusAccepted || !strings.Contains(rec.Body.String(), `"content":"안녕"`) {
		t.Fatalf("send = %d %s", rec.Code, rec.Body)
	}
	engine.Wait()
	v := getChat(t, router, 1)
	if len(v.Messages) != 2 || v.Messages[1].Role != "assistant" || v.Messages[1].Content != "re: 안녕" || v.Conversation.Title != "안녕" {
		t.Fatalf("chat after reply = %+v", v)
	}
	if other := getChat(t, router, 2); len(other.Messages) != 0 {
		t.Fatalf("user 2 sees %+v", other.Messages)
	}
}

// PRD-1 AC1 (2026-11-27 개정): ?after returns only the newer messages, a
// repeat with the last ETag is an empty 304, and a change in the conversation
// changes the tag.
func TestChat_IncrementalReadAndETag(t *testing.T) {
	router, _, engine := newChatRouter(t)
	chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "하나"}, 1)
	engine.Wait()

	full := chatDo(t, router, http.MethodGet, "/api/chat", nil, 1)
	var all struct {
		Messages []struct {
			ID int64 `json:"id"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(full.Body.Bytes(), &all); err != nil || len(all.Messages) != 2 {
		t.Fatalf("full read = %s (%v)", full.Body, err)
	}
	last := all.Messages[1].ID
	afterLast := "/api/chat?after=" + strconv.FormatInt(last, 10)

	rec := chatDo(t, router, http.MethodGet, afterLast, nil, 1)
	var v chatView
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil || rec.Code != http.StatusOK || len(v.Messages) != 0 || v.Conversation.ID == 0 {
		t.Fatalf("after last = %d %s", rec.Code, rec.Body)
	}
	etag := rec.Header().Get("ETag")
	if etag == "" || rec.Header().Get("Cache-Control") != "private, no-cache" {
		t.Fatalf("headers = %v", rec.Header())
	}

	req := authedRequest(http.MethodGet, afterLast, nil, 1)
	req.Header.Set("If-None-Match", etag)
	notModified := httptest.NewRecorder()
	router.ServeHTTP(notModified, req)
	if notModified.Code != http.StatusNotModified || notModified.Body.Len() != 0 || notModified.Header().Get("ETag") != etag {
		t.Fatalf("conditional repeat = %d %q %v", notModified.Code, notModified.Body, notModified.Header())
	}

	chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "둘"}, 1)
	engine.Wait()
	req = authedRequest(http.MethodGet, afterLast, nil, 1)
	req.Header.Set("If-None-Match", etag)
	changed := httptest.NewRecorder()
	router.ServeHTTP(changed, req)
	if err := json.Unmarshal(changed.Body.Bytes(), &v); err != nil || changed.Code != http.StatusOK || changed.Header().Get("ETag") == etag {
		t.Fatalf("after change = %d %s", changed.Code, changed.Body)
	}
	if len(v.Messages) != 2 || v.Messages[0].Content != "둘" || v.Messages[1].Content != "re: 둘" {
		t.Fatalf("only the newer messages = %+v", v.Messages)
	}
}

func TestChat_AfterMustBeAMessageID(t *testing.T) {
	router, _, _ := newChatRouter(t)
	for _, q := range []string{"abc", "-1", "1.5"} {
		if rec := chatDo(t, router, http.MethodGet, "/api/chat?after="+q, nil, 1); rec.Code != http.StatusBadRequest {
			t.Fatalf("after=%s = %d", q, rec.Code)
		}
	}
}

// PRD-1 AC1 (2026-11-09 개정): cancelling with nothing queued is a 404, and
// GET /api/chat reports no queued message while idle. The id says which
// message the cancel means, so a cancel without one is a 400.
func TestChat_CancelQueuedWithNothingQueued(t *testing.T) {
	router, _, _ := newChatRouter(t)
	if v := getChat(t, router, 1); v.Queued != nil {
		t.Fatalf("fresh chat queued = %+v, want nil", v.Queued)
	}
	if rec := chatDo(t, router, http.MethodDelete, "/api/chat/queue?id=whatever", nil, 1); rec.Code != http.StatusNotFound {
		t.Fatalf("cancel with nothing queued = %d %s", rec.Code, rec.Body)
	}
	if rec := chatDo(t, router, http.MethodDelete, "/api/chat/queue", nil, 1); rec.Code != http.StatusBadRequest {
		t.Fatalf("cancel without an id = %d %s", rec.Code, rec.Body)
	}
}

func TestChat_RejectsBadInput(t *testing.T) {
	router, _, _ := newChatRouter(t)
	for _, tc := range []struct {
		body any
		want int
	}{
		{map[string]string{"content": "   "}, http.StatusBadRequest},
		{"not an object", http.StatusBadRequest},
		{map[string]string{"content": strings.Repeat("a", 16*1024+1)}, http.StatusRequestEntityTooLarge},
	} {
		if rec := chatDo(t, router, http.MethodPost, "/api/chat/messages", tc.body, 1); rec.Code != tc.want {
			t.Errorf("body %v = %d, want %d", tc.body, rec.Code, tc.want)
		}
	}
}

// The platform's refusals reach the app as gateway errors, never as the
// app user's own 401/403.
func TestChat_PlatformErrors(t *testing.T) {
	router, platform, _ := newChatRouter(t)
	for _, tc := range []struct {
		status, want int
	}{
		{http.StatusForbidden, http.StatusBadGateway},
		{http.StatusUnauthorized, http.StatusBadGateway},
		{http.StatusServiceUnavailable, http.StatusServiceUnavailable},
	} {
		platform.mu.Lock()
		platform.status = tc.status
		platform.mu.Unlock()
		if rec := chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "hi"}, 1); rec.Code != tc.want {
			t.Errorf("platform %d: send = %d %s, want %d", tc.status, rec.Code, rec.Body, tc.want)
		}
	}
}

// Exhausted login slots are a 503 like a disabled platform, but the app gets
// auth_pool_exhausted so it can say so instead of "unreachable".
func TestChat_AuthPoolExhausted(t *testing.T) {
	router, platform, _ := newChatRouter(t)
	platform.mu.Lock()
	platform.status = http.StatusServiceUnavailable
	platform.statusBody = `{"error":"auth pool exhausted: all 1 slots are held"}`
	platform.mu.Unlock()
	rec := chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "hi"}, 1)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"code":"auth_pool_exhausted"`) {
		t.Fatalf("send = %d %s, want 503 with auth_pool_exhausted", rec.Code, rec.Body)
	}
	platform.mu.Lock()
	platform.statusBody = `{"error":"feature disabled"}`
	platform.mu.Unlock()
	rec = chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "hi"}, 1)
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "auth_pool_exhausted") {
		t.Fatalf("send = %d %s, want a plain 503", rec.Code, rec.Body)
	}
}

func TestChat_RefreshPlugins(t *testing.T) {
	router, platform, engine := newChatRouter(t)
	if rec := chatDo(t, router, http.MethodPost, "/api/chat/refresh-plugins", nil, 1); rec.Code != http.StatusConflict {
		t.Fatalf("refresh before the first message = %d", rec.Code)
	}
	chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "hi"}, 1)
	engine.Wait()
	rec := chatDo(t, router, http.MethodPost, "/api/chat/refresh-plugins", nil, 1)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"updated":["p--s"]`) {
		t.Fatalf("refresh = %d %s", rec.Code, rec.Body)
	}
	platform.mu.Lock()
	platform.refresh = http.StatusConflict
	platform.mu.Unlock()
	if rec := chatDo(t, router, http.MethodPost, "/api/chat/refresh-plugins", nil, 1); rec.Code != http.StatusConflict {
		t.Fatalf("refresh during a turn = %d", rec.Code)
	}
}

// archiveChat taps the chat tab's archive the way the apps do: naming the
// session the person was shown in the confirmation.
func archiveChat(t *testing.T, router http.Handler, sessionID string, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	return chatDo(t, router, http.MethodPost, "/api/chat/archive", map[string]string{"session_id": sessionID}, userID)
}

// PRD-1 AC11: the archive hands back the environment of the one session
// answering the caller's conversation, and says what became of it by reading
// the session back. The conversation and its history stay.
func TestChat_ArchiveTheSessionAnsweringTheConversation(t *testing.T) {
	router, _, engine := newChatRouter(t)
	chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "안녕"}, 1)
	engine.Wait()

	// The apps learn which session to name from the conversation itself.
	session := getChat(t, router, 1).Conversation.SessionID
	if session == "" {
		t.Fatal("the conversation must say which session answers it")
	}

	rec := archiveChat(t, router, session, 1)
	body := rec.Body.String()
	if rec.Code != http.StatusOK {
		t.Fatalf("archive = %d %s", rec.Code, body)
	}
	for _, want := range []string{`"archived":true`, `"state":"snapshot"`, `"session_id":"` + session + `"`, `"createdAt":"2026-11-28T06:00:00Z"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("archive body = %s, want %s", body, want)
		}
	}

	// Nothing of the conversation is thrown away, and it still points at the
	// same session, so the next message restores it (PRD-1 AC4).
	v := getChat(t, router, 1)
	if len(v.Messages) != 2 || v.Conversation.SessionID != session {
		t.Fatalf("chat after an archive = %+v, want the history and the session kept", v)
	}
}

// Each refusal the app explains differently gets its own code.
func TestChat_ArchiveRefusals(t *testing.T) {
	router, platform, engine := newChatRouter(t)

	// Before the first message there is no session to archive.
	rec := archiveChat(t, router, "cx1", 1)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"chat_archive_no_session"`) {
		t.Fatalf("archive before the first message = %d %s", rec.Code, rec.Body)
	}

	chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "hi"}, 1)
	engine.Wait()
	session := getChat(t, router, 1).Conversation.SessionID

	// The server will not archive "whatever is current" on the caller's behalf.
	for _, body := range []any{map[string]string{"session_id": ""}, map[string]string{"session_id": "   "}} {
		rec := chatDo(t, router, http.MethodPost, "/api/chat/archive", body, 1)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"code":"chat_archive_session_unnamed"`) {
			t.Fatalf("archive with %v = %d %s", body, rec.Code, rec.Body)
		}
	}

	// A session that took over between the confirmation and the tap.
	rec = archiveChat(t, router, "cx-stale", 1)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"chat_archive_session_changed"`) {
		t.Fatalf("archive naming a stale session = %d %s", rec.Code, rec.Body)
	}

	// The platform no longer has it: nothing archived and nothing lost.
	platform.mu.Lock()
	platform.snapshot = http.StatusNotFound
	platform.mu.Unlock()
	rec = archiveChat(t, router, session, 1)
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), `"code":"chat_archive_session_gone"`) {
		t.Fatalf("archive of a gone session = %d %s", rec.Code, rec.Body)
	}

	// The archive gate being off is a state to explain, not a failure — and
	// not the platform running out of login slots (PRD-11 AC8·AC11).
	platform.mu.Lock()
	platform.snapshot = http.StatusServiceUnavailable
	platform.snapshotBody = `{"error":"archive disabled"}`
	platform.mu.Unlock()
	rec = archiveChat(t, router, session, 1)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"code":"chat_archive_disabled"`) {
		t.Fatalf("archive with the gate off = %d %s", rec.Code, rec.Body)
	}

	platform.mu.Lock()
	platform.snapshotBody = `{"error":"auth pool exhausted for codex-chat"}`
	platform.mu.Unlock()
	rec = archiveChat(t, router, session, 1)
	if rec.Code != http.StatusServiceUnavailable || !strings.Contains(rec.Body.String(), `"code":"auth_pool_exhausted"`) {
		t.Fatalf("archive with no login slots = %d %s, want auth_pool_exhausted", rec.Code, rec.Body)
	}

	// A body that is not an archive request at all.
	rec = chatDo(t, router, http.MethodPost, "/api/chat/archive", nil, 1)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), `"code":"chat_archive_session_unnamed"`) {
		t.Fatalf("archive with no body = %d %s", rec.Code, rec.Body)
	}
}

// The session is derived from the signed in user's own conversation, so one
// user's tap cannot archive another's session.
func TestChat_ArchiveStaysWithTheSignedInUser(t *testing.T) {
	router, _, engine := newChatRouter(t)
	chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "mine"}, 1)
	engine.Wait()
	chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "theirs"}, 2)
	engine.Wait()
	first := getChat(t, router, 1).Conversation.SessionID
	second := getChat(t, router, 2).Conversation.SessionID
	if first == "" || first == second {
		t.Fatalf("users must not share a session: %q and %q", first, second)
	}

	rec := archiveChat(t, router, first, 2)
	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), `"code":"chat_archive_session_changed"`) {
		t.Fatalf("user 2 archiving user 1's session = %d %s", rec.Code, rec.Body)
	}
	if v := getChat(t, router, 1); len(v.Messages) != 2 || v.Conversation.SessionID != first {
		t.Fatalf("user 1's chat = %+v, want it untouched", v)
	}
}
