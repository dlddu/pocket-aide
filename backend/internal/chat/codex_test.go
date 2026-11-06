package chat_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
)

// fakePlatform is session-platform as the engine sees it: codex sessions
// whose transcript frames each turn with the platform's marker lines.
type fakePlatform struct {
	mu        sync.Mutex
	sessions  map[string]*fakeSession
	created   int
	hold      chan struct{} // when set, replies wait for it
	refreshed int
	refresh   int // status for plugin refresh; 0 = 200
}

type fakeSession struct {
	name, workload string
	transcript     []byte
	turns          int
	prompts        []string
}

func newFakePlatform() *fakePlatform {
	return &fakePlatform{sessions: map[string]*fakeSession{}}
}

func marker(kind string, seq int, extra string) string {
	return fmt.Sprintf("\x1esession-platform %s seq=%d%s\n", kind, seq, extra)
}

func (f *fakePlatform) Do(_ context.Context, method, path string, body []byte) (int, []byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	parts := strings.Split(strings.TrimPrefix(path, "/sessions"), "/")
	switch {
	case method == http.MethodPost && path == "/sessions":
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		f.created++
		id := fmt.Sprintf("cx%d", f.created)
		f.sessions[id] = &fakeSession{name: req["name"], workload: req["workloadType"]}
		raw, _ := json.Marshal(map[string]string{"id": id})
		return http.StatusCreated, raw, nil
	case len(parts) == 3 && parts[2] == "write":
		s, ok := f.sessions[parts[1]]
		if !ok {
			return http.StatusNotFound, nil, nil
		}
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		s.turns++
		seq := s.turns
		s.prompts = append(s.prompts, req["payload"])
		s.transcript = append(s.transcript, marker("input", seq, " at=x")+req["payload"]+"\n"+marker("output", seq, "")...)
		reply := func() {
			f.mu.Lock()
			defer f.mu.Unlock()
			s.transcript = append(s.transcript, "re: "+req["payload"]+"\n"+marker("end", seq, " status=ok at=x")...)
		}
		if f.hold != nil {
			hold := f.hold
			go func() { <-hold; reply() }()
		} else {
			go reply()
		}
		return http.StatusOK, []byte(`{}`), nil
	case len(parts) == 3 && parts[2] == "read":
		s, ok := f.sessions[parts[1]]
		if !ok {
			return http.StatusNotFound, nil, nil
		}
		var req struct {
			Offset int64 `json:"offset"`
		}
		_ = json.Unmarshal(body, &req)
		if req.Offset > int64(len(s.transcript)) {
			req.Offset = int64(len(s.transcript))
		}
		raw, _ := json.Marshal(map[string]any{"payload": string(s.transcript[req.Offset:]), "nextOffset": len(s.transcript)})
		return http.StatusOK, raw, nil
	case len(parts) == 4 && parts[2] == "plugins" && parts[3] == "refresh":
		if _, ok := f.sessions[parts[1]]; !ok {
			return http.StatusNotFound, nil, nil
		}
		f.refreshed++
		if f.refresh != 0 {
			return f.refresh, nil, nil
		}
		return http.StatusOK, []byte(`{"state":"ready","revision":"abc","added":["p--new"],"updated":[],"removed":[]}`), nil
	}
	return http.StatusTeapot, nil, nil
}

func (f *fakePlatform) session(id string) *fakeSession {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sessions[id]
}

func newEngine(t *testing.T) (*chat.Store, *chat.Engine, *fakePlatform) {
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
	platform := newFakePlatform()
	engine := chat.NewEngine(store, platform, nil).WithPoll(5 * time.Millisecond)
	t.Cleanup(engine.Wait)
	return store, engine, platform
}

