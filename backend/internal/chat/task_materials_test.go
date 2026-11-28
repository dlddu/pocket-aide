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

func tradeoffByName(materials chat.TaskMaterials, name string) (chat.TaskTradeoff, bool) {
	for _, t := range materials.Tradeoffs {
		if t.Alternative == name {
			return t, true
		}
	}
	return chat.TaskTradeoff{}, false
}

// --- the positive case ---------------------------------------------------
//
// An author weighed two options and wrote them in a table, recorded what was
// settled under 「결정 사항」, and left a question open. Two earlier turns
// produced the pull requests under review; the newest turn is a short follow-up
// that produced nothing.

const reviewedTaskTurns = `{
  "task_id": "t1", "task_version": 9,
  "task": {"id": "t1", "title": "검토 대상", "status": "review", "version": 9},
  "turns": [
    {"id": 1, "request": "구현해 줘", "instruction": "구현하라",
     "response": "두 가지로 갈 수 있습니다.\n\n| 대안 | 장점 | 단점 |\n| --- | --- | --- |\n| 백엔드에서 접기 | 모바일이 이력을 받지 않는다<br>작업당 1회로 끝난다 | 서버 메모리를 쓴다 |\n| 앱에서 접기 | 서버 상태가 없다 | 이력 전체를 내려받아야 한다<br>작업 수만큼 느려진다 |\n",
     "explanation": "두 PR 로 나눠 올렸어요.\n\n## 결정 사항\n- 접기는 백엔드에서 한다\n- 「주고받은 내용」은 접힌 채로 둔다\n",
     "question": "둘 중 어느 쪽으로 갈까요?",
     "choices": [{"label": "백엔드로", "request_text": "백엔드에서 접어 주세요"},
                 {"label": "앱에서", "request_text": "앱에서 접어 주세요"}],
     "artifacts": ["https://github.com/dlddu/pocket-aide/pull/425",
                   "https://github.com/dlddu/pocket-aide/pull/426"],
     "requested_at": "2026-11-20T00:00:00Z", "started_at": "2026-11-20T00:01:00Z",
     "ended_at": "2026-11-20T01:00:00Z"},
    {"id": 2, "request": "백엔드에서 접어 주세요", "instruction": "백엔드에서 접어라",
     "response": "백엔드에 투영을 넣었습니다.", "explanation": "반영했어요.",
     "question": null, "choices": [],
     "artifacts": ["https://github.com/dlddu/pocket-aide/pull/430"],
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

// Two alternatives, each with its own advantages *and* drawbacks, comparable
// side by side, each saying which turn and which field recorded it.
func TestTaskMaterialsReadsTwoAlternativesWithTheirOwnProsAndCons(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, reviewedTaskTurns)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	if !got.TradeoffsRecorded || len(got.Tradeoffs) != 2 {
		t.Fatalf("tradeoffs = %+v, want the two the table names", got.Tradeoffs)
	}
	backend, ok := tradeoffByName(got, "백엔드에서 접기")
	if !ok {
		t.Fatalf("the first alternative is missing: %+v", got.Tradeoffs)
	}
	app, ok := tradeoffByName(got, "앱에서 접기")
	if !ok {
		t.Fatalf("the second alternative is missing: %+v", got.Tradeoffs)
	}
	// Each side is the author's own words, and the two cells that held two
	// lines became two entries rather than one run-on string.
	if len(backend.Pros) != 2 || backend.Pros[0] != "모바일이 이력을 받지 않는다" {
		t.Errorf("backend pros = %+v", backend.Pros)
	}
	if len(backend.Cons) != 1 || backend.Cons[0] != "서버 메모리를 쓴다" {
		t.Errorf("backend cons = %+v", backend.Cons)
	}
	if len(app.Pros) != 1 || app.Pros[0] != "서버 상태가 없다" {
		t.Errorf("app pros = %+v", app.Pros)
	}
	if len(app.Cons) != 2 || app.Cons[1] != "작업 수만큼 느려진다" {
		t.Errorf("app cons = %+v", app.Cons)
	}
	// The two are actually different on both sides — that is what makes them
	// worth comparing, and a bug that copied one into the other would show up
	// here rather than looking plausible.
	if backend.Pros[0] == app.Pros[0] || backend.Cons[0] == app.Cons[0] {
		t.Error("the two alternatives must not share a side")
	}
	// Provenance: which turn, and which field of it.
	for _, entry := range got.Tradeoffs {
		if entry.TurnID != 1 || entry.Source != "response" {
			t.Errorf("%q came from turn %d %q, want turn 1 response", entry.Alternative, entry.TurnID, entry.Source)
		}
	}
}

// What was settled, what is still being asked, and the pull requests — three
// different things, each from where it is actually recorded.
func TestTaskMaterialsKeepsRecordedDecisionsApartFromWhatIsOpen(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, reviewedTaskTurns)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	// Decided: the two items the author listed under 「결정 사항」, verbatim and
	// in the order they were written.
	if !got.DecisionsRecorded || len(got.Decisions) != 2 {
		t.Fatalf("decisions = %+v, want the two recorded under the marker", got.Decisions)
	}
	if got.Decisions[0].Text != "접기는 백엔드에서 한다" {
		t.Errorf("first decision = %q", got.Decisions[0].Text)
	}
	if got.Decisions[1].Text != "「주고받은 내용」은 접힌 채로 둔다" {
		t.Errorf("second decision = %q", got.Decisions[1].Text)
	}
	for _, d := range got.Decisions {
		if d.TurnID != 1 || d.Source != "explanation" {
			t.Errorf("decision %q came from turn %d %q", d.Text, d.TurnID, d.Source)
		}
	}
	// Open: the latest turn's question only. Turn 1 asked one too and it has
	// turns after it, so offering it would ask for a decision already taken.
	if len(got.Open) != 1 || got.Open[0].TurnID != 3 {
		t.Fatalf("open = %+v, want only the latest turn's", got.Open)
	}
	if got.Open[0].Question != "어느 쪽을 고칠까요?" {
		t.Errorf("open question = %q", got.Open[0].Question)
	}
	// And nothing that was decided is also listed as open.
	for _, d := range got.Decisions {
		for _, o := range got.Open {
			if o.Question == d.Text {
				t.Errorf("%q is in both lists", d.Text)
			}
		}
	}
}

// The regression the feature exists for: the pull requests under review were
// produced on earlier turns, and the newest turn is a short follow-up that
// names nothing. Everything still has to be there, with its source.
func TestTaskMaterialsSurvivesAShortFollowUpAfterTheRealWork(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, reviewedTaskTurns)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	// Three distinct pull requests across two earlier turns, newest sighting
	// first, each pointing at the turn that produced it.
	if len(got.Artifacts) != 3 {
		t.Fatalf("artifacts = %+v, want all three of turns 1 and 2", got.Artifacts)
	}
	if got.Artifacts[0].Number != 430 || got.Artifacts[0].TurnID != 2 {
		t.Errorf("first artifact = %+v, want turn 2's 430 in front", got.Artifacts[0])
	}
	for _, a := range got.Artifacts[1:] {
		if a.TurnID != 1 {
			t.Errorf("artifact %q turn = %d, want 1", a.Ref, a.TurnID)
		}
		if a.Repo != "dlddu/pocket-aide" || a.URL != a.Ref {
			t.Errorf("artifact = %+v, want an openable pull request", a)
		}
	}
	// The trade-offs and decisions of turn 1 are still readable even though the
	// newest turn wrote neither.
	if len(got.Tradeoffs) != 2 || len(got.Decisions) != 2 {
		t.Errorf("materials thinned out behind the follow-up: %+v / %+v", got.Tradeoffs, got.Decisions)
	}
	// The newest turn's own request is a short follow-up, and it is not read as
	// a decision even though it is not one of the offered sentences either.
	for _, d := range got.Decisions {
		if strings.Contains(d.Text, "수정해 주세요") {
			t.Errorf("the follow-up request was read as a decision: %+v", d)
		}
	}
	if got.FoldedThrough != 3 {
		t.Errorf("foldedThrough = %d, want 3", got.FoldedThrough)
	}
}

// --- the negative case ---------------------------------------------------

// A task whose turns hold only ordinary prose and choice buttons records no
// trade-offs and no decisions, and says so. This is the case that would be
// easiest to fake: the explanation reads like a conclusion and the buttons read
// like options.
func TestTaskMaterialsReadsNeitherProseNorButtonsAsRecordedMaterial(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 1, "request": "해 줘",
	     "response": "A 가 더 빠르고 B 는 더 쌉니다. 장단점을 생각하면 A 가 좋아 보입니다.",
	     "explanation": "A 로 가는 게 좋겠습니다. 그래서 A 로 했습니다. 이게 결정이라고 보면 됩니다.",
	     "question": "A 로 할까요?",
	     "choices": [{"label": "A 로 간다 (빠름)", "request_text": "A 로 해 주세요"},
	                 {"label": "B 로 간다 (저렴)", "request_text": "B 로 해 주세요"}],
	     "artifacts": [], "requested_at": "2026-11-20T00:00:00Z",
	     "ended_at": "2026-11-20T01:00:00Z"},
	    {"id": 2, "request": "A 로 해 주세요", "artifacts": [], "choices": [],
	     "requested_at": "2026-11-20T02:00:00Z"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	// Prose that mentions 장단점 is not a marked trade-off, and button labels
	// that carry a reason in brackets are not advantages.
	if got.TradeoffsRecorded || len(got.Tradeoffs) != 0 {
		t.Errorf("tradeoffs = %+v, want none: nothing was marked", got.Tradeoffs)
	}
	// An explanation that says "이게 결정이라고 보면 됩니다" is still prose.
	if got.DecisionsRecorded || len(got.Decisions) != 0 {
		t.Errorf("decisions = %+v, want none: no 결정 사항 record", got.Decisions)
	}
	// And turn 2's request repeats an offered sentence word for word. That is
	// not evidence anybody pressed the button and not an acceptance of
	// anything, so it is not a decision.
	for _, d := range got.Decisions {
		if strings.Contains(d.Text, "A 로 해 주세요") {
			t.Errorf("an echoed sentence was read as a decision: %+v", d)
		}
	}
	if len(got.Artifacts) != 0 {
		t.Errorf("artifacts = %+v, want none", got.Artifacts)
	}
}

// --- the labelled-prose shape -------------------------------------------

// The other supported shape: a 「대안별 장단점」 section whose items carry
// 장점/단점 labels. A side the author did not write stays unrecorded for that
// alternative rather than being filled from the other one.
func TestTaskMaterialsReadsTheLabelledSectionAndLeavesAMissingSideUnrecorded(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 7, "request": "비교해 줘",
	     "explanation": "## 대안별 장단점\n- 지금 고치기\n  - 장점: 사용자가 바로 본다\n  - 단점: 테스트를 못 돌린다\n- 다음 PR 로 미루기\n  - 장점: CI 를 거친다\n\n## 결정 사항\n- 지금 고친다\n",
	     "artifacts": [], "choices": [],
	     "requested_at": "2026-11-20T00:00:00Z", "ended_at": "2026-11-20T01:00:00Z"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	if len(got.Tradeoffs) != 2 {
		t.Fatalf("tradeoffs = %+v, want both items of the section", got.Tradeoffs)
	}
	now, ok := tradeoffByName(got, "지금 고치기")
	if !ok || len(now.Pros) != 1 || len(now.Cons) != 1 {
		t.Fatalf("first alternative = %+v", now)
	}
	if now.Pros[0] != "사용자가 바로 본다" || now.Cons[0] != "테스트를 못 돌린다" {
		t.Errorf("first alternative sides = %+v / %+v", now.Pros, now.Cons)
	}
	later, ok := tradeoffByName(got, "다음 PR 로 미루기")
	if !ok {
		t.Fatalf("second alternative missing: %+v", got.Tradeoffs)
	}
	// The author wrote no drawback for this one, so none is reported — the
	// screen says that side is unrecorded rather than showing a blank that
	// reads as "no drawbacks".
	if len(later.Pros) != 1 || len(later.Cons) != 0 {
		t.Errorf("second alternative = %+v, want pros only", later)
	}
	if later.Source != "explanation" || later.TurnID != 7 {
		t.Errorf("provenance = turn %d %q", later.TurnID, later.Source)
	}
	// The 결정 사항 section after it is read as decisions, not swallowed into
	// the trade-offs above.
	if len(got.Decisions) != 1 || got.Decisions[0].Text != "지금 고친다" {
		t.Errorf("decisions = %+v", got.Decisions)
	}
}

// --- a turn edited in place ---------------------------------------------

// The platform lets a parent patch a turn without its id changing, so the
// task's newest turn can gain a result, a pull request and a trade-off table
// after the fold already read it. The next read has to show the new state —
// and must not keep the old values or list anything twice.
func TestTaskMaterialsPicksUpAPatchToTheTurnItAlreadyFolded(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	// First read: turn 2 is open, has no result and names one pull request.
	platform.page("worker-1", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 1, "request": "해 줘", "response": "시작합니다", "artifacts": [],
	     "choices": [], "requested_at": "2026-11-20T00:00:00Z",
	     "ended_at": "2026-11-20T01:00:00Z"},
	    {"id": 2, "request": "이어서 해 줘", "response": null, "explanation": null,
	     "artifacts": ["https://github.com/dlddu/pocket-aide/pull/500"],
	     "choices": [], "question": "이대로 갈까요?",
	     "requested_at": "2026-11-20T02:00:00Z", "ended_at": null}
	  ]
	}`)
	first, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	if len(first.Artifacts) != 1 || first.TradeoffsRecorded || first.DecisionsRecorded {
		t.Fatalf("first read = %+v", first)
	}
	calls := len(platform.calls)

	// turn_patch fills turn 2 in place: it finishes, writes a trade-off table
	// and a decision, swaps the pull request it named for a different one, and
	// answers its own question. Same id.
	platform.page("worker-1", "t1", 1, `{
	  "task_id": "t1", "turns": [
	    {"id": 2, "request": "이어서 해 줘",
	     "response": "| 대안 | 장점 | 단점 |\n| --- | --- | --- |\n| 유지 | 작게 끝난다 | 느리다 |\n| 교체 | 빠르다 | 위험하다 |\n",
	     "explanation": "### 결정 사항\n- 교체로 간다\n",
	     "artifacts": ["https://github.com/dlddu/pocket-aide/pull/501"],
	     "choices": [], "question": null,
	     "requested_at": "2026-11-20T02:00:00Z", "ended_at": "2026-11-20T05:00:00Z"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	// The resumed read asked for the turns after 1 — one before the newest it
	// had folded — so the patched turn was read again rather than skipped.
	resumed := platform.calls[calls:]
	if len(resumed) == 0 || !strings.Contains(resumed[0], "after_turn_id=1") {
		t.Fatalf("resumed calls = %v, want a re-read of turn 2", resumed)
	}
	// The newest state is there.
	if !got.TradeoffsRecorded || len(got.Tradeoffs) != 2 {
		t.Fatalf("tradeoffs = %+v, want the patched table", got.Tradeoffs)
	}
	if !got.DecisionsRecorded || len(got.Decisions) != 1 || got.Decisions[0].Text != "교체로 간다" {
		t.Fatalf("decisions = %+v, want the patched one", got.Decisions)
	}
	// Nothing is listed twice, and the ref the turn no longer names is gone —
	// the records do not state it any more, so neither does this.
	if len(got.Artifacts) != 1 {
		t.Fatalf("artifacts = %+v, want only the ref the patched turn names", got.Artifacts)
	}
	if got.Artifacts[0].Number != 501 {
		t.Errorf("artifact = %+v, want the patched 501", got.Artifacts[0])
	}
	// Re-reading the same turn did not double its decision or its table.
	seen := map[string]int{}
	for _, d := range got.Decisions {
		seen[d.Text]++
	}
	for text, n := range seen {
		if n > 1 {
			t.Errorf("decision %q appeared %d times", text, n)
		}
	}
	alts := map[string]int{}
	for _, tr := range got.Tradeoffs {
		alts[tr.Alternative]++
	}
	for alt, n := range alts {
		if n > 1 {
			t.Errorf("alternative %q appeared %d times", alt, n)
		}
	}
	// The question it answered is no longer open.
	if len(got.Open) != 0 {
		t.Errorf("open = %+v, want none once the turn ended with no question", got.Open)
	}
	// Turn 1 was not re-read — only the tail and anything after it.
	for _, call := range resumed {
		if strings.Contains(call, "after_turn_id=0") || !strings.Contains(call, "after_turn_id=") {
			t.Errorf("resumed read started over: %q", call)
		}
	}
}

// A later turn that restates one alternative supersedes the earlier statement
// for that alternative and leaves the other one alone.
func TestTaskMaterialsPrefersTheNewestStatementOfAnAlternative(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 1, "request": "비교",
	     "response": "| 대안 | 장점 | 단점 |\n| --- | --- | --- |\n| A | 처음 장점 | 처음 단점 |\n| B | B 장점 | B 단점 |\n",
	     "artifacts": [], "choices": [], "requested_at": "2026-11-20T00:00:00Z",
	     "ended_at": "2026-11-20T01:00:00Z"},
	    {"id": 2, "request": "다시 비교",
	     "response": "| 대안 | 장점 | 단점 |\n| --- | --- | --- |\n| A | 고친 장점 | 고친 단점 |\n",
	     "artifacts": [], "choices": [], "requested_at": "2026-11-20T02:00:00Z",
	     "ended_at": "2026-11-20T03:00:00Z"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	if len(got.Tradeoffs) != 2 {
		t.Fatalf("tradeoffs = %+v, want one row per alternative", got.Tradeoffs)
	}
	a, _ := tradeoffByName(got, "A")
	if len(a.Pros) != 1 || a.Pros[0] != "고친 장점" || a.TurnID != 2 {
		t.Errorf("A = %+v, want turn 2's restatement", a)
	}
	b, _ := tradeoffByName(got, "B")
	if len(b.Pros) != 1 || b.Pros[0] != "B 장점" || b.TurnID != 1 {
		t.Errorf("B = %+v, want turn 1's, untouched", b)
	}
}

// --- empty, old and isolated -------------------------------------------

// A task with one recent turn and nothing recorded on it reads as empty, which
// is the answer an old task with no marked material gives too.
func TestTaskMaterialsAnswersEmptyForATaskWithNothingRecorded(t *testing.T) {
	store, engine, platform := materialsEngine(t)
	bindSession(t, store, 1, "worker-1")
	platform.page("worker-1", "t1", 0, `{
	  "task_id": "t1", "turns": [
	    {"id": 91, "request": "오래된 작업의 마지막 턴만 남아 있다",
	     "response": "했습니다", "artifacts": [], "choices": [],
	     "requested_at": "2026-10-01T00:00:00Z", "ended_at": "2026-10-01T01:00:00Z"}
	  ]
	}`)

	got, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("materials: %v", err)
	}
	if len(got.Artifacts) != 0 || len(got.Tradeoffs) != 0 || len(got.Decisions) != 0 || len(got.Open) != 0 {
		t.Errorf("materials = %+v, want everything empty", got)
	}
	if got.TradeoffsRecorded || got.DecisionsRecorded {
		t.Error("nothing was recorded, so neither flag may claim otherwise")
	}
	// Folded through the one turn it has, so the next read resumes rather than
	// starting over.
	if got.FoldedThrough != 91 {
		t.Errorf("foldedThrough = %d, want 91", got.FoldedThrough)
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
	// Surrounding whitespace is not part of a ref.
	if byRef["https://github.com/dlddu/pocket-aide/pull/425"].Number != 425 {
		t.Errorf("trimmed ref missing: %+v", got.Artifacts)
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
	     "explanation": "## 결정 사항\n- 다른 세션의 결정\n",
	     "artifacts": ["https://github.com/dlddu/other/pull/9"],
	     "requested_at": "2026-11-20T00:00:00Z", "ended_at": "2026-11-20T01:00:00Z"}
	  ]
	}`)

	if _, err := engine.TaskMaterials(context.Background(), 1, "t1"); err != nil {
		t.Fatalf("first session: %v", err)
	}
	other, err := engine.TaskMaterials(context.Background(), 1, "t2")
	if err != nil {
		t.Fatalf("other task: %v", err)
	}
	if len(other.Artifacts) != 0 || len(other.Decisions) != 0 || len(other.Tradeoffs) != 0 {
		t.Errorf("other task = %+v, want nothing of t1's", other)
	}
	bindSession(t, store, 1, "worker-2")
	moved, err := engine.TaskMaterials(context.Background(), 1, "t1")
	if err != nil {
		t.Fatalf("moved session: %v", err)
	}
	if moved.SessionID != "worker-2" {
		t.Errorf("sessionID = %q", moved.SessionID)
	}
	if len(moved.Artifacts) != 1 || moved.Artifacts[0].Repo != "dlddu/other" {
		t.Errorf("artifacts = %+v, want only the new session's", moved.Artifacts)
	}
	if len(moved.Tradeoffs) != 0 {
		t.Errorf("tradeoffs = %+v, want none of the old session's", moved.Tradeoffs)
	}
	if len(moved.Decisions) != 1 || moved.Decisions[0].Text != "다른 세션의 결정" {
		t.Errorf("decisions = %+v, want only the new session's", moved.Decisions)
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
