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

// requestPlatform is session-platform's work write route as the engine sees it.
// It applies the platform's own rules — the task has to exist, the request id
// is the idempotency key, a pick has to name the task's latest turn — and
// records what it was asked, so a test can show that one user's request never
// reaches another's session and that storing one never touches the chat's
// message routes.
type requestPlatform struct {
	tasks map[string]map[string]bool // platform session id -> task ids it holds
	// latestTurn is each task's newest turn id. It is the one fact a pick is
	// checked against: the ledger's (choice id, basis version) pair is gone,
	// and what makes a pick current is the turn it was offered on still being
	// the latest.
	latestTurn map[string]int64
	stored     map[string]storedRequest // request id -> what was stored
	order      []string                 // request ids in arrival order
	calls      []string                 // method+path of every call
	// picked is the turn the body named, per request id. Zero means free text.
	picked map[string]int64
	// status, when set, is answered instead of storing.
	status int
	body   string
}

type storedRequest struct {
	ID, TaskID, Text, Author, Session string
	TurnID                            int64
}

func newRequestPlatform() *requestPlatform {
	return &requestPlatform{
		tasks: map[string]map[string]bool{}, stored: map[string]storedRequest{},
		latestTurn: map[string]int64{}, picked: map[string]int64{},
	}
}

func (p *requestPlatform) Do(_ context.Context, method, path string, body []byte) (int, []byte, error) {
	p.calls = append(p.calls, method+" "+path)
	if method != http.MethodPost || !strings.Contains(path, "/work/tasks/") {
		return http.StatusNotFound, nil, nil
	}
	if p.status != 0 {
		return p.status, []byte(p.body), nil
	}
	rest := strings.TrimPrefix(path, "/sessions/")
	session, rest, _ := strings.Cut(rest, "/work/tasks/")
	session, _ = url.PathUnescape(session)
	taskID, _, _ := strings.Cut(rest, "/requests")
	taskID, _ = url.PathUnescape(taskID)

	var in struct {
		ID           string `json:"id"`
		Text         string `json:"text"`
		Author       string `json:"author"`
		ChoiceTurnID int64  `json:"choice_turn_id"`
	}
	if err := json.Unmarshal(body, &in); err != nil {
		return http.StatusBadRequest, []byte(`{"error":"bad body"}`), nil
	}
	if existing, ok := p.stored[in.ID]; ok {
		if existing.TaskID != taskID || existing.Text != in.Text {
			return http.StatusConflict,
				[]byte(`{"code":"conflict","error":"request id already used with different content"}`), nil
		}
		return http.StatusOK, requestBody(existing, true), nil
	}
	if !p.tasks[session][taskID] {
		return http.StatusNotFound, []byte(`{"code":"not_found","error":"task not found"}`), nil
	}
	// The slot's own rule: a pick is accepted only from the task's latest turn.
	if in.ChoiceTurnID != 0 && in.ChoiceTurnID != p.latestTurn[taskID] {
		return http.StatusConflict,
			[]byte(`{"code":"stale_choice","error":"that choice is no longer offered for this task"}`), nil
	}
	p.picked[in.ID] = in.ChoiceTurnID
	// Storing appends a turn, which becomes the task's latest — so the screen
	// that just sent a pick is now one turn behind, exactly as on the slot.
	p.latestTurn[taskID]++
	stored := storedRequest{
		ID: in.ID, TaskID: taskID, Text: in.Text, Author: in.Author,
		Session: session, TurnID: p.latestTurn[taskID],
	}
	p.stored[in.ID] = stored
	p.order = append(p.order, in.ID)
	return http.StatusCreated, requestBody(stored, false), nil
}

func requestBody(s storedRequest, replayed bool) []byte {
	out, _ := json.Marshal(map[string]any{
		"turn": map[string]any{
			"id": s.TurnID, "request": s.Text, "requested_at": "2026-11-20T00:00:00Z",
			"choices": []any{}, "artifacts": []any{}, "origin": "app_request",
		},
		"replayed": replayed,
	})
	return out
}

