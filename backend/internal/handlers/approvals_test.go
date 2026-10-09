package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/approvals"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

type subVerifier struct{}

func (subVerifier) Verify(_ context.Context, raw string) (*oidc.IDToken, error) {
	sub, ok := strings.CutPrefix(raw, "oidc-")
	if !ok {
		return nil, errors.New("not an id token")
	}
	return &oidc.IDToken{Subject: sub}, nil
}

type approvalEnv struct {
	t      *testing.T
	conn   *sql.DB
	router http.Handler
	mu     sync.Mutex
	now    time.Time
}

func newApprovalEnv(t *testing.T) *approvalEnv {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "approvals.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	env := &approvalEnv{t: t, conn: conn, now: time.Unix(1_800_000_000, 0)}
	store := approvals.NewWithClock(conn, env.clock)
	r := chi.NewRouter()
	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(subVerifier{}, conn))
		p.Get("/api/approval-keys", handlers.ListApprovalKeys(store))
		p.Post("/api/approval-keys", handlers.IssueApprovalKey(store))
		p.Post("/api/approval-keys/{id}/revoke", handlers.RevokeApprovalKey(store))
		p.Get("/api/approvals", handlers.ListPendingApprovals(store))
		p.Get("/api/approvals/{id}", handlers.GetApproval(store))
		p.Post("/api/approvals/{id}/approve", handlers.DecideApproval(store, approvals.DecisionApprove))
		p.Post("/api/approvals/{id}/reject", handlers.DecideApproval(store, approvals.DecisionReject))
	})
	r.Group(func(x chi.Router) {
		x.Use(handlers.CallerKeyMiddleware(store))
		x.Post("/api/external/approvals", handlers.CreateExternalApproval(store))
		x.Get("/api/external/approvals/{id}", handlers.GetExternalApproval(store))
	})
	env.router = r
	return env
}

func (e *approvalEnv) clock() time.Time {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.now
}

func (e *approvalEnv) advance(d time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.now = e.now.Add(d)
}

func (e *approvalEnv) do(method, target, bearer string, body any) *httptest.ResponseRecorder {
	e.t.Helper()
	var reader *bytes.Reader
	switch b := body.(type) {
	case nil:
		reader = bytes.NewReader(nil)
	case string:
		reader = bytes.NewReader([]byte(b))
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			e.t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

func decodeApproval[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return v
}

func expectCode(t *testing.T, rec *httptest.ResponseRecorder, want int, what string) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("%s: got %d want %d (body=%s)", what, rec.Code, want, rec.Body.String())
	}
}

type issuedKeyBody struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Key        string `json:"key"`
	KeyPrefix  string `json:"key_prefix"`
	CreatedAt  int64  `json:"created_at"`
	LastUsedAt *int64 `json:"last_used_at"`
	RevokedAt  *int64 `json:"revoked_at"`
}

type externalBody struct {
	ID          string  `json:"id"`
	ExternalID  string  `json:"externalId"`
	Status      string  `json:"status"`
	CreatedAt   int64   `json:"createdAt"`
	ExpiresAt   *int64  `json:"expiresAt"`
	ProcessedAt *int64  `json:"processedAt"`
	UserID      *string `json:"userId"`
}

type appApprovalBody struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	ExternalID    string `json:"external_id"`
	Context       string `json:"context"`
	RequesterName string `json:"requester_name"`
	KeyName       string `json:"key_name"`
	KeyRevoked    bool   `json:"key_revoked"`
	DecisionMode  string `json:"decision_mode"`
	ProcessedAt   *int64 `json:"processed_at"`
}

func (e *approvalEnv) issue(sub, name string) issuedKeyBody {
	e.t.Helper()
	rec := e.do(http.MethodPost, "/api/approval-keys", "oidc-"+sub, map[string]any{"name": name})
	expectCode(e.t, rec, http.StatusCreated, "issue "+name)
	return decodeApproval[issuedKeyBody](e.t, rec)
}

