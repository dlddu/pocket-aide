// One task's turns, as the detail pane needs them (platform AC-N9, AC-N10).
//
// These are the whole history of a task now. There used to be a second pane
// underneath built on the ledger's history route, and both the route and the
// rows behind it are gone: a past management event was never a turn, and
// dressing one up as a turn would put words in a person's mouth. So a task's
// exchanges are its history, for a task carried over from the ledger exactly
// as for one written here — the pane no longer has a second source to offer,
// and does not imply one is missing.
package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// TaskTurnPageLimitMax bounds one page of turns, matching the platform's cap.
const TaskTurnPageLimitMax = 100

// TaskTurn is one exchange on a task.
//
// Request is the person's own words. Instruction is what the parent passed to
// the session doing the work. Response is that session's final answer verbatim
// — never a summary the parent wrote, which is Explanation. A field the
// platform left null has not happened or was never recorded, and the pane says
// so rather than showing a blank as if it were the answer, so every one of them
// is a pointer here too.
type TaskTurn struct {
	ID          int64   `json:"id"`
	Request     *string `json:"request"`
	Instruction *string `json:"instruction"`
	Response    *string `json:"response"`
	Explanation *string `json:"explanation"`
	Question    *string `json:"question"`
	// Choices are the sentences this turn offers for the next one. The app
	// shows the label and fills the request box with request_text; sending it
	// creates the next turn's request. There is no id — which turn a pick
	// answers is what decides whether it is still current, and the platform
	// refuses a pick from a turn that is no longer the latest.
	Choices   []TaskTurnChoice `json:"choices"`
	Artifacts []string         `json:"artifacts"`
	// RequestedAt/StartedAt/EndedAt are when the turn arrived, was picked up
	// and finished. A turn with no EndedAt is still somebody's to answer.
	RequestedAt string  `json:"requested_at"`
	StartedAt   *string `json:"started_at"`
	EndedAt     *string `json:"ended_at"`
	// Origin says where the turn came from: empty for one the session wrote,
	// "user_request" for one carried over from the retired ledger, and
	// "app_request" for one a person wrote through this app. The pane uses it
	// to avoid presenting a carried-over exchange as something written here —
	// the ledger's own words were a stored request, never an exchange the
	// session had.
	Origin string `json:"origin,omitempty"`
}

// TaskTurnChoice is one sentence a turn offers for the next one. It is a
// sentence to send, not a permission.
type TaskTurnChoice struct {
	Label       string `json:"label"`
	RequestText string `json:"request_text"`
}

// TaskTurnTask is the task those turns belong to, in the agreed shape. It
// travels with them so the pane does not need a second call that could land on
// a different version than the turns it is showing.
type TaskTurnTask struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Description        *string `json:"description"`
	CompletionCriteria *string `json:"completion_criteria"`
	Status             string  `json:"status"`
	Version            int64   `json:"version"`
	SessionID          *string `json:"session_id"`
	CreatedAt          *string `json:"created_at"`
	UpdatedAt          *string `json:"updated_at"`
}

// TaskTurnCursor says which page of a task's turns to read, and from which end.
//
// The pane shows the latest exchange at the top, so it reads Newest pages: the
// first one with no bound at all, then older ones from the cursor the page
// before it gave. Ascending is still what the fold behind the review materials
// walks, and is the zero value here so that path is unchanged.
type TaskTurnCursor struct {
	// After is an exclusive lower bound on the oldest-first page.
	After int64
	// Newest asks for the newest-first page. Before is then an exclusive upper
	// bound and zero is "from this task's newest turn", which is how the pane
	// gets the latest page without reading the history to find where it starts.
	Newest bool
	Before int64
}

// TaskTurnPage is the detail pane's payload.
type TaskTurnPage struct {
	TaskID      string       `json:"task_id"`
	TaskVersion int64        `json:"task_version"`
	Task        TaskTurnTask `json:"task"`
	Turns       []TaskTurn   `json:"turns"`
	// NextCursor continues the page in the direction it was read: older turns
	// on a newest-first page, newer ones on an oldest-first one.
	NextCursor int64 `json:"next_cursor,omitempty"`
	// Order is "desc" on a newest-first page. It is checked rather than
	// assumed: a slot that predates the parameter ignores it and answers the
	// oldest page, and a pane that read that page as if it were reversed would
	// put the task's first exchange where its latest belongs and then page in
	// a circle.
	Order string `json:"order,omitempty"`
}

// TaskTurnOrderDesc is what a newest-first page says it is.
const TaskTurnOrderDesc = "desc"

// TaskTurns reads one page of a task's turns from the platform session
// answering userID's conversation. The session is taken from the user's own
// row, so a caller cannot name someone else's.
func (e *Engine) TaskTurns(
	ctx context.Context, userID int64, taskID string, cursor TaskTurnCursor, limit int,
) (TaskTurnPage, error) {
	out := TaskTurnPage{Turns: []TaskTurn{}}
	if !workRecordID.MatchString(taskID) {
		return out, fmt.Errorf("%w: task id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	if cursor.After < 0 || cursor.Before < 0 || limit < 0 || limit > TaskTurnPageLimitMax {
		return out, fmt.Errorf("%w: the cursor and limit must be in range", ErrTaskRequestInvalid)
	}
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return out, err
	}
	if conv.platformSessionID == "" {
		return out, ErrNoConversation
	}
	query := url.Values{}
	switch {
	case cursor.Newest:
		// Set even when it is zero: on the platform's route the parameter's
		// presence is what asks for the newest-first page.
		query.Set("before_turn_id", strconv.FormatInt(cursor.Before, 10))
	case cursor.After > 0:
		query.Set("after_turn_id", strconv.FormatInt(cursor.After, 10))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	path := "/sessions/" + url.PathEscape(conv.platformSessionID) +
		"/work/tasks/" + url.PathEscape(taskID) + "/turns"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	status, raw, err := e.platform.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return out, fmt.Errorf("read task turns: %w", err)
	}
	switch status {
	case http.StatusOK:
		if err := json.Unmarshal(raw, &out); err != nil {
			return TaskTurnPage{Turns: []TaskTurn{}}, fmt.Errorf("decode task turns: %w", err)
		}
		if out.Turns == nil {
			out.Turns = []TaskTurn{}
		}
		if cursor.Newest && out.Order != TaskTurnOrderDesc {
			// The slot answered a page, but not the one that was asked for: it
			// does not have the parameter and served the oldest turns. Handing
			// that back would have the pane show the task's first exchange as
			// its latest and then page in a circle, so it is refused as a
			// contract the session cannot serve yet.
			return TaskTurnPage{Turns: []TaskTurn{}}, ErrTaskTurnsOrderUnsupported
		}
		return out, nil
	case http.StatusNotFound:
		// The task is not on this slot, or this platform has no work route at
		// all. Either way it is not "a task with no turns", which would read as
		// "nothing has happened yet" and hide a wrong id.
		return out, ErrTaskNotFound
	case http.StatusConflict:
		// The session is snapshotted, holds no pod, or a migration holds its
		// database. The turns are still on the slot, so this is not an empty
		// task — and reading must not wake the session to prove it.
		return out, ErrTaskRequestUnavailable
	case http.StatusNotImplemented, http.StatusMethodNotAllowed:
		// A platform or a session pod that keeps no work records. Not an empty
		// task: the pane says the records cannot be read here rather than that
		// there is no work.
		return out, ErrTaskRequestUnsupported
	case http.StatusBadRequest:
		return out, fmt.Errorf("%w: the platform refused the page request", ErrTaskRequestInvalid)
	default:
		return out, platformError("work turns", status, raw)
	}
}
