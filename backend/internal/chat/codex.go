package chat

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

// PRD-1 AC1·AC4 (2026-11-06 결정): the chat tab is one long-lived codex
// session per user on session-platform, called as the chat tab's tenant. The
// platform session holds the conversation; this package keeps the readable
// history in chat_messages and follows each turn in the platform transcript,
// whose turns are framed by `\x1esession-platform input|output|end seq=N` lines.

const (
	// workloadCodex is session-platform's chat Codex type. It was "codex"
	// until the platform split Codex into codex-chat and codex-task
	// (session-platform AC-J1, 2026-11-07); a session made before the rename
	// can still be listed under the old name, so adoption accepts both.
	workloadCodex       = "codex-chat"
	legacyWorkloadCodex = "codex"
	// chatPlugin is the marketplace bundle whose skills the chat session runs
	// with. The platform installs only what the create request names — it has
	// no default for codex-chat since session-platform AC-A5 (2026-11-14).
	chatPlugin   = "codex-session"
	turnSentinel = 0x1e
	turnLead     = "session-platform "
	// turnStatusOK is the one status= value on an end marker that means the
	// turn ran to its end (session-platform claudeTurnStatus: ok · timeout ·
	// canceled · error). Every other value, and a turn whose end marker never
	// came, is a reply the user must not read as the session's answer.
	turnStatusOK = "ok"
	defaultPoll  = 700 * time.Millisecond
	emptyReply   = "(응답이 비어 있습니다)"
	// brokenReply stands in when an interrupted turn produced nothing at all.
	// An empty failure has to read as a failure: the text for an empty but
	// finished turn would claim the session answered with nothing.
	brokenReply = "응답을 끝까지 받지 못했습니다. 잠시 뒤 다시 시도해 주세요."
	// interruptedNote closes a partial answer that stopped short. The apps that
	// predate Message.Interrupted draw content and nothing else, so a turn's
	// own mark is the only place the user can see it was cut off.
	interruptedNote = "중단됨"
	// interruptedTail is the note as a paragraph of its own. It is the only
	// form this server ever writes, which is what makes it safe to read back:
	// an answer whose own prose ends in the word ("다운로드가 중단됨") is not a
	// marked answer, and still gets a mark of its own.
	interruptedTail = "\n\n" + interruptedNote
	// steerApplied is session-platform's answer when the running turn took a
	// mid-turn message (AC-J9); anything else means it did not.
	steerApplied = "applied"
	// retryHold spaces out dispatches of a queued message whose last one
	// failed. The apps poll GET /api/chat about once a second while they
	// wait, and each poll could otherwise re-send to a refusing platform.
	retryHold = 30 * time.Second
)

var (
	ErrBusy            = errors.New("a reply is still in progress")
	ErrNoConversation  = errors.New("no codex session yet")
	ErrQueueFull       = errors.New("a message is already queued")
	ErrNoQueuedMessage = errors.New("no message queued")
)

// PlatformError is a session-platform answer the caller should relay.
// AuthPoolExhausted marks the 503 for exhausted login slots, which the app
// shows differently from an unreachable platform.
type PlatformError struct {
	Op                string
	Status            int
	AuthPoolExhausted bool
}

func platformError(op string, status int, body []byte) *PlatformError {
	return &PlatformError{Op: op, Status: status, AuthPoolExhausted: sessions.AuthPoolExhausted(status, body)}
}

func (e *PlatformError) Error() string {
	return fmt.Sprintf("session platform %s: status %d", e.Op, e.Status)
}

// Platform is the slice of *sessions.Client the engine uses.
type Platform interface {
	Do(ctx context.Context, method, path string, body []byte) (int, []byte, error)
}

// Pending is a reply still being written.
type Pending struct {
	Content   string `json:"content"`
	StartedAt int64  `json:"started_at"`
}

