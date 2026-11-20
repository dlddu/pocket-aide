// One task's turns, as the detail pane needs them (platform AC-N9).
//
// Why this exists beside TaskEvents: the task page and the history page are the
// ledger's shapes, and those have nowhere to put what the agreed model holds —
// a task's description and completion criteria, and a turn's instruction, the
// session's own final answer, the parent's explanation, its question and the
// artifacts. A detail pane built on them alone can say a task exists and that
// someone asked something, but not what was asked, what was done, or what came
// back. The turns route carries the records themselves.
//
// The two are not alternatives. A task migrated from the ledger keeps its
// ledger history and is read with TaskEvents; a task created after the cutover
// has no ledger history and none is invented for it — its turns *are* its
// history. The pane shows the turns, and the ledger history underneath only
// when there is one.
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
	// "user_request" for one carried over from the old ledger, "app_request"
	// for one a person wrote through this app after the cutover. The pane uses
	// it to avoid presenting a carried-over exchange as something written here.
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

// TaskTurnPage is the detail pane's payload.
type TaskTurnPage struct {
	TaskID      string       `json:"task_id"`
	TaskVersion int64        `json:"task_version"`
	Task        TaskTurnTask `json:"task"`
	Turns       []TaskTurn   `json:"turns"`
	NextCursor  int64        `json:"next_cursor,omitempty"`
	// LedgerHistory says this task was carried over from the old ledger, so the
	// pane also has a history to offer under the turns. A task created after
	// the cutover has none, and the pane does not imply one is missing.
	LedgerHistory bool `json:"ledger_history"`
}

// TaskTurns reads one page of a task's turns from the platform session
// answering userID's conversation. The session is taken from the user's own
// row, so a caller cannot name someone else's.
func (e *Engine) TaskTurns(
	ctx context.Context, userID int64, taskID string, afterTurnID int64, limit int,
) (TaskTurnPage, error) {
	out := TaskTurnPage{Turns: []TaskTurn{}}
	if !ledgerRequestID.MatchString(taskID) {
		return out, fmt.Errorf("%w: task id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	if afterTurnID < 0 || limit < 0 || limit > TaskTurnPageLimitMax {
		return out, fmt.Errorf("%w: after_turn_id and limit must be in range", ErrTaskRequestInvalid)
	}
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return out, err
	}
	if conv.platformSessionID == "" {
		return out, ErrNoConversation
	}
	query := url.Values{}
	if afterTurnID > 0 {
		query.Set("after_turn_id", strconv.FormatInt(afterTurnID, 10))
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
		// A carried-over task has an old ledger history worth offering; one
		// created here does not, and is not shown as if its history were
		// missing.
		for _, turn := range out.Turns {
			if turn.Origin == "user_request" {
				out.LedgerHistory = true
				break
			}
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
