package commitstatuspush_test

import (
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
	"github.com/dlddu/pocket-aide/backend/internal/commitstatuspush"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/push"
)

type sent struct {
	token, category, title, body string
	data                         map[string]any
}

type fakeSender struct {
	mu   sync.Mutex
	sent []sent
	err  map[string]error
}

func (f *fakeSender) SendWithCategory(_ context.Context, d devicetokens.Device, category, title, body string, data map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sent{d.Token, category, title, body, data})
	return f.err[d.Token]
}

func TestSendGoesToEveryOwnerDeviceWithDecisionCategory(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "push.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	res, err := conn.Exec(`INSERT INTO users (oidc_sub) VALUES ('alice')`)
	if err != nil {
		t.Fatal(err)
	}
	uid, _ := res.LastInsertId()
	tokens := devicetokens.New(conn)
	for _, d := range []devicetokens.Device{{Token: "phone", Platform: devicetokens.PlatformIOS}, {Token: "pixel", Platform: devicetokens.PlatformAndroid}} {
		if err := tokens.Upsert(context.Background(), uid, d.Token, d.Platform, ""); err != nil {
			t.Fatal(err)
		}
	}
	sender := &fakeSender{err: map[string]error{"pixel": push.ErrPlatformDisabled}}

	commitstatuspush.New(sender, tokens).Send(context.Background(), commitstatuspush.Request{
		UserID: uid, KeyName: "homelab-k3s-mcp", RequestID: "req-1", Owner: "dlddu", Repo: "pocket-aide",
		SHA: "0123456789abcdef0123456789abcdef01234567", State: "success", Context: "review/manual-approval", Requester: "Claude Code 세션",
	})

	if len(sender.sent) != 2 {
		t.Fatalf("sent %+v: want both devices", sender.sent)
	}
	got := sender.sent[0]
	if got.category != commitstatuspush.Category || got.title != "커밋 상태 요청 · homelab-k3s-mcp" {
		t.Fatalf("push %+v: want the commit status category and the caller in the title", got)
	}
	if !strings.Contains(got.body, "review/manual-approval → success") || !strings.Contains(got.body, "dlddu/pocket-aide@0123456") || !strings.Contains(got.body, "Claude Code 세션") {
		t.Fatalf("body %q: want context, state, short sha and requester", got.body)
	}
	if want := map[string]any{"kind": "commit_status_request", "commit_status_request_id": "req-1"}; !reflect.DeepEqual(got.data, want) {
		t.Fatalf("data %v, want %v", got.data, want)
	}
	if apns.PayloadSize(got.category, got.title, got.body, got.data) > apns.MaxPayloadBytes {
		t.Fatalf("payload over the APNs limit")
	}
	if left, _ := tokens.ListByUserID(context.Background(), uid); len(left) != 2 {
		t.Fatalf("a disabled platform must not drop the device token: %+v", left)
	}
}