func requestEngine(t *testing.T) (*chat.Store, *chat.Engine, *requestPlatform) {
	t.Helper()
	store, _, _ := workEngine(t)
	platform := newRequestPlatform()
	engine := chat.NewEngine(store, platform, nil)
	t.Cleanup(engine.Wait)
	return store, engine, platform
}

// A request is stored on the caller's own session, becomes a turn there with
// the author the server derived, and nothing goes through the chat's own
// routes. The turn is the whole answer: there is no stored-request row with a
// lifecycle to poll any more.
func TestCreateTaskRequest_BecomesATurnOnTheCallersOwnSession(t *testing.T) {
	store, engine, platform := requestEngine(t)
	bindSession(t, store, 1, "sess-1")
	bindSession(t, store, 2, "sess-2")
	platform.tasks["sess-1"] = map[string]bool{"t1": true}
	platform.tasks["sess-2"] = map[string]bool{"t1": true}

	stored, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "req-1", "이거 해 줘", 0)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	if stored.Replayed {
		t.Fatalf("a first store reported itself as a replay: %+v", stored)
	}
	if stored.Turn.Request == nil || *stored.Turn.Request != "이거 해 줘" {
		t.Fatalf("turn = %+v, want the person's words verbatim", stored.Turn)
	}
	if stored.Turn.Origin != "app_request" {
		t.Fatalf("origin = %q, want app_request so a reader can tell where it came from", stored.Turn.Origin)
	}
	// The session came from the caller's own row, and the author from the
	// authenticated user — neither is in the app's body.
	row := platform.stored["req-1"]
	if row.Session != "sess-1" {
		t.Fatalf("stored on %q, want the caller's own session", row.Session)
	}
	if row.Author != "user-1" {
		t.Fatalf("author = %q, want it derived from the authenticated user", row.Author)
	}
	// It called the work route and never the ledger route the platform retired.
	for _, call := range platform.calls {
		if strings.Contains(call, "/ledger/") {
			t.Fatalf("the engine called a retired route: %s", call)
		}
		if strings.Contains(call, "/messages") {
			t.Fatalf("storing a request went through a chat route: %s", call)
		}
	}
}

// The idempotency contract the app's retry depends on: the same id with the
// same words is the same turn, and different words under a used id is a
// conflict rather than a silent replacement of what the person asked.
func TestCreateTaskRequest_IsIdempotentOnTheRequestID(t *testing.T) {
	store, engine, platform := requestEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.tasks["sess-1"] = map[string]bool{"t1": true}

	first, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "req-1", "같은 말", 0)
	if err != nil {
		t.Fatal(err)
	}
	again, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "req-1", "같은 말", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !again.Replayed || again.Turn.ID != first.Turn.ID {
		t.Fatalf("retry = %+v, want the turn the first store made (%+v)", again, first.Turn)
	}
	if len(platform.order) != 1 {
		t.Fatalf("stored %d times, want one", len(platform.order))
	}
	_, err = engine.CreateTaskRequest(context.Background(), 1, "t1", "req-1", "다른 말", 0)
	if !errors.Is(err, chat.ErrTaskRequestConflict) {
		t.Fatalf("err = %v, want a conflict", err)
	}
}

// A pick names the turn it was offered on, and a pick from a turn the task has
// moved past is its own refusal — the screen is out of date, not the words, so
// the app keeps the draft and re-reads the task. Free text never carries a turn
// and is never refused for staleness.
func TestCreateTaskRequest_CarriesThePickAndRefusesAStaleOne(t *testing.T) {
	store, engine, platform := requestEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.tasks["sess-1"] = map[string]bool{"t1": true}
	platform.latestTurn["t1"] = 7

	if _, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "pick-1", "전부 해 줘", 7); err != nil {
		t.Fatalf("a pick from the latest turn: %v", err)
	}
	if platform.picked["pick-1"] != 7 {
		t.Fatalf("the turn the pick named did not reach the slot: %v", platform.picked)
	}
	// That store appended a turn, so the screen holding turn 7 is now behind.
	_, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "pick-2", "전부 해 줘", 7)
	if !errors.Is(err, chat.ErrTaskRequestStaleChoice) {
		t.Fatalf("err = %v, want the stale-pick class so the draft is kept", err)
	}
	// Free text goes through on the same task at the same moment.
	if _, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "free-1", "제 말로 쓸게요", 0); err != nil {
		t.Fatalf("free text was refused for staleness: %v", err)
	}
	if platform.picked["free-1"] != 0 {
		t.Fatalf("free text named a turn: %v", platform.picked)
	}
}

