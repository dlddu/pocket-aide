package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// The chat tab's wide layout shows the task ledger of the codex session that
// answers this conversation. The ledger is the chat slot's own SQLite
// bookkeeping, read through session-platform's read-only ledger route; nothing
// here parses chat text or asks a model to summarise, and nothing here writes.
const (
	// ledgerPageLimit is the platform's per-page cap for the ledger route.
	ledgerPageLimit = 100
	// ledgerMaxPages bounds a keyset walk. Reaching it is reported as a
	// truncated list rather than a finished one — a panel that silently drops
	// the tail would read as "this is all the work there is".
	ledgerMaxPages = 50
	// monitorTaskKind marks a record the parent's own monitoring loop keeps for
	// bookkeeping. The ledger schema has no kind column, so the kind lives in
	// details; that is the contract the parent ledger writes (the platform's
	// docs/session-ledger.md). Such rows are not the user's work and stay out
	// of the panel.
	monitorTaskKind = "monitor"
)

// TaskStatus values the ledger allows. Anything else — including a status a
// newer platform starts writing — is reported as TaskStatusUnknown so the
// panel can show it without guessing.
const (
	TaskStatusPlanned   = "planned"
	TaskStatusRunning   = "running"
	TaskStatusBlocked   = "blocked"
	TaskStatusReview    = "review"
	TaskStatusCompleted = "completed"
	TaskStatusUnknown   = "unknown"
)

// taskStatuses is the display order the panel counts in: what needs a person
// first, then what is merely waiting, then what is done.
var taskStatuses = []string{
	TaskStatusBlocked, TaskStatusReview, TaskStatusRunning,
	TaskStatusPlanned, TaskStatusCompleted, TaskStatusUnknown,
}

// ErrTasksUnsupported means this conversation's session keeps no task ledger —
// an older control plane without the route, or a workload type that gets no
// session DB. It is distinct from "no tasks yet" so the app can say so.
var ErrTasksUnsupported = errTasks{"session keeps no task ledger"}

// ErrTasksUnavailable means the session keeps a ledger but cannot be read right
// now: it holds no live pod. The rows are still on the slot's volume, so this
// is not an empty ledger — and reading never wakes the session to get at them.
var ErrTasksUnavailable = errTasks{"the chat session is not awake, so its task list cannot be read"}

type errTasks struct{ msg string }

func (e errTasks) Error() string { return e.msg }

// Task is one ledger record as the app needs it: what it is, where it stands,
// and — when the ledger carries them — why it is stuck and what comes next.
type Task struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Stage  string `json:"stage"`
	// Reason and NextAction come from the task's *current* state, not from the
	// first session it ever used. Empty means the ledger has none recorded; the
	// panel says so rather than leaving a blank.
	Reason     string `json:"reason,omitempty"`
	NextAction string `json:"next_action,omitempty"`
	// Requests is how many user requests are waiting on this task, absent when
	// none ever arrived. It comes from the ledger page itself, so the panel can
	// show "someone asked something" without a second call — and the parent's
	// own progress on them is what moves the counts.
	Requests *TaskRequestSummary `json:"requests,omitempty"`
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

// ledgerSessionNote is one session entry of a ledger task, narrowed to what the
// panel reads: whether the entry is still live, and the note it carries.
type ledgerSessionNote struct {
	Cleanup    string `json:"cleanup"`
	Reason     string `json:"reason"`
	NextAction string `json:"next_action"`
}

// ledgerPage is session-platform's answer to the read-only ledger route.
type ledgerPage struct {
	Tasks []struct {
		ID       string              `json:"id"`
		Requests *TaskRequestSummary `json:"requests"`
		Task     struct {
			Title    string              `json:"title"`
			Status   string              `json:"status"`
			Stage    string              `json:"stage"`
			Sessions []ledgerSessionNote `json:"sessions"`
			Details  map[string]any      `json:"details"`
		} `json:"task"`
	} `json:"tasks"`
	NextCursor string `json:"next_cursor"`
}

