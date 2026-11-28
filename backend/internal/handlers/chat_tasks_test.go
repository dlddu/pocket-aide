package handlers_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// tasksBody is the panel's payload as the app decodes it. It is declared here
// rather than reused from the chat package so that a field quietly leaving the
// response is a failing test rather than a shared struct change.
type tasksBody struct {
	SessionID string `json:"session_id"`
	Tasks     []struct {
		ID                 string `json:"id"`
		Title              string `json:"title"`
		Status             string `json:"status"`
		Description        string `json:"description"`
		CompletionCriteria string `json:"completion_criteria"`
		TurnsTotal         int64  `json:"turns_total"`
		Migrated           bool   `json:"migrated"`
		Waiting            *struct {
			TurnID   int64  `json:"turn_id"`
			Request  string `json:"request"`
			Question string `json:"question"`
			Running  bool   `json:"running"`
		} `json:"waiting"`
		Choices []struct {
			Label       string `json:"label"`
			RequestText string `json:"request_text"`
		} `json:"choices"`
		ChoiceTurnID int64 `json:"choice_turn_id"`
		// What the turn that offered those choices said, which the pane shows
		// directly above them. There is no response here on purpose — the
		// session's verbatim answer is the long field and belongs to the
		// exchanges, which are read only when somebody opens them.
		ChoiceTurn *struct {
			Request     string `json:"request"`
			Explanation string `json:"explanation"`
			Question    string `json:"question"`
			Response    string `json:"response"`
		} `json:"choice_turn"`
	} `json:"tasks"`
	Counts    map[string]int `json:"counts"`
	Truncated bool           `json:"truncated"`
	Code      string         `json:"code"`
}

func getTasks(t *testing.T, router http.Handler, userID int64) (int, tasksBody) {
	t.Helper()
	rec := chatDo(t, router, http.MethodGet, "/api/chat/tasks", nil, userID)
	var out tasksBody
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode %s: %v", rec.Body.String(), err)
		}
	}
	return rec.Code, out
}

// openSession gives the caller's conversation a platform session by sending one
// message, which is what opens it.
func openSession(t *testing.T, router http.Handler, userID int64) {
	t.Helper()
	if rec := chatDo(t, router, http.MethodPost, "/api/chat/messages", map[string]string{"content": "안녕"}, userID); rec.Code != http.StatusAccepted {
		t.Fatalf("open session: %d %s", rec.Code, rec.Body)
	}
}

// A conversation with no session yet has no tasks — 200 with an empty list and
// zeroed counts, which is the panel's "nothing yet" state.
func TestGetChatTasks_NoSessionIsAnEmptyList(t *testing.T) {
	router, _, _ := newChatRouter(t)
	status, body := getTasks(t, router, 1)
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	if len(body.Tasks) != 0 || body.SessionID != "" || body.Truncated {
		t.Fatalf("body = %+v", body)
	}
	if body.Counts["blocked"] != 0 || len(body.Counts) == 0 {
		t.Fatalf("counts = %v", body.Counts)
	}
}

