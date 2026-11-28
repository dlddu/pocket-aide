package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/notificationhistory"
	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

// materialsRouter wires the review-materials route onto the real chat engine,
// the real notification history store and a platform that answers the
// read-only work turns route. The history store is the real one so the title
// lookup's SQL is exercised rather than stubbed — a title that does not come
// out of the table is the whole point of the field.
func materialsRouter(t *testing.T, turnsPage string) (http.Handler, *notificationhistory.Store, []string) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	asked := []string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1")
		asked = append(asked, path)
		switch {
		case path == "/sessions" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"sessions":[]}`))
		case strings.HasSuffix(path, "/turns"):
			_, _ = w.Write([]byte(turnsPage))
		default:
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	t.Cleanup(srv.Close)
	store := chat.New(conn)
	engine := chat.NewEngine(store, sessions.NewClient(srv.URL), nil)
	t.Cleanup(engine.Wait)
	conv, err := store.Conversation(context.Background(), 1)
	if err != nil {
		t.Fatalf("conversation: %v", err)
	}
	if err := store.SetPlatform(context.Background(), 1, conv.ID, "worker-1", 0); err != nil {
		t.Fatalf("bind session: %v", err)
	}
	history := notificationhistory.New(conn)
	r := chi.NewRouter()
	r.Get("/api/chat/tasks/{taskID}/materials", handlers.GetChatTaskMaterials(engine, history))
	return r, history, asked
}

// materialsView is the route's answer as the app reads it.
type materialsView struct {
	TaskID    string `json:"task_id"`
	SessionID string `json:"session_id"`
	Artifacts []struct {
		Ref    string `json:"ref"`
		URL    string `json:"url"`
		Title  string `json:"title"`
		Repo   string `json:"repo"`
		Number int    `json:"number"`
		TurnID int64  `json:"turn_id"`
	} `json:"artifacts"`
	Decisions []struct {
		TurnID          int64  `json:"turn_id"`
		OfferedOnTurnID int64  `json:"offered_on_turn_id"`
		Label           string `json:"label"`
		Text            string `json:"text"`
	} `json:"decisions"`
	Open []struct {
		TurnID   int64  `json:"turn_id"`
		Question string `json:"question"`
		Choices  []struct {
			Label       string `json:"label"`
			RequestText string `json:"request_text"`
		} `json:"choices"`
	} `json:"open"`
	TradeoffsRecorded bool  `json:"tradeoffs_recorded"`
	FoldedThrough     int64 `json:"folded_through"`
	Truncated         bool  `json:"truncated"`
}

// The task under review: an earlier turn produced two pull requests and a pick
// was sent, and the newest turn is a short follow-up asking the next question.
const materialsTurnsPage = `{
  "task_id": "t1", "task_version": 7,
  "task": {"id": "t1", "title": "검토 대상", "status": "review", "version": 7},
  "turns": [
    {"id": 1, "request": "구현해 줘", "response": "다 했습니다",
     "explanation": "두 PR 로 올렸어요", "question": "둘 다 머지할까요?",
     "choices": [{"label": "둘 다 머지", "request_text": "둘 다 머지해 주세요"}],
     "artifacts": ["https://github.com/dlddu/pocket-aide/pull/425",
                   "https://github.com/dlddu/pocket-aide/pull/426",
                   "pull/42"],
     "requested_at": "2026-11-20T00:00:00Z", "ended_at": "2026-11-20T01:00:00Z"},
    {"id": 2, "request": "둘 다 머지해 주세요", "choices": [], "artifacts": [],
     "requested_at": "2026-11-20T02:00:00Z", "ended_at": "2026-11-20T03:00:00Z",
     "origin": "app_request"},
    {"id": 3, "request": "수정해 주세요", "question": "어느 쪽을 고칠까요?",
     "choices": [{"label": "앱만", "request_text": "앱만 고쳐 주세요"}],
     "artifacts": [], "requested_at": "2026-11-20T04:00:00Z", "origin": "app_request"}
  ]
}`

func getMaterials(t *testing.T, router http.Handler, taskID string) (*httptest.ResponseRecorder, materialsView) {
	t.Helper()
	rec := chatDo(t, router, http.MethodGet, "/api/chat/tasks/"+taskID+"/materials", nil, 1)
	var view materialsView
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &view); err != nil {
			t.Fatalf("decode: %v (%s)", err, rec.Body.String())
		}
	}
	return rec, view
}

// The producer and the consumer meet here: the route answers the pull requests
// of a task whose latest turn produced none, with the title this user's own
// notifications already recorded for the one that has it.
func TestChatTaskMaterialsAnswersThePullRequestsWithRecordedTitles(t *testing.T) {
	router, history, _ := materialsRouter(t, materialsTurnsPage)
	// One of the two pull requests has already sent this user a CI
	// notification, so its title is recorded; the other has not.
	if _, err := history.InsertBatchTx(context.Background(), []int64{1}, notificationhistory.Event{
		RepoFullName: "dlddu/pocket-aide",
		PRNumber:     425,
		PRTitle:      "채팅: 작업 상세의 「주고받은 내용」을 맨 아래로 두고 펼칠 때 읽는다",
		PRURL:        "https://github.com/dlddu/pocket-aide/pull/425",
		HeadSHA:      "ca0e9ea",
		Conclusion:   "success",
	}); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	rec, view := getMaterials(t, router, "t1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	if view.TaskID != "t1" || view.SessionID != "worker-1" {
		t.Errorf("view names = %q/%q", view.TaskID, view.SessionID)
	}
	if len(view.Artifacts) != 3 {
		t.Fatalf("artifacts = %+v, want all three of turn 1's", view.Artifacts)
	}
	byRef := map[string]string{}
	for _, a := range view.Artifacts {
		byRef[a.Ref] = a.Title
		if a.TurnID != 1 {
			t.Errorf("artifact %q turn = %d, want the turn that produced it", a.Ref, a.TurnID)
		}
	}
	// The recorded title is shown for the pull request the row names.
	if got := byRef["https://github.com/dlddu/pocket-aide/pull/425"]; !strings.Contains(got, "주고받은 내용") {
		t.Errorf("title for 425 = %q, want the recorded one", got)
	}
	// The one with no recorded title stays an openable link with no title
	// invented for it.
	if got := byRef["https://github.com/dlddu/pocket-aide/pull/426"]; got != "" {
		t.Errorf("title for 426 = %q, want none: no row names it", got)
	}
	// And the bare ref is neither titled nor linked — nothing guessed a
	// repository for it from the rows that exist.
	if got := byRef["pull/42"]; got != "" {
		t.Errorf("title for the bare ref = %q, want none", got)
	}
	for _, a := range view.Artifacts {
		if a.Ref == "pull/42" && (a.URL != "" || a.Repo != "") {
			t.Errorf("bare ref = %+v, want no link and no repository", a)
		}
	}
}

// The three conditions arrive as three different things, and the one the
// records cannot hold says so.
func TestChatTaskMaterialsKeepsDecidedAndOpenApart(t *testing.T) {
	router, _, _ := materialsRouter(t, materialsTurnsPage)
	rec, view := getMaterials(t, router, "t1")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if len(view.Decisions) != 1 || view.Decisions[0].Label != "둘 다 머지" {
		t.Fatalf("decisions = %+v, want the pick that was sent", view.Decisions)
	}
	if view.Decisions[0].TurnID != 2 || view.Decisions[0].OfferedOnTurnID != 1 {
		t.Errorf("decision turns = %+v", view.Decisions[0])
	}
	if len(view.Open) != 1 || view.Open[0].TurnID != 3 {
		t.Fatalf("open = %+v, want only the latest turn's", view.Open)
	}
	if view.Open[0].Question != "어느 쪽을 고칠까요?" {
		t.Errorf("open question = %q", view.Open[0].Question)
	}
	if len(view.Open[0].Choices) != 1 || view.Open[0].Choices[0].RequestText != "앱만 고쳐 주세요" {
		t.Errorf("open choices = %+v, want the sentence to send verbatim", view.Open[0].Choices)
	}
	if view.TradeoffsRecorded {
		t.Error("the agreed turn keeps no advantages and drawbacks")
	}
	if view.FoldedThrough != 3 {
		t.Errorf("foldedThrough = %d, want 3", view.FoldedThrough)
	}
}

// A task with no turns has no materials, and says so as an empty answer rather
// than as a refusal.
func TestChatTaskMaterialsAnswersEmptyForATaskWithNoTurns(t *testing.T) {
	router, _, _ := materialsRouter(t, `{"task_id":"t9","turns":[]}`)
	rec, view := getMaterials(t, router, "t9")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body.String())
	}
	if len(view.Artifacts) != 0 || len(view.Decisions) != 0 || len(view.Open) != 0 {
		t.Errorf("view = %+v, want everything empty", view)
	}
	// Empty arrays rather than nulls, so the app draws "none recorded" without
	// having to tell a missing field from an empty one.
	body := rec.Body.String()
	for _, want := range []string{`"artifacts":[]`, `"decisions":[]`, `"open":[]`} {
		if !strings.Contains(body, want) {
			t.Errorf("body %s is missing %s", body, want)
		}
	}
}

// An id the slot could never hold is refused before anything is read.
func TestChatTaskMaterialsRefusesAnImpossibleTaskID(t *testing.T) {
	router, _, asked := materialsRouter(t, materialsTurnsPage)
	rec := chatDo(t, router, http.MethodGet, "/api/chat/tasks/not%20an%20id!/materials", nil, 1)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	for _, path := range asked {
		if strings.Contains(path, "/turns") {
			t.Errorf("read %q for an id it had already refused", path)
		}
	}
}
