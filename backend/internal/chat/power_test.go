package chat_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
)

// powerPlatform is session-platform as the on/off routes see it: sessions with
// a state that `GET /sessions/{id}` reports and `POST /sessions/{id}/switch`
// moves back to active. It counts the transcript calls too, because the whole
// point of reading the record is that neither route touches the transcript —
// a read would restore the archived session it was asked about.
type powerPlatform struct {
	mu       sync.Mutex
	sessions map[string]*powerSession
	created  int
	// getStatus, when non-zero, is the status `GET /sessions/{id}` answers
	// with instead of the record.
	getStatus int
	// switchStatus and switchBody force the restore's answer; 0 = normal.
	switchStatus int
	switchBody   string
	// switches is every session id a restore was asked for, in order.
	switches []string
	reads    int
	writes   int
}

type powerSession struct {
	name       string
	state      string
	lastAccess string
	checkpoint string
	transcript []byte
	turns      int
}

func newPowerPlatform() *powerPlatform {
	return &powerPlatform{sessions: map[string]*powerSession{}}
}

func (p *powerPlatform) Do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	parts := strings.Split(strings.TrimPrefix(path, "/sessions"), "/")
	switch {
	case method == http.MethodPost && path == "/sessions":
		p.created++
		id := fmt.Sprintf("cx%d", p.created)
		var req struct {
			Name string `json:"name"`
		}
		_ = json.Unmarshal(body, &req)
		p.sessions[id] = &powerSession{name: req.Name, state: "active", lastAccess: "2026-11-28T00:00:00Z"}
		raw, _ := json.Marshal(map[string]string{"id": id})
		return http.StatusCreated, raw, nil
	case method == http.MethodGet && path == "/sessions":
		// Nothing to adopt: these tests open sessions through the same send
		// path the app uses, and an adoptable session would hide which one.
		return http.StatusOK, []byte(`{"sessions":[]}`), nil
	case method == http.MethodGet && len(parts) == 2:
		if p.getStatus != 0 {
			return p.getStatus, []byte(`{"error":"no"}`), nil
		}
		s, ok := p.sessions[parts[1]]
		if !ok {
			return http.StatusNotFound, nil, nil
		}
		return http.StatusOK, []byte(p.record(parts[1], s)), nil
	case len(parts) == 3 && parts[2] == "switch":
		p.switches = append(p.switches, parts[1])
		if p.switchStatus != 0 {
			return p.switchStatus, []byte(p.switchBody), nil
		}
		s, ok := p.sessions[parts[1]]
		if !ok {
			return http.StatusNotFound, nil, nil
		}
		s.state, s.checkpoint = "active", ""
		return http.StatusOK, []byte(p.record(parts[1], s)), nil
	case len(parts) == 3 && parts[2] == "write":
		s, ok := p.sessions[parts[1]]
		if !ok {
			return http.StatusNotFound, nil, nil
		}
		p.writes++
		var req map[string]string
		_ = json.Unmarshal(body, &req)
		s.turns++
		// The whole turn lands at once, so the follow finds it finished on its
		// first read — these tests are about the session, not about following.
		s.transcript = append(s.transcript, fmt.Sprintf(
			"\x1esession-platform input seq=%d at=x\n%s\n\x1esession-platform output seq=%d\nre: %s\n\x1esession-platform end seq=%d status=ok at=x\n",
			s.turns, req["payload"], s.turns, req["payload"], s.turns)...)
		return http.StatusOK, []byte(`{}`), nil
	case len(parts) == 3 && parts[2] == "read":
		s, ok := p.sessions[parts[1]]
		if !ok {
			return http.StatusNotFound, nil, nil
		}
		p.reads++
		var req struct {
			Offset int64 `json:"offset"`
		}
		_ = json.Unmarshal(body, &req)
		if req.Offset > int64(len(s.transcript)) {
			req.Offset = int64(len(s.transcript))
		}
		raw, _ := json.Marshal(map[string]any{"payload": string(s.transcript[req.Offset:]), "nextOffset": len(s.transcript)})
		return http.StatusOK, raw, nil
	}
	return http.StatusTeapot, nil, nil
}

func (p *powerPlatform) record(id string, s *powerSession) string {
	checkpoint := ""
	if s.checkpoint != "" {
		checkpoint = fmt.Sprintf(`,"checkpoint":{"createdAt":%q,"sizeBytes":2048}`, s.checkpoint)
	}
	return fmt.Sprintf(`{"id":%q,"name":%q,"workloadType":"codex-chat","state":%q,"lastAccess":%q%s}`,
		id, s.name, s.state, s.lastAccess, checkpoint)
}

