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
