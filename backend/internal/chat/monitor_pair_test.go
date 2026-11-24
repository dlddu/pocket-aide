package chat_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
)

// pairedPlatform is session-platform as the backend sees it when a chat
// session has a monitor: the tenant's session list carries the monitor's
// `monitorFor`, and the work routes answer for whichever session was asked
// for. It records every path so a test can show which session the panel read.
type pairedPlatform struct {
	// monitorOf is the pair the platform's backend-facing route reports:
	// parent session id -> its monitor.
	monitorOf map[string]string
	stateOf   map[string]string
	// frozen is the monitor that is snapshotted: its reads come from what it
	// published, and intake is held.
	frozen map[string]bool
	paths  []string
	bodies []string
	sends  int
}

func (p *pairedPlatform) Do(_ context.Context, method, path string, body []byte) (int, []byte, error) {
	p.paths = append(p.paths, method+" "+path)
	p.bodies = append(p.bodies, string(body))
	id := strings.TrimPrefix(path, "/sessions/")
	id, rest, _ := strings.Cut(id, "/")
	paired := strings.HasPrefix(rest, "monitor/")
	// The pair lookup: 404 for a chat session nobody monitors, which is how an
	// unpaired deployment reads.
	if strings.HasSuffix(path, "/monitor/pair") {
		mon, ok := p.monitorOf[id]
		if !ok {
			return http.StatusNotFound, []byte("this session has no monitor"), nil
		}
		intake := "available"
		if p.frozen[mon] {
			intake = "refused"
		}
		raw, _ := json.Marshal(map[string]any{
			"role":    "parent",
			"binding": map[string]any{"parentSessionId": id, "monitorSessionId": mon},
			"monitor": map[string]any{"state": p.stateOf[mon], "reads": "available", "intake": intake},
		})
		return http.StatusOK, raw, nil
	}
	// An unpaired read of a frozen session is what the old contract does.
	if !paired && p.frozen[id] {
		return http.StatusServiceUnavailable,
			[]byte(`{"error":"the session has no live pod to read its work records through"}`), nil
	}
	frozenMonitor := paired && p.frozen[p.monitorOf[id]]
	switch {
	case strings.HasSuffix(path, "/requests"):
		p.sends++
		if frozenMonitor {
			// Durably accepted, not delivered: the platform's 202.
			return http.StatusAccepted, []byte(`{"item":{"requestId":"req-app-2","state":"accepted"},"replayed":false,` +
				`"note":"accepted and held"}`), nil
		}
		return http.StatusOK, []byte(`{"item":{"requestId":"req-app-1","state":"forwarded","turnId":9},` +
			`"turn":{"id":9,"task_id":"task-1","requested_at":"2026-11-22T00:00:00Z"},"replayed":false}`), nil
	case strings.Contains(path, "/turns"):
		src := "slot"
		extra := ""
		if frozenMonitor {
			src, extra = "published", `,"asOf":"2026-11-22T01:00:00Z"`
		}
		return http.StatusOK, []byte(`{"task_id":"task-1","task":{"id":"task-1","title":"t","status":"running","version":1},` +
			`"turns":[{"id":9,"task_id":"task-1","response":"반영 끝","requested_at":"2026-11-22T00:00:00Z"}],` +
			`"source":"` + src + `"` + extra + `}`), nil
	case strings.Contains(path, "/work/tasks"):
		src := "slot"
		extra := ""
		if frozenMonitor {
			src, extra = "published", `,"asOf":"2026-11-22T01:00:00Z"`
		}
		return http.StatusOK, []byte(`{"tasks":[{"task":{"id":"task-1","title":"모니터 작업","status":"running","version":2},` +
			`"turns_total":1}],"next_cursor":"","source":"` + src + `"` + extra + `}`), nil
	}
	return http.StatusNotFound, nil, nil
}

func pairedEngine(t *testing.T, p *pairedPlatform, chatSession string) (*chat.Store, *chat.Engine) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := chat.New(conn)
	engine := chat.NewEngine(store, p, nil)
	t.Cleanup(engine.Wait)
	bindSession(t, store, 1, chatSession)
	return store, engine
}

