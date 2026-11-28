package chat_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// materialsPlatform answers the work-turns route a page at a time, so a test
// can show what the fold asked for as well as what it answered with. The pages
// are keyed by the after_turn_id the fold sends, which is how the keyset walk
// is observable: a fold that resumed asks for a cursor, and a fold that started
// over asks for none.
type materialsPlatform struct {
	pages  map[string]string
	calls  []string
	status int
	body   string
}

func newMaterialsPlatform() *materialsPlatform {
	return &materialsPlatform{pages: map[string]string{}}
}

// page registers the answer for one cursor. after 0 is the first page.
func (p *materialsPlatform) page(sessionID, taskID string, after int64, body string) {
	p.pages[fmt.Sprintf("%s/%s/%d", sessionID, taskID, after)] = body
}

func (p *materialsPlatform) Do(_ context.Context, method, path string, _ []byte) (int, []byte, error) {
	p.calls = append(p.calls, method+" "+path)
	if p.status != 0 {
		return p.status, []byte(p.body), nil
	}
	if method != http.MethodGet || !strings.Contains(path, "/work/tasks/") {
		return http.StatusNotFound, nil, nil
	}
	trimmed := strings.TrimPrefix(path, "/sessions/")
	sid := trimmed[:strings.Index(trimmed, "/")]
	rest := trimmed[strings.Index(trimmed, "/work/tasks/")+len("/work/tasks/"):]
	taskID := rest[:strings.Index(rest, "/")]
	after := int64(0)
	if cut := strings.Index(path, "after_turn_id="); cut >= 0 {
		_, _ = fmt.Sscanf(path[cut+len("after_turn_id="):], "%d", &after)
	}
	body, ok := p.pages[fmt.Sprintf("%s/%s/%d", sid, taskID, after)]
	if !ok {
		return http.StatusOK, []byte(`{"task_id":"x","turns":[]}`), nil
	}
	return http.StatusOK, []byte(body), nil
}

func materialsEngine(t *testing.T) (*chat.Store, *chat.Engine, *materialsPlatform) {
	t.Helper()
	store, _, _ := workEngine(t)
	platform := newMaterialsPlatform()
	engine := chat.NewEngine(store, platform, nil)
	t.Cleanup(engine.Wait)
	return store, engine, platform
}

// reviewedTaskTurns is the case the whole feature exists for: two earlier turns
// produced pull requests and settled a question, and the newest turn is a short
// follow-up that produced nothing and offers the next question.
//
// Turn 1 produced two pull requests. Turn 2's request is word for word the
// sentence turn 1 offered, which is a decision somebody took. Turn 3 is
// "수정해 주세요" and names no artifact — so a screen that showed only the latest
// turn would show no pull requests at all for a task whose review is entirely
// about two of them.
const reviewedTaskTurns = `{
  "task_id": "t1", "task_version": 7,
  "task": {"id": "t1", "title": "검토 대상", "status": "review", "version": 7},
  "turns": [
    {"id": 1, "request": "구현해 줘", "instruction": "구현하라",
     "response": "다 했습니다", "explanation": "두 PR 로 나눠 올렸어요",
     "question": "둘 다 머지할까요?",
     "choices": [{"label": "둘 다 머지", "request_text": "둘 다 머지해 주세요"},
                 {"label": "백엔드만", "request_text": "백엔드만 머지해 주세요"}],
     "artifacts": ["https://github.com/dlddu/pocket-aide/pull/425",
                   "https://github.com/dlddu/pocket-aide/pull/426"],
     "requested_at": "2026-11-20T00:00:00Z", "started_at": "2026-11-20T00:01:00Z",
     "ended_at": "2026-11-20T01:00:00Z"},
    {"id": 2, "request": "둘 다 머지해 주세요", "instruction": "둘 다 머지하라",
     "response": "머지했습니다", "explanation": "둘 다 들어갔어요",
     "question": null, "choices": [], "artifacts": [],
     "requested_at": "2026-11-20T02:00:00Z", "started_at": "2026-11-20T02:01:00Z",
     "ended_at": "2026-11-20T03:00:00Z", "origin": "app_request"},
    {"id": 3, "request": "수정해 주세요", "instruction": null, "response": null,
     "explanation": null, "question": "어느 쪽을 고칠까요?",
     "choices": [{"label": "앱만", "request_text": "앱만 고쳐 주세요"}],
     "artifacts": [],
     "requested_at": "2026-11-20T04:00:00Z", "started_at": null, "ended_at": null,
     "origin": "app_request"}
  ]
}`