func messages(t *testing.T, store *chat.Store, userID int64) []chat.Message {
	t.Helper()
	conv, err := store.Conversation(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	out, err := store.Messages(context.Background(), userID, conv.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func transcriptOf(msgs []chat.Message) string {
	var b strings.Builder
	for _, m := range msgs {
		b.WriteString(string(m.Role) + ":" + m.Content + "|")
	}
	return b.String()
}

// PRD-1 AC1·AC4: the first message opens the user's codex session; every
// reply is read out of that one session's transcript, turn by turn.
func TestEngine_OneCodexSessionPerUser(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()

	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	if _, err := engine.Send(ctx, 1, "again"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()

	if got, want := transcriptOf(messages(t, store, 1)), "user:hello|assistant:re: hello|user:again|assistant:re: again|"; got != want {
		t.Fatalf("history = %q, want %q", got, want)
	}
	s := platform.session("cx1")
	if platform.created != 1 || s.workload != "codex" || s.name != "pocket-aide-chat-1" || len(s.prompts) != 2 {
		t.Fatalf("platform = created %d, session %+v", platform.created, s)
	}
	if engine.Pending(1) != nil {
		t.Error("nothing should be pending after the turns ended")
	}

	// Another user gets a session of their own.
	if _, err := engine.Send(ctx, 2, "hi"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	if platform.created != 2 || len(messages(t, store, 1)) != 4 {
		t.Fatalf("users must not share a session: created %d", platform.created)
	}
}

func TestEngine_OneTurnAtATimeAndPendingReply(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}
	if engine.Pending(1) == nil {
		t.Fatal("a sent message must leave a pending reply")
	}
	if _, err := engine.Send(ctx, 1, "second"); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("second send = %v, want ErrBusy", err)
	}
	if _, err := engine.RefreshPlugins(ctx, 1); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("refresh during a reply = %v, want ErrBusy", err)
	}
	close(platform.hold)
	engine.Wait()
	if got := transcriptOf(messages(t, store, 1)); got != "user:slow|assistant:re: slow|" {
		t.Fatalf("history = %q", got)
	}
}

// A platform session that disappeared is replaced by a new one.
func TestEngine_RecreatesAMissingSession(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "one"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	platform.mu.Lock()
	delete(platform.sessions, "cx1")
	platform.mu.Unlock()
	if _, err := engine.Send(ctx, 1, "two"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	if platform.created != 2 {
		t.Fatalf("created = %d, want a new session", platform.created)
	}
	if got := transcriptOf(messages(t, store, 1)); !strings.HasSuffix(got, "user:two|assistant:re: two|") {
		t.Fatalf("history = %q", got)
	}
}

// After a restart mid-reply the engine no longer follows that turn. The next
// send saves the finished reply first; if the turn is still running it is
// picked up again and the new message waits.
func TestEngine_CatchesUpATurnItLostTrackOf(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "first"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()

	// A prompt the engine never followed: as if the server restarted.
	platform.hold = make(chan struct{})
	if status, _, _ := platform.Do(ctx, http.MethodPost, "/sessions/cx1/write", []byte(`{"payload":"lost"}`)); status != http.StatusOK {
		t.Fatal(status)
	}
	restarted := chat.NewEngine(store, platform, nil).WithPoll(5 * time.Millisecond)
	if _, err := restarted.Send(ctx, 1, "next"); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("send while the lost turn runs = %v, want ErrBusy", err)
	}
	close(platform.hold)
	restarted.Wait()
	platform.hold = nil
	if _, err := restarted.Send(ctx, 1, "next"); err != nil {
		t.Fatal(err)
	}
	restarted.Wait()
	want := "user:first|assistant:re: first|assistant:re: lost|user:next|assistant:re: next|"
	if got := transcriptOf(messages(t, store, 1)); got != want {
		t.Fatalf("history = %q, want %q", got, want)
	}
}

func TestEngine_RefreshPlugins(t *testing.T) {
	_, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.RefreshPlugins(ctx, 1); !errors.Is(err, chat.ErrNoConversation) {
		t.Fatalf("refresh before any message = %v", err)
	}
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	got, err := engine.RefreshPlugins(ctx, 1)
	if err != nil || got.State != "ready" || got.Added[0] != "p--new" {
		t.Fatalf("refresh = %+v, %v", got, err)
	}
	platform.refresh = http.StatusConflict
	if _, err := engine.RefreshPlugins(ctx, 1); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("refresh the platform refused = %v", err)
	}
	platform.refresh = http.StatusBadGateway
	var pe *chat.PlatformError
	if _, err := engine.RefreshPlugins(ctx, 1); !errors.As(err, &pe) || pe.Status != http.StatusBadGateway {
		t.Fatalf("failed refresh = %v", err)
	}
}
