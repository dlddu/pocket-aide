package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/approvals"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/commitstatus"
	"github.com/dlddu/pocket-aide/backend/internal/commitstatuspush"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

type fakeStatusWriter struct {
	mu     sync.Mutex
	writes []commitstatus.StatusWrite
	err    error
	delay  time.Duration
}

func (w *fakeStatusWriter) WriteStatus(_ context.Context, sw commitstatus.StatusWrite) (int64, error) {
	time.Sleep(w.delay)
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writes = append(w.writes, sw)
	if w.err != nil {
		return 0, w.err
	}
	return 9000 + int64(len(w.writes)), nil
}

func (w *fakeStatusWriter) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.writes)
}

func (w *fakeStatusWriter) setErr(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.err = err
}

type recordingStatusPusher struct {
	mu   sync.Mutex
	reqs []commitstatuspush.Request
}

func (p *recordingStatusPusher) Notify(req commitstatuspush.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reqs = append(p.reqs, req)
}

func (p *recordingStatusPusher) all() []commitstatuspush.Request {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]commitstatuspush.Request(nil), p.reqs...)
}

type commitStatusBody struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	Owner          string `json:"owner"`
	Repo           string `json:"repo"`
	SHA            string `json:"sha"`
	State          string `json:"state"`
	Context        string `json:"context"`
	Description    string `json:"description"`
	TargetURL      string `json:"target_url"`
	Requester      string `json:"requester"`
	Reason         string `json:"reason"`
	SessionID      string `json:"session_id"`
	KeyName        string `json:"key_name"`
	GitHubStatusID *int64 `json:"github_status_id"`
	LastError      string `json:"last_error"`
	CreatedAt      int64  `json:"created_at"`
	ProcessedAt    *int64 `json:"processed_at"`
}

type commitStatusList struct {
	Items   []commitStatusBody `json:"items"`
	History []commitStatusBody `json:"history"`
}

func newCommitStatusEnv(t *testing.T, writer commitstatus.Writer, push handlers.CommitStatusPusher) *approvalEnv {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "commit-status.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	env := &approvalEnv{t: t, conn: conn, reviewer: &fakeReviewer{tokens: map[string]string{}}, now: time.Unix(1_800_000_000, 0)}
	keys := approvals.NewWithClock(conn, env.clock)
	store := commitstatus.NewWithClock(conn, env.clock)
	decider := commitstatus.NewDecider(store, writer)
	r := chi.NewRouter()
	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(subVerifier{}, conn))
		p.Post("/api/approval-keys", handlers.IssueApprovalKey(keys))
		p.Post("/api/approval-keys/{id}/revoke", handlers.RevokeApprovalKey(keys))
		p.Get("/api/commit-status-requests", handlers.ListCommitStatusRequests(store))
		p.Get("/api/commit-status-requests/{id}", handlers.GetCommitStatusRequest(store))
		p.Post("/api/commit-status-requests/{id}/approve", handlers.ApproveCommitStatusRequest(decider))
		p.Post("/api/commit-status-requests/{id}/reject", handlers.RejectCommitStatusRequest(decider))
	})
	r.Group(func(x chi.Router) {
		x.Use(handlers.CallerKeyMiddleware(keys, env.reviewer))
		x.Post("/api/external/commit-status-requests", handlers.CreateExternalCommitStatusRequest(store, push, nil))
	})
	env.router = r
	return env
}

const testSHA = "0123456789abcdef0123456789abcdef01234567"

func validStatusRequest() map[string]any {
	return map[string]any{
		"owner":       "dlddu",
		"repo":        "pocket-aide",
		"sha":         testSHA,
		"state":       "success",
		"context":     "review/manual-approval",
		"description": "수동 승인",
		"targetUrl":   "https://github.com/dlddu/pocket-aide/pull/1",
		"requester":   "Claude Code 세션 — PR #1",
		"reason":      "PR 검토가 끝났다.",
	}
}

func (e *approvalEnv) fileStatus(key string, body map[string]any) string {
	e.t.Helper()
	rec := e.do(http.MethodPost, "/api/external/commit-status-requests", key, body)
	expectCode(e.t, rec, http.StatusCreated, "file commit status request")
	got := decodeApproval[struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		CreatedAt int64  `json:"createdAt"`
	}](e.t, rec)
	if got.ID == "" || got.Status != "PENDING" || got.CreatedAt != e.clock().Unix() {
		e.t.Fatalf("create answer %+v: want an id, PENDING and the creation time", got)
	}
	return got.ID
}

func (e *approvalEnv) statusList(sub string) commitStatusList {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/commit-status-requests", "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "list commit status requests")
	return decodeApproval[commitStatusList](e.t, rec)
}