// The headline case: the pull requests of earlier turns survive a latest turn
// that produced none, and they arrive as openable links with the repository and
// number read out of the recorded ref itself.
func TestTaskMaterialsKeepsEarlierPullRequestsWhenTheLatestTurnIsAShortFollowUp(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, reviewedTaskTurns)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	if len(got.Artifacts) != 2 {
		t.Fatalf("artifacts = %+v, want the two pull requests of turn 1", got.Artifacts)
	}
	// Newest sighting first, and both from the turn that actually produced them
	// rather than from the turn that happens to be latest.
	if got.Artifacts[0].Ref != "https://github.com/dlddu/pocket-aide/pull/426" {
		t.Errorf("first artifact = %q, want the newer of turn 1's two", got.Artifacts[0].Ref)
	}
	for _, a := range got.Artifacts {
		if a.TurnID != 1 {
			t.Errorf("artifact %q turn = %d, want 1", a.Ref, a.TurnID)
		}
		if a.URL != a.Ref {
			t.Errorf("artifact %q url = %q, want the ref itself", a.Ref, a.URL)
		}
		if a.Repo != "dlddu/pocket-aide" {
			t.Errorf("artifact %q repo = %q", a.Ref, a.Repo)
		}
	}
	if got.Artifacts[0].Number != 426 || got.Artifacts[1].Number != 425 {
		t.Errorf("numbers = %d,%d, want 426,425", got.Artifacts[0].Number, got.Artifacts[1].Number)
	}
	if got.FoldedThrough != 3 {
		t.Errorf("foldedThrough = %d, want 3", got.FoldedThrough)
	}
	if got.Truncated {
		t.Error("a task whose turns all fit is not truncated")
	}
}

// The three conditions are three different things on the same task, and each is
// read from where it is actually recorded.
func TestTaskMaterialsSeparatesDecisionsFromWhatIsStillOpen(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, reviewedTaskTurns)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	// Decided: the sentence turn 1 offered came back as turn 2's request. The
	// label is what the button said, so a person reads the words they picked.
	if len(got.Decisions) != 1 {
		t.Fatalf("decisions = %+v, want the one pick that was sent", got.Decisions)
	}
	decided := got.Decisions[0]
	if decided.Label != "둘 다 머지" || decided.Text != "둘 다 머지해 주세요" {
		t.Errorf("decision = %+v, want the offered sentence that was sent", decided)
	}
	if decided.TurnID != 2 || decided.OfferedOnTurnID != 1 {
		t.Errorf("decision turns = %d answering %d, want 2 answering 1",
			decided.TurnID, decided.OfferedOnTurnID)
	}
	// Open: the latest turn's question only. Turn 1 asked one too, and it was
	// answered — offering it as open would ask for a decision already taken.
	if len(got.Open) != 1 {
		t.Fatalf("open = %+v, want only the latest turn's question", got.Open)
	}
	open := got.Open[0]
	if open.TurnID != 3 || open.Question != "어느 쪽을 고칠까요?" {
		t.Errorf("open = %+v, want turn 3's question", open)
	}
	if len(open.Choices) != 1 || open.Choices[0].Label != "앱만" {
		t.Errorf("open choices = %+v, want turn 3's own", open.Choices)
	}
	// And the one thing the records cannot say is said to be unrecorded rather
	// than left to look like "there were no trade-offs".
	if got.TradeoffsRecorded {
		t.Error("the agreed turn has no field for advantages and drawbacks")
	}
}

