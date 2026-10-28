package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

type stubSessionNames struct {
	mu        sync.Mutex
	name      string
	err       error
	asked     []string
	deadlines []time.Duration
}

func (s *stubSessionNames) SessionName(ctx context.Context, id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.asked = append(s.asked, id)
	if d, ok := ctx.Deadline(); ok {
		s.deadlines = append(s.deadlines, time.Until(d))
	} else {
		s.deadlines = append(s.deadlines, -1)
	}
	return s.name, s.err
}

func (e *approvalEnv) appApproval(sub, id string) appApprovalBody {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/approvals/"+id, "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "get approval")
	return decodeApproval[appApprovalBody](e.t, rec)
}

func TestApprovalCreateStoresTrimmedSessionAndReasonWithLookedUpName(t *testing.T) {
	platform := newFakePlatform()
	sid := platform.seed("배포 파이프라인 정리", "minimax-code", "2026-10-25T00:00:00Z")
	srv := httptest.NewServer(platform)
	t.Cleanup(srv.Close)
	env := newApprovalEnvWith(t, nil, sessions.NewClient(srv.URL))
	k := env.issue("a", "homelab-k3s-mcp")

	created := env.create(k.Key, map[string]any{
		"externalId": "r-1", "context": "delete pod x", "sessionId": "  " + sid + "\n", "reason": "  재시작이 필요해서 지운다.  ",
	})
	if n := platform.callCount("GET /api/v1/sessions/" + sid); n != 1 {
		t.Fatalf("session lookups: got %d want 1 (calls=%v)", n, platform.calls)
	}

	got := env.appApproval("a", created.ID)
	if got.SessionID != sid || got.SessionName != "배포 파이프라인 정리" || got.Reason != "재시작이 필요해서 지운다." {
		t.Fatalf("detail: got session_id=%q session_name=%q reason=%q", got.SessionID, got.SessionName, got.Reason)
	}
	if listed := env.pending("a"); len(listed) != 1 || listed[0].SessionName != "배포 파이프라인 정리" || listed[0].Reason != got.Reason {
		t.Fatalf("pending list: got %+v", listed)
	}

	platform.mu.Lock()
	delete(platform.sessions, sid)
	platform.mu.Unlock()
	if again := env.appApproval("a", created.ID); again.SessionName != "배포 파이프라인 정리" {
		t.Fatalf("name after session deletion: got %q, want the name stored at creation", again.SessionName)
	}
}

func TestApprovalCreateWithoutSessionOrReasonKeepsFieldsEmptyAndSkipsLookup(t *testing.T) {
	names := &stubSessionNames{name: "never"}
	env := newApprovalEnvWith(t, nil, names)
	k := env.issue("a", "infra-ci")

	plain := env.create(k.Key, map[string]any{"externalId": "r-1", "context": "c"})
	blank := env.create(k.Key, map[string]any{"externalId": "r-2", "context": "c", "sessionId": "   ", "reason": " \t "})
	if len(names.asked) != 0 {
		t.Fatalf("lookup without a session id: asked %v", names.asked)
	}
	for _, id := range []string{plain.ID, blank.ID} {
		rec := env.do(http.MethodGet, "/api/approvals/"+id, "oidc-a", nil)
		expectCode(t, rec, http.StatusOK, "get approval")
		for _, key := range []string{`"session_id":""`, `"session_name":""`, `"reason":""`} {
			if !strings.Contains(rec.Body.String(), key) {
				t.Fatalf("app JSON missing %s: %s", key, rec.Body.String())
			}
		}
	}
}

func TestApprovalCreateSurvivesSessionLookupFailure(t *testing.T) {
	for _, tc := range []struct {
		name  string
		names *stubSessionNames
	}{
		{"error", &stubSessionNames{err: errors.New("control plane down")}},
		{"blank name", &stubSessionNames{name: "   "}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := newApprovalEnvWith(t, nil, tc.names)
			k := env.issue("a", "homelab-k3s-mcp")
			created := env.create(k.Key, map[string]any{"externalId": "r-1", "context": "c", "sessionId": "s-gone", "reason": "이유"})
			got := env.appApproval("a", created.ID)
			if got.SessionID != "s-gone" || got.SessionName != "" || got.Reason != "이유" {
				t.Fatalf("detail: got %+v", got)
			}
			if len(tc.names.deadlines) != 1 || tc.names.deadlines[0] <= 0 || tc.names.deadlines[0] > 2*time.Second {
				t.Fatalf("lookup deadline: got %v, want one bounded lookup of at most 2s", tc.names.deadlines)
			}
		})
	}
}

func TestApprovalCreateSurvivesUnknownAndUnreachableSession(t *testing.T) {
	platform := newFakePlatform()
	srv := httptest.NewServer(platform)
	t.Cleanup(srv.Close)
	down := httptest.NewServer(http.NotFoundHandler())
	downURL := down.URL
	down.Close()

	for name, client := range map[string]*sessions.Client{
		"404":         sessions.NewClient(srv.URL),
		"unreachable": sessions.NewClient(downURL),
	} {
		t.Run(name, func(t *testing.T) {
			env := newApprovalEnvWith(t, nil, client)
			k := env.issue("a", "homelab-k3s-mcp")
			created := env.create(k.Key, map[string]any{"externalId": "r-1", "context": "c", "sessionId": "s-missing"})
			if got := env.appApproval("a", created.ID); got.SessionID != "s-missing" || got.SessionName != "" {
				t.Fatalf("detail: got %+v", got)
			}
		})
	}
}

func TestApprovalCreateRejectsOverlongSessionOrReason(t *testing.T) {
	names := &stubSessionNames{name: "n"}
	env := newApprovalEnvWith(t, nil, names)
	k := env.issue("a", "homelab-k3s-mcp")

	env.create(k.Key, map[string]any{"externalId": "max", "context": "c", "sessionId": strings.Repeat("세", 128), "reason": strings.Repeat("이", 2000)})
	env.create(k.Key, map[string]any{"externalId": "trimmed", "context": "c", "sessionId": " " + strings.Repeat("s", 128) + " ", "reason": " " + strings.Repeat("r", 2000) + " "})
	for what, body := range map[string]map[string]any{
		"sessionId 129": {"externalId": "x1", "context": "c", "sessionId": strings.Repeat("세", 129)},
		"reason 2001":   {"externalId": "x2", "context": "c", "reason": strings.Repeat("이", 2001)},
		"sessionId num": {"externalId": "x3", "context": "c", "sessionId": 42},
		"reason object": {"externalId": "x4", "context": "c", "reason": map[string]any{"text": "x"}},
		"unknown field": {"externalId": "x5", "context": "c", "sessionId": "s", "reason": "r", "sessionName": "forged"},
	} {
		expectCode(t, env.do(http.MethodPost, "/api/external/approvals", k.Key, body), http.StatusBadRequest, what)
	}
	if got := env.pending("a"); len(got) != 2 {
		t.Fatalf("rejected bodies created rows: %d pending", len(got))
	}
	if len(names.asked) != 2 {
		t.Fatalf("lookups: got %v, want only the two accepted requests", names.asked)
	}
}
