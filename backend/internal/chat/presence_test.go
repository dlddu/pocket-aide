package chat_test

import (
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

// PRD-1 AC8: a chat read makes the user present, and presence expires one TTL
// after the last read — per user, so one screen does not vouch for another.
func TestPresence_TouchExpiresAfterTTL(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	p := chat.NewPresenceWithClock(func() time.Time { return now })

	if s := p.State(1); s.Present || s.LastSeenAt != nil {
		t.Fatalf("never-seen user = %+v, want absent with no last_seen_at", s)
	}
	if s := p.State(1); s.TTLSeconds != 45 {
		t.Fatalf("ttl_seconds = %d, want 45", s.TTLSeconds)
	}

	p.Touch(1)
	s := p.State(1)
	if !s.Present || s.LastSeenAt == nil || *s.LastSeenAt != now.Unix() {
		t.Fatalf("just read = %+v, want present at %d", s, now.Unix())
	}
	if other := p.State(2); other.Present {
		t.Fatalf("user 2 = %+v, want absent", other)
	}

	// One idle poll interval later the screen is still present; two idle
	// intervals (a dropped poll) too; past the TTL it is not.
	for _, elapsed := range []time.Duration{15 * time.Second, 30 * time.Second, 44 * time.Second} {
		now = now.Add(elapsed)
		if s := p.State(1); !s.Present {
			t.Errorf("after %s = absent, want present", elapsed)
		}
		now = now.Add(-elapsed)
	}

	now = now.Add(chat.PresenceTTL)
	s = p.State(1)
	if s.Present {
		t.Fatalf("after TTL = %+v, want absent", s)
	}
	if s.LastSeenAt == nil {
		t.Fatalf("after TTL: last_seen_at dropped, want the last read kept")
	}

	// A later read brings it back without any reset in between.
	p.Touch(1)
	if s := p.State(1); !s.Present || *s.LastSeenAt != now.Unix() {
		t.Fatalf("read after expiry = %+v, want present at %d", s, now.Unix())
	}
}

// A nil tracker (a server built without one) answers absent rather than
// panicking: presence is not load-bearing for the chat screen itself.
func TestPresence_NilIsAbsent(t *testing.T) {
	var p *chat.Presence
	p.Touch(1)
	if s := p.State(1); s.Present || s.LastSeenAt != nil || s.TTLSeconds != 45 {
		t.Fatalf("nil tracker = %+v", s)
	}
}
