package externalnotify_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/externalnotify"
)

type sent struct {
	token, title, body string
	data               map[string]any
}

type fakeSender struct {
	mu     sync.Mutex
	sent   []sent
	errFor map[string]error
}

func (f *fakeSender) SendWithData(_ context.Context, token, title, body string, data map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sent{token, title, body, data})
	return f.errFor[token]
}

func (f *fakeSender) all() []sent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sent(nil), f.sent...)
}

func setup(t *testing.T, tokens ...string) (*devicetokens.Store, int64) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "notify.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	res, err := conn.Exec(`INSERT INTO users (oidc_sub) VALUES ('alice')`)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	uid, _ := res.LastInsertId()
	if _, err := conn.Exec(`INSERT INTO users (oidc_sub) VALUES ('bob')`); err != nil {
		t.Fatalf("user: %v", err)
	}
	store := devicetokens.New(conn)
	for _, tok := range tokens {
		if err := store.Upsert(context.Background(), uid, tok); err != nil {
			t.Fatalf("token: %v", err)
		}
	}
	if err := store.Upsert(context.Background(), uid+1, "bob-device"); err != nil {
		t.Fatalf("token: %v", err)
	}
	return store, uid
}

func TestSendGoesToEveryDeviceOfTheOwnerAsIs(t *testing.T) {
	store, uid := setup(t, "phone", "ipad")
	sender := &fakeSender{}
	res, err := externalnotify.New(sender, store).Send(context.Background(), uid, 7, "백업 완료", "스냅샷 3개")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if res != (externalnotify.Result{Devices: 2, Sent: 2}) {
		t.Fatalf("result = %+v", res)
	}

	tokens := []string{}
	for _, s := range sender.all() {
		tokens = append(tokens, s.token)
		if s.title != "백업 완료" || s.body != "스냅샷 3개" || !reflect.DeepEqual(s.data, map[string]any{"kind": "external_notification", "notification_id": int64(7)}) {
			t.Fatalf("alert: title=%q body=%q data=%v, want the caller's text, the kind marker and the notification id", s.title, s.body, s.data)
		}
	}
	sort.Strings(tokens)
	if !reflect.DeepEqual(tokens, []string{"ipad", "phone"}) {
		t.Fatalf("devices: %v (another user's device must not receive it)", tokens)
	}
}

func TestSendWithNoDevicesReportsZero(t *testing.T) {
	store, uid := setup(t)
	sender := &fakeSender{}
	res, err := externalnotify.New(sender, store).Send(context.Background(), uid, 7, "x", "")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if res != (externalnotify.Result{}) || len(sender.all()) != 0 {
		t.Fatalf("result = %+v sent = %v, want nothing", res, sender.all())
	}
}

func TestLongBodyIsCutToFitThePayload(t *testing.T) {
	store, uid := setup(t, "phone")
	sender := &fakeSender{}
	long := strings.Repeat("야간 백업 로그 ", 800)
	if _, err := externalnotify.New(sender, store).Send(context.Background(), uid, 7, "백업", long); err != nil {
		t.Fatalf("send: %v", err)
	}
	got := sender.all()
	if len(got) != 1 {
		t.Fatalf("pushes: %d", len(got))
	}
	body := got[0].body
	if !strings.HasSuffix(body, "…") || !strings.HasPrefix(long, strings.TrimSuffix(body, "…")) {
		t.Fatalf("body should be a prefix of the input ending in …")
	}
	if size := apns.PayloadSize("", got[0].title, body, got[0].data); size > apns.MaxPayloadBytes {
		t.Fatalf("payload %d bytes > %d", size, apns.MaxPayloadBytes)
	}
}

func TestFailuresAreCountedAndDeadTokensDropped(t *testing.T) {
	store, uid := setup(t, "gone", "bad", "busy", "down", "fine")
	sender := &fakeSender{errFor: map[string]error{
		"gone": &apns.RejectedError{StatusCode: 410, Reason: "Unregistered"},
		"bad":  &apns.RejectedError{StatusCode: 400, Reason: "BadDeviceToken"},
		"busy": &apns.RejectedError{StatusCode: 429, Reason: "TooManyRequests"},
		"down": errors.New("network down"),
	}}
	res, err := externalnotify.New(sender, store).Send(context.Background(), uid, 7, "x", "y")
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if res != (externalnotify.Result{Devices: 5, Sent: 1}) {
		t.Fatalf("result = %+v, want 5 devices / 1 sent", res)
	}

	left, err := store.ListByUserID(context.Background(), uid)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	sort.Strings(left)
	if !reflect.DeepEqual(left, []string{"busy", "down", "fine"}) {
		t.Fatalf("tokens left: %v", left)
	}
	other, _ := store.ListByUserID(context.Background(), uid+1)
	if !reflect.DeepEqual(other, []string{"bob-device"}) {
		t.Fatalf("another user's tokens must stay: %v", other)
	}
}