// archive puts a session into the state an explicit archive or the platform's
// 60 minute idle freeze leaves it in — the state the chat tab calls 꺼짐.
func (p *powerPlatform) archive(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessions[id].state = "snapshot"
	p.sessions[id].checkpoint = "2026-11-28T01:00:00Z"
}

func (p *powerPlatform) setState(id, state string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessions[id].state = state
}

func (p *powerPlatform) counts() (reads, writes, created int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.reads, p.writes, p.created
}

// fail makes `GET /sessions/{id}` answer with this status instead of the
// record, and failSwitch does the same for the restore. Both go through the
// mutex: a background follow may still be reading this fake.
func (p *powerPlatform) fail(status int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.getStatus = status
}

func (p *powerPlatform) failSwitch(status int, body string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.switchStatus, p.switchBody = status, body
}

// forget drops a session the way the platform does when it is deleted or its
// record is lost — every later call about it answers 404.
func (p *powerPlatform) forget(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.sessions, id)
}

func (p *powerPlatform) switchedSessions() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]string{}, p.switches...)
}

// sessionOf is the session this platform holds for one user's chat, found by
// the name the engine creates it under. The conversation row's own field is
// not exported, so the name is how a test outside the package says "the
// session answering user N".
func (p *powerPlatform) sessionOf(t *testing.T, userID int64) string {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()
	want := fmt.Sprintf("pocket-aide-chat-%d", userID)
	for id, s := range p.sessions {
		if s.name == want {
			return id
		}
	}
	t.Fatalf("the platform has no session named %q", want)
	return ""
}

func newPowerEngine(t *testing.T) (*chat.Store, *chat.Engine, *powerPlatform) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "chat.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	store := chat.New(conn)
	platform := newPowerPlatform()
	engine := chat.NewEngine(store, platform, nil).WithPoll(5 * time.Millisecond)
	t.Cleanup(engine.Wait)
	return store, engine, platform
}

// openSession gives the user's conversation a session the way the app does —
// by sending a message — and returns the session the platform made for it.
func openSession(t *testing.T, engine *chat.Engine, platform *powerPlatform, userID int64) string {
	t.Helper()
	if _, err := engine.Send(context.Background(), userID, "안녕"); err != nil {
		t.Fatalf("send: %v", err)
	}
	engine.Wait()
	return platform.sessionOf(t, userID)
}

// PRD-1 AC12: before the first message there is no session to report on, and
// the tab says that rather than showing a session that is off.
func TestEngine_SessionPowerWithNoSessionYet(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	power, err := engine.SessionPowerOf(context.Background(), 1)
	if err != nil {
		t.Fatalf("session power: %v", err)
	}
	if power.Power != chat.PowerNone || power.State != "" {
		t.Fatalf("power = %+v, want power=none with no platform state", power)
	}
	if reads, writes, created := platform.counts(); reads+writes+created != 0 {
		t.Fatalf("platform was called with no session to ask about: reads=%d writes=%d created=%d", reads, writes, created)
	}
}

// PRD-1 AC12 · PRD-11 AC5: a session whose workload is up reads as on, whether
// it is answering right now or idling, and the status read never touches the
// transcript — a transcript read would restore an archived session.
func TestEngine_SessionPowerReportsTheStateWithoutTouchingTheTranscript(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	readsBefore, writesBefore, _ := platform.counts()

	for _, state := range []string{"active", "idle"} {
		platform.setState(id, state)
		power, err := engine.SessionPowerOf(context.Background(), 1)
		if err != nil {
			t.Fatalf("session power: %v", err)
		}
		if power.Power != chat.PowerOn || power.State != state {
			t.Fatalf("state %q read as %+v, want power=on", state, power)
		}
		if power.LastAccess == "" {
			t.Fatalf("state %q lost the last access the tab shows beside it", state)
		}
	}

	platform.archive(id)
	power, err := engine.SessionPowerOf(context.Background(), 1)
	if err != nil {
		t.Fatalf("session power: %v", err)
	}
	if power.Power != chat.PowerOff || power.State != "snapshot" {
		t.Fatalf("archived session read as %+v, want power=off", power)
	}
	if power.Checkpoint == nil || power.Checkpoint.CreatedAt == "" || power.Checkpoint.SizeBytes == nil {
		t.Fatalf("archived session lost its checkpoint detail: %+v", power.Checkpoint)
	}
	if power.Restored {
		t.Fatal("reading the state reported a restore")
	}
	if reads, writes, _ := platform.counts(); reads != readsBefore || writes != writesBefore {
		t.Fatalf("a status read touched the transcript: reads %d -> %d, writes %d -> %d",
			readsBefore, reads, writesBefore, writes)
	}
}