func TestCommitStatusRequestIsStoredAndReturnsAtOnce(t *testing.T) {
	writer := &fakeStatusWriter{}
	push := &recordingStatusPusher{}
	env := newCommitStatusEnv(t, writer, push)
	k := env.issue("a", "homelab-k3s-mcp")

	id := env.fileStatus(k.Key, validStatusRequest())

	if writer.count() != 0 {
		t.Fatalf("filing a request wrote %d statuses, want 0 before approval", writer.count())
	}
	list := env.statusList("a")
	if len(list.Items) != 1 || len(list.History) != 0 {
		t.Fatalf("owner list %+v: want one pending, no history", list)
	}
	got := list.Items[0]
	if got.ID != id || got.Status != "PENDING" || got.Owner != "dlddu" || got.Repo != "pocket-aide" || got.SHA != testSHA ||
		got.State != "success" || got.Context != "review/manual-approval" || got.Description != "수동 승인" ||
		got.TargetURL != "https://github.com/dlddu/pocket-aide/pull/1" || got.Requester != "Claude Code 세션 — PR #1" ||
		got.Reason != "PR 검토가 끝났다." || got.KeyName != "homelab-k3s-mcp" || got.GitHubStatusID != nil || got.ProcessedAt != nil {
		t.Fatalf("stored request %+v does not carry what was filed", got)
	}
	if other := env.statusList("b"); len(other.Items) != 0 {
		t.Fatalf("another user sees %+v", other.Items)
	}
	expectCode(t, env.do(http.MethodGet, "/api/commit-status-requests/"+id, "oidc-b", nil), http.StatusNotFound, "other user's detail")
	detail := decodeApproval[commitStatusBody](t, env.do(http.MethodGet, "/api/commit-status-requests/"+id, "oidc-a", nil))
	if detail.ID != id || detail.Context != "review/manual-approval" {
		t.Fatalf("detail %+v", detail)
	}

	pushed := push.all()
	if len(pushed) != 1 || pushed[0].RequestID != id || pushed[0].KeyName != "homelab-k3s-mcp" || pushed[0].UserID != env.userID("a") ||
		pushed[0].Context != "review/manual-approval" || pushed[0].State != "success" || pushed[0].Requester != "Claude Code 세션 — PR #1" {
		t.Fatalf("push %+v: want one push to the owner naming the caller and the status", pushed)
	}
}

func TestCommitStatusRequestRefusesWhatGitHubWouldRefuse(t *testing.T) {
	writer := &fakeStatusWriter{}
	push := &recordingStatusPusher{}
	env := newCommitStatusEnv(t, writer, push)
	k := env.issue("a", "homelab-k3s-mcp")

	cases := map[string]func(map[string]any){
		"missing owner":     func(b map[string]any) { delete(b, "owner") },
		"owner with slash":  func(b map[string]any) { b["owner"] = "dlddu/x" },
		"repo with owner":   func(b map[string]any) { b["repo"] = "dlddu/pocket-aide" },
		"short sha":         func(b map[string]any) { b["sha"] = "0123456" },
		"branch as sha":     func(b map[string]any) { b["sha"] = "main" },
		"unknown state":     func(b map[string]any) { b["state"] = "approved" },
		"missing context":   func(b map[string]any) { delete(b, "context") },
		"long context":      func(b map[string]any) { b["context"] = strings.Repeat("c", 256) },
		"long description":  func(b map[string]any) { b["description"] = strings.Repeat("가", 141) },
		"relative url":      func(b map[string]any) { b["targetUrl"] = "/pull/1" },
		"ftp url":           func(b map[string]any) { b["targetUrl"] = "ftp://example.com/x" },
		"missing requester": func(b map[string]any) { delete(b, "requester") },
		"blank requester":   func(b map[string]any) { b["requester"] = "  " },
		"long requester":    func(b map[string]any) { b["requester"] = strings.Repeat("r", 201) },
		"long reason":       func(b map[string]any) { b["reason"] = strings.Repeat("r", 2001) },
		"unknown field":     func(b map[string]any) { b["token"] = "x" },
	}
	for name, mutate := range cases {
		body := validStatusRequest()
		mutate(body)
		expectCode(t, env.do(http.MethodPost, "/api/external/commit-status-requests", k.Key, body), http.StatusBadRequest, name)
	}
	expectCode(t, env.do(http.MethodPost, "/api/external/commit-status-requests", "", validStatusRequest()), http.StatusUnauthorized, "no caller")
	expectCode(t, env.do(http.MethodPost, "/api/external/commit-status-requests", "pak_nope", validStatusRequest()), http.StatusUnauthorized, "unknown caller")

	if list := env.statusList("a"); len(list.Items) != 0 {
		t.Fatalf("refused requests were stored: %+v", list.Items)
	}
	if len(push.all()) != 0 {
		t.Fatalf("refused requests pushed: %+v", push.all())
	}
}

