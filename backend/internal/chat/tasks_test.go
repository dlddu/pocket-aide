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

// workPlatform is session-platform's read-only work task route as the engine
// sees it. It pages by id the way the platform does, so a test reads pages
// rather than one canned answer, and it records which sessions were asked for
// so a test can show one user's read never reaches another's session.
type workPlatform struct {
	rows map[string][]workRow // platform session id -> its tasks, in id order
	// status, when set for a session, is answered instead of a page.
	status map[string]int
	asked  []string
	// paths is every path that was asked for, so a test can show the engine
	// stopped calling a route that no longer exists.
	paths []string
	// stuckCursor makes every page answer the same cursor, as a misbehaving
	// platform would.
	stuckCursor bool
	pages       int
}

// workRow is one task as the platform's page carries it: the agreed task, its
// turn count, its latest turn and whether it came from the retired ledger.
type workRow struct {
	ID                 string
	Title              string
	Status             string
	Description        *string
	CompletionCriteria *string
	TurnsTotal         int64
	Migrated           bool
	Latest             *workLatest
}

type workLatest struct {
	ID        int64
	Request   *string
	Question  *string
	Choices   []chat.TaskChoice
	StartedAt *string
	EndedAt   *string
}

func newWorkPlatform() *workPlatform {
	return &workPlatform{rows: map[string][]workRow{}, status: map[string]int{}}
}