// PluginRefresh is session-platform's answer to a plugin refresh.
type PluginRefresh struct {
	State    string   `json:"state"`
	Revision string   `json:"revision,omitempty"`
	Added    []string `json:"added"`
	Updated  []string `json:"updated"`
	Removed  []string `json:"removed"`
}

type Engine struct {
	store    *Store
	platform Platform
	logger   *slog.Logger
	poll     time.Duration
	now      func() time.Time

	mu    sync.Mutex
	turns map[int64]*Pending
	// materials holds what has been folded of each task's turns for the review
	// answer (task_materials.go), keyed by task id. In memory only and a saving
	// rather than a source: a restart costs the next read one walk of the turns
	// it would otherwise have resumed.
	materials map[string]materialsFold
	// retryAfter holds off the next dispatch of a message whose last one
	// failed, so a platform that keeps refusing is not hammered once per
	// poll. In memory only: a restart just means one earlier retry.
	retryAfter map[int64]time.Time
	running    sync.WaitGroup
}

func NewEngine(store *Store, platform Platform, logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{
		store: store, platform: platform, logger: logger,
		poll: defaultPoll, now: time.Now,
		turns: map[int64]*Pending{}, retryAfter: map[int64]time.Time{},
	}
}

// WithPoll shortens the transcript poll for tests.
func (e *Engine) WithPoll(d time.Duration) *Engine {
	e.poll = d
	return e
}

// Wait blocks until every background turn has finished.
func (e *Engine) Wait() { e.running.Wait() }

// turnContext is what a background follow runs under. It is detached from the
// request that started the turn, because the reply outlives that HTTP call,
// and it carries no deadline of its own: a turn takes as long as the session
// takes to answer it, and this server must not declare an answer cut short
// while the platform is still writing it (PRD-1 AC1). Receiving a reply had a
// 40 minute limit until 2026-11-27, which is exactly that mistake — a session
// still working past it had the text it had produced stored as an interrupted
// answer, and the rest of the answer thrown away.
//
// What does end a follow is unchanged: the platform's own word on the turn (an
// end marker whose status is not ok — canceled, error, the platform's own
// timeout), or a transcript read that fails, each read bounded by the platform
// client's per-request timeout.
func turnContext(ctx context.Context) context.Context {
	return context.WithoutCancel(ctx)
}

// Pending returns the reply in progress for userID, if any.
func (e *Engine) Pending(userID int64) *Pending {
	e.mu.Lock()
	defer e.mu.Unlock()
	if p, ok := e.turns[userID]; ok {
		copied := *p
		return &copied
	}
	return nil
}

func (e *Engine) claim(userID int64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, busy := e.turns[userID]; busy {
		return false
	}
	e.turns[userID] = &Pending{StartedAt: e.now().Unix()}
	return true
}

func (e *Engine) release(userID int64) {
	e.mu.Lock()
	delete(e.turns, userID)
	e.mu.Unlock()
}

func (e *Engine) appendPending(userID int64, text string) {
	e.mu.Lock()
	if p, ok := e.turns[userID]; ok {
		p.Content += text
	}
	e.mu.Unlock()
}

// resetPending clears a still-claimed turn's accumulated preview text before
// a dequeued message starts a new one in its place.
func (e *Engine) resetPending(userID int64) {
	e.mu.Lock()
	if p, ok := e.turns[userID]; ok {
		p.Content = ""
		p.StartedAt = e.now().Unix()
	}
	e.mu.Unlock()
}

