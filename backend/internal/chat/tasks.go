package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// The chat tab's wide layout shows the work records of the codex session that
// answers this conversation, read through session-platform's read-only work
// route. Nothing here parses chat text or asks a model to summarise, and
// nothing here writes.
//
// This used to read the ledger page. The ledger is gone (platform AC-N10), and
// the shapes went with it rather than being aliased onto the new route: the
// ledger page had nowhere to put a task's description and completion criteria
// or anything a turn holds, so a panel built on it could say work existed but
// not what the work was. The agreed task and turn records are what this reads
// now, and the fields the panel shows come from them rather than from a
// translation layer that could disagree with either end.
const (
	// workPageLimit is the platform's per-page cap for the work task route.
	workPageLimit = 100
	// workMaxPages bounds a keyset walk. Reaching it is reported as a
	// truncated list rather than a finished one — a panel that silently drops
	// the tail would read as "this is all the work there is".
	workMaxPages = 50
)

// TaskStatus values this route answers with. They are the work records' own
// domain with one fold: there is a single status for "this is the person's to
// decide", and it is TaskStatusReview. Anything outside the domain — including
// a status a newer platform starts writing — is reported as TaskStatusUnknown
// so the panel can show it without guessing, and is never folded into a
// judgment status it has no evidence for.
const (
	TaskStatusPlanned   = "planned"
	TaskStatusRunning   = "running"
	TaskStatusReview    = "review"
	TaskStatusCompleted = "completed"
	TaskStatusUnknown   = "unknown"
)

// taskStatusBlocked is the records' other name for a task waiting on a person:
// the prerequisite it needs is an approval, a permission or a decision only the
// person can give, which is TaskStatusReview said twice. Two names meant two
// chips and two counts for one judgment, so this route answers the one.
//
// It is read, not retired. The upstream domain still holds it
// (dlddu/session-platform `docs/session-ledger.md` — the agreed shape's status
// domain is "the ledger's own domain unchanged", and a status outside the
// domain is refused rather than clamped), so every record written before the
// fold keeps this value and a writer that has not caught up keeps producing it.
// Reading it as anything but review would lose work a person was already shown
// as theirs to act on.
//
// It is deliberately not one of the statuses above: nothing downstream can
// build a count, a chip or a filter from it, because after normaliseStatus no
// task carries it.
const taskStatusBlocked = "blocked"

// taskStatuses is the display order the panel counts in: what needs a person
// first, then what is merely waiting, then what is done.
var taskStatuses = []string{
	TaskStatusReview, TaskStatusRunning,
	TaskStatusPlanned, TaskStatusCompleted, TaskStatusUnknown,
}

// ErrTasksUnsupported means this conversation's session keeps no work records —
// an older control plane without the route, or a workload type that gets no
// session DB. It is distinct from "no tasks yet" so the app can say so.
var ErrTasksUnsupported = errTasks{"session keeps no work records"}

// ErrTasksUnavailable means the session keeps work records but cannot be read
// right now: it holds no live pod. The rows are still on the slot's volume, so
// this is not an empty list — and reading never wakes the session to get at
// them.
var ErrTasksUnavailable = errTasks{"the chat session is not awake, so its task list cannot be read"}

type errTasks struct{ msg string }

func (e errTasks) Error() string { return e.msg }

