package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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

// chatPowerPlatform is session-platform for the chat tab's on/off routes: one
// session per create, with a state the record reports and the restore moves.
// It records which session each restore named, which is how these tests see
// that the route acts on the caller's own session and on no other.
type chatPowerPlatform struct {
	mu       sync.Mutex
	states   map[string]string
	names    map[string]string
	created  int
	switches []string
	// status and body force every call's answer when status is non-zero.
	status int
	body   string
	// switchStatus and switchBody force only the restore's answer.
	switchStatus int
	switchBody   string
	// transcript is each session's output, framed in the platform's turn
	// markers so a sent message's turn finishes and the follow ends.
	transcript map[string]string
	reads      int
}

func (p *chatPowerPlatform) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status != 0 {
		w.WriteHeader(p.status)
		_, _ = w.Write([]byte(p.body))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	parts := strings.Split(strings.TrimPrefix(path, "/sessions"), "/")
	record := func(id string) string {
		return fmt.Sprintf(`{"id":%q,"name":%q,"workloadType":"codex-chat","state":%q,"lastAccess":"2026-11-28T00:00:00Z","checkpoint":{"createdAt":"2026-11-28T01:00:00Z","sizeBytes":2048}}`,
			id, p.names[id], p.states[id])
	}
	switch {
	case path == "/sessions" && r.Method == http.MethodGet:
		_, _ = w.Write([]byte(`{"sessions":[]}`))
	case path == "/sessions":
		p.created++
		id := fmt.Sprintf("cx%d", p.created)
		p.states[id] = "active"
		p.names[id], _ = body["name"].(string)
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, `{"id":%q}`, id)
	case r.Method == http.MethodGet && len(parts) == 2:
		if _, ok := p.states[parts[1]]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(record(parts[1])))
	case len(parts) == 3 && parts[2] == "switch":
		p.switches = append(p.switches, parts[1])
		if p.switchStatus != 0 {
			w.WriteHeader(p.switchStatus)
			_, _ = w.Write([]byte(p.switchBody))
			return
		}
		p.states[parts[1]] = "active"
		_, _ = w.Write([]byte(record(parts[1])))
	case len(parts) == 3 && parts[2] == "write":
		id := parts[1]
		seq := strings.Count(p.transcript[id], "input seq=") + 1
		p.transcript[id] += fmt.Sprintf(
			"\x1esession-platform input seq=%d at=x\n%v\n\x1esession-platform output seq=%d\nre: %v\n\x1esession-platform end seq=%d status=ok at=x\n",
			seq, body["payload"], seq, body["payload"], seq)
		_, _ = w.Write([]byte(`{}`))
	case len(parts) == 3 && parts[2] == "read":
		p.reads++
		text := p.transcript[parts[1]]
		offset, _ := body["offset"].(float64)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"payload": text[min(int(offset), len(text)):], "nextOffset": len(text),
		})
	default:
		w.WriteHeader(http.StatusTeapot)
	}
}

func (p *chatPowerPlatform) archive(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.states[id] = "snapshot"
}

func (p *chatPowerPlatform) forget(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.states, id)
}

func (p *chatPowerPlatform) switchedSessions() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string{}, p.switches...)
}

// sessionOf is the session the platform holds for one user's chat, by the name
// the backend creates it under.
func (p *chatPowerPlatform) sessionOf(t *testing.T, userID int64) string {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()
	want := fmt.Sprintf("pocket-aide-chat-%d", userID)
	for id, name := range p.names {
		if name == want {
			return id
		}
	}
	t.Fatalf("the platform has no session named %q", want)
	return ""
}

func newChatPowerRouter(t *testing.T) (http.Handler, *chatPowerPlatform) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	platform := &chatPowerPlatform{
		states: map[string]string{}, names: map[string]string{}, transcript: map[string]string{},
	}
	srv := httptest.NewServer(platform)
	t.Cleanup(srv.Close)
	store := chat.New(conn)
	engine := chat.NewEngine(store, sessions.NewClient(srv.URL), nil).WithPoll(5 * time.Millisecond)
	t.Cleanup(engine.Wait)
	r := chi.NewRouter()
	r.Post("/api/chat/messages", handlers.SendChatMessage(engine))
	r.Get("/api/chat/session", handlers.GetChatSession(engine))
	r.Post("/api/chat/session/power", handlers.PowerOnChatSession(engine))
	return r, platform
}