// The agreed task and turn records reach the panel end to end, through the
// route the app actually calls. The ledger page this used to read could carry
// none of the description, the completion criteria, or the turn the choices
// belong to — a panel built on it could say work existed but not what it was.
func TestGetChatTasks_ServesTheWorkPage(t *testing.T) {
	router, platform, _ := newChatRouter(t)
	openSession(t, router, 1)
	platform.mu.Lock()
	platform.work = `{"tasks":[
      {"task":{"id":"t1","title":"작업","status":"blocked",
        "description":"무엇을 why","completion_criteria":"머지되면 끝"},
       "turns_total":2,"migrated":true,
       "latest_turn":{"id":7,"request":"이거 해 줘","question":"어디까지 할까요?",
         "explanation":"앞의 절반까지 했어요.","response":"세션이 스스로 쓴 아주 긴 원문 답",
         "choices":[{"label":"전부","request_text":"전부 해 줘"}],
         "started_at":null,"ended_at":null}}
    ],"next_cursor":""}`
	platform.mu.Unlock()

	status, body := getTasks(t, router, 1)
	if status != http.StatusOK {
		t.Fatalf("status = %d %+v", status, body)
	}
	if len(body.Tasks) != 1 || body.Tasks[0].ID != "t1" {
		t.Fatalf("tasks = %+v", body.Tasks)
	}
	got := body.Tasks[0]
	if got.Status != "blocked" || got.Description != "무엇을 why" || got.CompletionCriteria != "머지되면 끝" {
		t.Fatalf("the agreed task fields did not reach the panel: %+v", got)
	}
	if got.TurnsTotal != 2 || !got.Migrated {
		t.Fatalf("turn count or carried-over flag lost: %+v", got)
	}
	// What is waiting comes from the latest turn, and the choices carry the
	// turn they were offered on — which is what a pick is checked against.
	if got.Waiting == nil || got.Waiting.TurnID != 7 || got.Waiting.Question != "어디까지 할까요?" {
		t.Fatalf("waiting = %+v", got.Waiting)
	}
	// The words the person sent stay in this answer. The detail pane now draws
	// them from the task's exchanges instead, but this route has always carried
	// them and something we cannot see may still read them — so moving where
	// the pane reads from does not get to change what the route answers.
	if got.Waiting.Request != "이거 해 줘" {
		t.Fatalf("the task list stopped carrying the request it has always carried: %q", got.Waiting.Request)
	}
	if len(got.Choices) != 1 || got.ChoiceTurnID != 7 {
		t.Fatalf("choices = %+v turn=%d", got.Choices, got.ChoiceTurnID)
	}
	// The words the choices answer come down with them, from that same turn, so
	// the pane can put them above the choices without reading the task's
	// exchanges — which start closed at the bottom of the pane on purpose.
	if got.ChoiceTurn == nil {
		t.Fatal("the turn that offered the choices sent none of its own words")
	}
	if got.ChoiceTurn.Request != "이거 해 줘" || got.ChoiceTurn.Explanation != "앞의 절반까지 했어요." ||
		got.ChoiceTurn.Question != "어디까지 할까요?" {
		t.Fatalf("choice_turn = %+v", got.ChoiceTurn)
	}
	// The session's own long answer is not in it, and is nowhere else in the
	// payload either: the panel polls this route, and the answer belongs to the
	// exchanges a person opens.
	if got.ChoiceTurn.Response != "" {
		t.Fatalf("the choice turn carried the session's verbatim answer: %q", got.ChoiceTurn.Response)
	}
	if body.SessionID == "" {
		t.Fatal("the answer must name the session it read")
	}
}

// The panel's three "cannot show a list" cases are separate statuses, so none
// of them can be mistaken for "there is no work".
func TestGetChatTasks_DistinguishesUnsupportedAndUnavailable(t *testing.T) {
	for _, tc := range []struct {
		name     string
		platform int
		want     int
		code     string
	}{
		{"this session keeps no work records", http.StatusNotImplemented, http.StatusNotImplemented, "task_ledger_unsupported"},
		{"no such route or session", http.StatusNotFound, http.StatusNotImplemented, "task_ledger_unsupported"},
		{"the session is not awake", http.StatusConflict, http.StatusConflict, "task_ledger_unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router, platform, _ := newChatRouter(t)
			openSession(t, router, 1)
			platform.mu.Lock()
			platform.workStatus = tc.platform
			platform.mu.Unlock()

			status, body := getTasks(t, router, 1)
			if status != tc.want {
				t.Fatalf("status = %d, want %d (%+v)", status, tc.want, body)
			}
			if body.Code != tc.code {
				t.Fatalf("code = %q, want %q", body.Code, tc.code)
			}
			if len(body.Tasks) != 0 {
				t.Fatalf("unreadable work records must not answer tasks: %+v", body.Tasks)
			}
		})
	}
}

// The route is the caller's own conversation. Another user reading at the same
// time gets its own session's answer, never this one's.
func TestGetChatTasks_IsTheCallersOwnConversation(t *testing.T) {
	router, platform, _ := newChatRouter(t)
	openSession(t, router, 1)
	platform.mu.Lock()
	platform.work = `{"tasks":[{"task":{"id":"mine","title":"내 것","status":"running"}}],"next_cursor":""}`
	platform.mu.Unlock()

	_, mine := getTasks(t, router, 1)
	if len(mine.Tasks) != 1 || mine.Tasks[0].ID != "mine" {
		t.Fatalf("user 1 = %+v", mine)
	}
	// User 2 never opened a conversation, so it has no session and no tasks —
	// it does not inherit user 1's.
	status, theirs := getTasks(t, router, 2)
	if status != http.StatusOK || len(theirs.Tasks) != 0 || theirs.SessionID == mine.SessionID {
		t.Fatalf("user 2 = %d %+v (user 1 session %q)", status, theirs, mine.SessionID)
	}
}