func TestCommitStatusApprovalWritesExactlyWhatWasFiled(t *testing.T) {
	writer := &fakeStatusWriter{}
	env := newCommitStatusEnv(t, writer, nil)
	k := env.issue("a", "homelab-k3s-mcp")
	id := env.fileStatus(k.Key, validStatusRequest())

	expectCode(t, env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-b", nil), http.StatusNotFound, "other user approves")
	if writer.count() != 0 {
		t.Fatalf("another user's approval wrote a status")
	}

	env.advance(5 * time.Second)
	rec := env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-a", nil)
	expectCode(t, rec, http.StatusOK, "approve")
	got := decodeApproval[commitStatusBody](t, rec)
	if got.Status != "APPROVED" || got.GitHubStatusID == nil || *got.GitHubStatusID != 9001 || got.ProcessedAt == nil || *got.ProcessedAt != env.clock().Unix() {
		t.Fatalf("approved %+v: want APPROVED with GitHub's status id and the decision time", got)
	}
	want := commitstatus.StatusWrite{
		Owner: "dlddu", Repo: "pocket-aide", SHA: testSHA, State: "success", Context: "review/manual-approval",
		Description: "수동 승인", TargetURL: "https://github.com/dlddu/pocket-aide/pull/1",
	}
	if writer.count() != 1 || writer.writes[0] != want {
		t.Fatalf("writes %+v, want exactly %+v", writer.writes, want)
	}

	rec = env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-a", nil)
	expectCode(t, rec, http.StatusConflict, "approve twice")
	expectCode(t, env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/reject", "oidc-a", nil), http.StatusConflict, "reject after approve")
	if writer.count() != 1 {
		t.Fatalf("a decided request was written again (%d writes)", writer.count())
	}
	list := env.statusList("a")
	if len(list.Items) != 0 || len(list.History) != 1 || list.History[0].Status != "APPROVED" {
		t.Fatalf("after approval %+v: want no pending and the approval in history", list)
	}
}

func TestCommitStatusWriteFailureLeavesRequestPending(t *testing.T) {
	writer := &fakeStatusWriter{err: errors.New("GitHub answered 422: No commit found for SHA")}
	env := newCommitStatusEnv(t, writer, nil)
	k := env.issue("a", "homelab-k3s-mcp")
	id := env.fileStatus(k.Key, validStatusRequest())

	rec := env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-a", nil)
	expectCode(t, rec, http.StatusBadGateway, "approve with GitHub failing")
	failed := decodeApproval[struct {
		Error   string           `json:"error"`
		Request commitStatusBody `json:"request"`
	}](t, rec)
	if !strings.Contains(failed.Error, "No commit found") || failed.Request.Status != "PENDING" || !strings.Contains(failed.Request.LastError, "No commit found") {
		t.Fatalf("failure answer %+v: want GitHub's message and the request still pending with it recorded", failed)
	}
	if list := env.statusList("a"); len(list.Items) != 1 || list.Items[0].LastError == "" {
		t.Fatalf("pending after failure %+v: want it listed with its error", list.Items)
	}

	writer.setErr(nil)
	rec = env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-a", nil)
	expectCode(t, rec, http.StatusOK, "retry approve")
	if got := decodeApproval[commitStatusBody](t, rec); got.Status != "APPROVED" || got.LastError != "" {
		t.Fatalf("retried %+v: want APPROVED with the error cleared", got)
	}
}

func TestCommitStatusRejectNeverWrites(t *testing.T) {
	writer := &fakeStatusWriter{}
	env := newCommitStatusEnv(t, writer, nil)
	k := env.issue("a", "homelab-k3s-mcp")
	id := env.fileStatus(k.Key, validStatusRequest())

	rec := env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/reject", "oidc-a", nil)
	expectCode(t, rec, http.StatusOK, "reject")
	if got := decodeApproval[commitStatusBody](t, rec); got.Status != "REJECTED" || got.ProcessedAt == nil || got.GitHubStatusID != nil {
		t.Fatalf("rejected %+v", got)
	}
	expectCode(t, env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-a", nil), http.StatusConflict, "approve after reject")
	if writer.count() != 0 {
		t.Fatalf("a rejected request wrote %d statuses", writer.count())
	}
}

func TestCommitStatusWithoutGitHubAppStillRejects(t *testing.T) {
	env := newCommitStatusEnv(t, nil, nil)
	k := env.issue("a", "homelab-k3s-mcp")
	id := env.fileStatus(k.Key, validStatusRequest())

	rec := env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-a", nil)
	expectCode(t, rec, http.StatusServiceUnavailable, "approve without app")
	if got := env.statusList("a"); len(got.Items) != 1 {
		t.Fatalf("unconfigured approval changed the request: %+v", got)
	}
	expectCode(t, env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/reject", "oidc-a", nil), http.StatusOK, "reject without app")
}

func TestCommitStatusConcurrentApprovalsWriteOnce(t *testing.T) {
	writer := &fakeStatusWriter{delay: 50 * time.Millisecond}
	env := newCommitStatusEnv(t, writer, nil)
	k := env.issue("a", "homelab-k3s-mcp")
	id := env.fileStatus(k.Key, validStatusRequest())

	codes := make(chan int, 4)
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			codes <- env.do(http.MethodPost, "/api/commit-status-requests/"+id+"/approve", "oidc-a", nil).Code
		}()
	}
	wg.Wait()
	close(codes)
	ok, conflict := 0, 0
	for c := range codes {
		switch c {
		case http.StatusOK:
			ok++
		case http.StatusConflict:
			conflict++
		}
	}
	if ok != 1 || conflict != 3 || writer.count() != 1 {
		t.Fatalf("4 concurrent approvals: %d ok, %d conflict, %d writes — want 1, 3, 1", ok, conflict, writer.count())
	}
}
