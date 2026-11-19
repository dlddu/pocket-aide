package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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
	// ErrTaskRequestStaleChoice is a picked choice the session no longer
	// offers — the task moved under a screen that was already open, or the
	// choice was never offered. The words the person typed are still good, so
	// the app keeps the draft, re-reads the task, and says the choices changed.
	ErrTaskRequestStaleChoice = errors.New("that offered choice is no longer current; the task moved")
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
	ChoiceID  string `json:"choice_id,omitempty"`
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
func (e *Engine) CreateTaskRequest(ctx context.Context, userID int64, taskID, requestID, text, choiceID string, choiceBasisVersion int64) (TaskRequestStored, error) {
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
	if choiceID != "" && !ledgerRequestID.MatchString(choiceID) {
		return out, fmt.Errorf("%w: choice id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	if choiceBasisVersion < 0 || (choiceID == "" && choiceBasisVersion != 0) {
		return out, fmt.Errorf("%w: a choice basis version belongs to a picked choice", ErrTaskRequestInvalid)
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
	fields := map[string]any{"id": requestID, "text": body, "author": taskRequestAuthor(userID)}
	if choiceID != "" {
		// Which of the session's own offered sentences the person picked, and
		// the task version the screen saw it on. The slot checks both, so a
		// screen left open while the task moved has its pick refused rather
		// than stored. Empty means they typed their own words.
		fields["choice_id"] = choiceID
		if choiceBasisVersion > 0 {
			fields["choice_basis_version"] = choiceBasisVersion
		}
	}
	payload, err := json.Marshal(fields)
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
		// A pick the session no longer offers: the screen is out of date, not
		// the words. Its own class, because the answer is to re-read the task
		// and keep what the person already typed. The platform's sentences for
		// it are stable (session-platform docs/session-ledger.md).
		if lower := strings.ToLower(string(raw)); strings.Contains(lower, "no longer offered") ||
			strings.Contains(lower, "no longer current") {
			return out, ErrTaskRequestStaleChoice
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

// TaskEvent is one line of a task's history as the app shows it: a named fact
// the session or the person caused, in the order it happened.
type TaskEvent struct {
	Seq       int64          `json:"seq"`
	TaskID    string         `json:"task_id"`
	At        string         `json:"at"`
	Kind      string         `json:"kind"`
	Actor     string         `json:"actor"`
	Summary   string         `json:"summary"`
	Detail    map[string]any `json:"detail,omitempty"`
	RequestID string         `json:"request_id,omitempty"`
}

// TaskEventPage is one keyset page of that history. NextCursor is the seq to
// continue after and is zero on the last page.
type TaskEventPage struct {
	Events     []TaskEvent `json:"events"`
	NextCursor int64       `json:"next_cursor,omitempty"`
}

// TaskEventPageLimit bounds one page the app asks for.
const TaskEventPageLimitMax = 100

// TaskEvents reads one page of taskID's history from the ledger of the session
// that answers userID's own conversation.
//
// Read-only and derived from the user's own row, like the task list: a caller
// cannot read another user's history. The history is the ledger's, so a task
// with nothing recorded answers an empty page — the app says "no record from
// before this" rather than inventing earlier lines.
func (e *Engine) TaskEvents(ctx context.Context, userID int64, taskID string, afterSeq int64, limit int) (TaskEventPage, error) {
	out := TaskEventPage{Events: []TaskEvent{}}
	if !ledgerRequestID.MatchString(taskID) {
		return out, fmt.Errorf("%w: task id must be 1..128 letters, digits, _ or -", ErrTaskRequestInvalid)
	}
	if afterSeq < 0 || limit < 0 || limit > TaskEventPageLimitMax {
		return out, fmt.Errorf("%w: after_seq and limit must be in range", ErrTaskRequestInvalid)
	}
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return out, err
	}
	if conv.platformSessionID == "" {
		return out, ErrNoConversation
	}
	query := url.Values{}
	if afterSeq > 0 {
		query.Set("after_seq", strconv.FormatInt(afterSeq, 10))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	path := "/sessions/" + url.PathEscape(conv.platformSessionID) +
		"/ledger/tasks/" + url.PathEscape(taskID) + "/events"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	status, raw, err := e.platform.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return out, fmt.Errorf("read task history: %w", err)
	}
	switch status {
	case http.StatusOK:
		if err := json.Unmarshal(raw, &out); err != nil {
			return TaskEventPage{Events: []TaskEvent{}}, fmt.Errorf("decode task history: %w", err)
		}
		if out.Events == nil {
			out.Events = []TaskEvent{}
		}
		return out, nil
	case http.StatusNotFound:
		return out, ErrTaskNotFound
	case http.StatusConflict:
		return out, ErrTaskRequestUnavailable
	case http.StatusNotImplemented, http.StatusMethodNotAllowed:
		// A platform or a session pod without the history route. Not an empty
		// history: the app says the session cannot show one.
		return out, ErrTaskRequestUnsupported
	case http.StatusBadRequest:
		return out, fmt.Errorf("%w: the platform refused the history query", ErrTaskRequestInvalid)
	default:
		return out, platformError("ledger events", status, raw)
	}
}
