package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

// Whether the codex session answering the conversation is up, and turning it
// back on when it is not (PRD-1 AC12).
//
// Both are the platform's own answers about one session — the one the signed
// in user's conversation row points at — called with the chat tab's tenant,
// because a chat session is outside the sessions tab's tenant and that tab's
// routes answer 404 for it (PRD-11 AC2).
//
// Reading the state and turning the session on are the only two platform calls
// here, and neither touches the transcript: the platform's read restores an
// archived session (PRD-11 결정 「출력 전송 방식」), so a status poll built on
// read would turn on the very session it was asked about. `GET /sessions/{id}`
// is the one call that does not promote (PRD-11 AC5).
//
// Turning on is `POST /sessions/{id}/switch` — the same restore the sessions
// tab offers (PRD-11 AC9) — and it is only ever sent for the session the row
// already names. Nothing here creates or adopts a session, so no tap, and no
// two screens tapping at once, can leave the user with a second conversation
// session: the one in the row is the one that comes back, with its transcript
// and this server's cursor into it still valid (AC4).

// The platform's session states, as the chat tab reads them (PRD-11 AC5).
// Anything else is a state this server does not know, and it is relayed as
// such rather than guessed at.
const (
	platformStateActive   = "active"
	platformStateIdle     = "idle"
	platformStateSnapshot = "snapshot"
)

// What the chat tab says about the session answering the conversation.
const (
	// PowerNone is "there is no session yet" — the first message opens one
	// (AC4), so there is nothing to turn on and the tab says so instead of
	// offering a button that would have to invent a session.
	PowerNone = "none"
	// PowerOn is a session whose workload is up: active, or idle with its pod
	// still there. Both answer a message without a restore first, which is
	// what "on" means to the person reading the screen.
	PowerOn = "on"
	// PowerOff is an archived session — the environment was handed back and a
	// restore brings it, and the conversation, back (PRD-11 AC5·AC9).
	PowerOff = "off"
	// PowerGone is a session the platform no longer has. Nothing of the
	// conversation is lost — the history is this server's — and the next
	// message opens a new session (AC4), which is what the tab says.
	PowerGone = "gone"
	// PowerUnknown is for a state this server cannot place: a platform state
	// it does not know, or a turn-on the platform took without a state coming
	// back. Saying "unknown" is the honest answer; saying "on" would be a
	// guess the person would act on.
	PowerUnknown = "unknown"
)

var (
	// ErrPowerSessionGone means the platform has no such session. The
	// conversation and its history are untouched, and the next message opens a
	// new session, so this is a state to explain rather than a loss.
	ErrPowerSessionGone = errors.New("the session answering this conversation is already gone")
	// ErrPowerBusy is the platform holding the session for another lifecycle
	// operation (a restore already under way, an archive finishing). Worth
	// trying again shortly, and told apart from a reply in progress — which
	// cannot be the reason here, since a session with a running turn is on.
	ErrPowerBusy = errors.New("the session platform is busy with this session")
)

// SessionPower is the state of the session answering one conversation.
//
// Power is what the screen shows; State is the platform's own word, kept so an
// unknown state can be shown as what it is. LastAccess and Checkpoint are the
// detail the sessions tab shows for the same states (PRD-11 AC5) and are
// absent exactly when the platform did not give them.
//
// Restored is set only by a turn-on that actually restored the session, so the
// apps can tell "turned it on" from "it was already on" without guessing.
type SessionPower struct {
	Power      string               `json:"power"`
	State      string               `json:"state,omitempty"`
	LastAccess string               `json:"last_access,omitempty"`
	Checkpoint *sessions.Checkpoint `json:"checkpoint,omitempty"`
	Restored   bool                 `json:"restored,omitempty"`
}

// SessionPowerOf is what the chat tab shows for the session answering userID's
// conversation. It asks the platform for that one session's record and relays
// it; it does not list sessions, and it does not read the transcript, so a
// screen can poll it without promoting an idle session or restoring an
// archived one (PRD-11 AC5).
func (e *Engine) SessionPowerOf(ctx context.Context, userID int64) (SessionPower, error) {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return SessionPower{}, err
	}
	if conv.platformSessionID == "" {
		return SessionPower{Power: PowerNone}, nil
	}
	record, status, err := e.readSessionRecord(ctx, conv.platformSessionID)
	if err != nil {
		return SessionPower{}, err
	}
	if status == http.StatusNotFound {
		return SessionPower{Power: PowerGone}, nil
	}
	return powerOf(record), nil
}