// The panel of a paired chat session reads the monitor's work, not its own:
// the monitor owns the records.
func TestPanelReadsTheMonitorsWorkWhenPaired(t *testing.T) {
	p := &pairedPlatform{
		monitorOf: map[string]string{"chat-1": "monitor-1", "chat-2": "monitor-2"},
		stateOf:   map[string]string{"monitor-1": "active", "monitor-2": "active"},
	}
	_, engine := pairedEngine(t, p, "chat-1")

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if list.Monitor == nil || list.Monitor.SessionID != "monitor-1" {
		t.Fatalf("panel did not resolve the monitor: %+v", list.Monitor)
	}
	if len(list.Tasks) != 1 || list.Tasks[0].ID != "task-1" {
		t.Fatalf("tasks = %+v", list.Tasks)
	}
	var read string
	for _, path := range p.paths {
		if strings.Contains(path, "/work/tasks") {
			read = path
		}
	}
	if !strings.Contains(read, "/sessions/chat-1/monitor/work/tasks") {
		t.Fatalf("the panel read %q, want the pair surface of its own parent session", read)
	}
	if strings.Contains(read, "chat-2") || strings.Contains(read, "monitor-2") {
		t.Fatalf("the panel reached another pair: %q", read)
	}

	// The turn history comes from the same session.
	turns, err := engine.TaskTurns(context.Background(), 1, "task-1", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(turns.Turns) != 1 || turns.Turns[0].Response == nil || *turns.Turns[0].Response != "반영 끝" {
		t.Fatalf("turns = %+v", turns.Turns)
	}
}

// An instruction from the panel goes to the monitor under the app's own stable
// request id, and the task is not changed here.
func TestInstructionGoesToTheMonitorWithItsStableID(t *testing.T) {
	p := &pairedPlatform{
		monitorOf: map[string]string{"chat-1": "monitor-1"},
		stateOf:   map[string]string{"monitor-1": "idle"},
	}
	_, engine := pairedEngine(t, p, "chat-1")

	out, err := engine.CreateTaskRequest(context.Background(), 1, "task-1", "req-app-1", "코멘트 반영해 줘", 0)
	if err != nil {
		t.Fatal(err)
	}
	if out.Turn.ID != 9 {
		t.Fatalf("stored turn = %+v", out.Turn)
	}
	var sent, body string
	for i, path := range p.paths {
		if strings.HasSuffix(path, "/requests") {
			sent, body = path, p.bodies[i]
		}
	}
	if !strings.Contains(sent, "/sessions/chat-1/monitor/work/tasks/task-1/requests") {
		t.Fatalf("instruction went to %q, want the pair surface", sent)
	}
	if !strings.Contains(body, `"id":"req-app-1"`) {
		t.Fatalf("the stable id did not reach the platform: %s", body)
	}
	if p.sends != 1 {
		t.Fatalf("sends = %d, want one delegation", p.sends)
	}
}

// A frozen monitor is not an empty panel and not a silent success: reads and
// sends both say the monitor is not running, so the app can show the state and
// retry with the same id later.
func TestFrozenMonitorIsReportedNotFaked(t *testing.T) {
	p := &pairedPlatform{
		monitorOf: map[string]string{"chat-1": "monitor-1"},
		stateOf:   map[string]string{"monitor-1": "snapshot"},
		frozen:    map[string]bool{"monitor-1": true},
	}
	_, engine := pairedEngine(t, p, "chat-1")

	// The first visit is after the freeze: the panel still opens, from what
	// the monitor published, and says when that was.
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatalf("a frozen monitor must still serve the stored panel: %v", err)
	}
	if len(list.Tasks) != 1 || list.Source != "published" || list.AsOf == "" {
		t.Fatalf("frozen panel = %+v (source %q asOf %q)", list.Tasks, list.Source, list.AsOf)
	}
	if list.Monitor == nil || list.Monitor.IntakeLive {
		t.Fatalf("the panel must say intake is held: %+v", list.Monitor)
	}
	turns, err := engine.TaskTurns(context.Background(), 1, "task-1", 0, 0)
	if err != nil {
		t.Fatalf("frozen history: %v", err)
	}
	if len(turns.Turns) != 1 {
		t.Fatalf("frozen history = %+v", turns.Turns)
	}
	// Sending is durable: accepted, not done, and not a failure to retry blindly.
	stored, err := engine.CreateTaskRequest(context.Background(), 1, "task-1", "req-app-2", "나중에", 0)
	if err != nil {
		t.Fatalf("a frozen monitor must still accept the instruction durably: %v", err)
	}
	if !stored.Accepted || stored.State != "accepted" {
		t.Fatalf("held instruction = %+v, want accepted", stored)
	}
	if p.sends != 1 {
		t.Fatalf("sends = %d, want the one durable submit", p.sends)
	}
}