// Send stores the user's message, hands it to the codex session and follows
// the reply in the background; the reply lands in chat_messages when the
// turn ends. One turn per user at a time.
func (e *Engine) Send(ctx context.Context, userID int64, content string) (Message, error) {
	if !e.claim(userID) {
		return Message{}, ErrBusy
	}
	started := false
	defer func() {
		if !started {
			e.release(userID)
		}
	}()
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return Message{}, err
	}
	sessionID, offset := conv.platformSessionID, conv.transcriptOffset
	if sessionID != "" {
		// A turn this server lost track of (a restart mid-reply) is finished
		// first, so the new prompt's reply is never attributed to it.
		offset, err = e.catchUp(ctx, userID, conv.ID, sessionID, offset)
		if errors.Is(err, errBusyHandedOff) {
			started = true // the background follower releases the claim
			return Message{}, ErrBusy
		}
		if err != nil {
			return Message{}, err
		}
	}
	if sessionID == "" {
		if sessionID, offset, err = e.openSession(ctx, userID, conv.ID); err != nil {
			return Message{}, err
		}
	}
	status, body, err := e.write(ctx, sessionID, content)
	if err == nil && status == http.StatusNotFound {
		// The platform session is gone (deleted or reclaimed): start a new one.
		if sessionID, offset, err = e.openSession(ctx, userID, conv.ID); err != nil {
			return Message{}, err
		}
		status, body, err = e.write(ctx, sessionID, content)
	}
	if err != nil {
		return Message{}, err
	}
	if status != http.StatusOK && status != http.StatusAccepted {
		return Message{}, platformError("write", status, body)
	}
	// Stored only once the platform took it, so a failed send leaves no
	// message without a reply; and stored even if the client went away.
	msg, err := e.store.AppendMessage(context.WithoutCancel(ctx), userID, conv.ID, RoleUser, content)
	if err != nil {
		e.logger.Error("chat: save sent message", "user", userID, "err", err)
	}
	started = true
	e.running.Add(1)
	go func() {
		defer e.running.Done()
		defer e.release(userID)
		e.finishTurn(turnContext(ctx), userID, conv.ID, sessionID, offset)
	}()
	return msg, nil
}

func (e *Engine) finishTurn(ctx context.Context, userID, convID int64, sessionID string, offset int64) {
	f := &turnFollower{}
	next, err := e.follow(ctx, sessionID, offset, f, func(text string) { e.appendPending(userID, text) })
	reply, interrupted := turnReply(f, err)
	save := context.WithoutCancel(ctx)
	if _, saveErr := e.store.AppendReply(save, userID, convID, reply, interrupted); saveErr != nil {
		e.logger.Error("chat: save codex reply", "user", userID, "err", saveErr)
	}
	if err != nil {
		// The bytes this reply was made of are consumed even though the turn
		// has no end marker: a later catch-up reading them again would store
		// the same turn a second time. f.base is where the follow stopped,
		// inside that turn, and a catch-up from there looks for the next input
		// marker — so the abandoned turn is neither re-attributed nor able to
		// shift the offsets of the turns after it.
		if saveErr := e.store.SetPlatform(save, userID, convID, sessionID, f.base); saveErr != nil {
			e.logger.Error("chat: save transcript offset", "user", userID, "err", saveErr)
		}
		e.logger.Warn("chat: codex turn did not finish", "user", userID, "session", sessionID, "err", err)
		return
	}
	if saveErr := e.store.SetPlatform(save, userID, convID, sessionID, next); saveErr != nil {
		e.logger.Error("chat: save transcript offset", "user", userID, "err", saveErr)
	}
	e.resetPending(userID)
	e.drainQueued(save, userID, convID)
}

// steer offers content to the turn in progress so it can change where that
// turn is going (session-platform AC-J9). applied is true only when the
// platform says the running turn took it; anything else — a turn that ended
// first, a platform that does not take mid-turn messages — leaves the
// message for the caller to queue instead, so it is never lost.
func (e *Engine) steer(ctx context.Context, sessionID, content string) (bool, error) {
	body, _ := json.Marshal(map[string]string{"text": content})
	status, raw, err := e.platform.Do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(sessionID)+"/steer", body)
	if err != nil {
		return false, err
	}
	switch status {
	case http.StatusOK:
		var out struct {
			Status string `json:"status"`
		}
		if json.Unmarshal(raw, &out) == nil && out.Status == steerApplied {
			return true, nil
		}
		return false, nil
	case http.StatusConflict, http.StatusUnprocessableEntity, http.StatusNotFound, http.StatusMethodNotAllowed:
		// 409 the turn would not take it; 422/404/405 the platform has no
		// mid-turn path for this session at all. Both mean: queue it.
		return false, nil
	default:
		return false, platformError("steer", status, raw)
	}
}

