package chat_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
)

// ledgerPlatform is session-platform's read-only ledger route as the engine
// sees it. It pages by id the way the platform does, so a test reads pages
// rather than one canned answer, and it records which sessions were asked for
// so a test can show one user's read never reaches another's session.
type ledgerPlatform struct {
	rows map[string][]ledgerRow // platform session id -> its ledger, in id order
	// status, when set for a session, is answered instead of a page.
	status map[string]int
	asked  []string
	// stuckCursor makes every page answer the same cursor, as a misbehaving
	// platform would.
	stuckCursor bool
	pages       int
	// requests is the waiting-request summary the platform rides along with a
	// task, keyed by task id (AC-N7). A task absent from the map carries none.
	requests map[string]*chat.TaskRequestSummary
}

type ledgerRow struct {
	ID       string
	Title    string
	Status   string
	Stage    string
	Sessions []ledgerSess
	Details  map[string]any
}

type ledgerSess struct {
	Cleanup    string
	Reason     string
	NextAction string
}

func newLedgerPlatform() *ledgerPlatform {
	return &ledgerPlatform{rows: map[string][]ledgerRow{}, status: map[string]int{}}
}

func (p *ledgerPlatform) Do(_ context.Context, method, path string, _ []byte) (int, []byte, error) {
	if method != http.MethodGet || !strings.Contains(path, "/ledger/tasks") {
		return http.StatusNotFound, nil, nil
	}
	raw := strings.TrimPrefix(path, "/sessions/")
	id, query, _ := strings.Cut(raw, "/ledger/tasks?")
	id, _ = url.PathUnescape(id)
	p.asked = append(p.asked, id)
	if s, ok := p.status[id]; ok {
		return s, []byte(`{"error":"no"}`), nil
	}
	values, _ := url.ParseQuery(query)
	limit := 100
	if v := values.Get("limit"); v != "" {
		_, _ = fmt.Sscanf(v, "%d", &limit)
	}
	after := values.Get("after_id")

	type outSess struct {
		Cleanup    string `json:"cleanup"`
		Reason     string `json:"reason"`
		NextAction string `json:"next_action"`
	}
	type outTask struct {
		Title    string         `json:"title"`
		Status   string         `json:"status"`
		Stage    string         `json:"stage"`
		Sessions []outSess      `json:"sessions"`
		Details  map[string]any `json:"details"`
	}
	type outRow struct {
		ID       string                   `json:"id"`
		Requests *chat.TaskRequestSummary `json:"requests,omitempty"`
		Task     outTask                  `json:"task"`
	}
	page := struct {
		Tasks      []outRow `json:"tasks"`
		NextCursor string   `json:"next_cursor"`
	}{Tasks: []outRow{}}

	p.pages++
	for _, row := range p.rows[id] {
		if after != "" && row.ID <= after {
			continue
		}
		if len(page.Tasks) == limit {
			page.NextCursor = page.Tasks[len(page.Tasks)-1].ID
			break
		}
		sessions := []outSess{}
		for _, s := range row.Sessions {
			sessions = append(sessions, outSess(s))
		}
		page.Tasks = append(page.Tasks, outRow{
			ID:       row.ID,
			Requests: p.requests[row.ID],
			Task: outTask{
				Title: row.Title, Status: row.Status, Stage: row.Stage,
				Sessions: sessions, Details: row.Details,
			},
		})
	}
	if p.stuckCursor {
		page.NextCursor = after
		if page.NextCursor == "" {
			page.NextCursor = "stuck"
		}
	}
	body, _ := json.Marshal(page)
	return http.StatusOK, body, nil
}

// ledgerEngine is a chat engine over the ledger platform, with each user's
// conversation already bound to its own platform session.
func ledgerEngine(t *testing.T) (*chat.Store, *chat.Engine, *ledgerPlatform) {
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
	platform := newLedgerPlatform()
	engine := chat.NewEngine(store, platform, nil)
	t.Cleanup(engine.Wait)
	return store, engine, platform
}