// An unpaired chat session keeps reading and writing its own records, which is
// what every existing deployment does.
func TestUnpairedChatSessionIsUnchanged(t *testing.T) {
	p := &pairedPlatform{monitorOf: map[string]string{}, stateOf: map[string]string{}}
	_, engine := pairedEngine(t, p, "chat-1")

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if list.Monitor != nil && list.Monitor.Paired() {
		t.Fatalf("an unpaired session reported a monitor: %+v", list.Monitor)
	}
	var read string
	for _, path := range p.paths {
		if strings.Contains(path, "/work/tasks") {
			read = path
		}
	}
	if !strings.Contains(read, "/sessions/chat-1/work/tasks") || strings.Contains(read, "/monitor/") {
		t.Fatalf("an unpaired panel read %q, want its own session's work route", read)
	}
}

// pairLookup is a platform that answers the pair lookup with one fixed status
// and body, so each branch of the contract can be checked on its own.
type pairLookup struct {
	status int
	body   string
	err    error
	calls  []string
}

func (p *pairLookup) Do(_ context.Context, method, path string, _ []byte) (int, []byte, error) {
	p.calls = append(p.calls, method+" "+path)
	if strings.HasSuffix(path, "/monitor/pair") {
		return p.status, []byte(p.body), p.err
	}
	// Any work read in these tests would mean the panel fell back to the
	// parent's own records, which is exactly what must not happen.
	return http.StatusOK, []byte(`{"tasks":[],"next_cursor":""}`), nil
}

func lookupEngine(t *testing.T, p *pairLookup) *chat.Engine {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := chat.New(conn)
	engine := chat.NewEngine(store, p, nil)
	t.Cleanup(engine.Wait)
	bindSession(t, store, 1, "chat-1")
	return engine
}