func (e *approvalEnv) keys(sub string) []issuedKeyBody {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/approval-keys", "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "list keys")
	return decodeApproval[struct {
		Keys []issuedKeyBody `json:"keys"`
	}](e.t, rec).Keys
}

func (e *approvalEnv) create(key string, body map[string]any) externalBody {
	e.t.Helper()
	rec := e.do(http.MethodPost, "/api/external/approvals", key, body)
	expectCode(e.t, rec, http.StatusCreated, "create request")
	return decodeApproval[externalBody](e.t, rec)
}

func (e *approvalEnv) poll(key, id string) externalBody {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/external/approvals/"+id, key, nil)
	expectCode(e.t, rec, http.StatusOK, "poll request")
	return decodeApproval[externalBody](e.t, rec)
}

func (e *approvalEnv) pending(sub string) []appApprovalBody {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/approvals", "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "list pending")
	return decodeApproval[struct {
		Items []appApprovalBody `json:"items"`
	}](e.t, rec).Items
}

func TestApprovalKeyIssuedOnceAndHashedOnly(t *testing.T) {
	env := newApprovalEnv(t)

	k := env.issue("a", "infra-ci")
	if !strings.HasPrefix(k.Key, "pak_") || len(k.Key) < 40 {
		t.Fatalf("issued key %q: want pak_ prefix and a long random secret", k.Key)
	}
	if !strings.HasPrefix(k.Key, k.KeyPrefix) || len(k.KeyPrefix) >= len(k.Key) {
		t.Fatalf("key_prefix %q must be a strict prefix of the key", k.KeyPrefix)
	}

	listed := env.keys("a")
	if len(listed) != 1 || listed[0].Name != "infra-ci" || listed[0].Key != "" || listed[0].LastUsedAt != nil {
		t.Fatalf("key list: got %+v, want only infra-ci without raw key and never used", listed)
	}

	var stored int
	if err := env.conn.QueryRow(`SELECT COUNT(*) FROM approval_caller_keys WHERE key_hash = ? OR key_prefix = ?`, k.Key, k.Key).Scan(&stored); err != nil {
		t.Fatalf("count raw: %v", err)
	}
	if stored != 0 {
		t.Fatalf("raw key found in the database")
	}

	expectCode(t, env.do(http.MethodPost, "/api/approval-keys", "oidc-a", map[string]any{"name": "infra-ci"}), http.StatusConflict, "duplicate name")
	expectCode(t, env.do(http.MethodPost, "/api/approval-keys", "oidc-a", map[string]any{"name": "   "}), http.StatusBadRequest, "blank name")
	if got := env.keys("b"); len(got) != 0 {
		t.Fatalf("user b sees keys: %+v", got)
	}
	env.issue("b", "infra-ci")
}

func TestApprovalCreateAndPoll(t *testing.T) {
	env := newApprovalEnv(t)
	k := env.issue("a", "infra-ci")

	created := env.create(k.Key, map[string]any{
		"externalId": "deploy-1", "context": "prod 배포 v1.42", "requesterName": "deploy-bot", "timeoutSeconds": 600,
	})
	if created.Status != "PENDING" || created.ID == "" || created.ExpiresAt == nil || *created.ExpiresAt != created.CreatedAt+600 {
		t.Fatalf("create: got %+v, want PENDING with expiresAt = createdAt+600", created)
	}

	rec := env.do(http.MethodGet, "/api/external/approvals/"+created.ID, k.Key, nil)
	expectCode(t, rec, http.StatusOK, "poll")
	for _, leak := range []string{"oidc", "sub", "email", "user"} {
		if strings.Contains(rec.Body.String(), leak) {
			t.Fatalf("poll response leaks %q: %s", leak, rec.Body.String())
		}
	}
	got := decodeApproval[externalBody](t, rec)
	if got.Status != "PENDING" || got.ExternalID != "deploy-1" || got.ProcessedAt != nil {
		t.Fatalf("poll: got %+v, want PENDING deploy-1 without processedAt", got)
	}

	listed := env.keys("a")
	if listed[0].LastUsedAt == nil || *listed[0].LastUsedAt != created.CreatedAt {
		t.Fatalf("last_used_at: got %v, want %d", listed[0].LastUsedAt, created.CreatedAt)
	}
}

