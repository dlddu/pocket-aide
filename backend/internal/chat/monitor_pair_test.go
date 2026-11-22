package chat_test

import (
	"context"
	"encoding/json"
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
	sessions []map[string]any
	// frozen is the session whose work routes answer "no live pod", as the
	// platform does for a snapshotted session.
	frozen map[string]bool
	paths  []string
	bodies []string
	sends  int
}

func (p *pairedPlatform) Do(_ context.Context, method, path string, body []byte) (int, []byte, error) {
	p.paths = append(p.paths, method+" "+path)
	p.bodies = append(p.bodies, string(body))
	if method == http.MethodGet && path == "/sessions" {
		raw, _ := json.Marshal(p.sessions)
		return http.StatusOK, raw, nil
	}
	id := strings.TrimPrefix(path, "/sessions/")
	id, _, _ = strings.Cut(id, "/")
	if p.frozen[id] {
		return http.StatusServiceUnavailable,
			[]byte(`{"error":"the session has no live pod to read its work records through"}`), nil
	}
	switch {
	case strings.Contains(path, "/work/tasks/") && strings.HasSuffix(path, "/requests"):
		p.sends++
		return http.StatusOK, []byte(`{"turn":{"id":9,"task_id":"task-1","requested_at":"2026-11-22T00:00:00Z"},"replayed":false}`), nil
	case strings.Contains(path, "/turns"):
		return http.StatusOK, []byte(`{"task_id":"task-1","task":{"id":"task-1","title":"t","status":"running","version":1},` +
			`"turns":[{"id":9,"task_id":"task-1","response":"반영 끝","requested_at":"2026-11-22T00:00:00Z"}]}`), nil
	case strings.Contains(path, "/work/tasks"):
		return http.StatusOK, []byte(`{"tasks":[{"task":{"id":"task-1","title":"모니터 작업","status":"running","version":2},` +
			`"turns_total":1}],"next_cursor":""}`), nil
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
	p := &pairedPlatform{sessions: []map[string]any{
		{"id": "chat-1", "state": "active"},
		{"id": "monitor-1", "state": "active", "monitorFor": "chat-1"},
		// Another user's pair in the same tenant must not be picked up.
		{"id": "monitor-2", "state": "active", "monitorFor": "chat-2"},
	}}
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
	if !strings.Contains(read, "/sessions/monitor-1/") {
		t.Fatalf("the panel read %q, want the monitor's session", read)
	}
	if strings.Contains(read, "monitor-2") || strings.Contains(read, "chat-2") {
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
	p := &pairedPlatform{sessions: []map[string]any{
		{"id": "chat-1", "state": "active"},
		{"id": "monitor-1", "state": "idle", "monitorFor": "chat-1"},
	}}
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
	if !strings.Contains(sent, "/sessions/monitor-1/") {
		t.Fatalf("instruction went to %q, want the monitor", sent)
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
		sessions: []map[string]any{
			{"id": "chat-1", "state": "active"},
			{"id": "monitor-1", "state": "snapshot", "monitorFor": "chat-1"},
		},
		frozen: map[string]bool{"monitor-1": true},
	}
	_, engine := pairedEngine(t, p, "chat-1")

	if _, err := engine.Tasks(context.Background(), 1); err == nil {
		t.Fatal("a frozen monitor must not read as an empty panel")
	} else if !strings.Contains(err.Error(), "snapshot") {
		t.Fatalf("the error must name the monitor's state: %v", err)
	}
	_, err := engine.CreateTaskRequest(context.Background(), 1, "task-1", "req-app-2", "나중에", 0)
	if err == nil {
		t.Fatal("a frozen monitor must not report an instruction as delivered")
	}
	if p.sends != 0 {
		t.Fatalf("sends = %d, want none while the monitor is frozen", p.sends)
	}
}

// An unpaired chat session keeps reading and writing its own records, which is
// what every existing deployment does.
func TestUnpairedChatSessionIsUnchanged(t *testing.T) {
	p := &pairedPlatform{sessions: []map[string]any{{"id": "chat-1", "state": "active"}}}
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
	if !strings.Contains(read, "/sessions/chat-1/") {
		t.Fatalf("an unpaired panel read %q, want its own session", read)
	}
}