// Tasks reads the task ledger of the codex session answering userID's
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
		if page == ledgerMaxPages {
			// More pages remain. Say so rather than presenting a prefix as the
			// whole list.
			out.Truncated = true
			break
		}
		decoded, err := e.ledgerPage(ctx, sessionID, cursor)
		if err != nil {
			return TaskList{}, err
		}
		for _, row := range decoded.Tasks {
			// Keyset pages are not a snapshot, and a retried page can repeat a
			// row; the first sighting of an ID wins.
			if row.ID == "" || seen[row.ID] {
				continue
			}
			seen[row.ID] = true
			if detailsKind(row.Task.Details) == monitorTaskKind {
				continue
			}
			task := Task{
				ID:     row.ID,
				Title:  strings.TrimSpace(row.Task.Title),
				Status: normaliseStatus(row.Task.Status),
				Stage:  strings.TrimSpace(row.Task.Stage),
			}
			task.Reason, task.NextAction = currentNote(row.Task.Sessions)
			// Only when something is actually waiting: a zeroed summary on
			// every row would read as "nothing asked here" rather than
			// "nothing to say".
			if row.Requests != nil && row.Requests.Open() > 0 {
				summary := *row.Requests
				task.Requests = &summary
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

func (e *Engine) ledgerPage(ctx context.Context, sessionID, cursor string) (ledgerPage, error) {
	query := url.Values{}
	query.Set("limit", fmt.Sprint(ledgerPageLimit))
	if cursor != "" {
		query.Set("after_id", cursor)
	}
	path := "/sessions/" + url.PathEscape(sessionID) + "/ledger/tasks?" + query.Encode()
	status, raw, err := e.platform.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return ledgerPage{}, fmt.Errorf("read task ledger: %w", err)
	}
	switch status {
	case http.StatusOK:
	case http.StatusNotImplemented, http.StatusNotFound:
		// 501 is a session whose type keeps no ledger; 404 is a control plane
		// without the route at all, or a session this deployment no longer
		// has. None of them is an empty ledger, so the panel must not claim
		// there is no work.
		return ledgerPage{}, ErrTasksUnsupported
	case http.StatusConflict:
		// The session is snapshotted or holds no pod. Its rows are still on
		// the slot, so this is not an empty ledger — and the panel must not
		// wake the session to read them.
		return ledgerPage{}, ErrTasksUnavailable
	default:
		return ledgerPage{}, platformError("ledger", status, raw)
	}
	var decoded ledgerPage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return ledgerPage{}, fmt.Errorf("decode task ledger: %w", err)
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

func normaliseStatus(raw string) string {
	candidate := strings.ToLower(strings.TrimSpace(raw))
	for _, s := range taskStatuses {
		if candidate == s {
			return s
		}
	}
	return TaskStatusUnknown
}

// currentNote is why the task stands where it does now and what comes next. A
// task collects a session entry as it goes, so its present state is the *last*
// entry carrying a note — not the first, which is where the task started.
// Entries the parent has finished with (cleanup "deleted") are history, so a
// live entry wins over them even when it comes earlier in the list.
func currentNote(sessions []ledgerSessionNote) (reason, nextAction string) {
	for i := len(sessions) - 1; i >= 0; i-- {
		s := sessions[i]
		r, n := strings.TrimSpace(s.Reason), strings.TrimSpace(s.NextAction)
		if r == "" && n == "" {
			continue
		}
		if strings.ToLower(strings.TrimSpace(s.Cleanup)) == "deleted" {
			// Keep it as a fallback, but a live entry describes the present
			// better, so keep looking.
			if reason == "" && nextAction == "" {
				reason, nextAction = r, n
			}
			continue
		}
		return r, n
	}
	return reason, nextAction
}

// detailsKind reads details.kind, the contract that marks a monitoring
// bookkeeping row. A details map without it yields "".
func detailsKind(details map[string]any) string {
	if details == nil {
		return ""
	}
	kind, _ := details["kind"].(string)
	return strings.ToLower(strings.TrimSpace(kind))
}