func TestApprovalCreateRejectsMalformedAndUnauthenticated(t *testing.T) {
	env := newApprovalEnv(t)
	infra := env.issue("a", "infra-ci")
	staging := env.issue("a", "staging-ci")
	env.create(infra.Key, map[string]any{"externalId": "deploy-1", "context": "c"})

	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", infra.Key, map[string]any{"externalId": "deploy-2"}), http.StatusBadRequest, "missing context")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", infra.Key, map[string]any{"context": "c"}), http.StatusBadRequest, "missing externalId")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", infra.Key, map[string]any{"externalId": "deploy-2", "context": "c", "userId": "ckabc"}), http.StatusBadRequest, "legacy userId")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", infra.Key, map[string]any{"externalId": "deploy-2", "context": "c", "timeoutSeconds": 0}), http.StatusBadRequest, "zero timeout")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", infra.Key, "{"), http.StatusBadRequest, "broken json")
	if got := env.pending("a"); len(got) != 1 {
		t.Fatalf("rejected bodies created rows: %+v", got)
	}

	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", infra.Key, map[string]any{"externalId": "deploy-1", "context": "c"}), http.StatusConflict, "duplicate externalId on same key")
	env.create(staging.Key, map[string]any{"externalId": "deploy-1", "context": "c"})

	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", "", map[string]any{"externalId": "x", "context": "c"}), http.StatusUnauthorized, "no header")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", "pak_wrong", map[string]any{"externalId": "x", "context": "c"}), http.StatusUnauthorized, "wrong key")
	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", "oidc-a", map[string]any{"externalId": "x", "context": "c"}), http.StatusUnauthorized, "oidc token on caller route")
	expectCode(t, env.do(http.MethodGet, "/api/approvals", infra.Key, nil), http.StatusUnauthorized, "caller key on app route")
	expectCode(t, env.do(http.MethodGet, "/api/approval-keys", infra.Key, nil), http.StatusUnauthorized, "caller key on key route")
}

func TestApprovalOwnershipIsPrivate(t *testing.T) {
	env := newApprovalEnv(t)
	a := env.issue("a", "infra-ci")
	b := env.issue("b", "b-ci")
	req := env.create(a.Key, map[string]any{"externalId": "r", "context": "c"})

	if got := env.pending("a"); len(got) != 1 || got[0].ID != req.ID || got[0].KeyName != "infra-ci" {
		t.Fatalf("owner pending: %+v", got)
	}
	if got := env.pending("b"); len(got) != 0 {
		t.Fatalf("other user sees requests: %+v", got)
	}
	expectCode(t, env.do(http.MethodGet, "/api/approvals/"+req.ID, "oidc-b", nil), http.StatusNotFound, "other user detail")
	expectCode(t, env.do(http.MethodPost, "/api/approvals/"+req.ID+"/approve", "oidc-b", nil), http.StatusNotFound, "other user approve")
	expectCode(t, env.do(http.MethodPost, "/api/approvals/"+req.ID+"/reject", "oidc-b", nil), http.StatusNotFound, "other user reject")
	expectCode(t, env.do(http.MethodGet, "/api/approvals", "", nil), http.StatusUnauthorized, "anonymous list")
	expectCode(t, env.do(http.MethodGet, "/api/external/approvals/"+req.ID, b.Key, nil), http.StatusNotFound, "other key poll")
	if got := env.poll(a.Key, req.ID); got.Status != "PENDING" {
		t.Fatalf("status changed by foreign attempts: %+v", got)
	}
}