// Every refusal the platform can give, each as its own class — the app's next
// action differs for each, so a single "it failed" would be useless.
func TestCreateTaskRequest_DistinguishesRefusals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"a task the slot does not have", http.StatusNotFound, `{"code":"not_found","error":"no"}`, chat.ErrTaskNotFound},
		{"a session with no live pod", http.StatusConflict,
			`{"error":"the session has no live pod to read its work records through"}`, chat.ErrTaskRequestUnavailable},
		{"a task deleted under the panel", http.StatusConflict, `{"error":"that task was deleted"}`, chat.ErrTaskNotFound},
		{"a pick the task moved past", http.StatusConflict,
			`{"code":"stale_choice","error":"that choice is no longer current"}`, chat.ErrTaskRequestStaleChoice},
		{"a used id with different words", http.StatusConflict,
			`{"code":"conflict","error":"request id already used"}`, chat.ErrTaskRequestConflict},
		{"a session type with no work records", http.StatusNotImplemented, `{"error":"no"}`, chat.ErrTaskRequestUnsupported},
		{"a control plane without the write route", http.StatusMethodNotAllowed, `{"error":"no"}`, chat.ErrTaskRequestUnsupported},
		{"fields the platform refused", http.StatusBadRequest, `{"error":"no"}`, chat.ErrTaskRequestInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, engine, platform := requestEngine(t)
			bindSession(t, store, 1, "sess-1")
			platform.tasks["sess-1"] = map[string]bool{"t1": true}
			platform.status, platform.body = tc.status, tc.body
			_, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "req-1", "말", 0)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

// What the app sends wrong is refused here, before the hop: a malformed field
// should cost no call into the pod, and the message should be about the field
// rather than about the hop.
func TestCreateTaskRequest_RefusesBeforeTheHop(t *testing.T) {
	store, engine, platform := requestEngine(t)
	bindSession(t, store, 1, "sess-1")
	platform.tasks["sess-1"] = map[string]bool{"t1": true}

	for _, tc := range []struct {
		name              string
		taskID, requestID string
		text              string
		choiceTurnID      int64
	}{
		{"empty words", "t1", "r1", "   ", 0},
		{"words past the bound", "t1", "r1", strings.Repeat("가", chat.TaskRequestTextMax+1), 0},
		{"an id that is not an id", "t1", "has spaces", "말", 0},
		{"a task id that is not an id", "has spaces", "r1", "말", 0},
		{"a choice turn that is not a turn", "t1", "r1", "말", -1},
	} {
		if _, err := engine.CreateTaskRequest(
			context.Background(), 1, tc.taskID, tc.requestID, tc.text, tc.choiceTurnID,
		); !errors.Is(err, chat.ErrTaskRequestInvalid) {
			t.Errorf("%s: err = %v, want the invalid class", tc.name, err)
		}
	}
	if len(platform.calls) != 0 {
		t.Fatalf("a malformed request reached the platform: %v", platform.calls)
	}
}

// A conversation no session has answered yet has no task to leave a request on,
// and saying so is not the same as the request being wrong.
func TestCreateTaskRequest_NeedsASession(t *testing.T) {
	_, engine, platform := requestEngine(t)
	_, err := engine.CreateTaskRequest(context.Background(), 1, "t1", "r1", "말", 0)
	if !errors.Is(err, chat.ErrNoConversation) {
		t.Fatalf("err = %v, want the no-conversation class", err)
	}
	if len(platform.calls) != 0 {
		t.Fatalf("the platform was called for a conversation with no session: %v", platform.calls)
	}
}
