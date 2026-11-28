// What a person needs in order to review one task, gathered in one small
// answer (PRD-AI채팅 AC10, 검토 자료).
//
// A task in review is read to decide something, and what that takes is: the
// pull requests the work actually produced, what has already been settled, and
// what is still being asked. All three are in the task's turns — and the turns
// are the long part. A task that has run for a while has every verbatim session
// answer in them, which is exactly why the detail pane keeps them closed
// (AC10). So this reads them once, keeps only the few short facts a review
// needs, and answers from that.
//
// It is deliberately not part of the task list. The list is polled while the
// panel is on screen, and folding every task's turns on every poll would be one
// history read per task per poll — the panel would cost more the more work a
// person has. This is one task, asked for when somebody opens it.
//
// What it does not do is invent any of it. The agreed turn (platform AC-N9)
// holds `request`, `instruction`, `response`, `explanation`, `question`,
// `choices` and `artifacts`, and that schema is not ours to extend. It has no
// column for "this was decided" and none for an alternative's advantages and
// drawbacks — so those two are read out of what the author *marked* in the
// turn's own `response` or the parent's `explanation`, and nowhere else
// (task_materials_text.go). A paragraph of explanation that happens to read
// like a conclusion is not a decision, and a sentence on a choice button is not
// a trade-off.
//
// In particular, a request that repeats an offered sentence word for word is
// *not* read as a decision. It is not evidence that anybody pressed the button,
// and it is not an acceptance of anything — the same words can be typed, or
// quoted back to ask about them. Saying somebody decided it would be the
// clearest possible case of putting a position in their mouth.
package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// materialsFoldPage is how many turns one fold request reads at a time. It is
// the platform's own cap, so a first fold of a long task is as few calls as the
// route allows.
const materialsFoldPage = TaskTurnPageLimitMax

// materialsFoldMaxPages bounds a first fold. Reaching it is answered as a
// partial gathering rather than a complete one — materials presented as whole
// when they are a prefix would read as "this is every pull request there is".
const materialsFoldMaxPages = 50

// materialsCacheMax is how many tasks' folds are kept. A conversation has one
// session and a person opens a handful of tasks, so this is generous; it exists
// so a long-lived process cannot grow a fold per task it has ever seen.
const materialsCacheMax = 64

// TaskMaterials is the review answer for one task.
//
// Every field here is something the records state. Absent means the records do
// not have it, and the app says so rather than showing a blank that reads like
// an answer.
type TaskMaterials struct {
	TaskID    string `json:"task_id"`
	SessionID string `json:"session_id"`
	// Artifacts are what the work produced, newest sighting first. One entry
	// per distinct thing however many turns mentioned it: a pull request named
	// again on a later turn is the same pull request, and listing it twice
	// would read as two.
	Artifacts []TaskArtifact `json:"artifacts"`
	// Tradeoffs are the alternatives an author weighed, each with the
	// advantages and drawbacks they wrote for it, newest statement first. Only
	// what was written under a 「대안별 장단점」 marker or in a table whose header
	// names the columns gets here.
	Tradeoffs []TaskTradeoff `json:"tradeoffs"`
	// Decisions are what an author recorded under a 「결정 사항」 marker. A
	// decision is a thing somebody states they settled, so only a record that
	// says it is one counts.
	Decisions []TaskDecision `json:"decisions"`
	// Open is what is still being asked, taken from the task's latest turn
	// only. An earlier question that has turns after it is history, not an open
	// point, and mixing the two would ask somebody to answer something that has
	// moved on.
	Open []TaskOpenPoint `json:"open"`
	// TradeoffsRecorded is whether any alternative's advantages and drawbacks
	// were found written down. It is in the answer so the app can say *why* it
	// shows none — "the records do not have them" rather than a silence that
	// reads as "there were no trade-offs", which is a different claim and not
	// one the records make.
	TradeoffsRecorded bool `json:"tradeoffs_recorded"`
	// DecisionsRecorded is the same for settled points, for the same reason.
	DecisionsRecorded bool `json:"decisions_recorded"`
	// FoldedThrough is the newest turn these materials account for, so the app
	// can tell a fresh answer from one it already holds.
	FoldedThrough int64 `json:"folded_through"`
	// Truncated says the fold stopped at its page bound, so the app can say the
	// materials are partial rather than imply they are all there is.
	Truncated bool `json:"truncated"`
}

