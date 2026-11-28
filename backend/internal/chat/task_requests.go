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
//
// One thing did change about what the answer can say (platform AC-N14). A task
// that is waiting for the person's judgement is one the session's monitor has
// stopped looking at, so a request on it used to lie in the records until that
// monitor's next scheduled pass. The platform now hands it over the moment it
// lands, and says whether that worked — which is the third thing this file
// relays and the only new one: stored, handed over, and (still only from the
// turn) done.

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
	// Relay is what the platform did about the monitor that was waiting on
	// this answer, and it is present only when there was something to do: a
	// request on a task that was waiting for a person's judgement is handed to
	// the session's own monitor lane at once, because that lane had stopped
	// looking at the task for exactly that reason.
	//
	// Absent means no handover was due — an ordinary request on a task nobody
	// was waiting on, or a session whose platform does not run the pair. The
	// app must not show absence as a failure: the request is stored either way,
	// and the difference is only whether somebody was told now or will find it
	// on their next pass.
	Relay *TaskRequestRelay `json:"relay,omitempty"`
}

// TaskRequestRelay is that handover's outcome, relayed from the platform
// unchanged. It carries no receipt — the platform's receipt names a run of a
// conversation inside the session's pod, which is nothing the app can use or
// should hold.
//
// Three things stay separate here on purpose, because the person's next move
// differs for each: the request is stored (there is an answer at all), the
// monitor has been given it (Accepted), and the work is done (which only the
// turn ending says — ChatTaskTurnRef.settled). Collapsing them tells somebody
// their request has been dealt with when it has only been filed.
type TaskRequestRelay struct {
	// State is "accepted" when the monitor took it — straight away if that
	// conversation was idle, behind its current turn if it was busy — or
	// "pending" when the platform recorded the handover and could not reach
	// it. Pending is not lost: the record is on the session's own storage and
	// the monitor's standing round reads the same turn.
	State string `json:"state"`
	// Attempts is how many times the platform has handed this over, so a
	// second delivery after a lost acknowledgement is distinguishable from a
	// first one.
	Attempts int64 `json:"attempts"`
}

// The two states a handover is reported in. Keep in sync with the platform
// (session-platform control-plane/internal/session/work.go).
const (
	TaskRequestRelayAccepted = "accepted"
	TaskRequestRelayPending  = "pending"
)

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
	path := "/sessions/" + url.PathEscape(conv.platformSessionID) +
		"/work/tasks/" + url.PathEscape(taskID) + "/requests"
	status, raw, err := e.platform.Do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return out, fmt.Errorf("store task request: %w", err)
	}
	switch status {
	case http.StatusCreated, http.StatusOK:
		var answer TaskRequestStored
		if err := json.Unmarshal(raw, &answer); err != nil {
			return out, fmt.Errorf("decode stored task request: %w", err)
		}
		// The platform answers 200 for a replay and 201 for a first store; the
		// body says so too, and the status is what the app can rely on.
		answer.Replayed = answer.Replayed || status == http.StatusOK
		// A handover state this app has no sentence for is dropped rather than
		// passed on. The app's three cases are meant to be exhaustive, and a
		// fourth word arriving from a newer platform would otherwise reach a
		// screen that can only show it as the absence of a handover anyway —
		// which is what dropping it says, without inventing a meaning.
		if answer.Relay != nil &&
			answer.Relay.State != TaskRequestRelayAccepted &&
			answer.Relay.State != TaskRequestRelayPending {
			answer.Relay = nil
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
