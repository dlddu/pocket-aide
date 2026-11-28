// What a person needs in order to review one task, gathered in one small
// answer (PRD-AI채팅 AC11).
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
// What it does not do is invent any of the three. The agreed turn (platform
// AC-N9) holds `request`, `instruction`, `response`, `explanation`, `question`,
// `choices` and `artifacts`, and nothing else is ever added to it. There is no
// field for "this was decided" and none for an alternative's advantages and
// drawbacks, so this reports neither as recorded: an `explanation` is the
// parent's account of a result, not a decision somebody took, and a choice's
// label is a sentence on a button, not a trade-off. Reading either as more than
// it is would put a decision in a person's mouth that nobody made.
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
	// Decisions are the sentences the session offered that came back as the
	// next turn's request. That is a decision the records hold: the very words
	// a turn put on offer are the words the next turn was asked with. Nothing
	// else is read as one.
	Decisions []TaskDecision `json:"decisions"`
	// Open is what is still being asked, taken from the task's latest turn
	// only. An earlier question that has turns after it is history, not an open
	// point, and mixing the two would ask somebody to answer something that has
	// moved on.
	Open []TaskOpenPoint `json:"open"`
	// TradeoffsRecorded is always false, and is in the answer so the app can
	// say *why* it shows no advantages and drawbacks instead of leaving a
	// silence that reads as "there were none". The agreed turn has no field for
	// them; a label on a choice button is not one.
	TradeoffsRecorded bool `json:"tradeoffs_recorded"`
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

