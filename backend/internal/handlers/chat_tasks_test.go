package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type tasksBody struct {
	SessionID string `json:"session_id"`
	Tasks     []struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		Stage      string `json:"stage"`
		Reason     string `json:"reason"`
		NextAction string `json:"next_action"`
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

// The ledger's rows reach the panel with status, stage, reason and next action.
func TestGetChatTasks_ServesTheLedgerPage(t *testing.T) {
	router, platform, _ := newChatRouter(t)
	openSession(t, router, 1)
	platform.mu.Lock()
	platform.ledger = `{"tasks":[
      {"id":"t1","task":{"title":"작업","status":"blocked","stage":"review","sessions":[
        {"cleanup":"deleted","reason":"옛 사유","next_action":"옛 행동"},
        {"cleanup":"retained","reason":"지금 사유","next_action":"지금 행동"}],"details":{}}},
      {"id":"m1","task":{"title":"감시","status":"running","stage":"watch","sessions":[],"details":{"kind":"monitor"}}}
    ],"next_cursor":""}`
	platform.mu.Unlock()

	status, body := getTasks(t, router, 1)
	if status != http.StatusOK {
		t.Fatalf("status = %d %+v", status, body)
	}
	if len(body.Tasks) != 1 || body.Tasks[0].ID != "t1" {
		t.Fatalf("the monitor row must be excluded: %+v", body.Tasks)
	}
	got := body.Tasks[0]
	if got.Status != "blocked" || got.Stage != "review" {
		t.Fatalf("status/stage = %+v", got)
	}
	if got.Reason != "지금 사유" || got.NextAction != "지금 행동" {
		t.Fatalf("the current state's note must win: %+v", got)
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
		{"this session keeps no ledger", http.StatusNotImplemented, http.StatusNotImplemented, "task_ledger_unsupported"},
		{"no such route or session", http.StatusNotFound, http.StatusNotImplemented, "task_ledger_unsupported"},
		{"the session is not awake", http.StatusConflict, http.StatusConflict, "task_ledger_unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router, platform, _ := newChatRouter(t)
			openSession(t, router, 1)
			platform.mu.Lock()
			platform.ledgerStatus = tc.platform
			platform.mu.Unlock()

			status, body := getTasks(t, router, 1)
			if status != tc.want {
				t.Fatalf("status = %d, want %d (%+v)", status, tc.want, body)
			}
			if body.Code != tc.code {
				t.Fatalf("code = %q, want %q", body.Code, tc.code)
			}
			if len(body.Tasks) != 0 {
				t.Fatalf("an unreadable ledger must not answer tasks: %+v", body.Tasks)
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
	platform.ledger = `{"tasks":[{"id":"mine","task":{"title":"내 것","status":"running","stage":"impl","sessions":[],"details":{}}}],"next_cursor":""}`
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
