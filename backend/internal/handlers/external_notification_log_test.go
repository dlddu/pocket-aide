package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/externalnotify"
)

type externalNotificationsBody struct {
	Items       []externalnotify.Item `json:"items"`
	UnreadCount int                   `json:"unread_count"`
}

func (e *approvalEnv) externalNotifications(sub, query string) externalNotificationsBody {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/external-notifications"+query, "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "list external notifications")
	return decodeApproval[externalNotificationsBody](e.t, rec)
}

func (e *approvalEnv) notify(bearer string, body map[string]any) {
	e.t.Helper()
	expectCode(e.t, e.do(http.MethodPost, "/api/external/notifications", bearer, body), http.StatusOK, "send notification")
}

func TestExternalNotificationIsRecordedWithCallerAndRequester(t *testing.T) {
	push := &recordingNotifier{result: externalnotify.Result{Devices: 2, Sent: 1}}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")

	env.notify(key.Key, map[string]any{"title": " 백업 완료 ", "body": " 스냅샷 3개 ", "requester": "  jwt:user-123  "})
	env.notify(key.Key, map[string]any{"title": "두 번째"})

	got := env.externalNotifications("alice", "")
	if len(got.Items) != 2 || got.UnreadCount != 2 {
		t.Fatalf("list = %+v, want two unread items", got)
	}
	newest, first := got.Items[0], got.Items[1]
	if newest.Title != "두 번째" || newest.Requester != "" {
		t.Fatalf("newest = %+v, want the later notification first with no requester", newest)
	}
	want := externalnotify.Item{
		ID: first.ID, Title: "백업 완료", Body: "스냅샷 3개", CallerName: "infra-ci", CallerKind: "key",
		Requester: "jwt:user-123", Devices: 2, Sent: 1, CreatedAt: first.CreatedAt,
	}
	if first.ReadAt != nil || first.CreatedAt == 0 {
		t.Fatalf("first = %+v, want unread with a creation time", first)
	}
	first.ReadAt = nil
	if first != want {
		t.Fatalf("first = %+v, want %+v", first, want)
	}
}

func TestExternalNotificationRecordsServiceAccountCaller(t *testing.T) {
	push := &recordingNotifier{result: externalnotify.Result{Devices: 1, Sent: 1}}
	env := newNotificationEnv(t, push)
	env.reviewer.tokens["sa-token-mcp"] = "homelab-k3s-mcp/homelab-k3s-mcp"
	expectCode(t, env.registerSA("alice", "homelab-k3s-mcp", "homelab-k3s-mcp/homelab-k3s-mcp"), http.StatusCreated, "register SA")

	env.notify("sa-token-mcp", map[string]any{"title": "배포 완료", "requester": "github_oidc:dlddu/reconciler/.github/workflows/reconcile.yml"})

	items := env.externalNotifications("alice", "").Items
	if len(items) != 1 || items[0].CallerName != "homelab-k3s-mcp" || items[0].CallerKind != "k8s_sa" ||
		items[0].Requester != "github_oidc:dlddu/reconciler/.github/workflows/reconcile.yml" {
		t.Fatalf("items = %+v, want the SA caller and the reported requester", items)
	}
}

func TestExternalNotificationRequesterTooLongIsRejected(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")

	rec := env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": "x", "requester": strings.Repeat("가", 201)})
	expectCode(t, rec, http.StatusBadRequest, "requester too long")
	env.notify(key.Key, map[string]any{"title": "x", "requester": strings.Repeat("가", 200)})

	if n := len(push.all()); n != 1 {
		t.Fatalf("pushes = %d, want only the request within the limit", n)
	}
	if items := env.externalNotifications("alice", "").Items; len(items) != 1 {
		t.Fatalf("items = %+v, want only the sent notification recorded", items)
	}
}

func TestExternalNotificationsAreScopedToTheOwner(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")
	env.issue("bob", "bob-ci")

	env.notify(key.Key, map[string]any{"title": "alice only"})

	if got := env.externalNotifications("bob", ""); len(got.Items) != 0 || got.UnreadCount != 0 {
		t.Fatalf("bob sees %+v, want nothing", got)
	}
	expectCode(t, env.do(http.MethodGet, "/api/external-notifications", "", nil), http.StatusUnauthorized, "no token")
}