// Queue is what a send does while a reply is in progress. It tries three
// things in order, and the first that works wins (PRD-1 AC1):
//
//  1. send it as its own turn, when no reply is in progress;
//  2. put it into the reply in progress, so that reply changes course
//     (session-platform AC-J9) — the message is then part of that turn and
//     shows up as a sent message;
//  3. hold it for right after, as the queued message the user can cancel.
//
// At most one message waits at a time; a second arrival while one is already
// queued is ErrQueueFull.
func (e *Engine) Queue(ctx context.Context, userID int64, content string) (Message, *QueuedMessage, error) {
	msg, err := e.Send(ctx, userID, content)
	if err == nil {
		return msg, nil, nil
	}
	if !errors.Is(err, ErrBusy) {
		return Message{}, nil, err
	}
	conv, cErr := e.store.Conversation(ctx, userID)
	if cErr != nil {
		return Message{}, nil, cErr
	}
	// The reply in progress may take the message and change course with it.
	if conv.platformSessionID != "" {
		applied, sErr := e.steer(ctx, conv.platformSessionID, content)
		if sErr != nil {
			e.logger.Warn("chat: mid-turn delivery failed; queueing instead", "user", userID, "err", sErr)
		}
		if applied {
			// It belongs to the running turn now, so it reads as a message
			// that went out — not as one still waiting.
			sent, aErr := e.store.AppendMessage(context.WithoutCancel(ctx), userID, conv.ID, RoleUser, content)
			if aErr != nil {
				return Message{}, nil, aErr
			}
			return sent, nil, nil
		}
	}
	existing, gErr := e.store.GetQueued(ctx, userID, conv.ID)
	if gErr != nil {
		return Message{}, nil, gErr
	}
	if existing != nil {
		return Message{}, nil, ErrQueueFull
	}
	id, idErr := randomID()
	if idErr != nil {
		return Message{}, nil, idErr
	}
	q, sErr := e.store.SetQueued(ctx, userID, conv.ID, id, content, e.now().Unix())
	if sErr != nil {
		return Message{}, nil, sErr
	}
	// The turn in progress may have ended between the busy check above and
	// the write just made, in which case nothing is left running to drain
	// this message — claim it ourselves and start its turn if so. Whichever
	// of this and the turn's own drainQueued call sees the row first wins;
	// claim() makes that exclusive.
	if e.claim(userID) {
		e.running.Add(1)
		go func() {
			defer e.running.Done()
			defer e.release(userID)
			e.drainQueued(context.WithoutCancel(ctx), userID, conv.ID)
		}()
	}
	return Message{}, &q, nil
}

// CancelQueued drops the queued message named by id, if it is still waiting
// — once drainQueued has taken it there is nothing left to cancel, and the
// caller sees ErrNoQueuedMessage rather than losing whatever took its place.
func (e *Engine) CancelQueued(ctx context.Context, userID int64, id string) error {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return err
	}
	cleared, err := e.store.ClearQueued(ctx, userID, conv.ID, id)
	if err != nil {
		return err
	}
	if !cleared {
		return ErrNoQueuedMessage
	}
	e.forgetRetry(userID)
	return nil
}

