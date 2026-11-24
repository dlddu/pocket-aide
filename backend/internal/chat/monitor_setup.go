package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Turning the separated monitor on for one person's chat session.
//
// The monitor is a session of its own, created by this backend with the
// marketplace bundle that carries the monitor's role skills and bound to the
// chat session it serves. The platform records that pair at create time from
// `monitorFor`, so the binding is the control plane's, not something the app or
// a workload can claim later.
//
// Nothing here changes an existing pair. If the chat session already has a
// monitor, this reuses it — no second session, no second writer.
const (
	// monitorPlugin is the bundle whose skills the monitor session runs
	// (plugin-marketplace `monitor-session`): the pair runtime, the monitoring
	// and session lifecycles, and the next-task selector.
	monitorPlugin = "monitor-session"
	// monitorBootstrapPrompt is the first turn of a new monitor. It only tells
	// the session to run its own role skill, which is what binds the pair and
	// drains whatever is already waiting. The skills decide the rest.
	monitorBootstrapPrompt = "이 세션은 분리된 작업 모니터입니다. 설치된 모니터 역할 스킬을 읽고 그 순서대로 쌍을 확정하고, 인박스를 확인해 처리하세요."
	// monitorDrainPrompt wakes a live monitor so a newly accepted instruction
	// is discovered now rather than at the next scheduled run.
	monitorDrainPrompt = "새 지시가 인박스에 접수됐습니다. 모니터 역할 스킬의 순서대로 확인해 처리하세요."
)

// MonitorSetup is what the settings screen and the conversation both show
// after asking for the monitor.
type MonitorSetup struct {
	// SessionID is the monitor serving this chat session.
	SessionID string `json:"session_id"`
	// Created is false when an existing monitor was reused, which is the
	// common case on every call after the first.
	Created bool `json:"created"`
	// Bootstrapped is whether this call started the monitor's first turn.
	Bootstrapped bool   `json:"bootstrapped"`
	State        string `json:"state,omitempty"`
}

func monitorSessionName(userID int64) string { return fmt.Sprintf("%s-monitor", sessionName(userID)) }

// EnableMonitor turns the separated monitor on for userID's conversation. It is
// idempotent: an already bound monitor is returned as it is.
//
// The caller is this backend with its own tenant identity; the parent session
// is the one this user's conversation already owns. Neither the app nor the
// request body names a session, so nobody can bind someone else's chat session
// to a monitor.
func (e *Engine) EnableMonitor(ctx context.Context, userID int64) (MonitorSetup, error) {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return MonitorSetup{}, err
	}
	parent := conv.platformSessionID
	if parent == "" {
		// No chat session has answered this conversation yet, so there is
		// nothing to monitor. Opening one is the chat path's job, not this.
		return MonitorSetup{}, ErrNoConversation
	}

	// Already paired? Reuse. This is the guard against a second monitor and a
	// second writer, and it asks the platform rather than trusting local state.
	pair, err := e.monitorPair(ctx, parent)
	if err != nil {
		return MonitorSetup{}, err
	}
	if pair.Paired() {
		return MonitorSetup{SessionID: pair.SessionID, State: pair.State}, nil
	}

	// A monitor may exist from an earlier attempt whose reply was lost. Adopt
	// it by its name and its recorded parent instead of creating another.
	if id, state, ok := e.adoptMonitor(ctx, userID, parent); ok {
		e.forgetPair(parent)
		return MonitorSetup{SessionID: id, State: state,
			Bootstrapped: e.bootstrapMonitor(ctx, id, monitorBootstrapPrompt)}, nil
	}

	body, _ := json.Marshal(map[string]any{
		"name":         monitorSessionName(userID),
		"workloadType": workloadCodex,
		"plugins":      []string{monitorPlugin},
		// The platform checks this parent belongs to the same tenant and
		// records the pair. It refuses a parent that is itself a monitor.
		"monitorFor": parent,
	})
	status, raw, err := e.platform.Do(ctx, http.MethodPost, "/sessions", body)
	if err == nil && status != http.StatusCreated && status != http.StatusOK {
		err = platformError("create monitor", status, raw)
	}
	if err != nil {
		// A create whose reply was lost may still have landed. Adopt before
		// reporting, so a retry cannot make a second monitor.
		if id, state, ok := e.adoptMonitor(context.WithoutCancel(ctx), userID, parent); ok {
			e.logger.Warn("chat: adopted a monitor whose create failed", "user", userID, "session", id, "err", err)
			e.forgetPair(parent)
			return MonitorSetup{SessionID: id, State: state,
				Bootstrapped: e.bootstrapMonitor(ctx, id, monitorBootstrapPrompt)}, nil
		}
		return MonitorSetup{}, err
	}
	var created struct {
		ID    string `json:"id"`
		State string `json:"state"`
	}
	if err := json.Unmarshal(raw, &created); err != nil || created.ID == "" {
		return MonitorSetup{}, fmt.Errorf("create monitor: reply had no session id")
	}
	e.forgetPair(parent)
	return MonitorSetup{SessionID: created.ID, Created: true, State: created.State,
		Bootstrapped: e.bootstrapMonitor(ctx, created.ID, monitorBootstrapPrompt)}, nil
}

// MonitorStatus is the read the settings screen uses. It is a read: it does not
// create anything and does not wake the monitor.
func (e *Engine) MonitorStatus(ctx context.Context, userID int64) (MonitorSetup, error) {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return MonitorSetup{}, err
	}
	if conv.platformSessionID == "" {
		return MonitorSetup{}, nil
	}
	pair, err := e.monitorPair(ctx, conv.platformSessionID)
	if err != nil {
		return MonitorSetup{}, err
	}
	return MonitorSetup{SessionID: pair.SessionID, State: pair.State}, nil
}