// bindSession gives userID's conversation a platform session id.
func bindSession(t *testing.T, store *chat.Store, userID int64, sessionID string) {
	t.Helper()
	conv, err := store.Conversation(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetPlatform(context.Background(), userID, conv.ID, sessionID, 0); err != nil {
		t.Fatal(err)
	}
}

func row(id, status, stage string, sessions ...ledgerSess) ledgerRow {
	return ledgerRow{ID: id, Title: "작업 " + id, Status: status, Stage: stage, Sessions: sessions}
}

// The panel reads every page of its own session's ledger, keeps status and
// stage, and counts by status.
func TestTasks_WalksEveryPage(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	var rows []ledgerRow
	for i := 0; i < 250; i++ {
		rows = append(rows, row(fmt.Sprintf("t%03d", i), "running", "implementation"))
	}
	rows = append(rows, row("z-blocked", "blocked", "review", ledgerSess{Cleanup: "retained", Reason: "CI 가 막혔다", NextAction: "로그를 본다"}))
	platform.rows["cx1"] = rows

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 251 {
		t.Fatalf("read %d tasks, want 251 — the walk stopped early", len(list.Tasks))
	}
	if list.Truncated {
		t.Fatal("a finished walk must not be reported as truncated")
	}
	if list.Counts[chat.TaskStatusRunning] != 250 || list.Counts[chat.TaskStatusBlocked] != 1 {
		t.Fatalf("counts = %v", list.Counts)
	}
	last := list.Tasks[len(list.Tasks)-1]
	if last.Status != chat.TaskStatusBlocked || last.Stage != "review" {
		t.Fatalf("blocked row = %+v", last)
	}
	if last.Reason != "CI 가 막혔다" || last.NextAction != "로그를 본다" {
		t.Fatalf("reason/next action lost: %+v", last)
	}
	if list.SessionID != "cx1" {
		t.Fatalf("session id = %q, want the session the rows were read from", list.SessionID)
	}
}

// A repeated row across pages is one task, not two: keyset pages are not a
// snapshot.
func TestTasks_DeduplicatesByTaskID(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.rows["cx1"] = []ledgerRow{
		row("t1", "running", "impl"), row("t1", "completed", "done"), row("t2", "planned", "plan"),
	}
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 2 {
		t.Fatalf("tasks = %+v, want t1 once and t2", list.Tasks)
	}
	// The first sighting wins, so t1 stays running.
	if list.Tasks[0].ID != "t1" || list.Tasks[0].Status != chat.TaskStatusRunning {
		t.Fatalf("first sighting must win: %+v", list.Tasks[0])
	}
}

// The parent's monitoring bookkeeping is not the user's work, so it stays out.
func TestTasks_ExcludesMonitorRecords(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	mine := row("t1", "running", "impl")
	monitor := row("t2", "running", "impl")
	monitor.Details = map[string]any{"kind": "Monitor"} // case must not matter
	other := row("t3", "running", "impl")
	other.Details = map[string]any{"kind": "work"}
	platform.rows["cx1"] = []ledgerRow{mine, monitor, other}

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 2 || list.Tasks[0].ID != "t1" || list.Tasks[1].ID != "t3" {
		t.Fatalf("tasks = %+v, want the monitor row excluded", list.Tasks)
	}
	if list.Counts[chat.TaskStatusRunning] != 2 {
		t.Fatalf("an excluded row must not be counted: %v", list.Counts)
	}
}

// The reason shown is the task's state now — the last entry carrying a note —
// not the first session it ever used. A finished entry loses to a live one.
func TestTasks_ReasonComesFromTheCurrentState(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.rows["cx1"] = []ledgerRow{
		row("t1", "blocked", "review",
			ledgerSess{Cleanup: "deleted", Reason: "처음엔 스키마가 막혔다", NextAction: "스키마를 고친다"},
			ledgerSess{Cleanup: "retained", Reason: "지금은 리뷰 대기다", NextAction: "리뷰를 받는다"},
		),
		// Only history: nothing live carries a note, so the finished one is
		// better than nothing.
		row("t2", "blocked", "review",
			ledgerSess{Cleanup: "deleted", Reason: "옛 사유", NextAction: "옛 다음 행동"},
		),
		// Nothing recorded at all — the panel shows "not recorded", so the
		// fields stay empty rather than being invented.
		row("t3", "blocked", "review", ledgerSess{Cleanup: "pending"}),
	}
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if list.Tasks[0].Reason != "지금은 리뷰 대기다" || list.Tasks[0].NextAction != "리뷰를 받는다" {
		t.Fatalf("current state must win over the first session: %+v", list.Tasks[0])
	}
	if list.Tasks[1].Reason != "옛 사유" {
		t.Fatalf("history is better than nothing: %+v", list.Tasks[1])
	}
	if list.Tasks[2].Reason != "" || list.Tasks[2].NextAction != "" {
		t.Fatalf("an unrecorded note must stay empty: %+v", list.Tasks[2])
	}
}

// A walk that cannot finish says so. Reporting a prefix as the whole list would
// read as "this is all the work there is".
func TestTasks_UnfinishedWalkIsReportedTruncated(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.rows["cx1"] = []ledgerRow{row("t1", "running", "impl")}
	platform.stuckCursor = true

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !list.Truncated {
		t.Fatal("a cursor that does not move must end the walk as truncated")
	}
	if len(list.Tasks) != 1 {
		t.Fatalf("tasks = %+v", list.Tasks)
	}
}

// A conversation that has not opened a session has no tasks — and nothing is
// asked of the platform.
func TestTasks_NoSessionYetIsAnEmptyList(t *testing.T) {
	_, engine, platform := ledgerEngine(t)
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 0 || list.SessionID != "" {
		t.Fatalf("list = %+v", list)
	}
	if len(platform.asked) != 0 {
		t.Fatalf("a conversation with no session must ask the platform nothing, asked %v", platform.asked)
	}
	for _, status := range []string{chat.TaskStatusBlocked, chat.TaskStatusRunning, chat.TaskStatusCompleted} {
		if _, ok := list.Counts[status]; !ok {
			t.Fatalf("counts must be present and zero, got %v", list.Counts)
		}
	}
}

// "Cannot read" is never "no tasks": a session with no ledger and a session
// that is not awake are distinct errors, and both differ from an empty list.
func TestTasks_UnsupportedAndUnavailableAreNotEmptyLists(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		want   error
	}{
		{"no ledger on this type", http.StatusNotImplemented, chat.ErrTasksUnsupported},
		{"no such route or session", http.StatusNotFound, chat.ErrTasksUnsupported},
		{"session is not awake", http.StatusConflict, chat.ErrTasksUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, engine, platform := ledgerEngine(t)
			bindSession(t, store, 1, "cx1")
			platform.status["cx1"] = tc.status
			_, err := engine.Tasks(context.Background(), 1)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// A platform fault is an error, not an empty list.
func TestTasks_PlatformFaultIsAnError(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.status["cx1"] = http.StatusInternalServerError
	if _, err := engine.Tasks(context.Background(), 1); err == nil {
		t.Fatal("a 500 from the platform must not read as an empty ledger")
	}
}

// Each user reads its own conversation's session. The session is taken from the
// caller's own row, so one user's read never reaches another's ledger.
func TestTasks_ReadsOnlyTheCallersOwnSession(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	bindSession(t, store, 2, "cx2")
	platform.rows["cx1"] = []ledgerRow{row("mine", "running", "impl")}
	platform.rows["cx2"] = []ledgerRow{row("theirs", "running", "impl"), row("theirs2", "blocked", "review")}

	first, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Tasks) != 1 || first.Tasks[0].ID != "mine" || first.SessionID != "cx1" {
		t.Fatalf("user 1 read %+v", first)
	}
	second, err := engine.Tasks(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Tasks) != 2 || second.SessionID != "cx2" {
		t.Fatalf("user 2 read %+v", second)
	}
	for _, asked := range platform.asked {
		if asked != "cx1" && asked != "cx2" {
			t.Fatalf("the platform was asked for %q, which is nobody's conversation", asked)
		}
	}
}

// After the conversation moves to another session, a read reports the new
// session's id — so a page that arrives late can be told apart from the
// current one rather than shown under it.
func TestTasks_SessionSwitchIsIdentifiedInTheAnswer(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.rows["cx1"] = []ledgerRow{row("old", "running", "impl")}
	platform.rows["cx2"] = []ledgerRow{row("new", "planned", "plan")}

	before, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	bindSession(t, store, 1, "cx2")
	after, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if before.SessionID == after.SessionID {
		t.Fatal("a switched conversation must report a different session id")
	}
	if after.Tasks[0].ID != "new" {
		t.Fatalf("after the switch the panel must read the new session: %+v", after.Tasks)
	}
}

// A status the app does not know is shown as unknown rather than guessed at.
func TestTasks_UnknownStatusIsNotGuessed(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.rows["cx1"] = []ledgerRow{row("t1", "DEPLOYING", "impl"), row("t2", "  Review  ", "review")}
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if list.Tasks[0].Status != chat.TaskStatusUnknown {
		t.Fatalf("unknown status = %q", list.Tasks[0].Status)
	}
	if list.Tasks[1].Status != chat.TaskStatusReview {
		t.Fatalf("a known status in other casing must normalise: %q", list.Tasks[1].Status)
	}
}