// Only a clear answer may be read as unpaired, because unpaired is the path
// that reads and writes the parent's own records.
func TestPairLookupTreatsOnlyClearAnswersAsLegacy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"no monitor for this session", http.StatusNotFound, "this session has no monitor"},
		{"control plane without the route", http.StatusNotImplemented, ""},
		{"deployment says the feature is off", http.StatusServiceUnavailable,
			"the managed monitor pair is not enabled on this deployment"},
		{"a 200 that names no monitor", http.StatusOK, `{"role":"parent","binding":{}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &pairLookup{status: tc.status, body: tc.body}
			engine := lookupEngine(t, p)
			list, err := engine.Tasks(context.Background(), 1)
			if err != nil {
				t.Fatalf("a clear answer must read as unpaired: %v", err)
			}
			if list.Monitor != nil && list.Monitor.Paired() {
				t.Fatalf("unexpected pair: %+v", list.Monitor)
			}
			var read string
			for _, c := range p.calls {
				if strings.Contains(c, "/work/tasks") {
					read = c
				}
			}
			if !strings.Contains(read, "/sessions/chat-1/work/tasks") {
				t.Fatalf("legacy read = %q, want its own session", read)
			}
		})
	}
}

// An ambiguous or refused lookup is an error. It must not become "unpaired",
// because that hands the parent its own write path back.
func TestPairLookupRefusesToGuessOnAmbiguityOrRefusal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		err    error
		want   error
	}{
		{"unauthenticated", http.StatusUnauthorized, "no", nil, chat.ErrPairRefused},
		{"forbidden — another tenant", http.StatusForbidden, "no", nil, chat.ErrPairRefused},
		{"platform error", http.StatusInternalServerError, "boom", nil, chat.ErrPairUnknown},
		{"gateway timeout", http.StatusGatewayTimeout, "", nil, chat.ErrPairUnknown},
		{"service unavailable without the feature statement", http.StatusServiceUnavailable, "upstream down", nil, chat.ErrPairUnknown},
		{"transport failure", 0, "", errors.New("dial tcp: i/o timeout"), chat.ErrPairUnknown},
		{"a 200 we cannot read", http.StatusOK, "{not json", nil, chat.ErrPairUnknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &pairLookup{status: tc.status, body: tc.body, err: tc.err}
			engine := lookupEngine(t, p)
			if _, err := engine.Tasks(context.Background(), 1); !errors.Is(err, tc.want) {
				t.Fatalf("Tasks error = %v, want %v", err, tc.want)
			}
			// Neither the list nor a send may have reached the parent's own
			// work records.
			for _, c := range p.calls {
				if strings.Contains(c, "/work/tasks") {
					t.Fatalf("an unclear lookup still read work records: %q", c)
				}
			}
			if _, err := engine.CreateTaskRequest(context.Background(), 1, "task-1", "req-x", "하지 마", 0); !errors.Is(err, tc.want) {
				t.Fatalf("CreateTaskRequest error = %v, want %v", err, tc.want)
			}
			for _, c := range p.calls {
				if strings.Contains(c, "/requests") {
					t.Fatalf("an unclear lookup still wrote to the parent's records: %q", c)
				}
			}
		})
	}
}

// setupPlatform is the platform as the opt-in sees it: a session list, a
// create that records `monitorFor`, the pair lookup that then reports it, and
// writes it counts so a test can show the bootstrap turn really started.
type setupPlatform struct {
	sessions []map[string]any
	creates  []string
	writes   []string
	nextID   int
	// deletes counts rollbacks, and offReply lets a test make the platform
	// answer a half-done one.
	deletes  []string
	offReply int
}

func (p *setupPlatform) Do(_ context.Context, method, path string, body []byte) (int, []byte, error) {
	switch {
	case method == http.MethodGet && path == "/sessions":
		raw, _ := json.Marshal(map[string]any{"sessions": p.sessions})
		return http.StatusOK, raw, nil
	case method == http.MethodPost && path == "/sessions":
		p.creates = append(p.creates, string(body))
		var in map[string]any
		_ = json.Unmarshal(body, &in)
		p.nextID++
		id := "created-" + string(rune('0'+p.nextID))
		p.sessions = append(p.sessions, map[string]any{
			"id": id, "name": in["name"], "state": "active", "monitorFor": in["monitorFor"],
		})
		raw, _ := json.Marshal(map[string]any{"id": id, "state": "active"})
		return http.StatusCreated, raw, nil
	case method == http.MethodPost && strings.HasSuffix(path, "/write"):
		p.writes = append(p.writes, path+" "+string(body))
		return http.StatusOK, []byte(`{}`), nil
	case method == http.MethodDelete && strings.HasSuffix(path, "/monitor/pair"):
		p.deletes = append(p.deletes, path)
		parent := strings.TrimPrefix(path, "/sessions/")
		parent, _, _ = strings.Cut(parent, "/")
		if p.offReply == http.StatusForbidden || p.offReply == http.StatusInternalServerError {
			return p.offReply, []byte("refused"), nil
		}
		if p.offReply == http.StatusAccepted {
			// Half done: the monitor is closed, the records are still sealed.
			raw, _ := json.Marshal(map[string]any{"unbound": true,
				"note": "the records are still sealed to it"})
			return http.StatusAccepted, raw, nil
		}
		held := ""
		kept := p.sessions[:0]
		for _, s := range p.sessions {
			if s["monitorFor"] == parent {
				held, _ = s["id"].(string)
				// The platform unbinds the pair; the monitor session itself is
				// left alone, and so is every record.
				s = map[string]any{"id": s["id"], "name": s["name"], "state": s["state"]}
			}
			kept = append(kept, s)
		}
		p.sessions = kept
		if held == "" {
			return http.StatusNotFound, []byte("this session has no monitor"), nil
		}
		raw, _ := json.Marshal(map[string]any{"unbound": true,
			"writer": map[string]any{"writer": "session", "holder": held}})
		return http.StatusOK, raw, nil
	case strings.HasSuffix(path, "/monitor/pair"):
		parent := strings.TrimPrefix(path, "/sessions/")
		parent, _, _ = strings.Cut(parent, "/")
		for _, s := range p.sessions {
			if s["monitorFor"] == parent {
				raw, _ := json.Marshal(map[string]any{
					"role":    "parent",
					"binding": map[string]any{"parentSessionId": parent, "monitorSessionId": s["id"]},
					"monitor": map[string]any{"state": s["state"], "reads": "available", "intake": "available"},
				})
				return http.StatusOK, raw, nil
			}
		}
		return http.StatusNotFound, []byte("this session has no monitor"), nil
	}
	return http.StatusNotFound, nil, nil
}

func setupEngine(t *testing.T, p *setupPlatform, chatSession string) *chat.Engine {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := chat.New(conn)
	engine := chat.NewEngine(store, p, nil)
	t.Cleanup(engine.Wait)
	if chatSession != "" {
		bindSession(t, store, 1, chatSession)
	}
	return engine
}

// Turning the monitor on creates exactly one session, with the monitor bundle
// and this conversation's parent, and starts its first turn. Asking again
// reuses it.
func TestEnableMonitorCreatesOnceWithTheBundleAndBootstraps(t *testing.T) {
	p := &setupPlatform{sessions: []map[string]any{{"id": "chat-1", "name": "chat", "state": "active"}}}
	engine := setupEngine(t, p, "chat-1")

	first, err := engine.EnableMonitor(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Created || first.SessionID == "" {
		t.Fatalf("first enable = %+v, want a created monitor", first)
	}
	if !first.Bootstrapped {
		t.Fatalf("the monitor's first turn did not start: %+v", first)
	}
	if len(p.creates) != 1 {
		t.Fatalf("creates = %v, want one", p.creates)
	}
	// The bundle and the parent are what make it a monitor of this chat
	// session; the platform records the pair from monitorFor.
	if !strings.Contains(p.creates[0], `"plugins":["monitor-session"]`) {
		t.Fatalf("create body has the wrong bundle: %s", p.creates[0])
	}
	if !strings.Contains(p.creates[0], `"monitorFor":"chat-1"`) {
		t.Fatalf("create body does not bind the parent: %s", p.creates[0])
	}
	if len(p.writes) != 1 || !strings.Contains(p.writes[0], "/sessions/"+first.SessionID+"/write") {
		t.Fatalf("bootstrap write = %v", p.writes)
	}

	// Again: the pair already exists, so nothing is created and no second
	// writer appears.
	again, err := engine.EnableMonitor(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if again.Created || again.SessionID != first.SessionID {
		t.Fatalf("second enable = %+v, want the same monitor reused", again)
	}
	if len(p.creates) != 1 {
		t.Fatalf("creates after reuse = %v", p.creates)
	}
}

// Reading the status creates nothing and does not wake the monitor.
func TestMonitorStatusIsAReadOnly(t *testing.T) {
	p := &setupPlatform{sessions: []map[string]any{
		{"id": "chat-1", "name": "chat", "state": "active"},
		{"id": "mon-1", "name": "chat-monitor", "state": "snapshot", "monitorFor": "chat-1"},
	}}
	engine := setupEngine(t, p, "chat-1")
	got, err := engine.MonitorStatus(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.SessionID != "mon-1" || got.Created {
		t.Fatalf("status = %+v", got)
	}
	if len(p.creates) != 0 || len(p.writes) != 0 {
		t.Fatalf("a status read touched the platform: creates=%v writes=%v", p.creates, p.writes)
	}
}

// A monitor left over from a lost create reply is adopted, not duplicated —
// and only when the platform's own record says it monitors *this* parent.
func TestEnableMonitorAdoptsItsOwnAndNeverAnothersParent(t *testing.T) {
	p := &setupPlatform{sessions: []map[string]any{
		{"id": "chat-1", "name": "chat", "state": "active"},
		// Someone else's monitor, same tenant, same naming scheme.
		{"id": "mon-other", "name": "chat-monitor", "state": "active", "monitorFor": "chat-2"},
	}}
	engine := setupEngine(t, p, "chat-1")
	got, err := engine.EnableMonitor(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if got.SessionID == "mon-other" {
		t.Fatal("adopted another parent's monitor")
	}
	if !got.Created || len(p.creates) != 1 {
		t.Fatalf("want one fresh create, got %+v / %v", got, p.creates)
	}
}

// Without a chat session there is nothing to monitor, and the opt-in says so
// instead of creating a stray session.
func TestEnableMonitorNeedsAConversation(t *testing.T) {
	p := &setupPlatform{}
	engine := setupEngine(t, p, "")
	if _, err := engine.EnableMonitor(context.Background(), 1); !errors.Is(err, chat.ErrNoConversation) {
		t.Fatalf("enable without a chat session = %v", err)
	}
	if len(p.creates) != 0 {
		t.Fatalf("creates = %v, want none", p.creates)
	}
}

// Turning the monitor off from the normal path: the platform runs the rollback
// in order and this reports which half happened. Nothing is deleted — the work
// records were never moved, so the conversation keeps the same history — and
// repeating the call is safe.
func TestDisableMonitorRollsBackAndIsSafeToRepeat(t *testing.T) {
	p := &setupPlatform{sessions: []map[string]any{{"id": "chat-1", "name": "chat", "state": "active"}}}
	engine := setupEngine(t, p, "chat-1")
	ctx := t.Context()
	on, err := engine.EnableMonitor(ctx, 1)
	if err != nil || on.SessionID == "" {
		t.Fatalf("enable = %+v %v", on, err)
	}

	off, err := engine.DisableMonitor(ctx, 1)
	if err != nil {
		t.Fatalf("disable: %v", err)
	}
	if !off.Off || off.Retry {
		t.Fatalf("disable = %+v, want it off", off)
	}
	if off.Was != on.SessionID {
		t.Fatalf("the person was not told which monitor was turned off: %+v", off)
	}
	if len(p.deletes) != 1 {
		t.Fatalf("rollbacks = %v, want exactly one", p.deletes)
	}
	// The conversation reads its own work records again, and the monitor
	// session itself was not deleted by this.
	status, err := engine.MonitorStatus(ctx, 1)
	if err != nil || status.SessionID != "" {
		t.Fatalf("status after the rollback = %+v %v", status, err)
	}
	found := false
	for _, s := range p.sessions {
		if s["id"] == on.SessionID {
			found = true
		}
	}
	if !found {
		t.Fatalf("the rollback deleted the monitor session")
	}

	// Again, with nothing bound: not an error, and no second rollback claim.
	again, err := engine.DisableMonitor(ctx, 1)
	if err != nil || !again.Off {
		t.Fatalf("repeat = %+v %v", again, err)
	}

	// And turning it back on afterwards works, which is what makes this a
	// rollback rather than a one-way door.
	back, err := engine.EnableMonitor(ctx, 1)
	if err != nil || back.SessionID == "" {
		t.Fatalf("re-enable = %+v %v", back, err)
	}
}

// A rollback the platform could only half do is reported as something to
// retry, never as done: in that state nothing writes the records, which is the
// safe half to be stuck in, and calling again finishes it.
func TestDisableMonitorReportsAHalfDoneRollbackForRetry(t *testing.T) {
	p := &setupPlatform{sessions: []map[string]any{{"id": "chat-1", "name": "chat", "state": "active"}}}
	engine := setupEngine(t, p, "chat-1")
	ctx := t.Context()
	if _, err := engine.EnableMonitor(ctx, 1); err != nil {
		t.Fatal(err)
	}
	p.offReply = http.StatusAccepted
	off, err := engine.DisableMonitor(ctx, 1)
	if err != nil {
		t.Fatalf("disable: %v", err)
	}
	if off.Off || !off.Retry || off.Note == "" {
		t.Fatalf("a half-done rollback = %+v, want retry with a reason", off)
	}
	// The retry finishes it.
	p.offReply = 0
	done, err := engine.DisableMonitor(ctx, 1)
	if err != nil || !done.Off {
		t.Fatalf("retry = %+v %v", done, err)
	}
	if len(p.deletes) != 2 {
		t.Fatalf("rollback calls = %v, want the retry to have gone through", p.deletes)
	}
}

// A refused or unreadable rollback is an error the person is told about, not a
// quiet "it is off now" — because a pair that is still bound still owns the
// work records.
func TestDisableMonitorDoesNotClaimSuccessOnAnUnclearAnswer(t *testing.T) {
	p := &setupPlatform{sessions: []map[string]any{{"id": "chat-1", "name": "chat", "state": "active"}}}
	engine := setupEngine(t, p, "chat-1")
	ctx := t.Context()
	if _, err := engine.EnableMonitor(ctx, 1); err != nil {
		t.Fatal(err)
	}
	p.offReply = http.StatusForbidden
	if _, err := engine.DisableMonitor(ctx, 1); !errors.Is(err, chat.ErrPairRefused) {
		t.Fatalf("a refused rollback = %v, want ErrPairRefused", err)
	}
	p.offReply = http.StatusInternalServerError
	if _, err := engine.DisableMonitor(ctx, 1); !errors.Is(err, chat.ErrPairUnknown) {
		t.Fatalf("an unreadable rollback = %v, want ErrPairUnknown", err)
	}
}