// ResumeQueued dispatches a message left waiting with no turn running behind
// it — what a restart leaves, since the turn that would have drained it died
// with the process. The apps' GET /api/chat poll is what calls this, so the
// wait resumes without the user having to send anything.
func (e *Engine) ResumeQueued(ctx context.Context, userID, convID int64) {
	if !e.retryDue(userID) {
		return
	}
	queued, err := e.store.GetQueued(ctx, userID, convID)
	if err != nil || queued == nil {
		return
	}
	if !e.claim(userID) {
		return // a turn is running; its own drainQueued will take this one
	}
	e.running.Add(1)
	go func() {
		defer e.running.Done()
		defer e.release(userID)
		e.drainQueued(turnContext(ctx), userID, convID)
	}()
}

func (e *Engine) retryDue(userID int64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return !e.now().Before(e.retryAfter[userID])
}

func (e *Engine) holdRetry(userID int64) {
	e.mu.Lock()
	e.retryAfter[userID] = e.now().Add(retryHold)
	e.mu.Unlock()
}

func (e *Engine) forgetRetry(userID int64) {
	e.mu.Lock()
	delete(e.retryAfter, userID)
	e.mu.Unlock()
}

// keepQueued puts a taken message back under its own id after a dispatch
// that failed, so a message the user was told is waiting is never dropped on
// the floor: the next turn end or poll tries it again.
func (e *Engine) keepQueued(ctx context.Context, userID, convID int64, queued *QueuedMessage) {
	e.holdRetry(userID)
	if _, err := e.store.SetQueued(context.WithoutCancel(ctx), userID, convID, queued.ID, queued.Content, queued.QueuedAt); err != nil {
		e.logger.Error("chat: keep queued message after a failed dispatch", "user", userID, "err", err)
	}
}

// drainQueued sends the message waiting behind the turn that just ended, if
// any, and follows its reply the same way a fresh Send would. The claim for
// userID must already be held by the caller.
func (e *Engine) drainQueued(ctx context.Context, userID, convID int64) {
	queued, err := e.store.TakeQueued(ctx, userID, convID)
	if err != nil {
		e.logger.Error("chat: take queued message", "user", userID, "err", err)
		return
	}
	if queued == nil {
		return
	}
	conv, err := e.store.GetSession(ctx, userID, convID)
	if err != nil {
		e.logger.Error("chat: dequeue lookup conversation", "user", userID, "err", err)
		e.keepQueued(ctx, userID, convID, queued)
		return
	}
	sessionID, offset := conv.platformSessionID, conv.transcriptOffset
	if sessionID == "" {
		if sessionID, offset, err = e.openSession(ctx, userID, convID); err != nil {
			e.logger.Error("chat: dequeue open session", "user", userID, "err", err)
			e.keepQueued(ctx, userID, convID, queued)
			return
		}
	}
	status, body, err := e.write(ctx, sessionID, queued.Content)
	if err == nil && status == http.StatusNotFound {
		if sessionID, offset, err = e.openSession(ctx, userID, convID); err != nil {
			e.logger.Error("chat: dequeue reopen session", "user", userID, "err", err)
			e.keepQueued(ctx, userID, convID, queued)
			return
		}
		status, body, err = e.write(ctx, sessionID, queued.Content)
	}
	if err != nil {
		e.logger.Error("chat: dequeued message write failed", "user", userID, "err", err)
		e.keepQueued(ctx, userID, convID, queued)
		return
	}
	if status != http.StatusOK && status != http.StatusAccepted {
		e.logger.Error("chat: dequeued message refused", "user", userID, "status", status, "body", string(body))
		e.keepQueued(ctx, userID, convID, queued)
		return
	}
	e.forgetRetry(userID)
	if _, saveErr := e.store.AppendMessage(ctx, userID, convID, RoleUser, queued.Content); saveErr != nil {
		e.logger.Error("chat: save dequeued message", "user", userID, "err", saveErr)
	}
	e.finishTurn(ctx, userID, convID, sessionID, offset)
}