// TaskDecision is one sentence that was offered and then sent.
//
// OfferedOnTurnID is the turn that put the sentence on offer and TurnID is the
// turn it came back on. Both are here because a decision is the pair: what was
// asked, and what was answered. Label is the words the button showed, which is
// what a person read before picking it.
type TaskDecision struct {
	TurnID          int64  `json:"turn_id"`
	OfferedOnTurnID int64  `json:"offered_on_turn_id"`
	Label           string `json:"label"`
	Text            string `json:"text"`
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
// Artifacts and decisions are append-only facts: a turn that named a pull
// request named it, whatever happens afterwards. So they are kept and a later
// read only has to fold the turns that have arrived since — which is what keeps
// opening a task a cheap thing to do rather than a re-read of its whole
// history.
//
// The open points are not kept, because they are not append-only: a question
// that was the latest turn's stops being open the moment another turn lands.
// They are taken from the tail every time, and every read confirms the tail by
// asking for the turns after the one it already folded.
type materialsFold struct {
	sessionID string
	artifacts []TaskArtifact
	// artifactAt is where in artifacts a ref already sits, so a ref seen again
	// moves rather than repeats.
	artifactAt map[string]int
	decisions  []TaskDecision
	// tail is the newest turn folded, which is what the open points come from.
	tail      *TaskTurn
	through   int64
	truncated bool
	// offered is the latest folded turn's choices, held so the next read can
	// tell whether the turn after it answered one of them.
	offered []TaskTurnChoice
	// offeredOn is the turn those choices came from.
	offeredOn int64
}

// TaskMaterials gathers what userID's conversation has on one task for review.
//
// The session is taken from the user's own row, so a caller cannot name
// somebody else's, and nothing here writes or wakes anything: a session with no
// live pod answers that it cannot be read rather than answering an empty list.
func (e *Engine) TaskMaterials(ctx context.Context, userID int64, taskID string) (TaskMaterials, error) {
	out := TaskMaterials{
		TaskID: taskID, Artifacts: []TaskArtifact{},
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
	// Copied out of the fold, because the fold is shared: the caller adds
	// recorded titles to these entries, and writing into the slice the cache
	// holds would be a write to memory another request may be cloning.
	out.Artifacts = append(out.Artifacts, fold.artifacts...)
	out.Decisions = append(out.Decisions, fold.decisions...)
	out.Open = openPointsOf(fold.tail)
	out.FoldedThrough = fold.through
	out.Truncated = fold.truncated
	return out, nil
}

// foldTaskTurns reads the turns this fold has not seen and folds them in,
// returning a copy of the result. A first call for a task walks it; afterwards
// it asks for the turns after the newest one it folded, which is one small call
// that usually answers with none.
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
	for page := 0; ; page++ {
		if page == materialsFoldMaxPages {
			fold.truncated = true
			break
		}
		turns, next, err := e.taskTurnPage(ctx, sessionID, taskID, fold.through)
		if err != nil {
			return materialsFold{}, err
		}
		for i := range turns {
			fold.add(&turns[i])
		}
		if next == 0 {
			break
		}
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
	return materialsFold{
		sessionID: sessionID, artifacts: []TaskArtifact{},
		artifactAt: map[string]int{}, decisions: []TaskDecision{},
	}
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

// add folds one turn in. Turns arrive oldest first, which is why a pick can be
// recognised at all: the choices it answers are the ones the turn before it
// left on offer.
func (f *materialsFold) add(turn *TaskTurn) {
	if turn.ID <= f.through {
		// A page that repeated a turn, or one already folded. Folding it again
		// would double a decision and move an artifact that did not move.
		return
	}
	for _, ref := range turn.Artifacts {
		f.addArtifact(ref, turn.ID)
	}
	if decision, ok := decisionOf(turn, f.offered, f.offeredOn); ok {
		f.decisions = append(f.decisions, decision)
	}
	// This turn is now the newest, so its choices are what a later turn could
	// answer and its words are where the open points come from.
	f.offered, f.offeredOn = turn.Choices, turn.ID
	held := *turn
	f.tail = &held
	f.through = turn.ID
}

// addArtifact records one ref, newest sighting first. A ref already held is
// moved to the front under the newer turn rather than added again: it is the
// same pull request, and two rows for it would read as two pull requests.
func (f *materialsFold) addArtifact(ref string, turnID int64) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return
	}
	if f.artifactAt == nil {
		f.artifactAt = map[string]int{}
	}
	if at, held := f.artifactAt[ref]; held {
		f.artifacts[at].TurnID = turnID
		moved := f.artifacts[at]
		f.artifacts = append(f.artifacts[:at], f.artifacts[at+1:]...)
		f.artifacts = append([]TaskArtifact{moved}, f.artifacts...)
		f.reindexArtifacts()
		return
	}
	entry := TaskArtifact{Ref: ref, TurnID: turnID}
	if link, repo, number, ok := pullRequestRef(ref); ok {
		entry.URL, entry.Repo, entry.Number = link, repo, number
	} else if absoluteWebURL(ref) {
		entry.URL = ref
	}
	f.artifacts = append([]TaskArtifact{entry}, f.artifacts...)
	f.reindexArtifacts()
}

func (f *materialsFold) reindexArtifacts() {
	f.artifactAt = make(map[string]int, len(f.artifacts))
	for i, a := range f.artifacts {
		f.artifactAt[a.Ref] = i
	}
}

// clone is a copy nothing else holds a reference into, so a fold handed out
// cannot be changed under the next reader.
func (f materialsFold) clone() materialsFold {
	out := f
	out.artifacts = append([]TaskArtifact{}, f.artifacts...)
	out.decisions = append([]TaskDecision{}, f.decisions...)
	out.offered = append([]TaskTurnChoice{}, f.offered...)
	out.artifactAt = make(map[string]int, len(f.artifactAt))
	for k, v := range f.artifactAt {
		out.artifactAt[k] = v
	}
	if f.tail != nil {
		held := *f.tail
		out.tail = &held
	}
	return out
}

// decisionOf is the decision this turn carries, if it carries one: its request
// is a sentence the turn before it offered, word for word.
//
// Word for word is the whole test. The app sends a pick as the next turn's
// request unchanged and sends edited words as free text, so an exact match is
// the session's own sentence coming back and anything else is a person's own —
// which is a request, not a decision between recorded alternatives.
func decisionOf(turn *TaskTurn, offered []TaskTurnChoice, offeredOn int64) (TaskDecision, bool) {
	if turn.Request == nil || len(offered) == 0 {
		return TaskDecision{}, false
	}
	text := strings.TrimSpace(*turn.Request)
	if text == "" {
		return TaskDecision{}, false
	}
	for _, choice := range offered {
		if strings.TrimSpace(choice.RequestText) != text {
			continue
		}
		return TaskDecision{
			TurnID: turn.ID, OfferedOnTurnID: offeredOn,
			Label: strings.TrimSpace(choice.Label), Text: text,
		}, true
	}
	return TaskDecision{}, false
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
