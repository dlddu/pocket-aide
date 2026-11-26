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
	// createFails makes a create answer 504 after making the session, as a
	// gateway timing out on a slow create does.
	createFails bool
	writeStatus int // forced status for writes; 0 = normal
	// steerApplies makes /sessions/{id}/steer take the message into the
	// running turn. Off by default: a platform that refuses is the case the
	// queue tests exercise.
	steerApplies bool
	// steerStatus is the refusal status; 0 = 409.
	steerStatus int
	// steered is every mid-turn message the platform was handed, in order.
	steered []string
	// endStatus is the status= field of a turn's end marker. The platform
	// writes ok · timeout · canceled · error there (session-platform
	// claudeTurnStatus); newFakePlatform starts at ok.
	endStatus string
	// noEndStatus writes the end marker with no status= at all, as a platform
	// older than the field did — and as old stored history was framed.
	noEndStatus bool
	// silentReply makes a turn end without having written any output.
	silentReply bool
	// readsOK, when > 0, is how many reads answer 200 before every later one
	// answers 500 — a follow that loses the transcript part way through.
	readsOK int
	reads   int
}

type fakeSession struct {
	name, workload string
	plugins        []string
	createdAt      string
	transcript     []byte
	turns          int
	prompts        []string
}

func newFakePlatform() *fakePlatform {
	return &fakePlatform{sessions: map[string]*fakeSession{}, endStatus: "ok"}
}