func TestExternalNotificationSendFailureLeavesNoRecord(t *testing.T) {
	push := &recordingNotifier{err: errors.New("db locked")}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")

	expectCode(t, env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": "x"}), http.StatusInternalServerError, "notifier error")
	if items := env.externalNotifications("alice", "").Items; len(items) != 0 {
		t.Fatalf("items = %+v, want no record for a send that never started", items)
	}
}

func TestExternalNotificationWithoutPushIsNotRecorded(t *testing.T) {
	env := newNotificationEnv(t, nil)
	key := env.issue("alice", "infra-ci")

	expectCode(t, env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": "x"}), http.StatusServiceUnavailable, "APNs not configured")
	if items := env.externalNotifications("alice", "").Items; len(items) != 0 {
		t.Fatalf("items = %+v, want nothing recorded", items)
	}
}

type failingLog struct{}

func (failingLog) Insert(context.Context, int64, int64, externalnotify.Notification) (int64, error) {
	return 0, errors.New("disk full")
}
func (failingLog) SetDelivery(context.Context, int64, externalnotify.Result) error { return nil }
func (failingLog) Delete(context.Context, int64) error                             { return nil }

func TestExternalNotificationIsNotSentWhenItCannotBeRecorded(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnvWithLog(t, push, failingLog{})
	key := env.issue("alice", "infra-ci")

	expectCode(t, env.do(http.MethodPost, "/api/external/notifications", key.Key, map[string]any{"title": "x"}), http.StatusInternalServerError, "record failure")
	if n := len(push.all()); n != 0 {
		t.Fatalf("pushes = %d, want none without a record", n)
	}
}

func TestExternalNotificationPushCarriesTheRecordedID(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")
	env.notify(key.Key, map[string]any{"title": "one", "requester": "r"})
	env.notify(key.Key, map[string]any{"title": "two", "requester": "r"})

	items := env.externalNotifications("alice", "").Items
	push.mu.Lock()
	ids := append([]sentID(nil), push.ids...)
	push.mu.Unlock()
	if len(ids) != 2 || ids[0].notificationID != items[1].ID || ids[1].notificationID != items[0].ID {
		t.Fatalf("pushed ids = %v, want the recorded ids %d then %d", ids, items[1].ID, items[0].ID)
	}
}

func TestExternalNotificationDetailAndReadArePerItem(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")
	env.notify(key.Key, map[string]any{"title": "one", "requester": "긴 요청 주체"})
	env.notify(key.Key, map[string]any{"title": "two"})
	listed := env.externalNotifications("alice", "")
	if listed.UnreadCount != 2 {
		t.Fatalf("listing must not mark anything read: %+v", listed)
	}
	first := listed.Items[1]

	rec := env.do(http.MethodGet, "/api/external-notifications/"+strconv.FormatInt(first.ID, 10), "oidc-alice", nil)
	expectCode(t, rec, http.StatusOK, "get detail")
	got := decodeApproval[externalnotify.Item](t, rec)
	if got.ID != first.ID || got.Title != "one" || got.Requester != "긴 요청 주체" || got.CallerName != "infra-ci" || got.ReadAt != nil {
		t.Fatalf("detail = %+v, want the unread first notification", got)
	}

	path := "/api/external-notifications/" + strconv.FormatInt(first.ID, 10) + "/read"
	expectCode(t, env.do(http.MethodPost, path, "oidc-alice", nil), http.StatusNoContent, "mark read")
	expectCode(t, env.do(http.MethodPost, path, "oidc-alice", nil), http.StatusNoContent, "mark read again")
	after := env.externalNotifications("alice", "")
	if after.UnreadCount != 1 || after.Items[1].ReadAt == nil || after.Items[0].ReadAt != nil {
		t.Fatalf("after = %+v, want only the opened notification read", after)
	}

	expectCode(t, env.do(http.MethodGet, "/api/external-notifications/"+strconv.FormatInt(first.ID, 10), "oidc-bob", nil), http.StatusNotFound, "bob get")
	expectCode(t, env.do(http.MethodPost, path, "oidc-bob", nil), http.StatusNotFound, "bob mark")
	expectCode(t, env.do(http.MethodGet, "/api/external-notifications/999999", "oidc-alice", nil), http.StatusNotFound, "missing")
	expectCode(t, env.do(http.MethodGet, "/api/external-notifications/abc", "oidc-alice", nil), http.StatusBadRequest, "bad id")
}

func TestExternalNotificationsPageWithBefore(t *testing.T) {
	push := &recordingNotifier{}
	env := newNotificationEnv(t, push)
	key := env.issue("alice", "infra-ci")
	for i := 1; i <= 3; i++ {
		env.notify(key.Key, map[string]any{"title": "n" + strconv.Itoa(i)})
	}

	page := env.externalNotifications("alice", "?limit=2")
	if len(page.Items) != 2 || page.Items[0].Title != "n3" || page.UnreadCount != 3 {
		t.Fatalf("page 1 = %+v", page)
	}
	next := env.externalNotifications("alice", "?limit=2&before="+strconv.FormatInt(page.Items[1].ID, 10))
	if len(next.Items) != 1 || next.Items[0].Title != "n1" {
		t.Fatalf("page 2 = %+v", next)
	}
}