type chatPowerView struct {
	Power      string `json:"power"`
	State      string `json:"state"`
	LastAccess string `json:"last_access"`
	Restored   bool   `json:"restored"`
	Checkpoint *struct {
		CreatedAt string `json:"createdAt"`
		SizeBytes *int64 `json:"sizeBytes"`
	} `json:"checkpoint"`
	Code string `json:"code"`
}

func chatPower(t *testing.T, router http.Handler, method, target string, userID int64, wantStatus int) chatPowerView {
	t.Helper()
	var body any
	if method == http.MethodPost {
		body = map[string]any{}
	}
	rec := chatDo(t, router, method, target, body, userID)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s = %d %s, want %d", method, target, rec.Code, rec.Body, wantStatus)
	}
	var out chatPowerView
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return out
}

// openChatSession gives the user's conversation a session the way the app does
// — by sending a message — and returns the session the platform made for it.
func openChatSession(t *testing.T, router http.Handler, platform *chatPowerPlatform, userID int64) string {
	t.Helper()
	if rec := chatDo(t, router, http.MethodPost, "/api/chat/messages",
		map[string]string{"content": "안녕"}, userID); rec.Code != http.StatusAccepted {
		t.Fatalf("send = %d %s", rec.Code, rec.Body)
	}
	return platform.sessionOf(t, userID)
}

// PRD-1 AC12: the tab asks one route whether its conversation's session is on,
// and the answer carries the detail shown beside it.
func TestChat_SessionPowerShowsWhetherTheSessionIsOn(t *testing.T) {
	router, platform := newChatPowerRouter(t)
	if power := chatPower(t, router, http.MethodGet, "/api/chat/session", 1, http.StatusOK); power.Power != "none" {
		t.Fatalf("before the first message: %+v, want power=none", power)
	}
	id := openChatSession(t, router, platform, 1)

	power := chatPower(t, router, http.MethodGet, "/api/chat/session", 1, http.StatusOK)
	if power.Power != "on" || power.State != "active" || power.LastAccess == "" {
		t.Fatalf("power = %+v, want on/active with a last access", power)
	}

	platform.archive(id)
	power = chatPower(t, router, http.MethodGet, "/api/chat/session", 1, http.StatusOK)
	if power.Power != "off" || power.State != "snapshot" {
		t.Fatalf("power = %+v, want off/snapshot", power)
	}
	if power.Checkpoint == nil || power.Checkpoint.CreatedAt == "" {
		t.Fatalf("power = %+v, want the archive time the tab shows", power)
	}

	platform.forget(id)
	if power := chatPower(t, router, http.MethodGet, "/api/chat/session", 1, http.StatusOK); power.Power != "gone" {
		t.Fatalf("power = %+v, want power=gone", power)
	}
}

// PRD-1 AC12 · PRD-11 AC9: the button restores that one session and says it
// did; a session already on is answered without a restore.
func TestChat_PowerOnRestoresTheConversationsSession(t *testing.T) {
	router, platform := newChatPowerRouter(t)
	id := openChatSession(t, router, platform, 1)
	platform.archive(id)

	power := chatPower(t, router, http.MethodPost, "/api/chat/session/power", 1, http.StatusOK)
	if power.Power != "on" || !power.Restored {
		t.Fatalf("power = %+v, want on with restored=true", power)
	}
	if got := platform.switchedSessions(); len(got) != 1 || got[0] != id {
		t.Fatalf("restored %v, want exactly [%s]", got, id)
	}

	power = chatPower(t, router, http.MethodPost, "/api/chat/session/power", 1, http.StatusOK)
	if power.Power != "on" || power.Restored {
		t.Fatalf("power = %+v, want on with restored=false", power)
	}
	if got := platform.switchedSessions(); len(got) != 1 {
		t.Fatalf("restores sent %d, want the live session left alone", len(got))
	}
}

