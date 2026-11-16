package chat

import (
	"sync"
	"time"
)

// PresenceTTL is how long a chat-screen read keeps a user "present".
//
// Both apps poll GET /api/chat only while the chat screen is on top and the
// app is in the foreground (iOS ChatTab's .task(id: scenePhase) with
// scenePhase == .active, Android ChatViewModel.pollWhileVisible), every 15s
// when idle. Three idle intervals absorb one dropped poll plus network delay;
// a tighter TTL would blink present off between healthy polls, and a looser
// one would keep reporting someone who left.
const PresenceTTL = 45 * time.Second

// Presence answers "is the user looking at the chat screen right now?" from
// the chat reads the screen already makes — nothing in the apps has to send a
// heartbeat (PRD-1 AC8).
//
// It is process memory, not a table: the value dies within a minute, every
// poll would be a write, and the backend runs one replica (k8s/deployment.yaml).
// A replaced pod starts empty and reports absent until the next poll, which
// errs toward "not there" rather than inventing a presence.
type Presence struct {
	mu       sync.Mutex
	lastSeen map[int64]time.Time
	now      func() time.Time
}

// NewPresence builds an empty tracker.
func NewPresence() *Presence { return NewPresenceWithClock(time.Now) }

// NewPresenceWithClock builds a tracker reading time from now.
func NewPresenceWithClock(now func() time.Time) *Presence {
	return &Presence{lastSeen: map[int64]time.Time{}, now: now}
}

// State is what a caller learns about a user's chat screen (PRD-12 AC22).
type State struct {
	Present bool `json:"present"`
	// LastSeenAt is the unix second of the most recent chat read, or nil when
	// this process has never seen one.
	LastSeenAt *int64 `json:"last_seen_at"`
	TTLSeconds int    `json:"ttl_seconds"`
}

// Touch records that userID's chat screen just read the conversation.
func (p *Presence) Touch(userID int64) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.lastSeen[userID] = p.now()
}

// State reports whether userID read the chat screen within PresenceTTL.
func (p *Presence) State(userID int64) State {
	out := State{TTLSeconds: int(PresenceTTL / time.Second)}
	if p == nil {
		return out
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	seen, ok := p.lastSeen[userID]
	if !ok {
		return out
	}
	at := seen.Unix()
	out.LastSeenAt = &at
	out.Present = p.now().Sub(seen) < PresenceTTL
	return out
}