// catchUp reads what the transcript holds past offset. Finished turns are
// saved as replies; an unfinished one is followed in the background and the
// caller is told the user is busy.
func (e *Engine) catchUp(ctx context.Context, userID, convID int64, sessionID string, offset int64) (int64, error) {
	payload, _, status, err := e.read(ctx, sessionID, offset)
	if err != nil {
		return offset, err
	}
	if status == http.StatusNotFound {
		return offset, nil
	}
	if status != http.StatusOK {
		return offset, &PlatformError{Op: "read", Status: status}
	}
	pos := offset
	data := []byte(payload)
	for {
		f := &turnFollower{base: pos}
		done := f.feed(data, nil)
		if !f.started() {
			break
		}
		if !done {
			// Still running: follow it from where it started.
			turnStart := f.startedAt
			// The turns stored above are consumed before the hand-off, not
			// after it. The background follow can end in an error, which
			// leaves the offset where that follow found it — and the next
			// catch-up would then read these finished turns again and store
			// every one of them a second time.
			if turnStart != offset {
				if err := e.store.SetPlatform(ctx, userID, convID, sessionID, turnStart); err != nil {
					return offset, err
				}
			}
			e.running.Add(1)
			go func() {
				defer e.running.Done()
				defer e.release(userID)
				e.finishTurn(turnContext(ctx), userID, convID, sessionID, turnStart)
			}()
			return turnStart, errBusyHandedOff
		}
		// Read out of the transcript rather than followed live, but the same
		// turn either way: history and the live path get one rule.
		reply, interrupted := turnReply(f, nil)
		if _, err := e.store.AppendReply(ctx, userID, convID, reply, interrupted); err != nil {
			return offset, err
		}
		data = data[f.end-pos:]
		pos = f.end
	}
	if pos != offset {
		if err := e.store.SetPlatform(ctx, userID, convID, sessionID, pos); err != nil {
			return offset, err
		}
	}
	return pos, nil
}

// errBusyHandedOff is ErrBusy after the claim passed to a background follower.
var errBusyHandedOff = fmt.Errorf("%w (resuming an earlier turn)", ErrBusy)

func (e *Engine) follow(ctx context.Context, sessionID string, offset int64, f *turnFollower, emit func(string)) (int64, error) {
	f.base = offset
	cursor := offset
	for {
		payload, next, status, err := e.read(ctx, sessionID, cursor)
		if err != nil {
			return 0, err
		}
		if status != http.StatusOK {
			return 0, &PlatformError{Op: "read", Status: status}
		}
		if f.feed([]byte(payload), emit) {
			return f.end, nil
		}
		cursor = next
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(e.poll):
		}
	}
}