func (p *workPlatform) Do(_ context.Context, method, path string, _ []byte) (int, []byte, error) {
	p.paths = append(p.paths, method+" "+path)
	if method != http.MethodGet || !strings.Contains(path, "/work/tasks") {
		return http.StatusNotFound, nil, nil
	}
	raw := strings.TrimPrefix(path, "/sessions/")
	id, query, _ := strings.Cut(raw, "/work/tasks?")
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

	type outTurn struct {
		ID        int64             `json:"id"`
		Request   *string           `json:"request"`
		Question  *string           `json:"question"`
		Choices   []chat.TaskChoice `json:"choices"`
		StartedAt *string           `json:"started_at"`
		EndedAt   *string           `json:"ended_at"`
	}
	type outTask struct {
		ID                 string  `json:"id"`
		Title              string  `json:"title"`
		Description        *string `json:"description"`
		CompletionCriteria *string `json:"completion_criteria"`
		Status             string  `json:"status"`
	}
	type outRow struct {
		Task       outTask  `json:"task"`
		TurnsTotal int64    `json:"turns_total"`
		LatestTurn *outTurn `json:"latest_turn,omitempty"`
		Migrated   bool     `json:"migrated,omitempty"`
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
			page.NextCursor = page.Tasks[len(page.Tasks)-1].Task.ID
			break
		}
		out := outRow{
			Task: outTask{
				ID: row.ID, Title: row.Title, Status: row.Status,
				Description: row.Description, CompletionCriteria: row.CompletionCriteria,
			},
			TurnsTotal: row.TurnsTotal, Migrated: row.Migrated,
		}
		if row.Latest != nil {
			out.LatestTurn = &outTurn{
				ID: row.Latest.ID, Request: row.Latest.Request, Question: row.Latest.Question,
				Choices: row.Latest.Choices, StartedAt: row.Latest.StartedAt, EndedAt: row.Latest.EndedAt,
			}
		}
		page.Tasks = append(page.Tasks, out)
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

// workEngine is a chat engine over the work platform, with each user's
// conversation already bound to its own platform session.
func workEngine(t *testing.T) (*chat.Store, *chat.Engine, *workPlatform) {
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
	platform := newWorkPlatform()
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

func str(s string) *string { return &s }

func row(id, status string) workRow {
	return workRow{ID: id, Title: "작업 " + id, Status: status}
}

// taskByID finds one task in the answer, so an assertion names the task it is
// about rather than an index that shifts when the fixture grows.
func taskByID(t *testing.T, list chat.TaskList, id string) chat.Task {
	t.Helper()
	for _, task := range list.Tasks {
		if task.ID == id {
			return task
		}
	}
	t.Fatalf("task %q is not in %+v", id, list.Tasks)
	return chat.Task{}
}

// The whole read side in one test: the panel walks every page of its own
// session's work records, carries the agreed task fields the ledger page could
// not hold, and counts by status.
func TestTasks_CarryTheAgreedModelAndWalkEveryPage(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	rows := []workRow{}
	for i := 0; i < 150; i++ {
		rows = append(rows, row(fmt.Sprintf("t%03d", i), chat.TaskStatusRunning))
	}
	rows[0].Description = str("무엇을 why")
	rows[0].CompletionCriteria = str("머지되면 끝")
	rows[0].TurnsTotal = 3
	rows[0].Migrated = true
	rows[1].Status = chat.TaskStatusBlocked
	platform.rows["sess-1"] = rows

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatalf("tasks: %v", err)
	}
	if len(list.Tasks) != 150 || list.Truncated {
		t.Fatalf("tasks = %d truncated=%v, want the whole walk", len(list.Tasks), list.Truncated)
	}
	if platform.pages < 2 {
		t.Fatalf("pages = %d, want more than one page walked", platform.pages)
	}
	first := taskByID(t, list, "t000")
	// The two fields the ledger page had nowhere to put. Without them a panel
	// could say work existed but never what the work was.
	if first.Description != "무엇을 why" || first.CompletionCriteria != "머지되면 끝" {
		t.Fatalf("task = %+v, want the agreed description and completion criteria", first)
	}
	if first.TurnsTotal != 3 || !first.Migrated {
		t.Fatalf("task = %+v, want the turn count and the carried-over flag", first)
	}
	if list.Counts[chat.TaskStatusBlocked] != 1 || list.Counts[chat.TaskStatusRunning] != 149 {
		t.Fatalf("counts = %v", list.Counts)
	}
	// Every status is a key even at zero, so the panel's filter chips do not
	// have to decide between "no such status" and "none of these".
	for _, s := range []string{chat.TaskStatusPlanned, chat.TaskStatusReview,
		chat.TaskStatusCompleted, chat.TaskStatusUnknown} {
		if _, ok := list.Counts[s]; !ok {
			t.Fatalf("counts = %v, want %s present at zero", list.Counts, s)
		}
	}
	// It read the work route and never the ledger route the platform retired.
	for _, path := range platform.paths {
		if strings.Contains(path, "/ledger/") {
			t.Fatalf("the engine called a retired route: %s", path)
		}
	}
}

// Keyset pages are not a snapshot, so a retried page can repeat a row. The
// first sighting of an id wins; a second would show the task twice.
func TestTasks_DeduplicatesByTaskID(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.rows["sess-1"] = []workRow{
		row("t1", chat.TaskStatusRunning), row("t1", chat.TaskStatusCompleted),
		row("t2", chat.TaskStatusPlanned),
	}
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 2 {
		t.Fatalf("tasks = %+v, want the repeat dropped", list.Tasks)
	}
	if taskByID(t, list, "t1").Status != chat.TaskStatusRunning {
		t.Fatalf("the repeat replaced the first sighting: %+v", list.Tasks)
	}
}

// What the panel shows as "somebody is waiting" now comes from the task's
// latest turn, because that is where the agreed records keep it. A turn with
// words and no end is one somebody still owes an answer; started means the
// parent took it.
func TestTasks_WaitingComesFromTheLatestTurn(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	waiting := row("t-waiting", chat.TaskStatusBlocked)
	waiting.Latest = &workLatest{ID: 7, Request: str("이거 해 줘"), Question: str("어디까지 할까요?")}
	running := row("t-running", chat.TaskStatusRunning)
	running.Latest = &workLatest{ID: 8, Request: str("저것도"), StartedAt: str("2026-11-20T00:00:00Z")}
	done := row("t-done", chat.TaskStatusCompleted)
	done.Latest = &workLatest{
		ID: 9, Request: str("끝난 말"),
		StartedAt: str("2026-11-20T00:00:00Z"), EndedAt: str("2026-11-20T01:00:00Z"),
	}
	quiet := row("t-quiet", chat.TaskStatusPlanned)
	platform.rows["sess-1"] = []workRow{done, quiet, running, waiting}

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	w := taskByID(t, list, "t-waiting").Waiting
	if w == nil || w.TurnID != 7 || w.Running {
		t.Fatalf("waiting = %+v, want the open turn nobody has taken", w)
	}
	// The question is what the session is actually asking — the nearest thing
	// the agreed records have to the ledger's free-text reason.
	if w.Question != "어디까지 할까요?" {
		t.Fatalf("waiting = %+v, want the question", w)
	}
	if r := taskByID(t, list, "t-running").Waiting; r == nil || !r.Running {
		t.Fatalf("running = %+v, want the turn marked taken", r)
	}
	// A finished turn and a task with no turn at all are both "nothing
	// waiting", and neither gets a zeroed record that would read as "nothing
	// was ever asked here".
	if taskByID(t, list, "t-done").Waiting != nil {
		t.Fatal("a finished turn is shown as waiting")
	}
	if taskByID(t, list, "t-quiet").Waiting != nil {
		t.Fatal("a task with no turns is shown as waiting")
	}
}

// Only the latest turn's choices are current, and the turn they belong to
// travels with them — that turn id is what the platform checks a pick against.
func TestTasks_OfferedChoicesCarryTheirTurn(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	offered := row("t-offer", chat.TaskStatusReview)
	offered.Latest = &workLatest{
		ID: 42, Request: str("리뷰 반영해 줘"), Question: str("어디까지?"),
		Choices: []chat.TaskChoice{
			{Label: "전부", RequestText: "코멘트 전부 반영해 줘"},
			{Label: "필수만", RequestText: "필수 코멘트만 반영해 줘"},
		},
	}
	bare := row("t-bare", chat.TaskStatusRunning)
	bare.Latest = &workLatest{ID: 43, Request: str("진행 중")}
	platform.rows["sess-1"] = []workRow{bare, offered}

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	got := taskByID(t, list, "t-offer")
	if len(got.Choices) != 2 || got.Choices[0].RequestText != "코멘트 전부 반영해 줘" {
		t.Fatalf("choices = %+v", got.Choices)
	}
	if got.ChoiceTurnID != 42 {
		t.Fatalf("choice_turn_id = %d, want the turn the choices were offered on", got.ChoiceTurnID)
	}
	// A turn offering none carries no turn id either: a bare id with no
	// choices would let a screen send a pick nothing offered.
	if bareTask := taskByID(t, list, "t-bare"); len(bareTask.Choices) != 0 || bareTask.ChoiceTurnID != 0 {
		t.Fatalf("a turn with no choices = %+v", bareTask)
	}
}

// A walk that stops at its page bound says so. A panel that silently dropped
// the tail would read as "this is all the work there is".
func TestTasks_UnfinishedWalkIsReportedTruncated(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.rows["sess-1"] = []workRow{row("t1", chat.TaskStatusRunning)}
	platform.stuckCursor = true
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !list.Truncated {
		t.Fatal("a cursor that never moves was reported as a finished walk")
	}
}

// A conversation no session has answered yet simply has no tasks — not an
// error, and not a reason to open a session.
func TestTasks_NoSessionYetIsAnEmptyList(t *testing.T) {
	_, engine, platform := workEngine(t)
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatalf("tasks: %v", err)
	}
	if len(list.Tasks) != 0 || list.SessionID != "" {
		t.Fatalf("list = %+v, want an empty list for no session", list)
	}
	if len(platform.asked) != 0 {
		t.Fatalf("the platform was asked %v for a conversation with no session", platform.asked)
	}
}