// turnsBody is the detail pane's page as the app decodes it.
type turnsBody struct {
	TaskID string `json:"task_id"`
	Turns  []struct {
		ID int64 `json:"id"`
	} `json:"turns"`
	NextCursor int64  `json:"next_cursor"`
	Order      string `json:"order"`
	Code       string `json:"code"`
}

// The pane reads the latest exchanges first. before_turn_id — present even with
// no value — is what says so, and it reaches the platform that way: without it
// the slot answers the task's *oldest* turns, which is the opposite of what the
// pane shows.
func TestGetChatTaskTurns_ReadsTheNewestPageAndPagesBack(t *testing.T) {
	router, platform, _ := newChatRouter(t)
	openSession(t, router, 1)
	platform.mu.Lock()
	platform.work = `{"task_id":"t1","task_version":3,"task":{"id":"t1","title":"t","status":"blocked"},
      "turns":[{"id":9},{"id":7}],"next_cursor":7,"order":"desc"}`
	platform.workQueries = nil
	platform.mu.Unlock()

	read := func(query string) (int, turnsBody) {
		t.Helper()
		rec := chatDo(t, router, http.MethodGet, "/api/chat/tasks/t1/turns"+query, nil, 1)
		var out turnsBody
		if rec.Body.Len() > 0 {
			_ = json.Unmarshal(rec.Body.Bytes(), &out)
		}
		return rec.Code, out
	}

	code, page := read("?before_turn_id=&limit=2")
	if code != http.StatusOK || page.Order != "desc" || page.NextCursor != 7 {
		t.Fatalf("newest page = %d %+v", code, page)
	}
	if len(page.Turns) != 2 || page.Turns[0].ID != 9 {
		t.Fatalf("turns = %+v", page.Turns)
	}
	if code, _ = read("?before_turn_id=7&limit=2"); code != http.StatusOK {
		t.Fatalf("paging back = %d", code)
	}
	platform.mu.Lock()
	asked := append([]string{}, platform.workQueries...)
	platform.mu.Unlock()
	if len(asked) != 2 {
		t.Fatalf("work queries = %v", asked)
	}
	// An empty value still goes out as a bound of its own: on the platform's
	// route the parameter's presence is what picks the end.
	if !strings.Contains(asked[0], "before_turn_id=0") || strings.Contains(asked[0], "after_turn_id") {
		t.Fatalf("first page asked %q", asked[0])
	}
	if !strings.Contains(asked[1], "before_turn_id=7") {
		t.Fatalf("second page asked %q", asked[1])
	}

	// Naming both ends is refused rather than resolved, and a bad bound is a
	// refusal rather than a page from the wrong end.
	for _, query := range []string{
		"?after_turn_id=1&before_turn_id=2", "?before_turn_id=-1", "?before_turn_id=nope",
	} {
		if code, body := read(query); code != http.StatusBadRequest || body.Code != "task_request_invalid" {
			t.Fatalf("%s = %d %+v", query, code, body)
		}
	}
}

// A slot that does not have the parameter answers the oldest turns and says
// nothing about order. The pane must not be handed that page — it would show
// the task's first exchange as its latest — so it is its own status and its own
// code, distinct from "this session keeps no work records".
func TestGetChatTaskTurns_APageFromTheWrongEndIsItsOwnAnswer(t *testing.T) {
	router, platform, _ := newChatRouter(t)
	openSession(t, router, 1)
	platform.mu.Lock()
	platform.work = `{"task_id":"t1","task_version":3,"task":{"id":"t1","title":"t","status":"blocked"},
      "turns":[{"id":1},{"id":2}],"next_cursor":2}`
	platform.mu.Unlock()

	rec := chatDo(t, router, http.MethodGet, "/api/chat/tasks/t1/turns?before_turn_id=", nil, 1)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("wrong-end page = %d %s", rec.Code, rec.Body)
	}
	var body turnsBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Code != "task_turns_order_unsupported" {
		t.Fatalf("code = %q", body.Code)
	}
	if len(body.Turns) != 0 {
		t.Fatalf("a refused page still carried turns: %+v", body.Turns)
	}
}
