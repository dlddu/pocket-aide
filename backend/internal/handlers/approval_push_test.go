package handlers_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/approvalpush"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
)

type recordingPusher struct {
	mu   sync.Mutex
	reqs []approvalpush.Request
}

func (p *recordingPusher) Notify(req approvalpush.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reqs = append(p.reqs, req)
}

func (p *recordingPusher) all() []approvalpush.Request {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]approvalpush.Request(nil), p.reqs...)
}

type blockingSender struct {
	release chan struct{}
	mu      sync.Mutex
	titles  []string
	sent    chan struct{}
}

func (s *blockingSender) SendWithCategory(_ context.Context, _ devicetokens.Device, _, title, _ string, _ map[string]any) error {
	<-s.release
	s.mu.Lock()
	s.titles = append(s.titles, title)
	s.mu.Unlock()
	s.sent <- struct{}{}
	return nil
}

func (e *approvalEnv) userID(sub string) int64 {
	e.t.Helper()
	var id int64
	if err := e.conn.QueryRow(`SELECT id FROM users WHERE oidc_sub = ?`, sub).Scan(&id); err != nil {
		e.t.Fatalf("user %s: %v", sub, err)
	}
	return id
}

func TestApprovalCreatePushesKeyNameAndContext(t *testing.T) {
	pusher := &recordingPusher{}
	env := newApprovalEnvWithPush(t, pusher)
	key := env.issue("alice", "infra-ci")
	created := env.create(key.Key, map[string]any{
		"externalId":    "deploy-1",
		"context":       "prod 배포를 승인해 주세요",
		"requesterName": "Deploy Bot",
	})

	got := pusher.all()
	if len(got) != 1 {
		t.Fatalf("pushes after one create: got %d want 1", len(got))
	}
	want := approvalpush.Request{
		UserID:     env.userID("alice"),
		KeyName:    "infra-ci",
		ApprovalID: created.ID,
		Context:    "prod 배포를 승인해 주세요",
	}
	if got[0] != want {
		t.Fatalf("push request: got %+v want %+v", got[0], want)
	}

	rec := env.do(http.MethodPost, "/api/external/approvals", key.Key, map[string]any{"externalId": "deploy-1", "context": "again"})
	expectCode(t, rec, http.StatusConflict, "duplicate externalId")
	rec = env.do(http.MethodPost, "/api/external/approvals", key.Key, map[string]any{"externalId": "deploy-2"})
	expectCode(t, rec, http.StatusBadRequest, "missing context")
	if n := len(pusher.all()); n != 1 {
		t.Fatalf("rejected creates must not push: got %d pushes", n)
	}
}

func TestApprovalCreateDoesNotWaitForPush(t *testing.T) {
	sender := &blockingSender{release: make(chan struct{}), sent: make(chan struct{}, 1)}
	env := newApprovalEnvWithPush(t, approvalpush.New(sender, oneDevice{}))
	key := env.issue("alice", "infra-ci")
	time.AfterFunc(3*time.Second, func() { close(sender.release) })

	start := time.Now()
	env.create(key.Key, map[string]any{"externalId": "slow-apns", "context": "ctx"})
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("create waited for the push: %v", elapsed)
	}

	select {
	case <-sender.sent:
	case <-time.After(5 * time.Second):
		t.Fatal("the push should still go out after the response")
	}
	sender.mu.Lock()
	defer sender.mu.Unlock()
	if len(sender.titles) != 1 || sender.titles[0] != "승인 요청 · infra-ci" {
		t.Fatalf("push titles: %v", sender.titles)
	}
}

type oneDevice struct{}

func (oneDevice) ListByUserID(context.Context, int64) ([]devicetokens.Device, error) {
	return []devicetokens.Device{{Token: "phone", Platform: devicetokens.PlatformIOS}}, nil
}

func (oneDevice) Delete(context.Context, int64, string) error {
	return nil
}