// PRD-1 AC12: a platform state this server does not know is shown as unknown,
// with the platform's own word kept — not guessed into on or off.
func TestEngine_SessionPowerKeepsAnUnknownStateUnknown(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	platform.setState(id, "hibernating")
	power, err := engine.SessionPowerOf(context.Background(), 1)
	if err != nil {
		t.Fatalf("session power: %v", err)
	}
	if power.Power != chat.PowerUnknown || power.State != "hibernating" {
		t.Fatalf("power = %+v, want power=unknown with the platform's state kept", power)
	}
}

// PRD-1 AC12 · AC4: a session the platform no longer has is its own answer —
// nothing is lost, and the next message opens a new one.
func TestEngine_SessionPowerOfASessionThePlatformNoLongerHas(t *testing.T) {
	store, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	platform.forget(id)
	power, err := engine.SessionPowerOf(context.Background(), 1)
	if err != nil {
		t.Fatalf("session power: %v", err)
	}
	if power.Power != chat.PowerGone {
		t.Fatalf("power = %+v, want power=gone", power)
	}
	if got := len(messages(t, store, 1)); got == 0 {
		t.Fatal("the conversation's history went with the session")
	}
}

// PRD-1 AC12: the platform refusing the status read is a failure the caller
// hears about, not a session reported as off.
func TestEngine_SessionPowerRelaysAFailedRead(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	openSession(t, engine, platform, 1)
	platform.fail(http.StatusInternalServerError)
	if _, err := engine.SessionPowerOf(context.Background(), 1); err == nil {
		t.Fatal("a refused status read answered as a state")
	}
}

// PRD-1 AC12 · PRD-11 AC9: the one archived session the conversation points at
// is the one restored, and the conversation carries on in it — no second
// session is made, however the button is pressed.
func TestEngine_PowerOnRestoresTheSameSession(t *testing.T) {
	store, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	_, _, createdBefore := platform.counts()
	platform.archive(id)

	power, err := engine.PowerOnSession(context.Background(), 1)
	if err != nil {
		t.Fatalf("power on: %v", err)
	}
	if power.Power != chat.PowerOn || !power.Restored {
		t.Fatalf("power = %+v, want power=on restored=true", power)
	}
	if got := platform.switchedSessions(); len(got) != 1 || got[0] != id {
		t.Fatalf("restored %v, want exactly [%s]", got, id)
	}
	if got := platform.sessionOf(t, 1); got != id {
		t.Fatalf("the platform holds session %q for this chat, want the same %q", got, id)
	}

	// And the conversation really does carry on in it: the next message is a
	// turn of the same session, and no new one was created for it.
	if _, err := engine.Send(context.Background(), 1, "이어서"); err != nil {
		t.Fatalf("send after power on: %v", err)
	}
	engine.Wait()
	if _, _, created := platform.counts(); created != createdBefore {
		t.Fatalf("sessions created %d -> %d after a restore; a restore must not open a new session", createdBefore, created)
	}
	if got := transcriptOf(messages(t, store, 1)); !strings.Contains(got, "assistant:re: 이어서|") {
		t.Fatalf("the restored session did not answer the next message: %s", got)
	}
}

// PRD-1 AC12: a session that is already on is left alone — the restore is not
// sent at all, so a second tap (or a second screen) cannot disturb it.
func TestEngine_PowerOnLeavesALiveSessionAlone(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	platform.setState(id, "idle")
	power, err := engine.PowerOnSession(context.Background(), 1)
	if err != nil {
		t.Fatalf("power on: %v", err)
	}
	if power.Power != chat.PowerOn || power.Restored {
		t.Fatalf("power = %+v, want power=on restored=false", power)
	}
	if got := platform.switchedSessions(); len(got) != 0 {
		t.Fatalf("restored %v, want nothing", got)
	}
}

// PRD-1 AC12 · AC4: with no session yet there is nothing to turn on, and the
// button must not invent one — the first message is what opens a session.
func TestEngine_PowerOnWithNoSessionYet(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	if _, err := engine.PowerOnSession(context.Background(), 1); !errors.Is(err, chat.ErrNoConversation) {
		t.Fatalf("err = %v, want ErrNoConversation", err)
	}
	if _, _, created := platform.counts(); created != 0 {
		t.Fatalf("created %d sessions, want 0", created)
	}
}