// PowerOnSession turns the session answering userID's conversation back on.
//
// The session comes from that user's own conversation row and never from the
// request, which is what keeps one user from restoring another's session — the
// same rule the task routes follow. The row is not changed either: a restore
// brings back the session the row already names, so the conversation carries on
// in it (AC4) and no second session is created.
//
// A session that is already on is left alone — no platform write at all — so a
// second tap, or a second screen, costs nothing and cannot disturb a running
// turn.
func (e *Engine) PowerOnSession(ctx context.Context, userID int64) (SessionPower, error) {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return SessionPower{}, err
	}
	if conv.platformSessionID == "" {
		return SessionPower{}, ErrNoConversation
	}
	record, status, err := e.readSessionRecord(ctx, conv.platformSessionID)
	if err != nil {
		return SessionPower{}, err
	}
	if status == http.StatusNotFound {
		return SessionPower{}, ErrPowerSessionGone
	}
	// Only an archived session is restored. One that is already up needs
	// nothing, and a state this server does not know is not a state to act on
	// — in both cases the answer is what the platform just said, with
	// Restored left false.
	if record.State != platformStateSnapshot {
		return powerOf(record), nil
	}
	restored, err := e.switchSession(ctx, conv.platformSessionID)
	if err != nil {
		return SessionPower{}, err
	}
	out := powerOf(restored)
	if out.Power == PowerUnknown {
		// The restore was accepted — the platform answered it without a
		// refusal — but its answer carried no state this server could place.
		// Reading the record again is the one way to ask, and if that fails
		// too the turn-on is reported as accepted-and-unconfirmed rather than
		// as a session that is now on.
		if again, status, err := e.readSessionRecord(ctx, conv.platformSessionID); err == nil && status == http.StatusOK {
			out = powerOf(again)
		} else {
			e.logger.Warn("chat: turned the session on but could not read its state back",
				"user", userID, "session", conv.platformSessionID, "err", err)
		}
	}
	out.Restored = true
	return out, nil
}

// powerOf is the one place a platform state becomes what the screen shows, so
// the status route and the turn-on cannot disagree about what "on" is.
func powerOf(record sessions.Session) SessionPower {
	out := SessionPower{
		State:      record.State,
		LastAccess: record.LastAccess,
		Checkpoint: record.Checkpoint,
	}
	switch record.State {
	case platformStateActive, platformStateIdle:
		out.Power = PowerOn
	case platformStateSnapshot:
		out.Power = PowerOff
	default:
		out.Power = PowerUnknown
	}
	return out
}

// readSessionRecord reads one session's own record. The status is returned
// beside it because 404 is an answer the caller acts on — the session is gone
// — rather than a failure to relay.
func (e *Engine) readSessionRecord(ctx context.Context, sessionID string) (sessions.Session, int, error) {
	status, raw, err := e.platform.Do(ctx, http.MethodGet, "/sessions/"+url.PathEscape(sessionID), nil)
	if err != nil {
		return sessions.Session{}, 0, err
	}
	if status == http.StatusNotFound {
		return sessions.Session{}, status, nil
	}
	if status != http.StatusOK {
		return sessions.Session{}, status, platformError("get session", status, raw)
	}
	var out sessions.Session
	if err := json.Unmarshal(raw, &out); err != nil {
		return sessions.Session{}, status, fmt.Errorf("decode session: %w", err)
	}
	return out, status, nil
}

// switchSession restores an archived session into a fresh pod (PRD-11 AC9),
// translating the platform's refusals into the ones the chat tab tells apart.
// The session it answers with is the caller's evidence of what happened; an
// answer this server cannot decode leaves the state unset, and the caller
// reads the record again rather than claiming the session is on.
func (e *Engine) switchSession(ctx context.Context, sessionID string) (sessions.Session, error) {
	status, raw, err := e.platform.Do(ctx, http.MethodPost,
		"/sessions/"+url.PathEscape(sessionID)+"/switch", []byte("{}"))
	if err != nil {
		return sessions.Session{}, err
	}
	switch status {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		var out sessions.Session
		if json.Unmarshal(raw, &out) != nil {
			return sessions.Session{}, nil
		}
		return out, nil
	case http.StatusNotFound:
		return sessions.Session{}, ErrPowerSessionGone
	case http.StatusConflict:
		return sessions.Session{}, ErrPowerBusy
	default:
		// A restore's 503 is either the platform being out of login slots for
		// this workload type (AC7's refusal, which names itself in the body)
		// or the platform not being reachable — PRD-11 AC11 treats every other
		// 503 as unreachable, and platformError keeps the two apart.
		return sessions.Session{}, platformError("switch", status, raw)
	}
}