// Task is one work task as the panel needs it: what it is, where it stands,
// and — when the records carry them — what the session is asking and what it
// offers the person to say.
type Task struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	// Description and CompletionCriteria are the agreed task's own fields. The
	// ledger page had no column for either, so the panel could not show what a
	// task was for or how it would be known to be done. Empty means the record
	// has none, and the panel says so rather than leaving a blank that reads
	// like an answer.
	Description        string `json:"description,omitempty"`
	CompletionCriteria string `json:"completion_criteria,omitempty"`
	// TurnsTotal is how many exchanges this task has had, so the panel can
	// show that a task has history without opening it — and so the detail pane
	// can tell whether the page it already read is still the current one
	// without reading it again. The exchanges themselves are a route of their
	// own, which the pane calls only when somebody opens them.
	TurnsTotal int64 `json:"turns_total"`
	// Migrated says this task came from the ledger the platform retired. The
	// panel uses it to avoid presenting a carried-over task as one written
	// here; it is not a reason to offer a separate history, because the task's
	// turns are its history either way.
	Migrated bool `json:"migrated,omitempty"`
	// Waiting is the task's latest turn when that turn still needs somebody.
	// Absent when nothing is waiting: a zeroed record on every row would read
	// as "nothing asked here" rather than "nothing to say".
	Waiting *TaskWaiting `json:"waiting,omitempty"`
	// Choices are what the session offers a person to say next. They belong to
	// one turn — ChoiceTurnID — because what makes a pick current is that turn
	// still being the task's latest. Absent when it offers none; the app then
	// says so rather than inventing something.
	Choices []TaskChoice `json:"choices,omitempty"`
	// ChoiceTurnID is the turn Choices were offered on, sent back with a pick
	// so the platform can refuse one from a screen the task has moved past. It
	// replaced the ledger's (choice id, basis version) pair: the turn is the
	// one fact that decides currency, and a version carried beside a choice id
	// was a second way of saying it that could disagree with the first.
	ChoiceTurnID int64 `json:"choice_turn_id,omitempty"`
	// ChoiceTurn is what that same turn said, so the pane can show the words
	// the choices answer directly above them. Absent when it offers no choices
	// — there is then nothing on screen for it to belong to.
	ChoiceTurn *TaskChoiceTurn `json:"choice_turn,omitempty"`
}

// TaskChoiceTurn is the three things the turn that offered Choices said: the
// words that were sent, the parent's account of what came back, and the
// question the choices are answers to.
//
// It is projected from the very turn Choices and ChoiceTurnID come from, in one
// place, so what is shown above a set of choices cannot be some other turn's
// words. Nothing here is read with a second call: the turn the page already
// carries is the whole source, so a pane with the exchanges closed shows these
// without one line of the task's history being read.
//
// It deliberately carries no response: the session's verbatim answer is long,
// is already on screen when somebody opens the exchanges, and is not what a
// person needs to understand the question in front of them. An empty field is
// one the records do not have, and the pane leaves it out rather than filling
// it from another turn.
type TaskChoiceTurn struct {
	Request     string `json:"request,omitempty"`
	Explanation string `json:"explanation,omitempty"`
	Question    string `json:"question,omitempty"`
}

// TaskWaiting is the open turn on a task: what was asked, what the session is
// asking back, and whether anybody has picked it up. Running is a turn the
// parent has started and not finished; otherwise nobody has taken it yet.
//
// Request stays in this answer. The detail pane reads the words again with the
// task's exchanges — that turn's own request — and draws them from there, so
// nothing this route sends is what the pane renders. But this is the shape the
// route has always answered with, and a field is only ours to drop once we know
// what reads it; a client we cannot see is not the same as no client. Dropping
// it would save a message body per waiting task per poll and break anybody who
// reads it, which is the wrong trade for a route that is otherwise this small.
type TaskWaiting struct {
	TurnID  int64  `json:"turn_id"`
	Request string `json:"request,omitempty"`
	// Question is what the session is asking about this task, which is the
	// nearest thing the agreed records have to "why this is stuck". The ledger
	// carried a free-text reason and next action in its session relations;
	// those were the parent's own bookkeeping and are not part of what a person
	// reads, so the panel shows the question the session actually asked.
	Question string `json:"question,omitempty"`
	Running  bool   `json:"running,omitempty"`
}

// TaskChoice is one sentence a turn offers for the next one. The app shows
// Label and fills its request box with RequestText. It is a sentence to send,
// not a permission: picking one approves no tool and no merge.
type TaskChoice struct {
	Label       string `json:"label"`
	RequestText string `json:"request_text"`
}