func TestApprovalDecisionFirstWins(t *testing.T) {
	env := newApprovalEnv(t)
	k := env.issue("a", "infra-ci")
	req := env.create(k.Key, map[string]any{"externalId": "r", "context": "c", "requesterName": "self-declared"})

	detail := decodeApproval[appApprovalBody](t, env.do(http.MethodGet, "/api/approvals/"+req.ID, "oidc-a", nil))
	if detail.KeyName != "infra-ci" || detail.RequesterName != "self-declared" || detail.Status != "PENDING" {
		t.Fatalf("detail: %+v", detail)
	}

	env.advance(30 * time.Second)
	rec := env.do(http.MethodPost, "/api/approvals/"+req.ID+"/approve", "oidc-a", nil)
	expectCode(t, rec, http.StatusOK, "approve")
	approved := decodeApproval[appApprovalBody](t, rec)
	if approved.Status != "APPROVED" || approved.DecisionMode != "manual" || approved.ProcessedAt == nil {
		t.Fatalf("approve: %+v", approved)
	}

	rec = env.do(http.MethodPost, "/api/approvals/"+req.ID+"/reject", "oidc-a", nil)
	expectCode(t, rec, http.StatusConflict, "reject after approve")
	conflict := decodeApproval[struct {
		Error    string          `json:"error"`
		Approval appApprovalBody `json:"approval"`
	}](t, rec)
	if conflict.Error != "already processed" || conflict.Approval.Status != "APPROVED" {
		t.Fatalf("conflict body: %+v", conflict)
	}
	expectCode(t, env.do(http.MethodPost, "/api/approvals/"+req.ID+"/approve", "oidc-a", nil), http.StatusConflict, "approve twice")

	polled := env.poll(k.Key, req.ID)
	if polled.Status != "APPROVED" || polled.ProcessedAt == nil || *polled.ProcessedAt != req.CreatedAt+30 {
		t.Fatalf("poll after approve: %+v", polled)
	}
	if got := env.pending("a"); len(got) != 0 {
		t.Fatalf("processed request still pending: %+v", got)
	}
}

func TestApprovalConcurrentDecisionsOnlyOneSucceeds(t *testing.T) {
	env := newApprovalEnv(t)
	env.conn.SetMaxOpenConns(1)
	k := env.issue("a", "infra-ci")
	req := env.create(k.Key, map[string]any{"externalId": "r", "context": "c"})

	codes := make([]int, 8)
	var wg sync.WaitGroup
	for i := range codes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			action := "approve"
			if i%2 == 1 {
				action = "reject"
			}
			rec := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/approvals/"+req.ID+"/"+action, nil)
			r.Header.Set("Authorization", "Bearer oidc-a")
			env.router.ServeHTTP(rec, r)
			codes[i] = rec.Code
		}(i)
	}
	wg.Wait()
	ok := 0
	for _, c := range codes {
		switch c {
		case http.StatusOK:
			ok++
		case http.StatusConflict:
		default:
			t.Fatalf("unexpected status %d in %v", c, codes)
		}
	}
	if ok != 1 {
		t.Fatalf("successful decisions: got %d want 1 (%v)", ok, codes)
	}
}

