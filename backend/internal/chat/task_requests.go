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

// A request the user leaves on one task of the chat session's ledger.
//
// It is deliberately not a chat message. The app used to send follow-ups
// through POST /api/chat/messages, which meant a request could not be written
// while a reply was being composed and, once sent, lived only in the
// conversation. Stored on the slot's ledger instead, it survives the reply, a
// restart and a snapshot, and the session finds it the next time it reads its
// own ledger — the request's own status is what says whether that happened.
//
// Nothing here starts a turn and nothing here writes a task field: the platform
// appends a row of its own beside the task (session-platform AC-N7).

const (
	// TaskRequestTextMax bounds one request's words. Smaller than the ledger's
	// own 8 KiB cap, because this is a sentence a person types about one task.
	TaskRequestTextMax = 4000

	// Lifecycle values the ledger reports back. The app shows "stored and
	// waiting" for the first three and "the session answered" only for the
	// last two — storing a request is never the same as having it dealt with.
	TaskRequestPending     = "pending"
	TaskRequestClaimed     = "claimed"
	TaskRequestReconciling = "reconciling"
	TaskRequestProcessed   = "processed"
	TaskRequestFailed      = "failed"
)

var (
	// ErrTaskRequestUnsupported is a session whose platform has no ledger write
	// route (an older control plane) or whose type keeps no ledger at all. It is
	// distinct from a refusal so the app can say the session cannot take
	// requests rather than implying the request was wrong.
	ErrTaskRequestUnsupported = errors.New("this chat session cannot store task requests")
	// ErrTaskRequestUnavailable is a ledger that exists but cannot be written
	// right now: the session holds no live pod. The request was not stored.
	ErrTaskRequestUnavailable = errors.New("the chat session is not awake, so the request could not be stored")
	// ErrTaskNotFound is a task this session's ledger does not have — a row that
	// has since been deleted, or an id the panel no longer shows.
	ErrTaskNotFound = errors.New("that task is not on this session's ledger")
	// ErrTaskRequestConflict is the request id already used for different words.
	// A retry of the same words is not an error; different words under a used id
	// is, because it would otherwise silently replace what the person asked.
	ErrTaskRequestConflict = errors.New("that request id is already used for a different request")
	// ErrTaskRequestInvalid is a malformed id or text.
	ErrTaskRequestInvalid = errors.New("the request id or text is not usable")
)

// ledgerRequestID is the ledger's own id shape. The app's idempotency key has
// to fit it, so an id that would be refused on the slot is refused here with a
// message about the field.
var ledgerRequestID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// TaskRequest is one stored request as the app reads it. Status and result are
// the parent's own record: the app shows them and never infers them.
type TaskRequest struct {
	ID        string `json:"id"`
	TaskID    string `json:"task_id"`
	Seq       int64  `json:"seq"`
	Text      string `json:"text"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at"`
	Status    string `json:"status"`
	Result    string `json:"result,omitempty"`
	SettledAt string `json:"settled_at,omitempty"`
}

// TaskRequestStored is the answer to storing one. Replayed is true when the
// same id had already stored the same words, which is what makes the app's
// retry of an answer it never saw safe to send.
type TaskRequestStored struct {
	Request  TaskRequest `json:"request"`
	Replayed bool        `json:"replayed"`
}

// TaskRequestSummary is how many requests are waiting on a task, carried in the
// task list so the panel can show it without a second call.
type TaskRequestSummary struct {
	Pending     int   `json:"pending"`
	Claimed     int   `json:"claimed"`
	Reconciling int   `json:"reconciling"`
	LastSeq     int64 `json:"last_seq"`
}

// Open says the parent has not settled these yet — the number the panel shows
// as waiting.
func (s TaskRequestSummary) Open() int { return s.Pending + s.Claimed + s.Reconciling }

// CreateTaskRequest stores one request on taskID of the ledger belonging to the
// session that answers userID's own conversation.
//
// The session is taken from the user's own chat_sessions row, never from the
// request, so a caller cannot leave a request on someone else's ledger; and the
// author is derived here from the authenticated user rather than taken from the
// body, so the app cannot name who asked. The author is bookkeeping on the
// slot — the ledger grants no tool, approval or merge on the strength of it.
//
// requestID is the app's idempotency key: the same id with the same text stores
// nothing twice, and the same id with different text is a conflict rather than a
// second request. It does not claim a turn, so a request can be stored while a
// reply is still being written.
func (e *Engine) CreateTaskRequest(ctx context.Context, userID int64, taskID, requestID, text string) (TaskRequestStored, error) {
	var out TaskRequestStored
	body := strings.TrimSpace(text)
	if body == "" || len(body) > TaskRequestTextMax {
		return out, fmt.Errorf("%w: text must be 1..%d characters", ErrTaskRequestInvalid, TaskRequestTextMax)
	}
	if !ledgerRequestID.MatchString(requestID) {
		return out, fmt.Errorf("%w: request id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	if !ledgerRequestID.MatchString(taskID) {
		return out, fmt.Errorf("%w: task id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return out, err
	}
	if conv.platformSessionID == "" {
		// No session has answered this conversation yet, so there is no ledger
		// and no task to leave a request on.
		return out, ErrNoConversation
	}
	payload, err := json.Marshal(map[string]string{
		"id": requestID, "text": body, "author": taskRequestAuthor(userID),
	})
	if err != nil {
		return out, err
	}
	path := "/sessions/" + url.PathEscape(conv.platformSessionID) +
		"/ledger/tasks/" + url.PathEscape(taskID) + "/requests"
	status, raw, err := e.platform.Do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return out, fmt.Errorf("store task request: %w", err)
	}
	switch status {
	case http.StatusCreated, http.StatusOK:
		var answer struct {
			Request  TaskRequest `json:"request"`
			Replayed bool        `json:"replayed"`
		}
		if err := json.Unmarshal(raw, &answer); err != nil {
			return out, fmt.Errorf("decode stored task request: %w", err)
		}
		// The platform answers 200 for a replay and 201 for a first store; the
		// body says so too, and the status is what the app can rely on.
		answer.Replayed = answer.Replayed || status == http.StatusOK
		return TaskRequestStored{Request: answer.Request, Replayed: answer.Replayed}, nil
	case http.StatusNotFound:
		// Either the session is gone or the task is not on its ledger. Both mean
		// the panel is looking at something the ledger no longer has.
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
		return out, ErrTaskRequestConflict
	case http.StatusNotImplemented, http.StatusMethodNotAllowed:
		// 501 is a session that keeps no ledger; 405 is a control plane without
		// the write route yet. Neither is the user's mistake.
		return out, ErrTaskRequestUnsupported
	case http.StatusBadRequest:
		return out, fmt.Errorf("%w: the platform refused the request's fields", ErrTaskRequestInvalid)
	default:
		return out, platformError("ledger request", status, raw)
	}
}

// taskRequestAuthor is who asked, as the slot records it. It is derived from
// the authenticated user id and fits the ledger's id shape, so the app has no
// way to write a different author — and a reader of the slot can tell two
// users' requests apart without the ledger carrying anything personal.
func taskRequestAuthor(userID int64) string {
	return fmt.Sprintf("user-%d", userID)
}