// endMarker is how this platform closes a turn: with the status it was told to
// report, or — for noEndStatus — the way a platform with no status field did.
func (f *fakePlatform) endMarker(seq int) string {
	if f.noEndStatus {
		return marker("end", seq, " at=x")
	}
	return marker("end", seq, " status="+f.endStatus+" at=x")
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
		var req struct {
			Name         string   `json:"name"`
			WorkloadType string   `json:"workloadType"`
			Plugins      []string `json:"plugins"`
		}
		_ = json.Unmarshal(body, &req)
		f.created++
		id := fmt.Sprintf("cx%d", f.created)
		f.sessions[id] = &fakeSession{name: req.Name, workload: req.WorkloadType, plugins: req.Plugins, createdAt: fmt.Sprintf("2026-11-06T00:00:%02dZ", f.created)}
		if f.createFails {
			return http.StatusGatewayTimeout, nil, nil
		}
		raw, _ := json.Marshal(map[string]string{"id": id})
		return http.StatusCreated, raw, nil
	case method == http.MethodGet && path == "/sessions":
		var list []map[string]string
		for id, s := range f.sessions {
			list = append(list, map[string]string{"id": id, "name": s.name, "workloadType": s.workload, "createdAt": s.createdAt})
		}
		raw, _ := json.Marshal(map[string]any{"sessions": list})
		return http.StatusOK, raw, nil
	case len(parts) == 3 && parts[2] == "steer":
		if _, ok := f.sessions[parts[1]]; !ok {
			return http.StatusNotFound, nil, nil
		}
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		f.steered = append(f.steered, req["text"])
		if f.steerApplies {
			return http.StatusOK, []byte(`{"status":"applied","turnId":"t1"}`), nil
		}
		status := f.steerStatus
		if status == 0 {
			status = http.StatusConflict
		}
		return status, []byte(`{"status":"rejected","reason":"expectedTurnId does not match the active turn"}`), nil
	case len(parts) == 3 && parts[2] == "write":
		s, ok := f.sessions[parts[1]]
		if !ok {
			return http.StatusNotFound, nil, nil
		}
		if f.writeStatus != 0 {
			return f.writeStatus, nil, nil
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
			out := "re: " + req["payload"] + "\n"
			if f.silentReply {
				out = ""
			}
			s.transcript = append(s.transcript, out+f.endMarker(seq)...)
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
		f.reads++
		if f.readsOK > 0 && f.reads > f.readsOK {
			return http.StatusInternalServerError, nil, nil
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
	if platform.created != 1 || s.workload != "codex-chat" || s.name != "pocket-aide-chat-1" || len(s.prompts) != 2 {
		t.Fatalf("platform = created %d, session %+v", platform.created, s)
	}
	// The platform installs only the plugins the create names (session-platform
	// AC-A5): the chat's skills come from naming codex-session here.
	if len(s.plugins) != 1 || s.plugins[0] != "codex-session" {
		t.Fatalf("create plugins = %v, want [codex-session]", s.plugins)
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

// PRD-1 AC1 (2026-11-09 개정): a message that arrives while a reply is in
// progress waits instead of failing, and is sent right after that reply
// ends. Only one waits at a time.
func TestEngine_QueuesAMessageSentWhileBusyAndDrainsItAfter(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}

	msg, queued, err := engine.Queue(ctx, 1, "second")
	if err != nil || queued == nil || queued.Content != "second" || msg.ID != 0 {
		t.Fatalf("queue while busy = msg %+v, queued %+v, err %v", msg, queued, err)
	}
	if _, _, err := engine.Queue(ctx, 1, "third"); !errors.Is(err, chat.ErrQueueFull) {
		t.Fatalf("queue behind an already-queued message = %v, want ErrQueueFull", err)
	}

	close(platform.hold)
	engine.Wait()

	want := "user:slow|assistant:re: slow|user:second|assistant:re: second|"
	if got := transcriptOf(messages(t, store, 1)); got != want {
		t.Fatalf("history = %q, want %q", got, want)
	}
	s := platform.session("cx1")
	if len(s.prompts) != 2 || s.prompts[1] != "second" {
		t.Fatalf("prompts = %v, want the dequeued message sent after the first turn", s.prompts)
	}
	if engine.Pending(1) != nil {
		t.Error("nothing should be pending once the dequeued turn also ended")
	}
}

// A queued message cancelled before its turn starts is never sent.
func TestEngine_CancelsAQueuedMessage(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}
	_, queued, err := engine.Queue(ctx, 1, "second")
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.CancelQueued(ctx, 1, queued.ID); err != nil {
		t.Fatalf("cancel a queued message = %v", err)
	}
	if err := engine.CancelQueued(ctx, 1, queued.ID); !errors.Is(err, chat.ErrNoQueuedMessage) {
		t.Fatalf("cancel again = %v, want ErrNoQueuedMessage", err)
	}

	close(platform.hold)
	engine.Wait()

	if got := transcriptOf(messages(t, store, 1)); got != "user:slow|assistant:re: slow|" {
		t.Fatalf("a cancelled message must never be sent: history = %q", got)
	}
}

func TestEngine_CancelQueuedWithNothingQueued(t *testing.T) {
	_, engine, _ := newEngine(t)
	ctx := context.Background()
	if err := engine.CancelQueued(ctx, 1, "whatever"); !errors.Is(err, chat.ErrNoQueuedMessage) {
		t.Fatalf("cancel with nothing queued = %v, want ErrNoQueuedMessage", err)
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

// A create the gateway cut off still made the session; the engine adopts it
// instead of creating another, and a send the platform refused stores nothing.
func TestEngine_AdoptsASessionWhoseCreateTimedOut(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	platform.createFails = true
	if _, err := engine.Send(ctx, 1, "hello"); err != nil {
		t.Fatalf("send with a timed-out create = %v", err)
	}
	engine.Wait()
	if platform.created != 1 {
		t.Fatalf("created = %d, want the timed-out session adopted", platform.created)
	}
	if got := transcriptOf(messages(t, store, 1)); got != "user:hello|assistant:re: hello|" {
		t.Fatalf("history = %q", got)
	}

	platform.writeStatus = http.StatusInternalServerError
	if _, err := engine.Send(ctx, 1, "refused"); err == nil {
		t.Fatal("a refused write must fail the send")
	}
	if got := len(messages(t, store, 1)); got != 2 {
		t.Fatalf("a refused send left %d messages, want 2", got)
	}
}

// An orphan from an earlier failed create is adopted read from its end, so
// nothing already in it is taken for the new reply. The platform lists it as
// codex-chat, or as codex for a session made before the rename.
func TestEngine_AdoptsAnOrphanFromItsEnd(t *testing.T) {
	for _, workload := range []string{"codex-chat", "codex"} {
		t.Run(workload, func(t *testing.T) {
			store, engine, platform := newEngine(t)
			ctx := context.Background()
			platform.mu.Lock()
			platform.sessions["old"] = &fakeSession{
				name: "pocket-aide-chat-1", workload: workload, createdAt: "2026-11-01T00:00:00Z",
				transcript: []byte(marker("input", 1, "") + "stale\n" + marker("output", 1, "") + "stale reply\n" + marker("end", 1, " status=ok")),
				turns:      1,
			}
			platform.mu.Unlock()
			if _, err := engine.Send(ctx, 1, "fresh"); err != nil {
				t.Fatal(err)
			}
			engine.Wait()
			if platform.created != 0 {
				t.Fatalf("created = %d, want the orphan adopted", platform.created)
			}
			if got := transcriptOf(messages(t, store, 1)); got != "user:fresh|assistant:re: fresh|" {
				t.Fatalf("history = %q", got)
			}
		})
	}
}

// convID is the user's one conversation, as the store knows it.
func convID(t *testing.T, store *chat.Store, userID int64) int64 {
	t.Helper()
	conv, err := store.Conversation(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	return conv.ID
}

// restart is the engine a restarted server builds: the same store, none of
// the in-memory turn state.
func restart(t *testing.T, store *chat.Store, platform *fakePlatform) *chat.Engine {
	t.Helper()
	engine := chat.NewEngine(store, platform, nil).WithPoll(5 * time.Millisecond)
	t.Cleanup(engine.Wait)
	return engine
}

// A restart kills the turn that would have sent a waiting message, so the row
// is all that is left. The apps' poll resumes it, once (PRD-1 AC1 개정).
func TestEngine_ResumesAQueuedMessageAfterARestart(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "first"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	id := convID(t, store, 1)
	if _, err := store.SetQueued(ctx, 1, id, "q-1", "left over", 5); err != nil {
		t.Fatal(err)
	}

	resumed := restart(t, store, platform)
	resumed.ResumeQueued(ctx, 1, id)
	resumed.ResumeQueued(ctx, 1, id) // a second poll must not send it twice
	resumed.Wait()

	if got, want := transcriptOf(messages(t, store, 1)), "user:first|assistant:re: first|user:left over|assistant:re: left over|"; got != want {
		t.Fatalf("history = %q, want %q", got, want)
	}
	sent := 0
	for _, p := range platform.session("cx1").prompts {
		if p == "left over" {
			sent++
		}
	}
	if sent != 1 {
		t.Fatalf("the resumed message went out %d times, want exactly 1", sent)
	}
	queued, err := store.GetQueued(ctx, 1, id)
	if err != nil || queued != nil {
		t.Fatalf("queued after the resume = %v, %v, want none", queued, err)
	}
}

// Cancel names the message it means: one that arrives after its message went
// out is a 404, and the message that took its place stays queued.
func TestEngine_CancelWithAStaleIDLeavesTheNewerMessage(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}
	_, stale, err := engine.Queue(ctx, 1, "first queued")
	if err != nil {
		t.Fatal(err)
	}
	id := convID(t, store, 1)
	// The turn ended and took it, and the user queued another behind the next.
	if _, err := store.TakeQueued(ctx, 1, id); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetQueued(ctx, 1, id, "q-new", "second queued", 9); err != nil {
		t.Fatal(err)
	}

	if err := engine.CancelQueued(ctx, 1, stale.ID); !errors.Is(err, chat.ErrNoQueuedMessage) {
		t.Fatalf("cancel with a stale id = %v, want ErrNoQueuedMessage", err)
	}
	queued, err := store.GetQueued(ctx, 1, id)
	if err != nil {
		t.Fatal(err)
	}
	if queued == nil || queued.Content != "second queued" {
		t.Fatalf("queued = %v, want the newer message untouched", queued)
	}

	close(platform.hold)
}

// A dispatch the platform refuses must not swallow the message: it goes back
// to waiting, and the retry hold keeps a refusing platform from being polled
// once per second.
func TestEngine_KeepsAQueuedMessageWhenTheDispatchFails(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "first"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	id := convID(t, store, 1)
	if _, err := store.SetQueued(ctx, 1, id, "q-2", "refused", 7); err != nil {
		t.Fatal(err)
	}
	platform.mu.Lock()
	platform.writeStatus = http.StatusInternalServerError
	platform.mu.Unlock()

	resumed := restart(t, store, platform)
	resumed.ResumeQueued(ctx, 1, id)
	resumed.Wait()

	queued, err := store.GetQueued(ctx, 1, id)
	if err != nil {
		t.Fatal(err)
	}
	if queued == nil || queued.Content != "refused" || queued.ID != "q-2" || queued.QueuedAt != 7 {
		t.Fatalf("queued after a refused dispatch = %v, want it still waiting as itself", queued)
	}
	if got := transcriptOf(messages(t, store, 1)); got != "user:first|assistant:re: first|" {
		t.Fatalf("a message that never reached the session must not be in the history: %q", got)
	}
	before := len(platform.session("cx1").prompts)
	resumed.ResumeQueued(ctx, 1, id) // inside the retry hold
	resumed.Wait()
	if got := len(platform.session("cx1").prompts); got != before {
		t.Fatalf("writes = %d, want no retry inside the hold (%d)", got, before)
	}
}

// One user's queue is not another's: a cancel or a resume for user 2 never
// touches user 1's waiting message.
func TestEngine_QueuedMessagesAreIsolatedPerUser(t *testing.T) {
	store, engine, platform := newEngine(t)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "one"); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Send(ctx, 2, "two"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	one, two := convID(t, store, 1), convID(t, store, 2)
	if _, err := store.SetQueued(ctx, 1, one, "q-3", "mine", 11); err != nil {
		t.Fatal(err)
	}

	if err := engine.CancelQueued(ctx, 2, "q-3"); !errors.Is(err, chat.ErrNoQueuedMessage) {
		t.Fatalf("user 2 cancelling user 1's id = %v, want ErrNoQueuedMessage", err)
	}
	engine.ResumeQueued(ctx, 2, two)
	engine.Wait()

	queued, err := store.GetQueued(ctx, 1, one)
	if err != nil {
		t.Fatal(err)
	}
	if queued == nil || queued.Content != "mine" {
		t.Fatalf("user 1's queued message = %v, want it untouched", queued)
	}
	if got := transcriptOf(messages(t, store, 2)); got != "user:two|assistant:re: two|" {
		t.Fatalf("user 2's history = %q, want no trace of user 1's message", got)
	}
	_ = platform
}

// PRD-1 AC1 · session-platform AC-J9: a message sent while a reply is in
// progress goes into that reply when the platform takes it — it reads as a
// sent message, not as one still waiting.
func TestEngine_SteersIntoTheRunningTurn(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	platform.steerApplies = true
	// A test that fails before the reply is released would otherwise deadlock
	// in the engine.Wait() cleanup.
	defer func() {
		if platform.hold != nil {
			close(platform.hold)
		}
	}()
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}

	msg, queued, err := engine.Queue(ctx, 1, "actually, the newer paper")
	if err != nil {
		t.Fatal(err)
	}
	if queued != nil {
		t.Fatalf("queued = %+v, want the running turn to have taken it", queued)
	}
	if msg.Content != "actually, the newer paper" {
		t.Fatalf("message = %+v, want the steered message", msg)
	}
	platform.mu.Lock()
	steered := append([]string(nil), platform.steered...)
	platform.mu.Unlock()
	if len(steered) != 1 || steered[0] != "actually, the newer paper" {
		t.Fatalf("platform was handed %q, want the message once", steered)
	}
	// Steering is not a write: the turn still ran the one prompt it started with.
	if s := platform.session("cx1"); len(s.prompts) != 1 || s.prompts[0] != "slow" {
		t.Fatalf("prompts = %v, want only the original prompt", s.prompts)
	}
	if queuedRow, err := store.GetQueued(ctx, 1, convID(t, store, 1)); err != nil || queuedRow != nil {
		t.Fatalf("queued row = %v, %v, want none", queuedRow, err)
	}

	close(platform.hold)
	platform.hold = nil
	engine.Wait()
	if got := transcriptOf(messages(t, store, 1)); got != "user:slow|user:actually, the newer paper|assistant:re: slow|" {
		t.Fatalf("history = %q, want the steered message in the turn it joined", got)
	}
}

// The turn may end between the busy check and the steer; the platform says so
// and the message falls back to the queue rather than being lost.
func TestEngine_FallsBackToTheQueueWhenSteeringIsRefused(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}

	msg, queued, err := engine.Queue(ctx, 1, "second")
	if err != nil {
		t.Fatal(err)
	}
	if queued == nil || queued.Content != "second" || queued.ID == "" {
		t.Fatalf("queued = %+v, want the refused message waiting with an id", queued)
	}
	if msg.Content != "" {
		t.Fatalf("message = %+v, want nothing sent", msg)
	}

	close(platform.hold)
	engine.Wait()
	// It went out as its own turn once the first one ended.
	if got := transcriptOf(messages(t, store, 1)); got != "user:slow|assistant:re: slow|user:second|assistant:re: second|" {
		t.Fatalf("history = %q", got)
	}
}

// A platform with no mid-turn path at all (an older control plane answering
// 404/422) must not swallow the message either.
func TestEngine_QueuesWhenThePlatformHasNoSteerPath(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.hold = make(chan struct{})
	platform.steerStatus = http.StatusUnprocessableEntity
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "slow"); err != nil {
		t.Fatal(err)
	}
	_, queued, err := engine.Queue(ctx, 1, "second")
	if err != nil || queued == nil {
		t.Fatalf("queue = %+v, %v, want the message waiting", queued, err)
	}
	close(platform.hold)
	engine.Wait()
	if got := transcriptOf(messages(t, store, 1)); got != "user:slow|assistant:re: slow|user:second|assistant:re: second|" {
		t.Fatalf("history = %q", got)
	}
	_ = store
}

// The texts the engine stores for a turn that produced nothing. Repeated here
// rather than exported: a change to either one is a change to what the user
// reads, and this is the test that says so.
const (
	emptyReplyText  = "(응답이 비어 있습니다)"
	brokenReplyText = "응답을 끝까지 받지 못했습니다. 잠시 뒤 다시 시도해 주세요."
)

// seedCodexSession gives user 1 a conversation already bound to a platform
// session whose transcript holds exactly transcript, unread from offset 0 —
// the state a reconnecting or restarted server finds. Its turn counter starts
// past the seeded seqs so a later prompt cannot reuse one of them.
func seedCodexSession(t *testing.T, store *chat.Store, platform *fakePlatform, transcript string) string {
	t.Helper()
	ctx := context.Background()
	conv, err := store.Conversation(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	platform.mu.Lock()
	platform.created++
	id := fmt.Sprintf("cx%d", platform.created)
	platform.sessions[id] = &fakeSession{
		name:       "pocket-aide-chat-1",
		workload:   "codex-chat",
		createdAt:  "2026-11-06T00:00:01Z",
		transcript: []byte(transcript),
		turns:      9,
	}
	platform.mu.Unlock()
	if err := store.SetPlatform(ctx, 1, conv.ID, id, 0); err != nil {
		t.Fatal(err)
	}
	return id
}

func (f *fakePlatform) setReadsOK(n int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.readsOK = n
}

func (f *fakePlatform) appendTranscript(id, text string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sessions[id].transcript = append(f.sessions[id].transcript, text...)
}

func lastReply(t *testing.T, msgs []chat.Message) chat.Message {
	t.Helper()
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == chat.RoleAssistant {
			return msgs[i]
		}
	}
	t.Fatal("no reply stored")
	return chat.Message{}
}

func countReplies(msgs []chat.Message, content string) int {
	n := 0
	for _, m := range msgs {
		if m.Role == chat.RoleAssistant && m.Content == content {
			n++
		}
	}
	return n
}

// A turn session-platform ends with a status other than ok is not an answer.
// The partial output is kept exactly as it came — the interruption is the
// message's own mark, so neither the apps nor the next reader has to guess it
// out of the text.
func TestEngine_InterruptedTurnKeepsItsPartialAnswerAndSaysSo(t *testing.T) {
	for _, status := range []string{"error", "timeout", "canceled", "something-new"} {
		t.Run(status, func(t *testing.T) {
			store, engine, platform := newEngine(t)
			platform.endStatus = status
			if _, err := engine.Send(context.Background(), 1, "hi"); err != nil {
				t.Fatal(err)
			}
			engine.Wait()
			got := lastReply(t, messages(t, store, 1))
			if got.Content != "re: hi" {
				t.Fatalf("reply = %q, want the partial answer as it came", got.Content)
			}
			if !got.Interrupted {
				t.Fatalf("reply for a %q turn is not marked interrupted", status)
			}
		})
	}
}

// An interrupted turn that wrote nothing must not borrow the text of a turn
// that finished empty — that text claims the session answered.
func TestEngine_EmptyFailedTurnIsNotPresentedAsAnAnswer(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.endStatus = "error"
	platform.silentReply = true
	if _, err := engine.Send(context.Background(), 1, "hi"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	got := lastReply(t, messages(t, store, 1))
	if got.Content != brokenReplyText || !got.Interrupted {
		t.Fatalf("empty failed turn = %q interrupted=%v, want %q interrupted", got.Content, got.Interrupted, brokenReplyText)
	}
}

// A turn that ran to its end with no output is still a finished turn: it says
// the answer was empty and is not marked interrupted.
func TestEngine_EmptyFinishedTurnStaysAnEmptyAnswer(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.silentReply = true
	if _, err := engine.Send(context.Background(), 1, "hi"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	got := lastReply(t, messages(t, store, 1))
	if got.Content != emptyReplyText || got.Interrupted {
		t.Fatalf("empty finished turn = %q interrupted=%v, want %q not interrupted", got.Content, got.Interrupted, emptyReplyText)
	}
}

// Turns framed by a platform that wrote no status= — and the history it left
// behind — ran to their end. Reading a missing status as an interruption would
// mark every one of those replies as broken.
func TestEngine_EndMarkerWithoutStatusIsAFinishedTurn(t *testing.T) {
	store, engine, platform := newEngine(t)
	platform.noEndStatus = true
	if _, err := engine.Send(context.Background(), 1, "hi"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	got := lastReply(t, messages(t, store, 1))
	if got.Content != "re: hi" || got.Interrupted {
		t.Fatalf("legacy-marker turn = %q interrupted=%v", got.Content, got.Interrupted)
	}
}

// Turns read out of the transcript after the fact get the same judgement as
// the ones followed live: the status on each turn's own end marker decides.
func TestEngine_CatchUpJudgesEachTurnByItsOwnStatus(t *testing.T) {
	store, engine, platform := newEngine(t)
	seedCodexSession(t, store, platform,
		marker("input", 1, " at=x")+"one\n"+marker("output", 1, "")+"one answer\n"+
			marker("end", 1, " status=ok at=x")+
			marker("input", 2, " at=x")+"two\n"+marker("output", 2, "")+"two answer\n"+
			marker("end", 2, " status=timeout at=x"))
	if _, err := engine.Send(context.Background(), 1, "next"); err != nil {
		t.Fatal(err)
	}
	engine.Wait()
	msgs := messages(t, store, 1)
	var caught []chat.Message
	for _, m := range msgs {
		if m.Role == chat.RoleAssistant && (m.Content == "one answer" || m.Content == "two answer") {
			caught = append(caught, m)
		}
	}
	if len(caught) != 2 {
		t.Fatalf("caught-up replies = %v", transcriptOf(msgs))
	}
	if caught[0].Interrupted {
		t.Fatal("the ok turn was marked interrupted")
	}
	if !caught[1].Interrupted {
		t.Fatal("the timeout turn was not marked interrupted")
	}
}

// A catch-up that finds finished turns ahead of a still-running one has to
// consume them before handing that turn off. Otherwise the hand-off's own
// failure leaves the offset where it was, and the next catch-up stores every
// one of those finished turns a second time.
func TestEngine_CatchUpStoresAFinishedTurnOnceWhenTheHandOffFails(t *testing.T) {
	store, engine, platform := newEngine(t)
	running := marker("input", 2, " at=x") + "two\n" + marker("output", 2, "") + "partial\n"
	id := seedCodexSession(t, store, platform,
		marker("input", 1, " at=x")+"one\n"+marker("output", 1, "")+"one answer\n"+
			marker("end", 1, " status=ok at=x")+running)
	// Only the catch-up's own read succeeds; the handed-off follow loses the
	// transcript on its first one.
	platform.setReadsOK(1)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "next"); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("send while an earlier turn runs = %v, want ErrBusy", err)
	}
	engine.Wait()

	platform.setReadsOK(0)
	if _, err := engine.Send(ctx, 1, "next"); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("second send = %v, want ErrBusy — turn 2 is still running", err)
	}
	// Let the second hand-off finish so nothing is left polling.
	platform.appendTranscript(id, marker("end", 2, " status=error at=x"))
	engine.Wait()

	msgs := messages(t, store, 1)
	if n := countReplies(msgs, "one answer"); n != 1 {
		t.Fatalf("the finished turn was stored %d times: %v", n, transcriptOf(msgs))
	}
	if got := lastReply(t, msgs); got.Content != "partial" || !got.Interrupted {
		t.Fatalf("turn 2 = %q interrupted=%v, want the partial marked interrupted", got.Content, got.Interrupted)
	}
}

// A turn this server gave up on is consumed up to where it stopped reading.
// Reconnecting then does not attribute those same bytes a second time, and the
// offset is still a turn boundary for everything after it.
func TestEngine_AbandonedTurnIsNotStoredAgainAfterReconnect(t *testing.T) {
	store, engine, platform := newEngine(t)
	seedCodexSession(t, store, platform,
		marker("input", 1, " at=x")+"hi\n"+marker("output", 1, "")+"partial answer\n")
	// The catch-up read and the follow's first read land; the follow's next
	// one does not, so it gives up holding a partial answer.
	platform.setReadsOK(2)
	ctx := context.Background()
	if _, err := engine.Send(ctx, 1, "next"); !errors.Is(err, chat.ErrBusy) {
		t.Fatalf("send while the lost turn runs = %v, want ErrBusy", err)
	}
	engine.Wait()
	if got := lastReply(t, messages(t, store, 1)); got.Content != "partial answer" || !got.Interrupted {
		t.Fatalf("abandoned turn = %q interrupted=%v", got.Content, got.Interrupted)
	}

	platform.setReadsOK(0)
	if _, err := engine.Send(ctx, 1, "after"); err != nil {
		t.Fatalf("send after the abandoned turn = %v, want the new turn to go out", err)
	}
	engine.Wait()
	msgs := messages(t, store, 1)
	if n := countReplies(msgs, "partial answer"); n != 1 {
		t.Fatalf("the abandoned turn was stored %d times: %v", n, transcriptOf(msgs))
	}
	if got := lastReply(t, msgs); got.Content != "re: after" || got.Interrupted {
		t.Fatalf("the new turn = %q interrupted=%v", got.Content, got.Interrupted)
	}
}