// PRD-1 AC12: a session the platform lost cannot be turned back on, and saying
// so is not the same as losing the conversation.
func TestEngine_PowerOnOfASessionThePlatformNoLongerHas(t *testing.T) {
	store, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	platform.forget(id)
	if _, err := engine.PowerOnSession(context.Background(), 1); !errors.Is(err, chat.ErrPowerSessionGone) {
		t.Fatalf("err = %v, want ErrPowerSessionGone", err)
	}
	if got := len(messages(t, store, 1)); got == 0 {
		t.Fatal("the conversation's history went with the session")
	}
}

// PRD-1 AC12 · PRD-11 AC11: the platform holding the session for another
// lifecycle operation is its own refusal — try again shortly.
func TestEngine_PowerOnWhileThePlatformHoldsTheSession(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	platform.archive(id)
	platform.failSwitch(http.StatusConflict, "")
	if _, err := engine.PowerOnSession(context.Background(), 1); !errors.Is(err, chat.ErrPowerBusy) {
		t.Fatalf("err = %v, want ErrPowerBusy", err)
	}
}

// PRD-1 AC7 · PRD-11 AC11: a restore that finds every login slot taken says
// that, and is not folded into the plain unreachable platform.
func TestEngine_PowerOnWithNoLoginSlotFree(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	platform.archive(id)
	platform.failSwitch(http.StatusServiceUnavailable, `{"error":"auth pool exhausted for codex-chat"}`)
	_, err := engine.PowerOnSession(context.Background(), 1)
	var platformErr *chat.PlatformError
	if !errors.As(err, &platformErr) || !platformErr.AuthPoolExhausted {
		t.Fatalf("err = %v, want a platform error marked auth pool exhausted", err)
	}

	platform.failSwitch(http.StatusServiceUnavailable, `{"error":"control plane is down"}`)
	_, err = engine.PowerOnSession(context.Background(), 1)
	if !errors.As(err, &platformErr) || platformErr.AuthPoolExhausted {
		t.Fatalf("err = %v, want a plain 503 platform error", err)
	}
}

// PRD-1 AC12: a restore the platform took but whose state could not be read
// back is reported as accepted and unconfirmed — never as a session that is
// now on.
func TestEngine_PowerOnReportsAcceptedWhenTheStateCannotBeReadBack(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	id := openSession(t, engine, platform, 1)
	platform.archive(id)
	platform.failSwitch(http.StatusAccepted, "not json")

	power, err := engine.PowerOnSession(context.Background(), 1)
	if err != nil {
		t.Fatalf("power on: %v", err)
	}
	if !power.Restored {
		t.Fatal("an accepted restore was not reported as accepted")
	}
	// The record read back after the accept is what settles the state: the
	// session really is active again here, because the fake's switch moved it.
	if power.Power != chat.PowerOff {
		t.Fatalf("power = %+v, want the state read back from the record (off — the accept did not move it)", power)
	}

	// With the record unreadable too, there is no state to claim.
	platform.fail(http.StatusInternalServerError)
	power, err = engine.PowerOnSession(context.Background(), 1)
	if err == nil && power.Power != chat.PowerUnknown {
		t.Fatalf("power = %+v err = %v, want unknown or a relayed failure", power, err)
	}
}

// PRD-1 AC12: one user's button never reaches another user's session — both
// the conversation and the session come from the signed in user's own row.
func TestEngine_PowerOnStaysInTheCallersConversation(t *testing.T) {
	_, engine, platform := newPowerEngine(t)
	mine := openSession(t, engine, platform, 1)
	theirs := openSession(t, engine, platform, 2)
	platform.archive(mine)
	platform.archive(theirs)

	if _, err := engine.PowerOnSession(context.Background(), 2); err != nil {
		t.Fatalf("power on: %v", err)
	}
	if got := platform.switchedSessions(); len(got) != 1 || got[0] != theirs {
		t.Fatalf("restored %v, want only user 2's session %s", got, theirs)
	}
	power, err := engine.SessionPowerOf(context.Background(), 1)
	if err != nil {
		t.Fatalf("session power: %v", err)
	}
	if power.Power != chat.PowerOff {
		t.Fatalf("user 1's session is %+v, want still off", power)
	}
}
