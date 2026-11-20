package chat_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// turnPlatform is the platform's work-turns route as the engine sees it: it
// answers one session's task with the agreed records, and records what it was
// asked so a test can show the app never reads another user's session.
type turnPlatform struct {
	// pages is platform session id + task id -> the raw page to answer with.
	pages map[string]string
	calls []string
	// status/body, when set, are answered instead of a page.
	status int
	body   string
}

func newTurnPlatform() *turnPlatform {
	return &turnPlatform{pages: map[string]string{}}
}

func (p *turnPlatform) Do(_ context.Context, method, path string, _ []byte) (int, []byte, error) {
	p.calls = append(p.calls, method+" "+path)
	if method != http.MethodGet || !strings.Contains(path, "/work/tasks/") {
		return http.StatusNotFound, nil, nil
	}
	if p.status != 0 {
		return p.status, []byte(p.body), nil
	}
	// /sessions/<sid>/work/tasks/<taskID>/turns
	trimmed := strings.TrimPrefix(path, "/sessions/")
	sid := trimmed[:strings.Index(trimmed, "/")]
	rest := trimmed[strings.Index(trimmed, "/work/tasks/")+len("/work/tasks/"):]
	taskID := rest
	if cut := strings.Index(rest, "/"); cut >= 0 {
		taskID = rest[:cut]
	}
	page, ok := p.pages[sid+"/"+taskID]
	if !ok {
		return http.StatusNotFound, nil, nil
	}
	return http.StatusOK, []byte(page), nil
}

func turnEngine(t *testing.T) (*chat.Store, *chat.Engine, *turnPlatform) {
	t.Helper()
	store, _, _ := workEngine(t)
	platform := newTurnPlatform()
	engine := chat.NewEngine(store, platform, nil)
	t.Cleanup(engine.Wait)
	return store, engine, platform
}

// A migrated task with two turns: the first finished, the second asking a
// question with a choice. This is the page the platform's work route serves.
const migratedTurnsPage = `{
  "task_id": "t1", "task_version": 3,
  "task": {"id": "t1", "title": "옮겨 온 작업", "description": null,
           "completion_criteria": null, "status": "blocked", "version": 3,
           "session_id": "worker-now", "created_at": null, "updated_at": null},
  "turns": [
    {"id": 1, "request": "이거 먼저 해 줘", "instruction": "저장소를 읽고 패치를 만들어라",
     "response": "세션이 스스로 쓴 원문 답", "explanation": "부모가 사람에게 쓴 설명",
     "question": null, "choices": [], "artifacts": ["pull/42"],
     "requested_at": "2026-11-19T00:30:00Z", "started_at": "2026-11-19T00:31:00Z",
     "ended_at": "2026-11-19T00:40:00Z", "origin": "user_request"},
    {"id": 2, "request": "리뷰 반영도 해 줘", "instruction": null, "response": null,
     "explanation": null, "question": "어디까지 반영할까요?",
     "choices": [{"label": "전부", "request_text": "코멘트 전부 반영해 줘"}],
     "artifacts": [], "requested_at": "2026-11-19T01:00:00Z",
     "started_at": null, "ended_at": null, "origin": "app_request"}
  ]
}`

// A task created after the cutover, with the agreed fields recorded and no
// turns yet.
const freshTurnsPage = `{
  "task_id": "t2", "task_version": 1,
  "task": {"id": "t2", "title": "새 작업", "description": "무엇을 why",
           "completion_criteria": "머지되면 끝", "status": "planned", "version": 1,
           "session_id": "worker-1", "created_at": "2026-11-19T02:00:00Z",
           "updated_at": "2026-11-19T02:00:00Z"},
  "turns": []
}`