func (e *Engine) read(ctx context.Context, sessionID string, offset int64) (string, int64, int, error) {
	body, _ := json.Marshal(map[string]int64{"offset": offset})
	status, raw, err := e.platform.Do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(sessionID)+"/read", body)
	if err != nil || status != http.StatusOK {
		return "", 0, status, err
	}
	var out struct {
		Payload    string `json:"payload"`
		NextOffset int64  `json:"nextOffset"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", 0, 0, fmt.Errorf("decode read: %w", err)
	}
	return out.Payload, out.NextOffset, status, nil
}

func (e *Engine) write(ctx context.Context, sessionID, content string) (int, []byte, error) {
	body, _ := json.Marshal(map[string]string{"payload": content})
	return e.platform.Do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(sessionID)+"/write", body)
}

func sessionName(userID int64) string {
	return "pocket-aide-chat-" + strconv.FormatInt(userID, 10)
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// openSession gives the conversation a codex session. Creating one can
// outlast the gateway in front of the platform (a cold control plane, a pod
// pulling its image) while the session is still made, so a session already
// carrying this user's name is adopted — before creating, and again when a
// create fails — instead of leaving it orphaned. An adopted session is read
// from its current end.
func (e *Engine) openSession(ctx context.Context, userID, convID int64) (string, int64, error) {
	if id, end, ok := e.adoptSession(ctx, userID); ok {
		return id, end, e.store.SetPlatform(ctx, userID, convID, id, end)
	}
	body, _ := json.Marshal(map[string]any{"name": sessionName(userID), "workloadType": workloadCodex, "plugins": []string{chatPlugin}})
	status, raw, err := e.platform.Do(ctx, http.MethodPost, "/sessions", body)
	if err == nil && status != http.StatusCreated && status != http.StatusOK {
		err = platformError("create", status, raw)
	}
	if err != nil {
		if id, end, ok := e.adoptSession(context.WithoutCancel(ctx), userID); ok {
			e.logger.Warn("chat: adopted a codex session whose create failed", "user", userID, "session", id, "err", err)
			return id, end, e.store.SetPlatform(ctx, userID, convID, id, end)
		}
		return "", 0, err
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &created); err != nil || created.ID == "" {
		return "", 0, fmt.Errorf("decode created session: %w", err)
	}
	return created.ID, 0, e.store.SetPlatform(ctx, userID, convID, created.ID, 0)
}

// adoptSession finds the newest codex session named for userID among those
// the platform shows this caller, and where its transcript ends.
func (e *Engine) adoptSession(ctx context.Context, userID int64) (string, int64, bool) {
	status, raw, err := e.platform.Do(ctx, http.MethodGet, "/sessions", nil)
	if err != nil || status != http.StatusOK {
		return "", 0, false
	}
	var list struct {
		Sessions []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			WorkloadType string `json:"workloadType"`
			CreatedAt    string `json:"createdAt"`
		} `json:"sessions"`
	}
	if json.Unmarshal(raw, &list) != nil {
		return "", 0, false
	}
	id, newest := "", time.Time{}
	for _, s := range list.Sessions {
		created, _ := time.Parse(time.RFC3339Nano, s.CreatedAt)
		if s.Name == sessionName(userID) && (s.WorkloadType == workloadCodex || s.WorkloadType == legacyWorkloadCodex) && (id == "" || created.After(newest)) {
			id, newest = s.ID, created
		}
	}
	if id == "" {
		return "", 0, false
	}
	_, end, status, err := e.read(ctx, id, 0)
	if err != nil || status != http.StatusOK {
		return "", 0, false
	}
	return id, end, true
}

// RefreshPlugins asks the codex session to fetch its plugin skills again,
// keeping the conversation (session-platform AC-J7). Between turns only.
func (e *Engine) RefreshPlugins(ctx context.Context, userID int64) (PluginRefresh, error) {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return PluginRefresh{}, err
	}
	if conv.platformSessionID == "" {
		return PluginRefresh{}, ErrNoConversation
	}
	if !e.claim(userID) {
		return PluginRefresh{}, ErrBusy
	}
	defer e.release(userID)
	status, raw, err := e.platform.Do(ctx, http.MethodPost,
		"/sessions/"+url.PathEscape(conv.platformSessionID)+"/plugins/refresh", nil)
	if err != nil {
		return PluginRefresh{}, err
	}
	switch status {
	case http.StatusOK:
	case http.StatusConflict:
		return PluginRefresh{}, ErrBusy
	case http.StatusNotFound:
		return PluginRefresh{}, ErrNoConversation
	default:
		return PluginRefresh{}, &PlatformError{Op: "refresh plugins", Status: status}
	}
	var out PluginRefresh
	if err := json.Unmarshal(raw, &out); err != nil {
		return PluginRefresh{}, fmt.Errorf("decode plugin refresh: %w", err)
	}
	return out, nil
}

// turnFollower picks the first turn out of transcript bytes fed to it in
// order. Marker lines start with the sentinel byte, which the platform strips
// from everything else it writes.
type turnFollower struct {
	stage     int // 0 before input, 1 before output, 2 in output, 3 done
	seq       string
	buf       []byte
	base      int64 // transcript offset of buf[0]
	startedAt int64 // offset of the turn's input marker
	end       int64 // offset just past the end marker
	// status is the status= field of the end marker that closed this turn,
	// taken only from the marker carrying the seq the input marker opened —
	// another turn's end says nothing about this one. Empty when the turn has
	// not ended, and also when it ended on a platform old enough to write no
	// status at all.
	status string
	reply  strings.Builder
}

func (f *turnFollower) started() bool { return f.stage > 0 }

// feed consumes p and reports whether the turn ended. emit, when set, gets
// each piece of the turn's output as it arrives.
func (f *turnFollower) feed(p []byte, emit func(string)) bool {
	f.buf = append(f.buf, p...)
	for f.stage < 3 {
		i := bytes.IndexByte(f.buf, turnSentinel)
		if i < 0 {
			f.output(f.buf, emit)
			f.consume(len(f.buf))
			return false
		}
		f.output(f.buf[:i], emit)
		nl := bytes.IndexByte(f.buf[i:], '\n')
		if nl < 0 {
			f.consume(i)
			return false
		}
		markerAt := f.base + int64(i)
		kind, seq, status := parseMarker(string(f.buf[i+1 : i+nl]))
		f.consume(i + nl + 1)
		switch {
		case f.stage == 0 && kind == "input":
			f.stage, f.seq, f.startedAt = 1, seq, markerAt
		case f.stage == 1 && kind == "output" && seq == f.seq:
			f.stage = 2
		case f.stage == 2 && kind == "end" && seq == f.seq:
			f.stage, f.end, f.status = 3, f.base, status
		}
	}
	return true
}

func (f *turnFollower) output(p []byte, emit func(string)) {
	if f.stage != 2 || len(p) == 0 {
		return
	}
	f.reply.Write(p)
	if emit != nil {
		emit(string(p))
	}
}

func (f *turnFollower) consume(n int) {
	f.buf = f.buf[n:]
	f.base += int64(n)
}

func parseMarker(header string) (kind, seq, status string) {
	if !strings.HasPrefix(header, turnLead) {
		return "", "", ""
	}
	fields := strings.Fields(strings.TrimPrefix(header, turnLead))
	if len(fields) == 0 {
		return "", "", ""
	}
	kind = fields[0]
	for _, field := range fields[1:] {
		if v, ok := strings.CutPrefix(field, "seq="); ok {
			seq = v
		}
		if v, ok := strings.CutPrefix(field, "status="); ok {
			status = v
		}
	}
	return kind, seq, status
}

// turnReply is what one turn becomes in the history. followErr is set when
// this server stopped following before the turn's end marker arrived — a read
// that failed. A turn merely running long is not one of those: the follow has
// no deadline of its own (turnContext).
//
// An interrupted turn keeps every byte it received and says so twice: in
// Message.Interrupted for readers that model the state, and in a trailing
// interruptedNote paragraph for the apps that draw content and nothing else.
func turnReply(f *turnFollower, followErr error) (string, bool) {
	partial := strings.TrimSpace(f.reply.String())
	// A turn with no status= ended on a platform that wrote none, which only
	// ever framed turns that ran to the end — so does old stored history.
	finished := followErr == nil && (f.status == "" || f.status == turnStatusOK)
	switch {
	case finished && partial == "":
		return emptyReply, false
	case finished:
		return partial, false
	case partial == "":
		return brokenReply, true
	default:
		return noteInterrupted(partial), true
	}
}

// noteInterrupted marks a partial answer as cut short without rewriting what
// arrived before it. Only a note standing as the last paragraph counts as one
// already there, so a turn judged twice keeps the single mark it has while an
// answer whose prose merely ends in the word still gets its own.
func noteInterrupted(partial string) string {
	if partial == interruptedNote || strings.HasSuffix(partial, interruptedTail) {
		return partial
	}
	return partial + interruptedTail
}