// TaskList is the panel's whole payload.
type TaskList struct {
	// SessionID is the platform session these tasks were read from. The app
	// keeps it so a page that arrives after the conversation moved to another
	// session is discarded instead of shown under the new one.
	SessionID string         `json:"session_id"`
	Tasks     []Task         `json:"tasks"`
	Counts    map[string]int `json:"counts"`
	// Truncated says the walk stopped at its page bound with more to come, so
	// the panel can say the list is partial instead of implying it is whole.
	Truncated bool `json:"truncated"`
}

// workTaskPage is session-platform's answer to the read-only work task route.
// It is decoded into the fields the panel reads and no others — a struct that
// mirrored the whole platform shape would quietly grow fields nothing renders.
type workTaskPage struct {
	Tasks []struct {
		Task struct {
			ID                 string  `json:"id"`
			Title              string  `json:"title"`
			Description        *string `json:"description"`
			CompletionCriteria *string `json:"completion_criteria"`
			Status             string  `json:"status"`
		} `json:"task"`
		TurnsTotal int64 `json:"turns_total"`
		LatestTurn *struct {
			ID      int64   `json:"id"`
			Request *string `json:"request"`
			// Explanation is the parent's own account of what came back on this
			// turn, which the pane shows above the choices the same turn
			// offered. The platform already sends it on this page, so reading
			// it costs nothing more than decoding it — the alternative was a
			// second call per open task for words the page was carrying all
			// along.
			Explanation *string      `json:"explanation"`
			Question    *string      `json:"question"`
			Choices     []TaskChoice `json:"choices"`
			StartedAt   *string      `json:"started_at"`
			EndedAt     *string      `json:"ended_at"`
		} `json:"latest_turn"`
		Migrated bool `json:"migrated"`
	} `json:"tasks"`
	NextCursor string `json:"next_cursor"`
}

// Tasks reads the work records of the codex session answering userID's
// conversation. The session is taken from the user's own chat_sessions row, so
// a caller cannot name someone else's session; a conversation that has not
// opened a session yet simply has no tasks.
func (e *Engine) Tasks(ctx context.Context, userID int64) (TaskList, error) {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return TaskList{}, err
	}
	if conv.platformSessionID == "" {
		return emptyTaskList(""), nil
	}
	return e.tasksOf(ctx, conv.platformSessionID)
}

func (e *Engine) tasksOf(ctx context.Context, sessionID string) (TaskList, error) {
	out := emptyTaskList(sessionID)
	seen := map[string]bool{}
	cursor := ""
	for page := 0; ; page++ {
		if page == workMaxPages {
			// More pages remain. Say so rather than presenting a prefix as the
			// whole list.
			out.Truncated = true
			break
		}
		decoded, err := e.workTaskPage(ctx, sessionID, cursor)
		if err != nil {
			return TaskList{}, err
		}
		for _, row := range decoded.Tasks {
			// Keyset pages are not a snapshot, and a retried page can repeat a
			// row; the first sighting of an ID wins.
			if row.Task.ID == "" || seen[row.Task.ID] {
				continue
			}
			seen[row.Task.ID] = true
			// There is no monitoring row to filter out any more: the parent's
			// loops are a record of their own on the slot and were never tasks
			// (platform AC-N10). A filter here would be dead code that reads
			// like a live rule.
			task := Task{
				ID:                 row.Task.ID,
				Title:              strings.TrimSpace(row.Task.Title),
				Status:             normaliseStatus(row.Task.Status),
				Description:        strings.TrimSpace(derefString(row.Task.Description)),
				CompletionCriteria: strings.TrimSpace(derefString(row.Task.CompletionCriteria)),
				TurnsTotal:         row.TurnsTotal,
				Migrated:           row.Migrated,
			}
			if turn := row.LatestTurn; turn != nil {
				// Only the latest turn's choices are current, so those are the
				// only ones the panel may offer — and the turn they belong to
				// travels with them.
				if len(turn.Choices) > 0 {
					task.Choices = turn.Choices
					task.ChoiceTurnID = turn.ID
					// What that turn said travels with them, from this one
					// read of this one turn. Deciding it anywhere else — the
					// open turn below, the last row of a history page — could
					// put one turn's words over another turn's choices.
					task.ChoiceTurn = choiceTurnOf(
						derefString(turn.Request),
						derefString(turn.Explanation),
						derefString(turn.Question),
					)
				}
				// A turn with words and no end is one somebody still owes an
				// answer. Started means the parent took it; otherwise nobody
				// has yet. This is the platform's own definition of an open
				// turn, applied to the one turn this page carries.
				if turn.Request != nil && turn.EndedAt == nil {
					task.Waiting = &TaskWaiting{
						TurnID:   turn.ID,
						Request:  strings.TrimSpace(derefString(turn.Request)),
						Question: strings.TrimSpace(derefString(turn.Question)),
						Running:  turn.StartedAt != nil,
					}
				}
			}
			out.Tasks = append(out.Tasks, task)
			out.Counts[task.Status]++
		}
		// An empty cursor means the walk is done.
		if decoded.NextCursor == "" {
			break
		}
		// A cursor that does not move would loop. It also means the walk did
		// not reach the end, so it ends as truncated rather than as complete.
		if decoded.NextCursor == cursor {
			out.Truncated = true
			break
		}
		cursor = decoded.NextCursor
	}
	return out, nil
}

