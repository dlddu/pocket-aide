package approvalpush_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
	"github.com/dlddu/pocket-aide/backend/internal/approvalpush"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/fcm"
	"github.com/dlddu/pocket-aide/backend/internal/push"
)

type sent struct {
	token, platform, category, title, body string
	data                                   map[string]any
}

type fakeSender struct {
	mu     sync.Mutex
	sent   []sent
	errFor map[string]error
}

func (f *fakeSender) SendWithCategory(_ context.Context, d devicetokens.Device, category, title, body string, data map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sent{d.Token, d.Platform, category, title, body, data})
	return f.errFor[d.Token]
}

func (f *fakeSender) all() []sent {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]sent(nil), f.sent...)
}

func setup(t *testing.T, tokens ...string) (*devicetokens.Store, int64) {
	t.Helper()
	devices := make([]devicetokens.Device, 0, len(tokens))
	for _, tok := range tokens {
		devices = append(devices, devicetokens.Device{Token: tok, Platform: devicetokens.PlatformIOS})
	}
	return setupDevices(t, devices...)
}

func setupDevices(t *testing.T, devices ...devicetokens.Device) (*devicetokens.Store, int64) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "push.db"))
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
	for _, d := range devices {
		if err := store.Upsert(context.Background(), uid, d.Token, d.Platform); err != nil {
			t.Fatalf("token: %v", err)
		}
	}
	if err := store.Upsert(context.Background(), uid+1, "bob-device", devicetokens.PlatformIOS); err != nil {
		t.Fatalf("token: %v", err)
	}
	return store, uid
}

func TestSendGoesToEveryDeviceOfTheOwner(t *testing.T) {
	store, uid := setup(t, "phone", "ipad")
	sender := &fakeSender{}
	approvalpush.New(sender, store).Send(context.Background(), approvalpush.Request{
		UserID: uid, KeyName: "infra-ci", ApprovalID: "req-1", Context: "prod 배포",
	})

	got := sender.all()
	tokens := []string{}
	for _, s := range got {
		tokens = append(tokens, s.token)
		if s.title != "승인 요청 · infra-ci" || s.body != "prod 배포" {
			t.Fatalf("alert: title=%q body=%q", s.title, s.body)
		}
		if !reflect.DeepEqual(s.data, map[string]any{"approval_id": "req-1"}) {
			t.Fatalf("data: %v", s.data)
		}
		if s.category != "APPROVAL_REQUEST" {
			t.Fatalf("category: %q", s.category)
		}
	}
	sort.Strings(tokens)
	if !reflect.DeepEqual(tokens, []string{"ipad", "phone"}) {
		t.Fatalf("devices: %v", tokens)
	}
}

func TestLongContextIsCutToFitThePayload(t *testing.T) {
	store, uid := setup(t, "phone")
	sender := &fakeSender{}
	ctx := strings.Repeat("배포 승인 요청 context ", 600)
	approvalpush.New(sender, store).Send(context.Background(), approvalpush.Request{
		UserID: uid, KeyName: "infra-ci", ApprovalID: "req-1", Context: ctx,
	})

	got := sender.all()
	if len(got) != 1 {
		t.Fatalf("pushes: %d", len(got))
	}
	body := got[0].body
	if !strings.HasSuffix(body, "…") || !strings.HasPrefix(ctx, strings.TrimSuffix(body, "…")) {
		t.Fatalf("body should be a prefix of the context ending in …: %q", body[len(body)-20:])
	}
	if size := apns.PayloadSize(got[0].category, got[0].title, body, got[0].data); size > apns.MaxPayloadBytes {
		t.Fatalf("payload %d bytes > %d", size, apns.MaxPayloadBytes)
	}
}

func TestInvalidTokensAreDroppedOthersKept(t *testing.T) {
	store, uid := setup(t, "gone", "bad", "busy", "fine")
	sender := &fakeSender{errFor: map[string]error{
		"gone": &apns.RejectedError{StatusCode: 410, Reason: "Unregistered"},
		"bad":  &apns.RejectedError{StatusCode: 400, Reason: "BadDeviceToken"},
		"busy": &apns.RejectedError{StatusCode: 429, Reason: "TooManyRequests"},
		"fine": errors.New("network down"),
	}}
	approvalpush.New(sender, store).Send(context.Background(), approvalpush.Request{
		UserID: uid, KeyName: "infra-ci", ApprovalID: "req-1", Context: "ctx",
	})

	if left := tokensOf(t, store, uid); !reflect.DeepEqual(left, []string{"busy", "fine"}) {
		t.Fatalf("tokens left: %v", left)
	}
	if other := tokensOf(t, store, uid+1); !reflect.DeepEqual(other, []string{"bob-device"}) {
		t.Fatalf("another user's tokens must stay: %v", other)
	}
}

func TestAndroidDevicesGetTheSamePushAndDeadFCMTokensAreDropped(t *testing.T) {
	store, uid := setupDevices(t,
		devicetokens.Device{Token: "phone", Platform: devicetokens.PlatformIOS},
		devicetokens.Device{Token: "pixel", Platform: devicetokens.PlatformAndroid},
		devicetokens.Device{Token: "old-pixel", Platform: devicetokens.PlatformAndroid},
		devicetokens.Device{Token: "no-fcm", Platform: devicetokens.PlatformAndroid},
	)
	sender := &fakeSender{errFor: map[string]error{
		"old-pixel": &fcm.RejectedError{StatusCode: 404, Status: "NOT_FOUND", ErrorCode: "UNREGISTERED"},
		"no-fcm":    push.ErrPlatformDisabled,
	}}
	approvalpush.New(sender, store).Send(context.Background(), approvalpush.Request{
		UserID: uid, KeyName: "infra-ci", ApprovalID: "req-1", Context: "ctx",
	})

	platforms := map[string]string{}
	for _, s := range sender.all() {
		platforms[s.token] = s.platform
		if s.category != approvalpush.Category {
			t.Fatalf("category for %s: %q", s.token, s.category)
		}
	}
	if platforms["pixel"] != devicetokens.PlatformAndroid || platforms["phone"] != devicetokens.PlatformIOS {
		t.Fatalf("sends: %v", platforms)
	}
	if left := tokensOf(t, store, uid); !reflect.DeepEqual(left, []string{"no-fcm", "phone", "pixel"}) {
		t.Fatalf("tokens left: %v (a disabled platform must not drop its tokens)", left)
	}
}

func tokensOf(t *testing.T, store *devicetokens.Store, uid int64) []string {
	t.Helper()
	devices, err := store.ListByUserID(context.Background(), uid)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	out := make([]string, 0, len(devices))
	for _, d := range devices {
		out = append(out, d.Token)
	}
	sort.Strings(out)
	return out
}

type gate struct {
	release chan struct{}
	done    chan struct{}
}

func (g *gate) SendWithCategory(context.Context, devicetokens.Device, string, string, string, map[string]any) error {
	<-g.release
	close(g.done)
	return nil
}

func TestNotifyReturnsBeforeTheSendFinishes(t *testing.T) {
	store, uid := setup(t, "phone")
	g := &gate{release: make(chan struct{}), done: make(chan struct{})}
	time.AfterFunc(3*time.Second, func() { close(g.release) })
	start := time.Now()
	approvalpush.New(g, store).Notify(approvalpush.Request{UserID: uid, KeyName: "k", ApprovalID: "r", Context: "c"})
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Notify blocked for %v", elapsed)
	}
	select {
	case <-g.done:
	case <-time.After(5 * time.Second):
		t.Fatal("the push should be sent in the background")
	}
}