// A generic explanation and a choice label are not a decision. A task whose
// turns hold both but no pick has no decided material, and says so.
func TestTaskMaterialsDoesNotReadAnExplanationOrALabelAsADecision(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 1, "request": "해 줘", "explanation": "A 로 가는 게 좋겠습니다. 그렇게 했습니다.",
	     "question": "A 로 할까요?",
	     "choices": [{"label": "A 로 간다", "request_text": "A 로 해 주세요"}],
	     "artifacts": [], "requested_at": "2026-11-20T00:00:00Z",
	     "ended_at": "2026-11-20T01:00:00Z"},
	    {"id": 2, "request": "B 가 더 낫지 않나요?", "artifacts": [], "choices": [],
	     "requested_at": "2026-11-20T02:00:00Z"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	// Turn 2's request is the person's own words, not the offered sentence, so
	// it decided nothing between recorded alternatives.
	if len(got.Decisions) != 0 {
		t.Errorf("decisions = %+v, want none: free text is not a pick", got.Decisions)
	}
	if len(got.Artifacts) != 0 {
		t.Errorf("artifacts = %+v, want none", got.Artifacts)
	}
	// Turn 2 asked nothing and offers nothing, so there is no open point either
	// — an empty answer a person can read at once.
	if len(got.Open) != 0 {
		t.Errorf("open = %+v, want none", got.Open)
	}
}

// Opening a task again folds only the turns that have arrived since. This is
// what keeps the review answer from being a re-read of the whole history every
// time somebody steps into a task.
func TestTaskMaterialsResumesInsteadOfRereadingTheWholeHistory(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, reviewedTaskTurns)

	if _, err := engine.TaskMaterials(context.Background(), 1, "t1"); err != nil {
		t.Fatalf("first: %v", err)
	}
	first := len(platform.calls)
	// A fourth turn arrives, naming one more pull request and finishing the
	// question turn 3 asked.
	platform.page("worker-1", "t1", 3, `{
	  "task_id": "t1", "turns": [
	    {"id": 4, "request": "앱만 고쳐 주세요", "explanation": "앱만 고쳤어요",
	     "question": null, "choices": [],
	     "artifacts": ["https://github.com/dlddu/pocket-aide/pull/430"],
	     "requested_at": "2026-11-20T05:00:00Z", "ended_at": "2026-11-20T06:00:00Z",
	     "origin": "app_request"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	// The resumed read asked for the turns after 3 — not for the task's first
	// page again.
	resumed := platform.calls[first:]
	if len(resumed) != 1 {
		t.Fatalf("resumed calls = %v, want one page", resumed)
	}
	if !strings.Contains(resumed[0], "after_turn_id=3") {
		t.Errorf("resumed call = %q, want it to resume after turn 3", resumed[0])
	}
	// Everything earlier is still there, with the new one in front.
	if len(got.Artifacts) != 3 || got.Artifacts[0].Number != 430 {
		t.Fatalf("artifacts = %+v, want the new one in front of the kept two", got.Artifacts)
	}
	// Turn 4 sent the sentence turn 3 offered, so the pick made across two
	// separate reads is recognised just the same.
	if len(got.Decisions) != 2 || got.Decisions[1].TurnID != 4 {
		t.Fatalf("decisions = %+v, want turn 4's pick after turn 2's", got.Decisions)
	}
	if got.Decisions[1].OfferedOnTurnID != 3 || got.Decisions[1].Label != "앱만" {
		t.Errorf("decision = %+v, want the sentence turn 3 offered", got.Decisions[1])
	}
	// Turn 4 asks nothing, so nothing is open any more — the open points are
	// taken from the tail each time rather than kept.
	if len(got.Open) != 0 {
		t.Errorf("open = %+v, want none once the question was answered", got.Open)
	}
}

// A ref that names no host is shown as the identifier it is. Nothing guesses a
// repository for it, because a wrong one would send somebody to a pull request
// this task never mentioned.
func TestTaskMaterialsLeavesARefWithoutAHostAsAnIdentifier(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 1, "request": "해 줘", "choices": [],
	     "artifacts": ["pull/42", "https://github.com/dlddu/pocket-aide/pull/425/files",
	                   "  https://github.com/dlddu/pocket-aide/pull/425  "],
	     "requested_at": "2026-11-20T00:00:00Z", "ended_at": "2026-11-20T01:00:00Z"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	if len(got.Artifacts) != 3 {
		t.Fatalf("artifacts = %+v, want all three refs", got.Artifacts)
	}
	byRef := map[string]chat.TaskArtifact{}
	for _, a := range got.Artifacts {
		byRef[a.Ref] = a
	}
	bare, ok := byRef["pull/42"]
	if !ok {
		t.Fatalf("the bare ref was dropped: %+v", got.Artifacts)
	}
	if bare.URL != "" || bare.Repo != "" || bare.Number != 0 {
		t.Errorf("bare ref = %+v, want no link and no guessed repository", bare)
	}
	// A page *about* a pull request is openable but is not keyed as the pull
	// request itself — titling it by number would title the wrong thing.
	files := byRef["https://github.com/dlddu/pocket-aide/pull/425/files"]
	if files.URL == "" || files.Number != 0 {
		t.Errorf("deep link = %+v, want openable but not keyed as the pull request", files)
	}
	// Surrounding whitespace is not part of a ref, so the trimmed form is what
	// is recorded and what a title can be keyed by.
	trimmed := byRef["https://github.com/dlddu/pocket-aide/pull/425"]
	if trimmed.Number != 425 {
		t.Errorf("trimmed ref = %+v, want the pull request it names", trimmed)
	}
}

// One task's materials never arrive under another's, and a conversation that
// moved to a different session does not inherit the old one's fold.
func TestTaskMaterialsAreNotSharedAcrossTasksOrSessions(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, reviewedTaskTurns)
	platform.page("worker-2", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 1, "request": "다른 세션의 일", "choices": [],
	     "artifacts": ["https://github.com/dlddu/other/pull/9"],
	     "requested_at": "2026-11-20T00:00:00Z", "ended_at": "2026-11-20T01:00:00Z"}
	  ]
	}`)

	if _, err := engine.TaskMaterials(context.Background(), 1, "t1"); err != nil {
		t.Fatalf("first session: %v", err)
	}
	// A task the session has no turns for is empty, not the other task's.
	other, err := engine.TaskMaterials(context.Background(), 1, "t2")
	if err != nil {
		t.Fatalf("other task: %v", err)
	}
	if len(other.Artifacts) != 0 || len(other.Decisions) != 0 {
		t.Errorf("other task = %+v, want nothing of t1's", other)
	}
	// The conversation moves to another session. The same task id there is
	// other work, so the fold is not reused.
	bindSession(t, store, 1, "worker-2")
	moved, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("moved session: %v", err)
	}
	if moved.SessionID != "worker-2" {
		t.Errorf("sessionID = %q, want the session it was read from", moved.SessionID)
	}
	if len(moved.Artifacts) != 1 || moved.Artifacts[0].Repo != "dlddu/other" {
		t.Errorf("artifacts = %+v, want only the new session's", moved.Artifacts)
	}
	if len(moved.Decisions) != 0 {
		t.Errorf("decisions = %+v, want none of the old session's", moved.Decisions)
	}
}

