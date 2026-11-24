package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
// monitorPairDisabledMarker is what the control plane says when the pair
// feature is off on that deployment. Matching the statement — rather than the
// bare 503 — is what keeps a transient outage from being read as "no monitor".
const monitorPairDisabledMarker = "monitor pair is not enabled"

var (
	// ErrPairUnknown is a pair lookup that neither found a monitor nor proved
	// there is none: a timeout, a 5xx, an unreadable reply. The panel reports
	// it instead of falling back, because falling back would mean reading and
	// writing the parent's own records — which a paired parent may not do.
	ErrPairUnknown = errors.New("whether this chat session has a monitor could not be determined")
	// ErrPairRefused is an authorization failure on that lookup. It is kept
	// apart from ErrPairUnknown so a permission problem is never mistaken for
	// a deployment without the feature.
	ErrPairRefused = errors.New("the platform refused this backend's identity for the monitor pair")
)

// monitorPairTTL keeps the pair lookup off every read. The platform's answer
// changes only when a monitor is bound or unbound, which is rare.
const monitorPairTTL = 30 * time.Second

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
	// IntakeLive is whether the monitor can take an instruction right now. It
	// is false while the monitor is frozen: the platform still accepts the
	// instruction durably, and the app shows it as waiting rather than done.
	IntakeLive bool `json:"intake_live"`
}

// monitorBase is the backend-facing pair surface for the parent session this
// backend's tenant owns. The monitor is never named here — the platform takes
// it from the binding, so nothing the app or this process says can point the
// read at another pair.
func monitorBase(parentSessionID string) string {
	return "/sessions/" + url.PathEscape(parentSessionID) + "/monitor"
}

// Paired reports whether the panel should address a monitor.
func (p MonitorPair) Paired() bool { return p.SessionID != "" }

type monitorPairCache struct {
	mu   sync.Mutex
	at   time.Time
	for_ string
	pair MonitorPair
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

	status, raw, err := e.platform.Do(ctx, http.MethodGet, monitorBase(chatSessionID)+"/pair", nil)
	unpaired := func() (MonitorPair, error) {
		e.pairs.mu.Lock()
		e.pairs.for_, e.pairs.pair, e.pairs.at = chatSessionID, MonitorPair{}, e.now()
		e.pairs.mu.Unlock()
		return MonitorPair{}, nil
	}
	// Only a *clear* answer may become "unpaired", because unpaired means the
	// panel reads and writes this session's own work records — the very path a
	// paired parent must not use. An ambiguous or refused lookup is an error
	// the caller reports, never a quiet downgrade into writing here.
	if err != nil {
		return MonitorPair{}, fmt.Errorf("%w: %v", ErrPairUnknown, err)
	}
	switch {
	case status == http.StatusOK:
	case status == http.StatusNotFound:
		// This session has no monitor. Clear, and the common case.
		return unpaired()
	case status == http.StatusNotImplemented:
		// A control plane without the route: clearly an older deployment.
		return unpaired()
	case status == http.StatusServiceUnavailable && strings.Contains(string(raw), monitorPairDisabledMarker):
		// The deployment says the feature is off. That is a statement, not a
		// hiccup, so it is legacy too.
		return unpaired()
	case status == http.StatusUnauthorized, status == http.StatusForbidden:
		// A refused lookup must never look like "no monitor": that would hand
		// the parent its own write path back on an authorization failure.
		return MonitorPair{}, fmt.Errorf("%w: the platform refused this backend's identity (%d)", ErrPairRefused, status)
	default:
		return MonitorPair{}, fmt.Errorf("%w: the platform answered %d", ErrPairUnknown, status)
	}
	var view struct {
		Binding struct {
			MonitorSessionID string `json:"monitorSessionId"`
		} `json:"binding"`
		Monitor *struct {
			State  string `json:"state"`
			Reads  string `json:"reads"`
			Intake string `json:"intake"`
		} `json:"monitor"`
	}
	if err := json.Unmarshal(raw, &view); err != nil {
		// A 200 we cannot read is ambiguous, not legacy.
		return MonitorPair{}, fmt.Errorf("%w: the pair reply could not be read", ErrPairUnknown)
	}
	if view.Binding.MonitorSessionID == "" {
		// A 200 that names no monitor is the platform saying there is none.
		return unpaired()
	}
	pair := MonitorPair{SessionID: view.Binding.MonitorSessionID}
	if view.Monitor != nil {
		pair.State = view.Monitor.State
		// Reads stay available while the monitor is away: the platform serves
		// the state the monitor published, with the time it published it.
		// Intake is what stops, and the platform says so.
		pair.Reachable = true
		pair.IntakeLive = view.Monitor.Intake == "available"
	}
	e.pairs.mu.Lock()
	e.pairs.for_, e.pairs.pair, e.pairs.at = chatSessionID, pair, e.now()
	e.pairs.mu.Unlock()
	return pair, nil
}

// workPrefix is the path the panel reads and writes under, and the pair it
// came from. A paired chat session uses the backend-facing monitor surface —
// which answers while the monitor is frozen, from what it published — and an
// unpaired one keeps reading its own session's work records as before.
func (e *Engine) workPrefix(ctx context.Context, chatSessionID string) (string, MonitorPair, error) {
	pair, err := e.monitorPair(ctx, chatSessionID)
	if err != nil {
		return "", MonitorPair{}, err
	}
	if pair.Paired() {
		return monitorBase(chatSessionID), pair, nil
	}
	return "/sessions/" + url.PathEscape(chatSessionID), pair, nil
}
