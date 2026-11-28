package handlers_test

import (
	"encoding/json"
	"net/http"
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