// The three answers the panel must tell apart. Neither refusal may arrive as
// an empty list, which would claim there is no work.
func TestTasks_UnsupportedAndUnavailableAreNotEmptyLists(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		want   error
	}{
		{"a session type that keeps no work records", http.StatusNotImplemented, chat.ErrTasksUnsupported},
		{"a control plane without the route", http.StatusNotFound, chat.ErrTasksUnsupported},
		{"a session with no live pod", http.StatusConflict, chat.ErrTasksUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, engine, platform := workEngine(t)
			bindSession(t, store, 1, "sess-1")
			platform.status["sess-1"] = tc.status
			if _, err := engine.Tasks(context.Background(), 1); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// A fault is an error, not an empty panel.
func TestTasks_PlatformFaultIsAnError(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.status["sess-1"] = http.StatusInternalServerError
	_, err := engine.Tasks(context.Background(), 1)
	if err == nil || errors.Is(err, chat.ErrTasksUnsupported) || errors.Is(err, chat.ErrTasksUnavailable) {
		t.Fatalf("err = %v, want a plain failure", err)
	}
}

// The session comes from the caller's own row, so one user's panel can never
// reach another's slot.
func TestTasks_ReadsOnlyTheCallersOwnSession(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	bindSession(t, store, 2, "sess-2")
	platform.rows["sess-1"] = []workRow{row("mine", chat.TaskStatusRunning)}
	platform.rows["sess-2"] = []workRow{row("theirs", chat.TaskStatusRunning)}

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 1 || list.Tasks[0].ID != "mine" {
		t.Fatalf("tasks = %+v", list.Tasks)
	}
	for _, asked := range platform.asked {
		if asked != "sess-1" {
			t.Fatalf("user 1's read reached %q", asked)
		}
	}
}

// The answer names the session it came from, so a page that arrives after the
// conversation moved is discarded rather than shown under the new session.
func TestTasks_SessionIsIdentifiedInTheAnswer(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.rows["sess-1"] = []workRow{row("t1", chat.TaskStatusRunning)}
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if list.SessionID != "sess-1" {
		t.Fatalf("session_id = %q", list.SessionID)
	}
}

// A status this app does not know is reported as unknown rather than guessed,
// so a newer platform cannot make the panel show a status it cannot render.
func TestTasks_UnknownStatusIsNotGuessed(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.rows["sess-1"] = []workRow{
		row("t1", "something-new"), row("t2", "  RUNNING  "), row("t3", ""),
	}
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if taskByID(t, list, "t1").Status != chat.TaskStatusUnknown {
		t.Fatalf("an unknown status was guessed: %+v", list.Tasks)
	}
	// Case and padding are the platform's to write however it likes.
	if taskByID(t, list, "t2").Status != chat.TaskStatusRunning {
		t.Fatalf("a padded status was not normalised: %+v", list.Tasks)
	}
	if taskByID(t, list, "t3").Status != chat.TaskStatusUnknown {
		t.Fatalf("an empty status = %+v", list.Tasks)
	}
}

// A monitoring loop is not a task and never arrives as one: the parent's loops
// are a record of their own on the slot (platform AC-N10). The panel therefore
// has no kind filter, and a task whose details once said "monitor" is just a
// task — the filter that used to be here would now hide a person's work.
func TestTasks_NoMonitoringRowsArriveAsTasks(t *testing.T) {
	store, engine, platform := workEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.rows["sess-1"] = []workRow{
		row("t-work", chat.TaskStatusRunning), row("monitor-loop", chat.TaskStatusRunning),
	}
	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 2 {
		t.Fatalf("tasks = %+v, want both — the platform sends no monitoring rows here", list.Tasks)
	}
}
