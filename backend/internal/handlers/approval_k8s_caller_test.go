package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/k8sauth"
)

type fakeReviewer struct {
	tokens      map[string]string
	unavailable bool
}

func (f *fakeReviewer) Review(_ context.Context, token string) (string, error) {
	if f.unavailable {
		return "", k8sauth.ErrUnavailable
	}
	sa, ok := f.tokens[token]
	if !ok {
		return "", k8sauth.ErrUnauthenticated
	}
	return sa, nil
}

func (e *approvalEnv) registerSA(sub, name, sa string) *httptest.ResponseRecorder {
	e.t.Helper()
	return e.do(http.MethodPost, "/api/approval-keys", "oidc-"+sub, map[string]any{"name": name, "kind": "k8s_sa", "service_account": sa})
}

func TestApprovalServiceAccountCallerRegistrationAndUse(t *testing.T) {
	env := newApprovalEnv(t)
	env.reviewer.tokens["sa-token-mcp"] = "homelab-k3s-mcp/homelab-k3s-mcp"
	env.reviewer.tokens["sa-token-other"] = "other/runner"
	body := map[string]any{"externalId": "secret-get-1", "context": "get Secret prod/db"}

	rec := env.do(http.MethodPost, "/api/external/approvals", "sa-token-mcp", body)
	expectCode(t, rec, http.StatusUnauthorized, "unregistered service account")

	rec = env.registerSA("alice", "homelab-k3s-mcp", "homelab-k3s-mcp/homelab-k3s-mcp")
	expectCode(t, rec, http.StatusCreated, "register service account")
	registered := decodeApproval[issuedKeyBody](t, rec)
	if registered.Key != "" || registered.Kind != "k8s_sa" || registered.ServiceAccount == nil || *registered.ServiceAccount != "homelab-k3s-mcp/homelab-k3s-mcp" || registered.KeyPrefix != "" {
		t.Fatalf("registration = %+v, want k8s_sa row with no secret", registered)
	}

	created := env.create("sa-token-mcp", body)
	if created.Status != "PENDING" {
		t.Fatalf("status = %q, want PENDING", created.Status)
	}
	if got := env.poll("sa-token-mcp", created.ID); got.Status != "PENDING" {
		t.Fatalf("poll status = %q", got.Status)
	}
	pending := env.pending("alice")
	if len(pending) != 1 || pending[0].KeyName != "homelab-k3s-mcp" {
		t.Fatalf("alice pending = %+v, want the SA request under its caller name", pending)
	}
	if keys := env.keys("alice"); len(keys) != 1 || keys[0].LastUsedAt == nil || keys[0].Kind != "k8s_sa" {
		t.Fatalf("alice keys = %+v, want the SA row with last_used_at", keys)
	}
	if rec := env.do(http.MethodGet, "/api/external/approvals/"+created.ID, "sa-token-other", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("other SA poll = %d, want 401", rec.Code)
	}

	expectCode(t, env.registerSA("bob", "mcp", "homelab-k3s-mcp/homelab-k3s-mcp"), http.StatusConflict, "same SA, other user")
	expectCode(t, env.registerSA("alice", "mcp-2", "homelab-k3s-mcp/homelab-k3s-mcp"), http.StatusConflict, "same SA, other name")
	expectCode(t, env.registerSA("alice", "homelab-k3s-mcp", "other/runner"), http.StatusConflict, "duplicate caller name")
	for _, bad := range []string{"", "homelab-k3s-mcp", "/name", "ns/", "NS/name", "ns/-name", "ns/name/extra", "ns/na_me"} {
		expectCode(t, env.registerSA("alice", "bad", bad), http.StatusBadRequest, "invalid service account "+bad)
	}
	expectCode(t, env.do(http.MethodPost, "/api/approval-keys", "oidc-alice", map[string]any{"name": "x", "kind": "token"}), http.StatusBadRequest, "unknown kind")
	expectCode(t, env.do(http.MethodPost, "/api/approval-keys", "oidc-alice", map[string]any{"name": "x", "service_account": "ns/name"}), http.StatusBadRequest, "service_account on a key")

	expectCode(t, env.do(http.MethodPost, "/api/approval-keys/"+strconv.FormatInt(registered.ID, 10)+"/revoke", "oidc-alice", nil), http.StatusOK, "revoke SA caller")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", "sa-token-mcp", map[string]any{"externalId": "secret-get-2", "context": "x"}), http.StatusUnauthorized, "create after revoke")
	expectCode(t, env.do(http.MethodGet, "/api/external/approvals/"+created.ID, "sa-token-mcp", nil), http.StatusUnauthorized, "poll after revoke")
	if pending := env.pending("alice"); len(pending) != 1 || !pending[0].KeyRevoked || pending[0].Status != "PENDING" {
		t.Fatalf("pending after revoke = %+v, want the request kept with the revoked warning", pending)
	}
	expectCode(t, env.registerSA("bob", "mcp", "homelab-k3s-mcp/homelab-k3s-mcp"), http.StatusCreated, "SA free again after revoke")
}

func TestApprovalServiceAccountCallerVerificationFailures(t *testing.T) {
	env := newApprovalEnv(t)
	env.reviewer.tokens["sa-token"] = "ci/runner"
	expectCode(t, env.registerSA("alice", "runner", "ci/runner"), http.StatusCreated, "register")
	body := map[string]any{"externalId": "e-1", "context": "x"}

	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", "not-a-reviewed-token", body), http.StatusUnauthorized, "unauthenticated token")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", "pak_not-issued", body), http.StatusUnauthorized, "unknown caller key bypasses review")

	env.reviewer.unavailable = true
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", "sa-token", body), http.StatusServiceUnavailable, "apiserver unavailable")
	issued := env.issue("alice", "infra-ci")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", issued.Key, body), http.StatusCreated, "caller keys do not depend on the apiserver")
}

func TestApprovalExternalResponsesCarryDecisionMode(t *testing.T) {
	env := newApprovalEnv(t)
	key := env.issue("alice", "infra-ci").Key

	manual := env.create(key, map[string]any{"externalId": "m", "context": "x"})
	if manual.DecisionMode != nil {
		t.Fatalf("pending decisionMode = %v, want null", *manual.DecisionMode)
	}
	expectCode(t, env.do(http.MethodPost, "/api/approvals/"+manual.ID+"/approve", "oidc-alice", nil), http.StatusOK, "approve")
	if got := env.poll(key, manual.ID); got.DecisionMode == nil || *got.DecisionMode != "manual" {
		t.Fatalf("approved decisionMode = %v, want manual", got.DecisionMode)
	}

	env.setMode("alice", "AUTO_APPROVE")
	auto := env.create(key, map[string]any{"externalId": "a", "context": "x"})
	if auto.Status != "APPROVED" || auto.DecisionMode == nil || *auto.DecisionMode != "auto" {
		t.Fatalf("auto create = %+v, want APPROVED/auto", auto)
	}

	env.setMode("alice", "NONE")
	expiring := env.create(key, map[string]any{"externalId": "e", "context": "x", "timeoutSeconds": 5})
	env.advance(6 * time.Second)
	if got := env.poll(key, expiring.ID); got.Status != "EXPIRED" || got.DecisionMode == nil || *got.DecisionMode != "expired" {
		t.Fatalf("expired poll = %+v, want EXPIRED/expired", got)
	}
}
