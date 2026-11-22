package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// The work panel's target when a chat session has a monitor.
//
// The platform's separated monitor owns the work records: the chat session
// (the parent) keeps the conversation and reads, and its monitor registers,
// changes and finishes the tasks. So the panel must read the *monitor's*
// records, and an instruction the person sends must go to the monitor rather
// than change a task here.
//
// Nothing about that is asserted by the app. The backend asks the platform for
// the sessions this deployment's tenant owns and keeps the one whose record
// says it monitors this chat session — `monitorFor` is written by the control
// plane at create time, so it cannot be claimed by a workload or forged in a
// header. A deployment whose platform does not have the field yet, or a chat
// session nobody monitors, resolves to no monitor and every existing path
// keeps working unchanged.
const (
	monitorPairTTL = 30 * time.Second
	// monitorStateActive and monitorStateIdle are the states whose session can
	// answer a read. Anything else is frozen, stopped or gone, and the panel
	// says so instead of showing an empty list.
	monitorStateActive = "active"
	monitorStateIdle   = "idle"
)

// MonitorPair is what the panel needs to know about the monitor behind a chat
// session: whether there is one, and whether it can answer right now.
type MonitorPair struct {
	// SessionID is the monitor session. Empty means this chat session is
	// unpaired and the panel reads its own records, as it always has.
	SessionID string `json:"session_id,omitempty"`
	State     string `json:"state,omitempty"`
	// Reachable is whether a read of the monitor's records can succeed now.
	// False is a frozen or stopped monitor: the panel shows the last state it
	// has with its time, and an instruction is reported as not delivered.
	Reachable bool `json:"reachable"`
}

// Paired reports whether the panel should address a monitor.
func (p MonitorPair) Paired() bool { return p.SessionID != "" }

type monitorPairCache struct {
	mu   sync.Mutex
	at   time.Time
	for_ string
	pair MonitorPair
}

type platformSession struct {
	ID         string `json:"id"`
	State      string `json:"state"`
	MonitorFor string `json:"monitorFor"`
}

// monitorPair resolves the monitor of chatSessionID. The lookup is the
// platform's own session list, which is already scoped to this deployment's
// tenant, so a reply can only ever contain sessions this backend may see.
func (e *Engine) monitorPair(ctx context.Context, chatSessionID string) (MonitorPair, error) {
	if chatSessionID == "" {
		return MonitorPair{}, nil
	}
	e.pairs.mu.Lock()
	if e.pairs.for_ == chatSessionID && e.now().Sub(e.pairs.at) < monitorPairTTL {
		cached := e.pairs.pair
		e.pairs.mu.Unlock()
		return cached, nil
	}
	e.pairs.mu.Unlock()

	status, raw, err := e.platform.Do(ctx, http.MethodGet, "/sessions", nil)
	if err != nil {
		return MonitorPair{}, fmt.Errorf("list sessions: %w", err)
	}
	switch status {
	case http.StatusOK:
	case http.StatusNotFound, http.StatusNotImplemented:
		// A control plane without the session list, or without the pair field
		// on it, cannot have a pair either: this chat session is unpaired and
		// every existing path keeps working. Cache it so the lookup is not
		// repeated per read.
		e.pairs.mu.Lock()
		e.pairs.for_, e.pairs.pair, e.pairs.at = chatSessionID, MonitorPair{}, e.now()
		e.pairs.mu.Unlock()
		return MonitorPair{}, nil
	default:
		// Anything else is a real failure. The panel must not guess "unpaired"
		// from it, because that would read the wrong session's work.
		return MonitorPair{}, fmt.Errorf("list sessions: platform answered %d", status)
	}
	var sessions []platformSession
	if err := json.Unmarshal(raw, &sessions); err != nil {
		// Some deployments wrap the list; try that shape before giving up.
		var wrapped struct {
			Sessions []platformSession `json:"sessions"`
		}
		if err2 := json.Unmarshal(raw, &wrapped); err2 != nil {
			return MonitorPair{}, fmt.Errorf("decode sessions: %w", err)
		}
		sessions = wrapped.Sessions
	}
	var pair MonitorPair
	for _, s := range sessions {
		if s.MonitorFor != chatSessionID {
			continue
		}
		pair = MonitorPair{SessionID: s.ID, State: s.State,
			Reachable: s.State == monitorStateActive || s.State == monitorStateIdle}
		break
	}
	e.pairs.mu.Lock()
	e.pairs.for_, e.pairs.pair, e.pairs.at = chatSessionID, pair, e.now()
	e.pairs.mu.Unlock()
	return pair, nil
}

// workTarget is the session whose work records the panel shows, and the pair
// it came from. For an unpaired chat session it is the chat session itself.
func (e *Engine) workTarget(ctx context.Context, chatSessionID string) (string, MonitorPair, error) {
	pair, err := e.monitorPair(ctx, chatSessionID)
	if err != nil {
		return "", MonitorPair{}, err
	}
	if pair.Paired() {
		return pair.SessionID, pair, nil
	}
	return chatSessionID, pair, nil
}
