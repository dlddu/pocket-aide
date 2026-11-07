package chat

import (
	"bytes"
	"context"
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
	turnSentinel        = 0x1e
	turnLead            = "session-platform "
	defaultPoll         = 700 * time.Millisecond
	defaultTimeout      = 40 * time.Minute
	emptyReply          = "(응답이 비어 있습니다)"
)

var (
	ErrBusy           = errors.New("a reply is still in progress")
	ErrNoConversation = errors.New("no codex session yet")
)

// PlatformError is a session-platform answer the caller should relay.
type PlatformError struct {
	Op     string
	Status int
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
	timeout  time.Duration
	now      func() time.Time

	mu      sync.Mutex
	turns   map[int64]*Pending
	running sync.WaitGroup
}

func NewEngine(store *Store, platform Platform, logger *slog.Logger) *Engine {
	if logger == nil {
		logger = slog.Default()
	}
	return &Engine{
		store: store, platform: platform, logger: logger,
		poll: defaultPoll, timeout: defaultTimeout, now: time.Now,
		turns: map[int64]*Pending{},
	}
}

// WithPoll shortens the transcript poll for tests.
func (e *Engine) WithPoll(d time.Duration) *Engine {
	e.poll = d
	return e
}

// Wait blocks until every background turn has finished.
func (e *Engine) Wait() { e.running.Wait() }

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
	status, err := e.write(ctx, sessionID, content)
	if err == nil && status == http.StatusNotFound {
		// The platform session is gone (deleted or reclaimed): start a new one.
		if sessionID, offset, err = e.openSession(ctx, userID, conv.ID); err != nil {
			return Message{}, err
		}
		status, err = e.write(ctx, sessionID, content)
	}
	if err != nil {
		return Message{}, err
	}
	if status != http.StatusOK && status != http.StatusAccepted {
		return Message{}, &PlatformError{Op: "write", Status: status}
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
		turnCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), e.timeout)
		defer cancel()
		e.finishTurn(turnCtx, userID, conv.ID, sessionID, offset)
	}()
	return msg, nil
}

func (e *Engine) finishTurn(ctx context.Context, userID, convID int64, sessionID string, offset int64) {
	f := &turnFollower{}
	next, err := e.follow(ctx, sessionID, offset, f, func(text string) { e.appendPending(userID, text) })
	reply := strings.TrimSpace(f.reply.String())
	switch {
	case err != nil && reply == "":
		reply = "응답을 끝까지 받지 못했습니다. 잠시 뒤 다시 시도해 주세요."
	case err != nil:
		reply += "\n\n(응답이 중간에 끊겼습니다)"
	case reply == "":
		reply = emptyReply
	}
	save := context.WithoutCancel(ctx)
	if _, saveErr := e.store.AppendMessage(save, userID, convID, RoleAssistant, reply); saveErr != nil {
		e.logger.Error("chat: save codex reply", "user", userID, "err", saveErr)
	}
	if err != nil {
		e.logger.Warn("chat: codex turn did not finish", "user", userID, "session", sessionID, "err", err)
		return
	}
	if saveErr := e.store.SetPlatform(save, userID, convID, sessionID, next); saveErr != nil {
		e.logger.Error("chat: save transcript offset", "user", userID, "err", saveErr)
	}
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
			e.running.Add(1)
			go func() {
				defer e.running.Done()
				defer e.release(userID)
				turnCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), e.timeout)
				defer cancel()
				e.finishTurn(turnCtx, userID, convID, sessionID, turnStart)
			}()
			return offset, errBusyHandedOff
		}
		reply := strings.TrimSpace(f.reply.String())
		if reply == "" {
			reply = emptyReply
		}
		if _, err := e.store.AppendMessage(ctx, userID, convID, RoleAssistant, reply); err != nil {
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

func (e *Engine) write(ctx context.Context, sessionID, content string) (int, error) {
	body, _ := json.Marshal(map[string]string{"payload": content})
	status, _, err := e.platform.Do(ctx, http.MethodPost, "/sessions/"+url.PathEscape(sessionID)+"/write", body)
	return status, err
}

func sessionName(userID int64) string {
	return "pocket-aide-chat-" + strconv.FormatInt(userID, 10)
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
	body, _ := json.Marshal(map[string]string{"name": sessionName(userID), "workloadType": workloadCodex})
	status, raw, err := e.platform.Do(ctx, http.MethodPost, "/sessions", body)
	if err == nil && status != http.StatusCreated && status != http.StatusOK {
		err = &PlatformError{Op: "create", Status: status}
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
	reply     strings.Builder
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
		kind, seq := parseMarker(string(f.buf[i+1 : i+nl]))
		f.consume(i + nl + 1)
		switch {
		case f.stage == 0 && kind == "input":
			f.stage, f.seq, f.startedAt = 1, seq, markerAt
		case f.stage == 1 && kind == "output" && seq == f.seq:
			f.stage = 2
		case f.stage == 2 && kind == "end" && seq == f.seq:
			f.stage, f.end = 3, f.base
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

func parseMarker(header string) (kind, seq string) {
	if !strings.HasPrefix(header, turnLead) {
		return "", ""
	}
	fields := strings.Fields(strings.TrimPrefix(header, turnLead))
	if len(fields) == 0 {
		return "", ""
	}
	kind = fields[0]
	for _, field := range fields[1:] {
		if v, ok := strings.CutPrefix(field, "seq="); ok {
			seq = v
		}
	}
	return kind, seq
}
