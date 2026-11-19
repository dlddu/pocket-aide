package chat_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// requestPlatform is session-platform's ledger write route as the engine sees
// it. It applies the platform's own rules — the task has to exist, the request
// id is the idempotency key — and records what it was asked, so a test can show
// that one user's request never reaches another's session and that storing one
// never touches the chat's message routes.
type requestPlatform struct {
	tasks  map[string]map[string]bool // platform session id -> task ids it holds
	stored map[string]storedRequest   // request id -> what was stored
	order  []string                   // request ids in arrival order
	calls  []string                   // method+path of every call
	// status, when set, is answered instead of storing.
	status int
	body   string
}

type storedRequest struct {
	ID, TaskID, Text, Author, Session string
	Seq                               int64
}

func newRequestPlatform() *requestPlatform {
	return &requestPlatform{tasks: map[string]map[string]bool{}, stored: map[string]storedRequest{}}
}

func (p *requestPlatform) Do(_ context.Context, method, path string, body []byte) (int, []byte, error) {
	p.calls = append(p.calls, method+" "+path)
	if method != http.MethodPost || !strings.Contains(path, "/ledger/tasks/") {
		return http.StatusNotFound, nil, nil
	}
	if p.status != 0 {
		return p.status, []byte(p.body), nil
	}
	rest := strings.TrimPrefix(path, "/sessions/")
	session, rest, _ := strings.Cut(rest, "/ledger/tasks/")
	session, _ = url.PathUnescape(session)
	taskID, _, _ := strings.Cut(rest, "/requests")
	taskID, _ = url.PathUnescape(taskID)

	var in struct {
		ID     string `json:"id"`
		Text   string `json:"text"`
		Author string `json:"author"`
	}
	if err := json.Unmarshal(body, &in); err != nil {
		return http.StatusBadRequest, []byte(`{"error":"bad body"}`), nil
	}
	if existing, ok := p.stored[in.ID]; ok {
		if existing.TaskID != taskID || existing.Text != in.Text {
			return http.StatusConflict, []byte(`{"error":"request id already used with different content"}`), nil
		}
		return http.StatusOK, requestBody(existing, true), nil
	}
	if !p.tasks[session][taskID] {
		return http.StatusNotFound, []byte(`{"error":"task not found"}`), nil
	}
	stored := storedRequest{
		ID: in.ID, TaskID: taskID, Text: in.Text, Author: in.Author,
		Session: session, Seq: int64(len(p.order) + 1),
	}
	p.stored[in.ID] = stored
	p.order = append(p.order, in.ID)
	return http.StatusCreated, requestBody(stored, false), nil
}

func requestBody(s storedRequest, replayed bool) []byte {
	out, _ := json.Marshal(map[string]any{
		"request": map[string]any{
			"id": s.ID, "task_id": s.TaskID, "seq": s.Seq, "text": s.Text,
			"author": s.Author, "created_at": "2026-11-18T00:00:00Z", "status": "pending",
		},
		"replayed": replayed,
	})
	return out
}

func requestEngine(t *testing.T) (*chat.Store, *chat.Engine, *requestPlatform) {
	t.Helper()
	store, _, _ := ledgerEngine(t)
	platform := newRequestPlatform()
	engine := chat.NewEngine(store, platform, nil)
	t.Cleanup(engine.Wait)
	return store, engine, platform
}

// A request is stored on the caller's own session ledger, as pending, with the
// author the server derived — and nothing goes through the chat's own routes.
func TestCreateTaskRequest_StoresOnTheCallersOwnLedger(t *testing.T) {
	store, engine, platform := requestEngine(t)
	bindSession(t, store, 1, "cx1")
	bindSession(t, store, 2, "cx2")
	platform.tasks["cx1"] = map[string]bool{"t1": true}
	platform.tasks["cx2"] = map[string]bool{"t1": true}

	stored, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "req-1", "  이 작업 진행해 줘  ", "")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Replayed || stored.Request.Status != chat.TaskRequestPending {
		t.Fatalf("stored = %+v, want a first store that is pending", stored)
	}
	if stored.Request.Text != "이 작업 진행해 줘" {
		t.Fatalf("text = %q, want the trimmed words the person typed", stored.Request.Text)
	}
	// The session came from the user's own row, and the author from the
	// authenticated user — neither was in the request.
	if got := platform.stored["req-1"]; got.Session != "cx1" || got.Author != "user-1" {
		t.Fatalf("stored = %+v, want session cx1 and author user-1", got)
	}
	// Another user's request lands on their own ledger, never on this one.
	if _, err := engine.CreateTaskRequest(context.Background(), 2, "t1", "req-2", "다른 사용자", ""); err != nil {
		t.Fatal(err)
	}
	if got := platform.stored["req-2"]; got.Session != "cx2" {
		t.Fatalf("second user's session = %q, want cx2", got.Session)
	}
	for _, call := range platform.calls {
		if strings.Contains(call, "/write") || strings.Contains(call, "/read") || strings.Contains(call, "messages") {
			t.Fatalf("storing a request must not touch the chat routes, got %q", call)
		}
	}
}

// The request id is the idempotency key: the same words store once, and
// different words under a used id are a conflict rather than a silent replace.
func TestCreateTaskRequest_IsIdempotentOnTheRequestID(t *testing.T) {
	store, engine, platform := requestEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.tasks["cx1"] = map[string]bool{"t1": true}
	ctx := context.Background()

	first, err := engine.CreateTaskRequest(ctx, 1, "t1", "req-1", "진행해 줘", "")
	if err != nil {
		t.Fatal(err)
	}
	again, err := engine.CreateTaskRequest(ctx, 1, "t1", "req-1", "진행해 줘", "")
	if err != nil {
		t.Fatal(err)
	}
	if !again.Replayed || again.Request.Seq != first.Request.Seq {
		t.Fatalf("replay = %+v, want the stored row marked replayed", again)
	}
	if len(platform.order) != 1 {
		t.Fatalf("stored rows = %d, want the retry to have stored nothing twice", len(platform.order))
	}
	if _, err := engine.CreateTaskRequest(ctx, 1, "t1", "req-1", "다른 말", ""); !errors.Is(err, chat.ErrTaskRequestConflict) {
		t.Fatalf("reused id with other words = %v, want a conflict", err)
	}
}