// adoptMonitor finds a monitor this backend already made for userID. The match
// is the name *and* the platform's own record of which parent it monitors, so
// a session that monitors someone else is never adopted.
func (e *Engine) adoptMonitor(ctx context.Context, userID int64, parent string) (string, string, bool) {
	status, raw, err := e.platform.Do(ctx, http.MethodGet, "/sessions", nil)
	if err != nil || status != http.StatusOK {
		return "", "", false
	}
	var list struct {
		Sessions []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			State      string `json:"state"`
			MonitorFor string `json:"monitorFor"`
		} `json:"sessions"`
	}
	if json.Unmarshal(raw, &list) != nil {
		return "", "", false
	}
	for _, s := range list.Sessions {
		if s.Name == monitorSessionName(userID) && s.MonitorFor == parent {
			return s.ID, s.State, true
		}
	}
	return "", "", false
}

// bootstrapMonitor starts the monitor's first turn, which is what makes the
// role skills run: bind the pair, drain the inbox, schedule the next run. A
// monitor that is not live is left alone — the platform holds instructions
// durably and the person restores it when they choose.
func (e *Engine) bootstrapMonitor(ctx context.Context, monitorID, prompt string) bool {
	body, _ := json.Marshal(map[string]any{"payload": prompt})
	status, raw, err := e.platform.Do(ctx, http.MethodPost,
		"/sessions/"+url.PathEscape(monitorID)+"/write", body)
	if err != nil || (status != http.StatusOK && status != http.StatusAccepted && status != http.StatusCreated) {
		e.logger.Warn("chat: monitor did not take its prompt", "session", monitorID,
			"status", status, "err", err, "body", string(raw))
		return false
	}
	return true
}

// NudgeMonitor wakes a live monitor after an instruction was accepted, so the
// drain happens now instead of at the next scheduled run. A frozen monitor is
// left to the platform, which resumes it on intake through its own write path
// and says so in the reply (`woken`) — this backend does not have to arrange a
// recovery of its own.
func (e *Engine) NudgeMonitor(ctx context.Context, parent string, pair MonitorPair) bool {
	if !pair.Paired() || !pair.IntakeLive {
		return false
	}
	return e.bootstrapMonitor(ctx, pair.SessionID, monitorDrainPrompt)
}

// forgetPair drops the cached pair so the next read sees the binding that was
// just made rather than waiting out the TTL.
func (e *Engine) forgetPair(parent string) {
	e.pairs.mu.Lock()
	if e.pairs.for_ == parent {
		e.pairs.for_, e.pairs.pair, e.pairs.at = "", MonitorPair{}, time.Time{}
	}
	e.pairs.mu.Unlock()
}

// MonitorOff is what the settings screen gets back when the monitor is turned
// off, including the case where only half of it could be done.
type MonitorOff struct {
	// Was is the monitor that was bound, so the person is told what was turned
	// off rather than just "done".
	Was string `json:"was,omitempty"`
	// Off is true when the pair is gone and the conversation writes its own
	// work records again.
	Off bool `json:"off"`
	// Retry is true when the monitor was closed but the records are still
	// sealed to it. Nothing writes them in that state — which is the safe half
	// to be stuck in — and calling again finishes the rollback.
	Retry bool   `json:"retry,omitempty"`
	Note  string `json:"note,omitempty"`
}

// DisableMonitor turns the separated monitor off for userID's conversation.
//
// The order is the platform's and it matters: the binding goes first, which
// closes the monitor, and only then are the work records handed back to the
// chat session. So there is never a moment when both could write. A partial
// rollback — monitor closed, records still sealed — is reported as something to
// retry rather than as done, and repeating the call is safe.
//
// Nothing is deleted. The tasks, their turns and the accepted instructions stay
// exactly where they are: the records were never moved, so turning the monitor
// off leaves the same history the panel was showing.
func (e *Engine) DisableMonitor(ctx context.Context, userID int64) (MonitorOff, error) {
	conv, err := e.store.Conversation(ctx, userID)
	if err != nil {
		return MonitorOff{}, err
	}
	parent := conv.platformSessionID
	if parent == "" {
		return MonitorOff{}, ErrNoConversation
	}
	status, raw, err := e.platform.Do(ctx, http.MethodDelete, monitorBase(parent)+"/pair", nil)
	if err != nil {
		return MonitorOff{}, fmt.Errorf("%w: %v", ErrPairUnknown, err)
	}
	e.forgetPair(parent)
	switch status {
	case http.StatusOK:
		var view struct {
			Writer struct {
				Writer string `json:"writer"`
				Holder string `json:"holder"`
			} `json:"writer"`
		}
		_ = json.Unmarshal(raw, &view)
		return MonitorOff{Off: true, Was: view.Writer.Holder}, nil
	case http.StatusAccepted:
		// The monitor is closed — it cannot write any more — but the records
		// were not handed back. Say so plainly: this is not "off" yet.
		return MonitorOff{Retry: true,
			Note: "모니터는 닫혔지만 작업 기록이 아직 모니터에 봉인돼 있습니다. 다시 시도하면 마무리됩니다."}, nil
	case http.StatusNotFound:
		// There was no pair. Turning off what is already off is not an error.
		return MonitorOff{Off: true}, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return MonitorOff{}, fmt.Errorf("%w: the platform refused this backend's identity (%d)",
			ErrPairRefused, status)
	case http.StatusNotImplemented:
		// A control plane without the route: there is nothing bound either.
		return MonitorOff{Off: true}, nil
	default:
		return MonitorOff{}, fmt.Errorf("%w: the platform answered %d", ErrPairUnknown, status)
	}
}