// PRD-1 AC12: the refusals the app explains differently each have their own
// status and code.
func TestChat_PowerOnRefusals(t *testing.T) {
	t.Run("no session yet", func(t *testing.T) {
		router, _ := newChatPowerRouter(t)
		power := chatPower(t, router, http.MethodPost, "/api/chat/session/power", 1, http.StatusConflict)
		if power.Code != "chat_power_no_session" {
			t.Fatalf("code = %q", power.Code)
		}
	})
	t.Run("session gone", func(t *testing.T) {
		router, platform := newChatPowerRouter(t)
		platform.forget(openChatSession(t, router, platform, 1))
		power := chatPower(t, router, http.MethodPost, "/api/chat/session/power", 1, http.StatusNotFound)
		if power.Code != "chat_power_session_gone" {
			t.Fatalf("code = %q", power.Code)
		}
	})
	t.Run("platform busy with the session", func(t *testing.T) {
		router, platform := newChatPowerRouter(t)
		platform.archive(openChatSession(t, router, platform, 1))
		platform.switchStatus = http.StatusConflict
		power := chatPower(t, router, http.MethodPost, "/api/chat/session/power", 1, http.StatusConflict)
		if power.Code != "chat_power_busy" {
			t.Fatalf("code = %q", power.Code)
		}
	})
	t.Run("no login slot free", func(t *testing.T) {
		router, platform := newChatPowerRouter(t)
		platform.archive(openChatSession(t, router, platform, 1))
		platform.switchStatus = http.StatusServiceUnavailable
		platform.switchBody = `{"error":"auth pool exhausted for codex-chat"}`
		power := chatPower(t, router, http.MethodPost, "/api/chat/session/power", 1, http.StatusServiceUnavailable)
		if power.Code != sessions.ErrorCodeAuthPoolExhausted {
			t.Fatalf("code = %q, want %q", power.Code, sessions.ErrorCodeAuthPoolExhausted)
		}
	})
	t.Run("platform unreachable", func(t *testing.T) {
		router, platform := newChatPowerRouter(t)
		platform.archive(openChatSession(t, router, platform, 1))
		platform.switchStatus = http.StatusServiceUnavailable
		platform.switchBody = `{"error":"control plane is down"}`
		// A restore has no feature gate of its own, so this must not borrow the
		// sessions tab's "archiving is disabled here" wording (PRD-11 AC11).
		rec := chatDo(t, router, http.MethodPost, "/api/chat/session/power", map[string]any{}, 1)
		if rec.Code != http.StatusBadGateway || !strings.Contains(rec.Body.String(), "unreachable") {
			t.Fatalf("= %d %s, want 502 unreachable", rec.Code, rec.Body)
		}
	})
	t.Run("status read refused", func(t *testing.T) {
		router, platform := newChatPowerRouter(t)
		openChatSession(t, router, platform, 1)
		platform.status = http.StatusInternalServerError
		if rec := chatDo(t, router, http.MethodGet, "/api/chat/session", nil, 1); rec.Code == http.StatusOK {
			t.Fatalf("a refused status read answered 200: %s", rec.Body)
		}
	})
}

// PRD-1 AC12: the session comes from the signed in user's own conversation, so
// one user's button never reaches another user's session.
func TestChat_PowerOnStaysWithTheSignedInUser(t *testing.T) {
	router, platform := newChatPowerRouter(t)
	mine := openChatSession(t, router, platform, 1)
	theirs := openChatSession(t, router, platform, 2)
	platform.archive(mine)
	platform.archive(theirs)

	chatPower(t, router, http.MethodPost, "/api/chat/session/power", 2, http.StatusOK)
	if got := platform.switchedSessions(); len(got) != 1 || got[0] != theirs {
		t.Fatalf("restored %v, want only user 2's session %s", got, theirs)
	}
	if power := chatPower(t, router, http.MethodGet, "/api/chat/session", 1, http.StatusOK); power.Power != "off" {
		t.Fatalf("user 1's session is %+v, want still off", power)
	}
}