// TaskArtifact is one thing the work produced, as the records name it.
//
// Ref is the recorded string verbatim and is always present — it is the fact.
// URL is set only when Ref is itself an absolute http(s) address, because that
// is the only case where there is something to open; a bare `pull/42` names no
// host and no repository, and guessing either from somewhere else would send
// somebody to a page this task never mentioned. Title likewise is filled only
// from a record that already ties that exact repository and number to a title.
type TaskArtifact struct {
	Ref   string `json:"ref"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
	// Repo and Number are the pull request Ref points at, when it is one in a
	// shape that says so. They are what a title lookup is keyed by, and they
	// are only ever read out of Ref itself.
	Repo   string `json:"repo,omitempty"`
	Number int    `json:"number,omitempty"`
	// TurnID is the newest turn that named this, so a person can see which
	// exchange produced it and in what order they arrived.
	TurnID int64 `json:"turn_id"`
}

// TaskDecision is one settled point, in the author's own words.
//
// Text is kept verbatim — this never summarises. TurnID and Source are where it
// was written, so a reviewer can go to that exchange and read the whole thing
// rather than taking this line's word for it.
type TaskDecision struct {
	Text   string `json:"text"`
	TurnID int64  `json:"turn_id"`
	Source string `json:"source"`
}

// TaskTradeoff is one alternative as its author weighed it.
//
// Pros and Cons are their own words, and an empty one means that side was not
// written — the screen says so for that side rather than leaving a blank that
// reads as "no drawbacks". Alternative is the name the author gave the option,
// which is what makes two of these comparable.
type TaskTradeoff struct {
	Alternative string   `json:"alternative"`
	Pros        []string `json:"pros,omitempty"`
	Cons        []string `json:"cons,omitempty"`
	TurnID      int64    `json:"turn_id"`
	Source      string   `json:"source"`
}

// TaskOpenPoint is something the task's latest turn is still asking.
//
// Question is what was asked and Choices are the sentences offered as answers,
// each of which is a sentence to send and not a permission. A point with
// choices and no question, or a question with no choices, is the ordinary case:
// a turn is written as it happens.
type TaskOpenPoint struct {
	TurnID   int64        `json:"turn_id"`
	Question string       `json:"question,omitempty"`
	Choices  []TaskChoice `json:"choices,omitempty"`
}

// materialsFold is what has been read of one task's turns so far.
//
// It keeps each turn's own contribution under that turn's id rather than one
// merged pile. That is what makes a turn being *edited* safe: the platform lets
// a parent patch a turn in place (turn_patch), so the task's newest turn can
// gain a result, a pull request or a trade-off table without its id changing.
// Re-folding that turn replaces its entry, so the new reading cannot sit beside
// the old one as a duplicate and the old values cannot survive as facts the
// records no longer state.
//
// The lists a reader sees are derived from these entries every time (see
// materials), so nothing is cached in the shape it is displayed in.
type materialsFold struct {
	sessionID string
	// facts is turn id -> what that turn contributed.
	facts map[int64]turnFacts
	// order is the turn ids in the order they were folded, which is the order
	// they happened in: the platform pages turns by ascending id.
	order []int64
	// tail is the newest turn folded, which is what the open points come from.
	tail      *TaskTurn
	through   int64
	truncated bool
}

// turnFacts is one turn's contribution. The artifact refs are kept as strings
// rather than as built entries so the display order and the pull-request
// parsing are decided in one place, when the answer is derived.
type turnFacts struct {
	artifacts []string
	tradeoffs []TaskTradeoff
	decisions []TaskDecision
}

// TaskMaterials gathers what userID's conversation has on one task for review.
//
// The session is taken from the user's own row, so a caller cannot name
// somebody else's, and nothing here writes or wakes anything: a session with no
// live pod answers that it cannot be read rather than answering an empty list.
func (e *Engine) TaskMaterials(ctx context.Context, userID int64, taskID string) (TaskMaterials, error) {
	out := TaskMaterials{
		TaskID: taskID, Artifacts: []TaskArtifact{}, Tradeoffs: []TaskTradeoff{},
		Decisions: []TaskDecision{}, Open: []TaskOpenPoint{},
	}
	if !workRecordID.MatchString(taskID) {
		return out, fmt.Errorf("%w: task id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return out, err
	}
	if conv.platformSessionID == "" {
		return out, ErrNoConversation
	}
	out.SessionID = conv.platformSessionID
	fold, err := e.foldTaskTurns(ctx, conv.platformSessionID, taskID)
	if err != nil {
		return out, err
	}
	return fold.materials(out), nil
}

// materials is the reader's answer, derived from the per-turn entries.
func (f materialsFold) materials(out TaskMaterials) TaskMaterials {
	// Artifacts newest sighting first: walk the turns backwards and keep the
	// first time each ref appears, which is the newest turn that named it. One
	// entry per distinct ref — the same pull request named on three turns is
	// one pull request, and three rows would read as three.
	seenRef := map[string]bool{}
	for i := len(f.order) - 1; i >= 0; i-- {
		id := f.order[i]
		for _, ref := range f.facts[id].artifacts {
			if seenRef[ref] {
				continue
			}
			seenRef[ref] = true
			out.Artifacts = append(out.Artifacts, artifactOf(ref, id))
		}
	}
	// Trade-offs newest statement first, one per alternative. An author who
	// restates an option's advantages on a later turn has superseded the
	// earlier statement, and showing both would put two answers on screen for
	// the same question.
	seenAlt := map[string]bool{}
	for i := len(f.order) - 1; i >= 0; i-- {
		for _, t := range f.facts[f.order[i]].tradeoffs {
			key := strings.ToLower(strings.TrimSpace(t.Alternative))
			if key == "" || seenAlt[key] {
				continue
			}
			seenAlt[key] = true
			out.Tradeoffs = append(out.Tradeoffs, t)
		}
	}
	// Decisions oldest first, which is the order they were settled in. A
	// decision restated later is the same decision, kept at the turn that
	// first recorded it.
	seenDecision := map[string]bool{}
	for _, id := range f.order {
		for _, d := range f.facts[id].decisions {
			key := strings.ToLower(strings.Join(strings.Fields(d.Text), " "))
			if key == "" || seenDecision[key] {
				continue
			}
			seenDecision[key] = true
			out.Decisions = append(out.Decisions, d)
		}
	}
	out.Open = openPointsOf(f.tail)
	out.TradeoffsRecorded = len(out.Tradeoffs) > 0
	out.DecisionsRecorded = len(out.Decisions) > 0
	out.FoldedThrough = f.through
	out.Truncated = f.truncated
	return out
}

// foldTaskTurns reads the turns this fold has not seen and folds them in,
// returning a copy of the result.
//
// A resumed read starts one before the newest turn it already folded, so the
// newest turn is read again. That is deliberate: that turn is the one a parent
// patches in place, and a walk that started strictly after it would never see a
// result, a pull request or a trade-off table added to the exchange that is
// still open. Everything older cannot change without a new turn, so this is one
// turn of re-reading rather than a re-read of the history.
func (e *Engine) foldTaskTurns(ctx context.Context, sessionID, taskID string) (materialsFold, error) {
	// The fold is read out, advanced outside the lock, and written back. The
	// platform calls must not be made holding it, and a second reader of the
	// same task meanwhile just folds the same turns again and writes the same
	// answer — wasted work in a case that does not happen, rather than a lock
	// held across a network call.
	fold := e.materialsFoldFor(sessionID, taskID)
	// Re-decided by this run rather than carried: a fold that stopped at the
	// bound last time and reaches the end this time has gathered everything,
	// and saying it is still partial would understate what is on screen.
	fold.truncated = false
	cursor := int64(0)
	if fold.through > 0 {
		cursor = fold.through - 1
	}
	for page := 0; ; page++ {
		if page == materialsFoldMaxPages {
			fold.truncated = true
			break
		}
		turns, next, err := e.taskTurnPage(ctx, sessionID, taskID, cursor)
		if err != nil {
			return materialsFold{}, err
		}
		for i := range turns {
			fold.add(&turns[i])
		}
		if next == 0 {
			break
		}
		cursor = next
	}
	e.storeMaterialsFold(taskID, fold)
	return fold, nil
}

// materialsFoldFor is this task's fold, or an empty one. A fold read under a
// different session than the one asked for is dropped rather than reused: a
// conversation that moved to another session has other work, and folding the
// two together would show one session's pull requests under the other's task.
func (e *Engine) materialsFoldFor(sessionID, taskID string) materialsFold {
	e.mu.Lock()
	defer e.mu.Unlock()
	if held, ok := e.materials[taskID]; ok && held.sessionID == sessionID {
		return held.clone()
	}
	return materialsFold{sessionID: sessionID, facts: map[int64]turnFacts{}}
}

func (e *Engine) storeMaterialsFold(taskID string, fold materialsFold) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.materials == nil {
		e.materials = map[string]materialsFold{}
	}
	// Full means the process has folded more tasks than one person opens. The
	// folds are a saving and not a source, so dropping them all costs the next
	// read one walk and keeps this bounded — rather than picking a victim by an
	// order this map does not have.
	if _, held := e.materials[taskID]; !held && len(e.materials) >= materialsCacheMax {
		e.materials = map[string]materialsFold{}
	}
	e.materials[taskID] = fold
}

// taskTurnPage reads one page of a task's turns from the platform, starting
// after afterTurnID. The statuses are told apart the same way the detail pane's
// own read tells them apart, because the same answers mean the same things.
func (e *Engine) taskTurnPage(
	ctx context.Context, sessionID, taskID string, afterTurnID int64,
) ([]TaskTurn, int64, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(materialsFoldPage))
	if afterTurnID > 0 {
		query.Set("after_turn_id", strconv.FormatInt(afterTurnID, 10))
	}
	path := "/sessions/" + url.PathEscape(sessionID) +
		"/work/tasks/" + url.PathEscape(taskID) + "/turns?" + query.Encode()
	status, raw, err := e.platform.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("read task turns: %w", err)
	}
	switch status {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, 0, ErrTaskNotFound
	case http.StatusConflict:
		return nil, 0, ErrTaskRequestUnavailable
	case http.StatusNotImplemented, http.StatusMethodNotAllowed:
		return nil, 0, ErrTaskRequestUnsupported
	default:
		return nil, 0, platformError("work turns", status, raw)
	}
	var decoded TaskTurnPage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, 0, fmt.Errorf("decode task turns: %w", err)
	}
	return decoded.Turns, decoded.NextCursor, nil
}

// add folds one turn in, replacing whatever that turn contributed before.
//
// Replacing rather than appending is what makes a patched turn correct: the new
// reading of that turn is the whole of what it contributes now, so a pull
// request it no longer names stops being listed and one it gained appears once.
func (f *materialsFold) add(turn *TaskTurn) {
	if f.facts == nil {
		f.facts = map[int64]turnFacts{}
	}
	if _, seen := f.facts[turn.ID]; !seen {
		f.order = append(f.order, turn.ID)
	}
	facts := turnFacts{}
	for _, ref := range turn.Artifacts {
		if ref = strings.TrimSpace(ref); ref != "" {
			facts.artifacts = append(facts.artifacts, ref)
		}
	}
	// The two marked shapes, read from the turn's own answer and from the
	// parent's account of it. Both are fields the records already have; neither
	// is a new column on the agreed turn.
	for _, field := range []struct {
		text, source string
	}{
		{derefString(turn.Response), "response"},
		{derefString(turn.Explanation), "explanation"},
	} {
		read := readMaterialsText(field.text, turn.ID, field.source)
		facts.tradeoffs = append(facts.tradeoffs, read.tradeoffs...)
		facts.decisions = append(facts.decisions, read.decisions...)
	}
	f.facts[turn.ID] = facts
	if turn.ID >= f.through {
		f.through = turn.ID
		held := *turn
		f.tail = &held
	}
}

// artifactOf is one recorded ref as the reader needs it.
func artifactOf(ref string, turnID int64) TaskArtifact {
	entry := TaskArtifact{Ref: ref, TurnID: turnID}
	if link, repo, number, ok := pullRequestRef(ref); ok {
		entry.URL, entry.Repo, entry.Number = link, repo, number
	} else if absoluteWebURL(ref) {
		entry.URL = ref
	}
	return entry
}

// clone is a copy nothing else holds a reference into, so a fold handed out
// cannot be changed under the next reader.
func (f materialsFold) clone() materialsFold {
	out := f
	out.order = append([]int64{}, f.order...)
	out.facts = make(map[int64]turnFacts, len(f.facts))
	for id, facts := range f.facts {
		out.facts[id] = turnFacts{
			artifacts: append([]string{}, facts.artifacts...),
			tradeoffs: append([]TaskTradeoff{}, facts.tradeoffs...),
			decisions: append([]TaskDecision{}, facts.decisions...),
		}
	}
	if f.tail != nil {
		held := *f.tail
		out.tail = &held
	}
	return out
}

// openPointsOf is what the latest turn is still asking — nothing when it asks
// nothing. Only the latest turn is read: a question with turns after it has
// been moved past, and offering it as open would ask somebody to decide
// something the task no longer waits on.
func openPointsOf(tail *TaskTurn) []TaskOpenPoint {
	out := []TaskOpenPoint{}
	if tail == nil {
		return out
	}
	question := strings.TrimSpace(derefString(tail.Question))
	choices := make([]TaskChoice, 0, len(tail.Choices))
	for _, c := range tail.Choices {
		choices = append(choices, TaskChoice{
			Label: strings.TrimSpace(c.Label), RequestText: c.RequestText,
		})
	}
	if question == "" && len(choices) == 0 {
		return out
	}
	return append(out, TaskOpenPoint{TurnID: tail.ID, Question: question, Choices: choices})
}

// pullRequestRef reads a GitHub pull request address out of a ref, when the ref
// is one. It is the one shape recognised, because it is the one this project's
// work produces and the one a recorded title can be keyed by; everything else
// stays a ref and is shown as what it says.
func pullRequestRef(ref string) (link, repo string, number int, ok bool) {
	if !absoluteWebURL(ref) {
		return "", "", 0, false
	}
	parsed, err := url.Parse(ref)
	if err != nil || parsed.Host != "github.com" {
		return "", "", 0, false
	}
	// /{owner}/{repo}/pull/{number} and nothing after it. A longer path is a
	// page about the pull request — a file, a commit — rather than the pull
	// request, and keying a title by it would title that page wrongly.
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 4 || parts[2] != "pull" {
		return "", "", 0, false
	}
	n, err := strconv.Atoi(parts[3])
	if err != nil || n <= 0 || parts[0] == "" || parts[1] == "" {
		return "", "", 0, false
	}
	return ref, parts[0] + "/" + parts[1], n, true
}

// absoluteWebURL is whether this is something a tap can open. A ref that is not
// is shown as the identifier it is.
func absoluteWebURL(ref string) bool {
	parsed, err := url.Parse(ref)
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "https" || parsed.Scheme == "http"
}
