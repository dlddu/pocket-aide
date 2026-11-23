package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// A request the user leaves on one task of the chat session's work records.
//
// It is deliberately not a chat message. The app used to send follow-ups
// through POST /api/chat/messages, which meant a request could not be written
// while a reply was being composed and, once sent, lived only in the
// conversation. Stored on the slot instead, it survives the reply, a restart
// and a snapshot, and the parent finds it the next time it reads the task's
// open turns.
//
// What it becomes changed with the ledger's retirement (platform AC-N10): a
// person's words are now the task's **next turn** rather than a row beside it.
// That is why there is no stored-request record to read back and no request
// lifecycle to poll — the turn is the record, and whether it has been dealt
// with is whether that turn has ended. Nothing here writes a task field and
// nothing here starts a turn of the session's own.

const (
	// TaskRequestTextMax bounds one request's words. Smaller than the slot's
	// own cap, because this is a sentence a person types about one task.
	TaskRequestTextMax = 4000
)

var (
	// ErrTaskRequestUnsupported is a session whose platform has no work write
	// route (an older control plane) or whose type keeps no work records at
	// all. It is distinct from a refusal so the app can say the session cannot
	// take requests rather than implying the request was wrong.
	ErrTaskRequestUnsupported = errors.New("this chat session cannot store task requests")
	// ErrTaskRequestUnavailable is a slot that exists but cannot be written
	// right now: the session holds no live pod. The request was not stored.
	ErrTaskRequestUnavailable = errors.New("the chat session is not awake, so the request could not be stored")
	// ErrTaskNotFound is a task this session's records do not have — one that
	// has since been deleted, or an id the panel no longer shows.
	ErrTaskNotFound = errors.New("that task is not among this session's work records")
	// ErrTaskRequestConflict is the request id already used for different words.
	// A retry of the same words is not an error; different words under a used id
	// is, because it would otherwise silently replace what the person asked.
	ErrTaskRequestConflict = errors.New("that request id is already used for a different request")
	// ErrTaskRequestStaleChoice is a pick from a turn the task has moved past —
	// a screen that was already open when a newer turn arrived. The words the
	// person typed are still good, so the app keeps the draft, re-reads the
	// task, and says the choices changed.
	ErrTaskRequestStaleChoice = errors.New("that offered choice is no longer current; the task moved")
	// ErrTaskRequestInvalid is a malformed id or text.
	ErrTaskRequestInvalid = errors.New("the request id or text is not usable")
)

// workRecordID is the slot's own id shape. The app's idempotency key has to
// fit it, so an id that would be refused on the slot is refused here with a
// message about the field.
var workRecordID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// TaskRequestStored is the answer to storing one: the turn the request became.
// Replayed is true when the same id had already stored the same words, which is
// what makes the app's retry of an answer it never saw safe to send.
//
// There is no stored-request row here any more. The ledger kept one beside the
// task with a lifecycle of its own (pending, claimed, reconciling, processed,
// failed) that the app polled; the turn replaced all of it, and whether a
// request has been dealt with is whether its turn has ended. Keeping the old
// shape would have meant showing a status nothing on the slot writes.
type TaskRequestStored struct {
	Turn     TaskTurn `json:"turn"`
	Replayed bool     `json:"replayed"`
	// State and Accepted describe a paired send: the monitor owns the task, so
	// an instruction is first accepted durably ("accepted") and becomes a turn
	// when the monitor takes it ("forwarded"). Accepted true means the app
	// should show it as waiting, not done.
	State    string `json:"state,omitempty"`
	Accepted bool   `json:"accepted,omitempty"`
}

