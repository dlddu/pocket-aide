package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

// Archiving the conversation's codex session on demand (PRD-1 AC4). It is the
// platform's own snapshot — the same `POST /sessions/{id}/snapshot` the
// sessions tab calls (PRD-11 AC8) — called with the chat tab's tenant, because
// a chat session is outside the sessions tab's tenant and that tab's proxy
// answers 404 for it (PRD-11 AC2).
//
// Nothing of pocket-aide's own is thrown away: chat_messages stays, and so do
// the row's platform_session_id and transcript offset. That is what lets the
// existing entry points carry on — the next message restores this very session
// and the conversation continues where it left off (AC4).

// sessionStateSnapshot is the platform state that means the workload is
// archived and its environment handed back ("아카이브됨" to the user, PRD-11
// AC5). The platform has no state for "archiving", so it is either this or
// not archived yet (PRD-11 결정 2026-10-23).
const sessionStateSnapshot = "snapshot"

var (
	// ErrArchiveSessionUnnamed means the request did not say which session it
	// means. The server will not archive "whatever is current" on a caller's
	// behalf: the point of naming it is that the person saw that session.
	ErrArchiveSessionUnnamed = errors.New("say which session to archive")
	// ErrArchiveSessionChanged means the named session is no longer the one
	// answering this conversation — the conversation switched, or a new
	// session took over, between the confirmation and this request.
	ErrArchiveSessionChanged = errors.New("another session answers this conversation now")
	// ErrArchiveQueued means a message is already waiting to go out. The next
	// poll would send it, and sending restores the session, so the archive
	// would be undone by something the person did not ask for just now.
	ErrArchiveQueued = errors.New("a message is waiting to be sent")
	// ErrArchiveSessionGone means the platform has no such session any more.
	// Nothing was archived and nothing is lost: the conversation's history is
	// this server's, and the next message opens a new session (AC4).
	ErrArchiveSessionGone = errors.New("the session answering this conversation is already gone")
	// ErrArchiveDisabled is the platform's archive gate being off. The pod
	// stays up and no data is lost, so it is a state to explain rather than a
	// failure (PRD-11 AC8).
	ErrArchiveDisabled = errors.New("archiving is turned off on the session platform")
)

// ArchiveResult is what became of the session, read back from the platform
// rather than assumed from the request.
//
// Archived is true only when the session really reads as archived now. A
// platform that took the request without having finished leaves it false with
// the state it reported, so the apps can say the request was accepted without
// claiming the environment is already handed back.
type ArchiveResult struct {
	SessionID  string               `json:"session_id"`
	State      string               `json:"state,omitempty"`
	Archived   bool                 `json:"archived"`
	Checkpoint *sessions.Checkpoint `json:"checkpoint,omitempty"`
}

// ArchiveSession archives the codex session answering userID's conversation.
//
// expectedSessionID is the session the person was shown and confirmed. The
// conversation's own row decides which session is archived, and a request
// naming a different one is refused — showing the confirmation and acting on
// it are two steps, and between them the conversation can switch or a new
// session can take over the answering. Deriving the session from the signed in
// user's row (never from the request) is also what keeps one user from
// archiving another's, the same way the task routes do it.
func (e *Engine) ArchiveSession(ctx context.Context, userID int64, expectedSessionID string) (ArchiveResult, error) {
	expected := strings.TrimSpace(expectedSessionID)
	if expected == "" {
		return ArchiveResult{}, ErrArchiveSessionUnnamed
	}
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return ArchiveResult{}, err
	}
	if conv.PlatformSessionID == "" {
		return ArchiveResult{}, ErrNoConversation
	}
	if conv.PlatformSessionID != expected {
		return ArchiveResult{}, ErrArchiveSessionChanged
	}
	// Held before the platform call, so no turn starts under the archive and
	// the poll's own resume cannot restore the session while it runs.
	if !e.claimQuiet(userID) {
		return ArchiveResult{}, ErrBusy
	}
	defer e.release(userID)
	// Checked under the hold, so a message that arrives after this point is
	// the person's own next send — and that send restoring the session is AC4
	// working as it should, not a race.
	queued, err := e.store.GetQueued(ctx, userID, conv.ID)
	if err != nil {
		return ArchiveResult{}, err
	}
	if queued != nil {
		return ArchiveResult{}, ErrArchiveQueued
	}
	if err := e.snapshot(ctx, conv.PlatformSessionID); err != nil {
		return ArchiveResult{}, err
	}
	out := ArchiveResult{SessionID: conv.PlatformSessionID}
	// What the platform says now, not what the request hoped for. `GET` is the
	// one way to ask: reading the transcript would restore the session that
	// was just archived (PRD-11 AC9 · T-11 시나리오 10). A re-read that fails
	// leaves the archive reported as accepted-and-unconfirmed — it did happen,
	// and saying otherwise would be a lie in the other direction.
	state, err := e.sessionState(ctx, conv.PlatformSessionID)
	if err != nil {
		e.logger.Warn("chat: archived but could not re-read the session",
			"user", userID, "session", conv.PlatformSessionID, "err", err)
		return out, nil
	}
	out.State = state.State
	out.Archived = state.State == sessionStateSnapshot
	out.Checkpoint = state.Checkpoint
	return out, nil
}

// snapshot asks the platform to archive sessionID, translating its refusals
// into the ones the chat tab tells apart.
func (e *Engine) snapshot(ctx context.Context, sessionID string) error {
	status, raw, err := e.platform.Do(ctx, http.MethodPost,
		"/sessions/"+url.PathEscape(sessionID)+"/snapshot", nil)
	if err != nil {
		return err
	}
	switch status {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusNotFound:
		return ErrArchiveSessionGone
	case http.StatusConflict:
		return ErrBusy
	case http.StatusServiceUnavailable:
		// A snapshot's 503 is the archive gate being off (PRD-11 AC11) —
		// except for the one 503 that names exhausted login slots, which is
		// about the platform's pool and not about this feature.
		if platformErr := platformError("snapshot", status, raw); platformErr.AuthPoolExhausted {
			return platformErr
		}
		return ErrArchiveDisabled
	default:
		return platformError("snapshot", status, raw)
	}
}

// sessionState reads one session's own record — its state and, once archived,
// when the checkpoint was taken and how big it is. Only those two fields are
// used; the rest of the record says nothing the chat tab shows.
func (e *Engine) sessionState(ctx context.Context, sessionID string) (sessions.Session, error) {
	status, raw, err := e.platform.Do(ctx, http.MethodGet, "/sessions/"+url.PathEscape(sessionID), nil)
	if err != nil {
		return sessions.Session{}, err
	}
	if status != http.StatusOK {
		return sessions.Session{}, platformError("get session", status, raw)
	}
	var out sessions.Session
	if err := json.Unmarshal(raw, &out); err != nil {
		return sessions.Session{}, fmt.Errorf("decode session: %w", err)
	}
	return out, nil
}