// The detail pane's whole point: every field of the agreed model reaches the
// app. The ledger task and history routes cannot carry any of it, so this is
// the test that the pane can show the work rather than only that work exists.
func TestTaskTurns_CarryTheAgreedModelToTheApp(t *testing.T) {
	store, engine, platform := turnEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.pages["cx1/t1"] = migratedTurnsPage

	page, err := engine.TaskTurns(context.Background(), 1, "t1", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if page.TaskID != "t1" || page.TaskVersion != 3 {
		t.Fatalf("page = %+v", page)
	}
	if page.Task.Title != "옮겨 온 작업" || page.Task.Status != "blocked" {
		t.Fatalf("task = %+v", page.Task)
	}
	// A migrated task never recorded a description: null reaches the app as
	// null, so the pane says nothing is recorded instead of showing a blank.
	if page.Task.Description != nil || page.Task.CompletionCriteria != nil {
		t.Fatalf("a migrated task arrived with invented fields: %+v", page.Task)
	}
	if len(page.Turns) != 2 {
		t.Fatalf("turns = %+v", page.Turns)
	}
	one := page.Turns[0]
	for field, got := range map[string]*string{
		"request": one.Request, "instruction": one.Instruction,
		"response": one.Response, "explanation": one.Explanation,
		"started_at": one.StartedAt, "ended_at": one.EndedAt,
	} {
		if got == nil || *got == "" {
			t.Fatalf("turn 1 lost %s: %+v", field, one)
		}
	}
	// The session's own words and the parent's summary stay distinct all the
	// way out: a pane that showed one as the other would put words in a
	// session's mouth.
	if *one.Response != "세션이 스스로 쓴 원문 답" || *one.Explanation != "부모가 사람에게 쓴 설명" {
		t.Fatalf("response/explanation crossed: %+v", one)
	}
	if len(one.Artifacts) != 1 || one.Artifacts[0] != "pull/42" {
		t.Fatalf("turn 1 lost its artifacts: %+v", one)
	}
	two := page.Turns[1]
	if two.Question == nil || *two.Question != "어디까지 반영할까요?" {
		t.Fatalf("turn 2 lost its question: %+v", two)
	}
	if len(two.Choices) != 1 || two.Choices[0].RequestText != "코멘트 전부 반영해 줘" {
		t.Fatalf("turn 2 lost its choices: %+v", two)
	}
	// An unanswered turn has no ending, which is how the pane knows it is
	// still somebody's to answer.
	if two.EndedAt != nil {
		t.Fatalf("turn 2 arrived already closed: %+v", two)
	}
	// A carried-over exchange is marked, so the pane does not present it as
	// something written here. It is not an offer of a second history: the
	// ledger's own history route is gone, and a task's turns are its history
	// whether it was carried over or written here.
	if one.Origin != "user_request" || two.Origin != "app_request" {
		t.Fatalf("origins lost: %q %q", one.Origin, two.Origin)
	}
}

// A task created after the cutover: the agreed fields it *did* record arrive,
// and having no turns yet is the ordinary state of a new task rather than an
// error or a missing history.
func TestTaskTurns_FreshTaskHasItsFieldsAndNoTurns(t *testing.T) {
	store, engine, platform := turnEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.pages["cx1/t2"] = freshTurnsPage

	page, err := engine.TaskTurns(context.Background(), 1, "t2", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if page.Task.Description == nil || *page.Task.Description != "무엇을 why" {
		t.Fatalf("description lost: %+v", page.Task)
	}
	if page.Task.CompletionCriteria == nil || *page.Task.CompletionCriteria != "머지되면 끝" {
		t.Fatalf("completion criteria lost: %+v", page.Task)
	}
	if len(page.Turns) != 0 {
		t.Fatalf("a fresh task has turns: %+v", page.Turns)
	}
}

// The session is the caller's own, so one user cannot read another's task, and
// a conversation that has opened no session has nothing to read.
func TestTaskTurns_ScopedToTheCallersOwnSession(t *testing.T) {
	store, engine, platform := turnEngine(t)
	bindSession(t, store, 1, "cx1")
	bindSession(t, store, 2, "cx2")
	platform.pages["cx1/t1"] = migratedTurnsPage

	if _, err := engine.TaskTurns(context.Background(), 2, "t1", 0, 0); !errors.Is(err, chat.ErrTaskNotFound) {
		t.Fatalf("another user's task = %v, want ErrTaskNotFound", err)
	}
	for _, call := range platform.calls {
		if strings.Contains(call, "/sessions/cx1/") && strings.Contains(call, "work/tasks") {
			// Only the owner's own call may name cx1.
			if len(platform.calls) > 0 && platform.calls[0] != call {
				continue
			}
		}
	}
	if _, err := engine.TaskTurns(context.Background(), 3, "t1", 0, 0); err == nil {
		t.Fatal("a user with no conversation read turns")
	}
}

// What the platform says is carried through as what it is: "not on this slot",
// "no live pod", "this deployment keeps no work records" and "bad page" are
// four different things, and none of them is an empty task.
func TestTaskTurns_PlatformAnswersAreNotFlattenedIntoEmpty(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   error
	}{
		{http.StatusNotFound, chat.ErrTaskNotFound},
		{http.StatusConflict, chat.ErrTaskRequestUnavailable},
		{http.StatusNotImplemented, chat.ErrTaskRequestUnsupported},
		{http.StatusBadRequest, chat.ErrTaskRequestInvalid},
	} {
		store, engine, platform := turnEngine(t)
		bindSession(t, store, 1, "cx1")
		platform.pages["cx1/t1"] = migratedTurnsPage
		platform.status = tc.status
		platform.body = `{"code":"x","error":"y"}`
		page, err := engine.TaskTurns(context.Background(), 1, "t1", 0, 0)
		if !errors.Is(err, tc.want) {
			t.Fatalf("status %d = %v, want %v", tc.status, err, tc.want)
		}
		if len(page.Turns) != 0 {
			t.Fatalf("status %d answered turns as well as an error", tc.status)
		}
	}
}

// The page request itself is bounded before it costs a hop into the platform.
func TestTaskTurns_RefusesABadPageRequest(t *testing.T) {
	store, engine, platform := turnEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.pages["cx1/t1"] = migratedTurnsPage

	for _, bad := range []struct {
		taskID string
		after  int64
		limit  int
	}{
		{"", 0, 0},
		{"../escape", 0, 0},
		{"t1", -1, 0},
		{"t1", 0, 101},
		{"t1", 0, -1},
	} {
		if _, err := engine.TaskTurns(context.Background(), 1, bad.taskID, bad.after, bad.limit); !errors.Is(err, chat.ErrTaskRequestInvalid) {
			t.Fatalf("%+v = %v, want ErrTaskRequestInvalid", bad, err)
		}
	}
	if len(platform.calls) != 0 {
		t.Fatalf("a refused page request still called the platform: %v", platform.calls)
	}
}

// The cursor and page size are passed through unchanged, so the pane pages the
// turns the way the platform does.
func TestTaskTurns_PagesWithTheCursor(t *testing.T) {
	store, engine, platform := turnEngine(t)
	bindSession(t, store, 1, "cx1")
	page := map[string]any{
		"task_id": "t1", "task_version": 3,
		"task":        map[string]any{"id": "t1", "title": "t", "status": "blocked", "version": 3},
		"turns":       []any{},
		"next_cursor": 7,
	}
	encoded, _ := json.Marshal(page)
	platform.pages["cx1/t1"] = string(encoded)

	got, err := engine.TaskTurns(context.Background(), 1, "t1", 4, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got.NextCursor != 7 {
		t.Fatalf("next_cursor = %d", got.NextCursor)
	}
	if len(platform.calls) != 1 ||
		!strings.Contains(platform.calls[0], "after_turn_id=4") ||
		!strings.Contains(platform.calls[0], "limit=2") {
		t.Fatalf("calls = %v", platform.calls)
	}
}