func (e *Engine) workTaskPage(ctx context.Context, sessionID, cursor string) (workTaskPage, error) {
	query := url.Values{}
	query.Set("limit", fmt.Sprint(workPageLimit))
	if cursor != "" {
		query.Set("after_id", cursor)
	}
	path := "/sessions/" + url.PathEscape(sessionID) + "/work/tasks?" + query.Encode()
	status, raw, err := e.platform.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return workTaskPage{}, fmt.Errorf("read work tasks: %w", err)
	}
	switch status {
	case http.StatusOK:
	case http.StatusNotImplemented, http.StatusNotFound:
		// 501 is a session whose type keeps no work records; 404 is a control
		// plane without the route at all, or a session this deployment no
		// longer has. None of them is an empty list, so the panel must not
		// claim there is no work.
		return workTaskPage{}, ErrTasksUnsupported
	case http.StatusConflict:
		// The session is snapshotted or holds no pod. Its rows are still on
		// the slot, so this is not an empty list — and the panel must not wake
		// the session to read them.
		return workTaskPage{}, ErrTasksUnavailable
	default:
		return workTaskPage{}, platformError("work tasks", status, raw)
	}
	var decoded workTaskPage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return workTaskPage{}, fmt.Errorf("decode work tasks: %w", err)
	}
	return decoded, nil
}

func emptyTaskList(sessionID string) TaskList {
	counts := make(map[string]int, len(taskStatuses))
	for _, s := range taskStatuses {
		counts[s] = 0
	}
	return TaskList{SessionID: sessionID, Tasks: []Task{}, Counts: counts}
}

// normaliseStatus is the one place a work record's status becomes the status
// this route answers with. Every read goes through it — the task list and a
// task's own turn page alike — so a second route cannot answer with a status
// the first one folded away.
func normaliseStatus(raw string) string {
	candidate := strings.ToLower(strings.TrimSpace(raw))
	if candidate == taskStatusBlocked {
		return TaskStatusReview
	}
	for _, s := range taskStatuses {
		if candidate == s {
			return s
		}
	}
	return TaskStatusUnknown
}

// choiceTurnOf is the turn's three words as the pane needs them, or nil when
// the records have none of them — a turn can offer sentences without having
// said anything itself, and an object of three empty strings would tell the app
// there is something to draw.
func choiceTurnOf(request, explanation, question string) *TaskChoiceTurn {
	out := TaskChoiceTurn{
		Request:     strings.TrimSpace(request),
		Explanation: strings.TrimSpace(explanation),
		Question:    strings.TrimSpace(question),
	}
	if out == (TaskChoiceTurn{}) {
		return nil
	}
	return &out
}

// derefString is "not recorded" read as the empty string. The platform sends
// null for a field it never had, which is deliberately not an empty string on
// the wire; the panel shows nothing for either, so they collapse here and the
// distinction is not carried further than it means anything.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