// Requests on one task keep the order they were left in, which is the order the
// session is expected to take them in.
func TestCreateTaskRequest_KeepsArrivalOrder(t *testing.T) {
	store, engine, platform := requestEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.tasks["cx1"] = map[string]bool{"t1": true, "t2": true}
	ctx := context.Background()
	for _, r := range []struct{ task, id, text string }{
		{"t1", "req-1", "첫 번째"}, {"t2", "req-2", "다른 작업"}, {"t1", "req-3", "두 번째"},
	} {
		if _, err := engine.CreateTaskRequest(ctx, 1, r.task, r.id, r.text, ""); err != nil {
			t.Fatal(err)
		}
	}
	if got := strings.Join(platform.order, ","); got != "req-1,req-2,req-3" {
		t.Fatalf("order = %s, want arrival order", got)
	}
	if platform.stored["req-3"].Seq != 3 {
		t.Fatalf("seq = %d, want the ledger's arrival seq", platform.stored["req-3"].Seq)
	}
}

// What the app has to tell the person apart, each its own error: the task is
// gone, the session cannot be written to, this platform has no such route, and
// a malformed field. None of them stores anything.
func TestCreateTaskRequest_DistinguishesRefusals(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"unknown task", http.StatusNotFound, `{"error":"task not found"}`, chat.ErrTaskNotFound},
		{"deleted task", http.StatusConflict, `{"error":"that task was deleted"}`, chat.ErrTaskNotFound},
		{"no live pod", http.StatusConflict, `{"error":"the session has no live pod to store through"}`, chat.ErrTaskRequestUnavailable},
		{"id reused", http.StatusConflict, `{"error":"request id already used with different content"}`, chat.ErrTaskRequestConflict},
		{"no ledger", http.StatusNotImplemented, `{"error":"keeps no ledger"}`, chat.ErrTaskRequestUnsupported},
		{"older platform", http.StatusMethodNotAllowed, ``, chat.ErrTaskRequestUnsupported},
		{"platform refused fields", http.StatusBadRequest, `{"error":"bad author"}`, chat.ErrTaskRequestInvalid},
	} {
		store, engine, platform := requestEngine(t)
		bindSession(t, store, 1, "cx1")
		platform.status, platform.body = tc.status, tc.body
		if _, err := engine.CreateTaskRequest(ctx, 1, "t1", "req-1", "요청", ""); !errors.Is(err, tc.want) {
			t.Errorf("%s = %v, want %v", tc.name, err, tc.want)
		}
		if len(platform.stored) != 0 {
			t.Errorf("%s stored something", tc.name)
		}
	}
}

// A malformed id or text never leaves the app, and a conversation with no
// session yet has no ledger to leave a request on.
func TestCreateTaskRequest_RefusesBeforeTheHop(t *testing.T) {
	store, engine, platform := requestEngine(t)
	ctx := context.Background()
	bindSession(t, store, 1, "cx1")
	platform.tasks["cx1"] = map[string]bool{"t1": true}

	for _, tc := range []struct{ taskID, id, text string }{
		{"t1", "req-1", "   "},
		{"t1", "req-1", strings.Repeat("가", chat.TaskRequestTextMax+1)},
		{"t1", "요청 아이디", "요청"},
		{"작업 아이디", "req-1", "요청"},
	} {
		if _, err := engine.CreateTaskRequest(ctx, 1, tc.taskID, tc.id, tc.text, ""); !errors.Is(err, chat.ErrTaskRequestInvalid) {
			t.Errorf("%+v = %v, want invalid", tc, err)
		}
	}
	if len(platform.calls) != 0 {
		t.Fatalf("calls = %v, want nothing to have left the app", platform.calls)
	}

	// User 2 has no session bound yet.
	if _, err := engine.CreateTaskRequest(ctx, 2, "t1", "req-9", "요청", ""); !errors.Is(err, chat.ErrNoConversation) {
		t.Fatalf("no session = %v, want ErrNoConversation", err)
	}
}

// The panel learns that requests are waiting from the task list itself, and a
// task with none carries no summary at all — so "nothing waiting" and "nothing
// to say" never look the same.
func TestTasks_CarryWaitingRequestCounts(t *testing.T) {
	store, engine, platform := ledgerEngine(t)
	bindSession(t, store, 1, "cx1")
	platform.rows["cx1"] = []ledgerRow{
		row("t1", "blocked", "review"),
		row("t2", "running", "implementation"),
	}
	platform.requests = map[string]*chat.TaskRequestSummary{
		"t1": {Pending: 2, Claimed: 1, LastSeq: 7},
	}

	list, err := engine.Tasks(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tasks) != 2 {
		t.Fatalf("tasks = %d", len(list.Tasks))
	}
	waiting := list.Tasks[0].Requests
	if waiting == nil || waiting.Pending != 2 || waiting.Claimed != 1 || waiting.Open() != 3 || waiting.LastSeq != 7 {
		t.Fatalf("t1 requests = %+v, want 2 pending and 1 claimed", waiting)
	}
	if list.Tasks[1].Requests != nil {
		t.Fatalf("t2 requests = %+v, want no summary when none arrived", list.Tasks[1].Requests)
	}
}
