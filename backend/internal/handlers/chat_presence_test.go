package handlers_test

import (
	"net/http"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/approvals"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

type presenceEnv struct {
	*approvalEnv
	presenceNow time.Time
}

// newPresenceEnv wires the user-facing chat read and the caller-facing
// presence read onto one router, so a test can poll as the app and then ask as
// homelab-k3s-mcp does.
func newPresenceEnv(t *testing.T) *presenceEnv {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "presence.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	base := &approvalEnv{t: t, conn: conn, reviewer: &fakeReviewer{tokens: map[string]string{}}, now: time.Unix(1_800_000_000, 0)}
	env := &presenceEnv{approvalEnv: base, presenceNow: time.Unix(1_800_000_000, 0)}
	store := approvals.NewWithClock(conn, base.clock)
	chatStore := chat.New(conn)
	presence := chat.NewPresenceWithClock(func() time.Time { return env.presenceNow })

	r := chi.NewRouter()
	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(subVerifier{}, conn))
		p.Post("/api/approval-keys", handlers.IssueApprovalKey(store))
		p.Post("/api/approval-keys/{id}/revoke", handlers.RevokeApprovalKey(store))
		p.Get("/api/chat", handlers.GetChat(chatStore, chat.NewEngine(chatStore, nil, nil), presence))
	})
	r.Group(func(x chi.Router) {
		x.Use(handlers.CallerKeyMiddleware(store, base.reviewer))
		x.Get("/api/external/chat-presence", handlers.GetExternalChatPresence(presence))
	})
	base.router = r
	return env
}

type presenceBody struct {
	Present    bool   `json:"present"`
	LastSeenAt *int64 `json:"last_seen_at"`
	TTLSeconds int    `json:"ttl_seconds"`
}

func (e *presenceEnv) openChat(sub string) {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/chat", "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "chat read by "+sub)
}

func (e *presenceEnv) ask(key string) presenceBody {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/external/chat-presence", key, nil)
	expectCode(e.t, rec, http.StatusOK, "read presence")
	return decodeApproval[presenceBody](e.t, rec)
}

// PRD-12 AC22: the caller learns the presence of the user who registered it —
// the chat screen's own poll is the signal, and it expires one TTL later.
func TestChatPresenceFollowsCallerOwnersChatReads(t *testing.T) {
	env := newPresenceEnv(t)
	aliceKey := env.issue("alice", "homelab-k3s-mcp")
	bobKey := env.issue("bob", "bob-ci")

	before := env.ask(aliceKey.Key)
	if before.Present || before.LastSeenAt != nil || before.TTLSeconds != 45 {
		t.Fatalf("before any read = %+v, want absent with no last_seen_at and ttl 45", before)
	}

	env.openChat("alice")
	got := env.ask(aliceKey.Key)
	if !got.Present || got.LastSeenAt == nil || *got.LastSeenAt != env.presenceNow.Unix() {
		t.Fatalf("after alice's read = %+v, want present at %d", got, env.presenceNow.Unix())
	}

	// bob's caller sees bob, who has not opened the chat screen.
	if other := env.ask(bobKey.Key); other.Present || other.LastSeenAt != nil {
		t.Fatalf("bob's caller = %+v, want absent — alice's read is not bob's", other)
	}

	// Still present across one dropped idle poll, absent past the TTL, and the
	// last read stays readable so a caller can tell "just left" from "never here".
	env.presenceNow = env.presenceNow.Add(30 * time.Second)
	if mid := env.ask(aliceKey.Key); !mid.Present {
		t.Fatalf("30s after the read = %+v, want present", mid)
	}
	seenAt := env.presenceNow.Add(-30 * time.Second).Unix()
	env.presenceNow = env.presenceNow.Add(16 * time.Second)
	after := env.ask(aliceKey.Key)
	if after.Present || after.LastSeenAt == nil || *after.LastSeenAt != seenAt {
		t.Fatalf("46s after the read = %+v, want absent with last_seen_at %d", after, seenAt)
	}

	// The screen coming back makes the user present again.
	env.openChat("alice")
	if back := env.ask(aliceKey.Key); !back.Present {
		t.Fatalf("after returning to the chat screen = %+v, want present", back)
	}
}

// The route sits behind the same caller middleware as every other
// /api/external route: no key, an unknown key, a revoked key and an OIDC
// token are all 401, and it never takes a user as a parameter.
func TestChatPresenceRequiresLiveCaller(t *testing.T) {
	env := newPresenceEnv(t)
	key := env.issue("alice", "homelab-k3s-mcp")
	env.openChat("alice")

	for _, c := range []struct{ name, bearer string }{
		{"no credential", ""},
		{"unknown key", "nope"},
		{"oidc token", "oidc-alice"},
	} {
		rec := env.do(http.MethodGet, "/api/external/chat-presence", c.bearer, nil)
		expectCode(t, rec, http.StatusUnauthorized, c.name)
	}

	expectCode(t, env.do(http.MethodPost, "/api/approval-keys/"+strconv.FormatInt(key.ID, 10)+"/revoke", "oidc-alice", nil), http.StatusOK, "revoke")
	expectCode(t, env.do(http.MethodGet, "/api/external/chat-presence", key.Key, nil), http.StatusUnauthorized, "revoked key")
}

// homelab-k3s-mcp reaches this route with its projected ServiceAccount token,
// the same credential as the notification route (PRD-12 AC15·AC16).
func TestChatPresenceAcceptsServiceAccountCaller(t *testing.T) {
	env := newPresenceEnv(t)
	env.reviewer.tokens["sa-token-mcp"] = "homelab-k3s-mcp/homelab-k3s-mcp"
	expectCode(t, env.registerSA("alice", "homelab-k3s-mcp", "homelab-k3s-mcp/homelab-k3s-mcp"), http.StatusCreated, "register SA")

	if got := env.ask("sa-token-mcp"); got.Present {
		t.Fatalf("before any read = %+v, want absent", got)
	}
	env.openChat("alice")
	if got := env.ask("sa-token-mcp"); !got.Present || got.LastSeenAt == nil {
		t.Fatalf("after alice's read = %+v, want present", got)
	}
}
