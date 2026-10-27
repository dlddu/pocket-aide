package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/approvals"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/externalnotify"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

type sentNotification struct {
	userID      int64
	title, body string
}

type sentID struct{ notificationID int64 }

type recordingNotifier struct {
	mu     sync.Mutex
	sent   []sentNotification
	ids    []sentID
	result externalnotify.Result
	err    error
}

func (n *recordingNotifier) Send(_ context.Context, userID, notificationID int64, title, body string) (externalnotify.Result, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.sent = append(n.sent, sentNotification{userID, title, body})
	n.ids = append(n.ids, sentID{notificationID})
	return n.result, n.err
}

func (n *recordingNotifier) all() []sentNotification {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]sentNotification(nil), n.sent...)
}

func newNotificationEnv(t *testing.T, push handlers.NotificationSender) *approvalEnv {
	t.Helper()
	return newNotificationEnvWithLog(t, push, nil)
}

func newNotificationEnvWithLog(t *testing.T, push handlers.NotificationSender, log handlers.NotificationLog) *approvalEnv {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "notifications.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	env := &approvalEnv{t: t, conn: conn, reviewer: &fakeReviewer{tokens: map[string]string{}}, now: time.Unix(1_800_000_000, 0)}
	store := approvals.NewWithClock(conn, env.clock)
	notes := externalnotify.NewStore(conn)
	if log == nil {
		log = notes
	}
	r := chi.NewRouter()
	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(subVerifier{}, conn))
		p.Get("/api/approval-keys", handlers.ListApprovalKeys(store))
		p.Post("/api/approval-keys", handlers.IssueApprovalKey(store))
		p.Post("/api/approval-keys/{id}/revoke", handlers.RevokeApprovalKey(store))
		p.Get("/api/external-notifications", handlers.ListExternalNotifications(notes))
		p.Get("/api/external-notifications/{id}", handlers.GetExternalNotification(notes))
		p.Post("/api/external-notifications/{id}/read", handlers.MarkExternalNotificationRead(notes))
	})
	r.Group(func(x chi.Router) {
		x.Use(handlers.CallerKeyMiddleware(store, env.reviewer))
		x.Post("/api/external/notifications", handlers.CreateExternalNotification(push, log))
	})
	env.router = r
	return env
}

func TestExternalNotificationGoesToCallerOwner(t *testing.T) {
	push := &recordingNotifier{result: externalnotify.Result{Devices: 2, Sent: 2}}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")
	env.issue("bob", "bob-ci")

	rec := env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{
		"title": "  백업 완료  ",
		"body":  "nightly 스냅샷 3개 생성",
	})
	expectCode(t, rec, http.StatusOK, "send notification")
	got := decodeApproval[externalnotify.Result](t, rec)
	if got != (externalnotify.Result{Devices: 2, Sent: 2}) {
		t.Fatalf("response = %+v, want the notifier's counts", got)
	}

	sent := push.all()
	want := sentNotification{userID: env.userID("alice"), title: "백업 완료", body: "nightly 스냅샷 3개 생성"}
	if len(sent) != 1 || sent[0] != want {
		t.Fatalf("sent = %+v, want one push %+v to the key owner", sent, want)
	}
	if keys := env.keys("alice"); len(keys) != 1 || keys[0].LastUsedAt == nil {
		t.Fatalf("caller last_used_at should be updated: %+v", keys)
	}
}

func TestExternalNotificationBodyIsOptional(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")

	rec := env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": "ping"})
	expectCode(t, rec, http.StatusOK, "title only")
	if sent := push.all(); len(sent) != 1 || sent[0].body != "" {
		t.Fatalf("sent = %+v, want one push with an empty body", sent)
	}
}

func TestExternalNotificationRejectsBadBodies(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")

	cases := []struct {
		name string
		body any
	}{
		{"missing title", map[string]any{"body": "x"}},
		{"blank title", map[string]any{"title": "   "}},
		{"title too long", map[string]any{"title": strings.Repeat("가", 121)}},
		{"unknown field", map[string]any{"title": "x", "userId": "alice"}},
		{"not json", "{"},
	}
	for _, c := range cases {
		rec := env.do(http.MethodPost, "/api/external/notifications", key.Key, c.body)
		expectCode(t, rec, http.StatusBadRequest, c.name)
	}
	rec := env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": strings.Repeat("가", 120)})
	expectCode(t, rec, http.StatusOK, "title at the limit")
	if n := len(push.all()); n != 1 {
		t.Fatalf("pushes = %d, want only the valid request to send", n)
	}
}

func TestExternalNotificationRequiresLiveCaller(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")
	body := map[string]any{"title": "x"}

	expectCode(t, env.do(http.MethodPost, "/api/external/notifications", "", body), http.StatusUnauthorized, "no key")
	expectCode(t, env.do(http.MethodPost, "/api/external/notifications", "pak_nope", body), http.StatusUnauthorized, "unknown key")
	expectCode(t, env.do(http.MethodPost, "/api/external/notifications", "oidc-alice", body), http.StatusUnauthorized, "OIDC token")

	expectCode(t, env.do(http.MethodPost, "/api/approval-keys/"+strconv.FormatInt(key.ID, 10)+"/revoke", "oidc-alice", nil), http.StatusOK, "revoke")
	expectCode(t, env.do(http.MethodPost, "/api/external/notifications", key.Key, body), http.StatusUnauthorized, "revoked key")

	if n := len(push.all()); n != 0 {
		t.Fatalf("pushes = %d, want none from rejected callers", n)
	}
}

func TestExternalNotificationAcceptsServiceAccountCaller(t *testing.T) {
	push := &recordingNotifier{result: externalnotify.Result{Devices: 1, Sent: 1}}
	env := newNotificationEnv(t, push)
	env.reviewer.tokens["sa-token-mcp"] = "homelab-k3s-mcp/homelab-k3s-mcp"
	expectCode(t, env.registerSA("alice", "homelab-k3s-mcp", "homelab-k3s-mcp/homelab-k3s-mcp"), http.StatusCreated, "register SA")

	rec := env.do(http.MethodPost, "/api/external/notifications", "sa-token-mcp", map[string]any{"title": "배포 완료"})
	expectCode(t, rec, http.StatusOK, "SA caller")
	if sent := push.all(); len(sent) != 1 || sent[0].userID != env.userID("alice") {
		t.Fatalf("sent = %+v, want one push to the SA's registrant", sent)
	}
}

func TestExternalNotificationWithoutPushIsUnavailable(t *testing.T) {
	env := newNotificationEnv(t, nil)
	key := env.issue("alice", "infra-ci")

	rec := env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": "x"})
	expectCode(t, rec, http.StatusServiceUnavailable, "APNs not configured")
}

func TestExternalNotificationDeviceListFailureIs500(t *testing.T) {
	push := &recordingNotifier{err: errors.New("db locked")}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")

	rec := env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": "x"})
	expectCode(t, rec, http.StatusInternalServerError, "notifier error")
}