func TestApprovalExpiryIsConsistentEverywhere(t *testing.T) {
	env := newApprovalEnv(t)
	k := env.issue("a", "infra-ci")
	expiring := env.create(k.Key, map[string]any{"externalId": "e", "context": "c", "timeoutSeconds": 5})
	forever := env.create(k.Key, map[string]any{"externalId": "n", "context": "c"})
	if forever.ExpiresAt != nil {
		t.Fatalf("request without timeoutSeconds has expiresAt: %+v", forever)
	}

	env.advance(6 * time.Second)

	pending := env.pending("a")
	if len(pending) != 1 || pending[0].ID != forever.ID {
		t.Fatalf("pending after expiry: got %+v, want only the request without timeout", pending)
	}

	rec := env.do(http.MethodPost, "/api/approvals/"+expiring.ID+"/approve", "oidc-a", nil)
	expectCode(t, rec, http.StatusConflict, "approve expired")
	if body := decodeApproval[struct {
		Error    string          `json:"error"`
		Approval appApprovalBody `json:"approval"`
	}](t, rec); body.Error != "expired" || body.Approval.Status != "EXPIRED" {
		t.Fatalf("approve expired body: %+v", body)
	}
	expectCode(t, env.do(http.MethodPost, "/api/approvals/"+expiring.ID+"/reject", "oidc-a", nil), http.StatusConflict, "reject expired")

	detail := decodeApproval[appApprovalBody](t, env.do(http.MethodGet, "/api/approvals/"+expiring.ID, "oidc-a", nil))
	if detail.Status != "EXPIRED" || detail.DecisionMode != "expired" {
		t.Fatalf("expired detail: %+v", detail)
	}
	if got := env.poll(k.Key, expiring.ID); got.Status != "EXPIRED" || got.ProcessedAt == nil || *got.ProcessedAt != *expiring.ExpiresAt {
		t.Fatalf("expired poll: %+v", got)
	}

	env.advance(30 * 24 * time.Hour)
	if got := env.poll(k.Key, forever.ID); got.Status != "PENDING" {
		t.Fatalf("request without timeout expired: %+v", got)
	}
	var stored string
	if err := env.conn.QueryRow(`SELECT status FROM approval_requests WHERE id = ?`, expiring.ID).Scan(&stored); err != nil {
		t.Fatalf("select status: %v", err)
	}
	if stored != "PENDING" {
		t.Fatalf("expiry wrote a status change: %q", stored)
	}
}

func TestApprovalKeyRevocation(t *testing.T) {
	env := newApprovalEnv(t)
	infra := env.issue("a", "infra-ci")
	staging := env.issue("a", "staging-ci")
	req := env.create(infra.Key, map[string]any{"externalId": "r", "context": "c"})

	expectCode(t, env.do(http.MethodPost, "/api/approval-keys/"+strconv.FormatInt(infra.ID, 10)+"/revoke", "oidc-b", nil), http.StatusNotFound, "revoke foreign key")
	rec := env.do(http.MethodPost, "/api/approval-keys/"+strconv.FormatInt(infra.ID, 10)+"/revoke", "oidc-a", nil)
	expectCode(t, rec, http.StatusOK, "revoke")
	if got := decodeApproval[issuedKeyBody](t, rec); got.RevokedAt == nil {
		t.Fatalf("revoke response without revoked_at: %+v", got)
	}
	expectCode(t, env.do(http.MethodPost, "/api/approval-keys/"+strconv.FormatInt(infra.ID, 10)+"/revoke", "oidc-a", nil), http.StatusConflict, "revoke twice")

	expectCode(t, env.do(http.MethodPost, "/api/external/approvals", infra.Key, map[string]any{"externalId": "x", "context": "c"}), http.StatusUnauthorized, "create with revoked key")
	expectCode(t, env.do(http.MethodGet, "/api/external/approvals/"+req.ID, infra.Key, nil), http.StatusUnauthorized, "poll with revoked key")
	env.create(staging.Key, map[string]any{"externalId": "x", "context": "c"})

	pending := env.pending("a")
	var found *appApprovalBody
	for i := range pending {
		if pending[i].ID == req.ID {
			found = &pending[i]
		}
	}
	if found == nil || found.Status != "PENDING" || !found.KeyRevoked || found.KeyName != "infra-ci" {
		t.Fatalf("request from revoked key: got %+v, want PENDING with key_revoked and key name", found)
	}

	keys := env.keys("a")
	if len(keys) != 2 {
		t.Fatalf("revoked key dropped from the list: %+v", keys)
	}
	reissued := env.issue("a", "infra-ci")
	if reissued.ID == infra.ID {
		t.Fatalf("reissue reused the revoked key row")
	}

	rec = env.do(http.MethodPost, "/api/approvals/"+req.ID+"/reject", "oidc-a", nil)
	expectCode(t, rec, http.StatusOK, "reject request from revoked key")
	if got := decodeApproval[appApprovalBody](t, rec); got.KeyName != "infra-ci" || got.Status != "REJECTED" {
		t.Fatalf("rejected request lost its key name: %+v", got)
	}
}