// CreateTaskRequest stores one request on taskID of the work records belonging
// to the session that answers userID's own conversation.
//
// The session is taken from the user's own chat_sessions row, never from the
// request, so a caller cannot leave a request on someone else's slot; and the
// author is derived here from the authenticated user rather than taken from the
// body, so the app cannot name who asked. The author is bookkeeping on the
// slot — nothing there grants a tool, approval or merge on the strength of it.
//
// requestID is the app's idempotency key: the same id with the same text stores
// nothing twice, and the same id with different text is a conflict rather than a
// second request. It does not claim a turn, so a request can be stored while a
// reply is still being written.
func (e *Engine) CreateTaskRequest(
	ctx context.Context, userID int64, taskID, requestID, text string, choiceTurnID int64,
) (TaskRequestStored, error) {
	var out TaskRequestStored
	body := strings.TrimSpace(text)
	if body == "" || len(body) > TaskRequestTextMax {
		return out, fmt.Errorf("%w: text must be 1..%d characters", ErrTaskRequestInvalid, TaskRequestTextMax)
	}
	if !workRecordID.MatchString(requestID) {
		return out, fmt.Errorf("%w: request id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	if !workRecordID.MatchString(taskID) {
		return out, fmt.Errorf("%w: task id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	if choiceTurnID < 0 {
		return out, fmt.Errorf("%w: a choice turn id must be a positive turn id", ErrTaskRequestInvalid)
	}
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return out, err
	}
	if conv.platformSessionID == "" {
		// No session has answered this conversation yet, so there are no work
		// records and no task to leave a request on.
		return out, ErrNoConversation
	}
	fields := map[string]any{"id": requestID, "text": body, "author": taskRequestAuthor(userID)}
	if choiceTurnID > 0 {
		// The turn whose offered sentences the person picked from. The slot
		// refuses a pick from a turn that is no longer the task's latest, so a
		// screen left open while the task moved on cannot act on it. Free text
		// omits it and is never refused for staleness.
		fields["choice_turn_id"] = choiceTurnID
	}
	payload, err := json.Marshal(fields)
	if err != nil {
		return out, err
	}
	// A paired chat session does not change a task: the instruction goes to
	// the monitor that owns it, under the app's own stable request id, and the
	// monitor decides and records. The platform accepts it durably even while
	// the monitor is frozen, so a retry with the same id is safe and never a
	// second delegation. An unpaired session keeps leaving the request on its
	// own task, as before.
	prefix, _, err := e.workPrefix(ctx, conv.platformSessionID)
	if err != nil {
		return out, err
	}
	path := prefix + "/work/tasks/" + url.PathEscape(taskID) + "/requests"
	status, raw, err := e.platform.Do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return out, fmt.Errorf("store task request: %w", err)
	}
	switch status {
	case http.StatusCreated, http.StatusOK, http.StatusAccepted:
		var answer TaskRequestStored
		if err := json.Unmarshal(raw, &answer); err != nil {
			return out, fmt.Errorf("decode stored task request: %w", err)
		}
		// The platform answers 200 for a replay and 201 for a first store; the
		// body says so too, and the status is what the app can rely on.
		answer.Replayed = answer.Replayed || status == http.StatusOK
		// The paired surface also answers 202: the instruction is durably
		// accepted but the monitor has not taken it as a turn yet. The app
		// shows that as waiting — not as done, and not as a failure to retry
		// blindly, because the same request id is already held.
		var pairedAnswer struct {
			Item *struct {
				State  string `json:"state"`
				TurnID *int64 `json:"turnId"`
			} `json:"item"`
		}
		if json.Unmarshal(raw, &pairedAnswer) == nil && pairedAnswer.Item != nil {
			answer.State = pairedAnswer.Item.State
			answer.Accepted = status == http.StatusAccepted || pairedAnswer.Item.TurnID == nil
		}
		return answer, nil
	case http.StatusNotFound:
		// Either the session is gone or the task is not among its records. Both
		// mean the panel is looking at something the slot no longer has.
		return out, ErrTaskNotFound
	case http.StatusConflict:
		// A used id with different words, a deleted task, or a session with no
		// live pod — the platform's own message distinguishes them.
		if strings.Contains(strings.ToLower(string(raw)), "no live pod") {
			return out, ErrTaskRequestUnavailable
		}
		if strings.Contains(strings.ToLower(string(raw)), "deleted") {
			return out, ErrTaskNotFound
		}
		// A pick from a turn the task has moved past: the screen is out of
		// date, not the words. Its own class, because the answer is to re-read
		// the task and keep what the person already typed.
		if lower := strings.ToLower(string(raw)); strings.Contains(lower, "no longer offered") ||
			strings.Contains(lower, "no longer current") {
			return out, ErrTaskRequestStaleChoice
		}
		return out, ErrTaskRequestConflict
	case http.StatusNotImplemented, http.StatusMethodNotAllowed:
		// 501 is a session that keeps no work records; 405 is a control plane
		// without the write route yet. Neither is the user's mistake.
		return out, ErrTaskRequestUnsupported
	case http.StatusBadRequest:
		return out, fmt.Errorf("%w: the platform refused the request's fields", ErrTaskRequestInvalid)
	default:
		return out, platformError("work request", status, raw)
	}
}

// taskRequestAuthor is who asked, as the slot records it. It is derived from
// the authenticated user id and fits the slot's id shape, so the app has no way
// to write a different author — and a reader of the slot can tell two users'
// requests apart without the records carrying anything personal.
func taskRequestAuthor(userID int64) string {
	return fmt.Sprintf("user-%d", userID)
}