// The platform's answers stay the different things they are. None of them is
// "this task produced nothing", which would read as a finished review.
func TestTaskMaterialsTellsTheRefusalsApart(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		want   error
	}{
		{"no such task", http.StatusNotFound, chat.ErrTaskNotFound},
		{"session not awake", http.StatusConflict, chat.ErrTaskRequestUnavailable},
		{"keeps no work records", http.StatusNotImplemented, chat.ErrTaskRequestUnsupported},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store, engine, platform := materialsEngine(t)
			bindSession(t, store, 1, "worker-1")
			platform.status = tc.status
			got, err := engine.TaskMaterials(context.Background(), 1, "t1")
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if len(got.Artifacts) != 0 {
				t.Errorf("artifacts = %+v, want none on a refusal", got.Artifacts)
			}
		})
	}
}

// A conversation with no session has no work to review, and an id the slot
// could never hold is refused before anything is read.
func TestTaskMaterialsRefusesBeforeReadingWhenThereIsNothingToRead(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	if _, err := engine.TaskMaterials(context.Background(), 1, "t1"); !errors.Is(err, chat.ErrNoConversation) {
		t.Fatalf("err = %v, want ErrNoConversation", err)
	}
	bindSession(t, store, 1, "worker-1")
	if _, err := engine.TaskMaterials(context.Background(), 1, "bad id!"); !errors.Is(err, chat.ErrTaskRequestInvalid) {
		t.Fatalf("err = %v, want ErrTaskRequestInvalid", err)
	}
	if len(platform.calls) != 0 {
		t.Errorf("calls = %v, want none", platform.calls)
	}
}
