package mail_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapserver"
	"github.com/emersion/go-imap/v2/imapserver/imapmemserver"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/mail"
)

func testKey(b byte) string { return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{b}, 32)) }

func TestCipher_RoundTripAndVersioning(t *testing.T) {
	c, err := mail.NewCipher("1", testKey(1))
	if err != nil {
		t.Fatal(err)
	}
	a, _ := c.Encrypt("app-password")
	b, _ := c.Encrypt("app-password")
	if a == b {
		t.Error("two encryptions of the same password must differ (random nonce)")
	}
	if !strings.HasPrefix(a, "1:") || strings.Contains(a, "app-password") {
		t.Errorf("sealed form should carry the key version and hide the plaintext: %q", a)
	}
	if got, err := c.Decrypt(a); err != nil || got != "app-password" {
		t.Errorf("Decrypt = %q, %v", got, err)
	}

	other, _ := mail.NewCipher("1", testKey(2))
	if _, err := other.Decrypt(a); !errors.Is(err, mail.ErrBadCiphertext) {
		t.Errorf("wrong key: want ErrBadCiphertext, got %v", err)
	}
	v2, _ := mail.NewCipher("2", testKey(1))
	if _, err := v2.Decrypt(a); !errors.Is(err, mail.ErrBadCiphertext) {
		t.Errorf("unknown key version: want ErrBadCiphertext, got %v", err)
	}
	if _, err := c.Decrypt("garbage"); !errors.Is(err, mail.ErrBadCiphertext) {
		t.Errorf("malformed: want ErrBadCiphertext, got %v", err)
	}
}

func TestNewCipher_RejectsBadKeys(t *testing.T) {
	for name, key := range map[string]string{
		"not base64": "***",
		"too short":  base64.StdEncoding.EncodeToString([]byte("short")),
	} {
		if _, err := mail.NewCipher("1", key); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if _, err := mail.NewCipher("", testKey(1)); err == nil {
		t.Error("empty version: expected error")
	}
}

func newStore(t *testing.T) *mail.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "mail.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	c, _ := mail.NewCipher("1", testKey(1))
	return mail.NewStore(conn, c)
}

func TestStore_PasswordSealedAtRestAndScopedToUser(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)
	acct, err := store.Add(ctx, 1, mail.Credentials{Email: "a@example.com", Password: "secret-pw", Host: "imap.example.com", Port: 993})
	if err != nil {
		t.Fatal(err)
	}
	if acct.NeedsReconnect || acct.IMAPPort != 993 {
		t.Errorf("unexpected account: %+v", acct)
	}
	if _, err := store.Add(ctx, 1, mail.Credentials{Email: "a@example.com", Password: "x", Host: "h", Port: 993}); !errors.Is(err, mail.ErrAlreadyConnected) {
		t.Errorf("duplicate: want ErrAlreadyConnected, got %v", err)
	}
	cred, err := store.Credentials(ctx, 1, acct.ID)
	if err != nil || cred.Password != "secret-pw" {
		t.Fatalf("Credentials = %+v, %v", cred, err)
	}
	if _, err := store.Credentials(ctx, 2, acct.ID); !errors.Is(err, mail.ErrNotFound) {
		t.Errorf("other user's credentials: want ErrNotFound, got %v", err)
	}
	if err := store.Delete(ctx, 2, acct.ID); !errors.Is(err, mail.ErrNotFound) {
		t.Errorf("other user's delete: want ErrNotFound, got %v", err)
	}

	if err := store.SetNeedsReconnect(ctx, 1, acct.ID, true); err != nil {
		t.Fatal(err)
	}
	// A revoked password keeps its row and ciphertext; only the flag moves.
	if got, _ := store.Account(ctx, 1, acct.ID); !got.NeedsReconnect {
		t.Error("expected needs_reconnect")
	}
	if err := store.ReplacePassword(ctx, 1, acct.ID, "new-pw"); err != nil {
		t.Fatal(err)
	}
	got, _ := store.Account(ctx, 1, acct.ID)
	cred, _ = store.Credentials(ctx, 1, acct.ID)
	if got.NeedsReconnect || cred.Password != "new-pw" {
		t.Errorf("after reconnect: %+v / %q", got, cred.Password)
	}

	if err := store.Delete(ctx, 1, acct.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Credentials(ctx, 1, acct.ID); !errors.Is(err, mail.ErrNotFound) {
		t.Errorf("deleted account must take its password with it, got %v", err)
	}
}

// --- against a real (in-memory) IMAP server ---------------------------------

type literal struct{ *bytes.Reader }

func (l literal) Size() int64 { return int64(l.Len()) }

type imapFixture struct {
	user *imapmemserver.User
	host string
	port int
}

func startIMAP(t *testing.T, username, password string) *imapFixture {
	t.Helper()
	mem := imapmemserver.New()
	user := imapmemserver.NewUser(username, password)
	if err := user.Create("INBOX", nil); err != nil {
		t.Fatal(err)
	}
	mem.AddUser(user)
	srv := imapserver.New(&imapserver.Options{
		NewSession: func(*imapserver.Conn) (imapserver.Session, *imapserver.GreetingData, error) {
			return mem.NewSession(), nil, nil
		},
		Caps:         imap.CapSet{imap.CapIMAP4rev1: {}, imap.CapIMAP4rev2: {}},
		InsecureAuth: true,
	})
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	addr := ln.Addr().(*net.TCPAddr)
	return &imapFixture{user: user, host: addr.IP.String(), port: addr.Port}
}

func (f *imapFixture) deliver(t *testing.T, from, subject, body string, at time.Time, flags ...imap.Flag) {
	t.Helper()
	raw := fmt.Sprintf("From: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-Id: <%d@x>\r\n\r\n%s\r\n",
		from, subject, at.Format(time.RFC1123Z), at.UnixNano(), body)
	if _, err := f.user.Append("INBOX", literal{bytes.NewReader([]byte(raw))}, &imap.AppendOptions{Flags: flags, Time: at}); err != nil {
		t.Fatal(err)
	}
}

func (f *imapFixture) creds(password string) mail.Credentials {
	return mail.Credentials{Email: "me@example.com", Password: password, Host: f.host, Port: f.port}
}

func TestClient_ProbeDistinguishesAuthFromConnectFailure(t *testing.T) {
	f := startIMAP(t, "me@example.com", "good")
	c := &mail.Client{Insecure: true, Timeout: 5 * time.Second}
	ctx := context.Background()

	if err := c.Probe(ctx, f.creds("good")); err != nil {
		t.Fatalf("good login: %v", err)
	}
	if err := c.Probe(ctx, f.creds("bad")); !errors.Is(err, mail.ErrAuth) {
		t.Errorf("bad password: want ErrAuth, got %v", err)
	}
	dead := f.creds("good")
	dead.Port = 1 // nothing listens here
	if err := c.Probe(ctx, dead); err == nil || errors.Is(err, mail.ErrAuth) {
		t.Errorf("unreachable server must be a plain error, got %v", err)
	}
}

func TestClient_UnseenIsReadOnlyNewestFirstAndCapped(t *testing.T) {
	f := startIMAP(t, "me@example.com", "pw")
	base := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	f.deliver(t, "Old <old@example.com>", "oldest", "first body", base)
	f.deliver(t, "Seen <seen@example.com>", "already read", "x", base.Add(time.Hour), imap.FlagSeen)
	f.deliver(t, "plain@example.com", "newest", "hello\r\n  world", base.Add(2*time.Hour))
	f.deliver(t, "Mid <mid@example.com>", "middle", "mid body", base.Add(90*time.Minute))

	c := &mail.Client{Insecure: true, Timeout: 5 * time.Second}
	ctx := context.Background()
	msgs, err := c.Unseen(ctx, f.creds("pw"), 50)
	if err != nil {
		t.Fatal(err)
	}
	var subjects []string
	for _, m := range msgs {
		subjects = append(subjects, m.Subject)
	}
	if got := strings.Join(subjects, ","); got != "newest,middle,oldest" {
		t.Fatalf("subjects = %s (seen mail must be excluded, newest first)", got)
	}
	if msgs[0].From != "plain@example.com" || msgs[1].From != "Mid" {
		t.Errorf("from = %q / %q (display name wins, else address)", msgs[0].From, msgs[1].From)
	}
	if msgs[0].Preview != "hello world" {
		t.Errorf("preview = %q", msgs[0].Preview)
	}

	// Reading the list must not mark anything read: a second fetch sees the same set.
	again, err := c.Unseen(ctx, f.creds("pw"), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 2 || again[0].Subject != "newest" || again[1].Subject != "middle" {
		t.Errorf("capped second fetch = %+v", again)
	}
	all, _ := c.Unseen(ctx, f.creds("pw"), 50)
	if len(all) != 3 {
		t.Errorf("unseen count after two reads = %d, want 3 (client must not set \\Seen)", len(all))
	}
}

func TestService_ConnectMessagesReconnectDisconnect(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)
	f1 := startIMAP(t, "one@example.com", "pw1")
	f2 := startIMAP(t, "two@example.com", "pw2")
	base := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)
	f1.deliver(t, "a@x.com", "from-one", "b", base)
	f2.deliver(t, "b@x.com", "from-two", "b", base.Add(time.Hour))

	svc := &mail.Service{Store: store, IMAP: &mail.Client{Insecure: true, Timeout: 5 * time.Second}}
	// Preset hosts are TLS-only (port 993), so tests pass an explicit host and
	// override the port after the fact through the store row.
	connect := func(email, pw string, f *imapFixture) mail.Account {
		t.Helper()
		_, err := store.Add(ctx, 1, mail.Credentials{Email: email, Password: pw, Host: f.host, Port: f.port})
		if err != nil {
			t.Fatal(err)
		}
		accts, _ := store.List(ctx, 1)
		return accts[len(accts)-1]
	}
	a1 := connect("one@example.com", "pw1", f1)
	a2 := connect("two@example.com", "pw2", f2)

	msgs, statuses, err := svc.Messages(ctx, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 2 || msgs[0].Subject != "from-two" || msgs[0].AccountID != a2.ID || msgs[0].AccountEmail != "two@example.com" {
		t.Errorf("merged list = %+v", msgs)
	}
	if len(statuses) != 2 {
		t.Errorf("statuses = %+v", statuses)
	}
	only, _, _ := svc.Messages(ctx, 1, a1.ID)
	if len(only) != 1 || only[0].Subject != "from-one" {
		t.Errorf("filtered list = %+v", only)
	}
	if _, _, err := svc.Messages(ctx, 2, a1.ID); !errors.Is(err, mail.ErrNotFound) {
		t.Errorf("another user's account filter: want ErrNotFound, got %v", err)
	}

	// Password revoked on the server: the account is flagged, kept, and the
	// other account still lists.
	if err := store.ReplacePassword(ctx, 1, a1.ID, "revoked"); err != nil {
		t.Fatal(err)
	}
	msgs, statuses, _ = svc.Messages(ctx, 1, 0)
	if len(msgs) != 1 || msgs[0].AccountID != a2.ID {
		t.Errorf("list with one revoked account = %+v", msgs)
	}
	got := map[int64]string{}
	for _, s := range statuses {
		got[s.AccountID] = s.Status
	}
	if got[a1.ID] != mail.StatusReauth || got[a2.ID] != mail.StatusOK {
		t.Errorf("statuses = %v", got)
	}
	if acct, _ := store.Account(ctx, 1, a1.ID); !acct.NeedsReconnect {
		t.Error("revoked account should be flagged needs_reconnect, not deleted")
	}

	// Reconnect: a wrong password is rejected and changes nothing; the right one clears the flag.
	if err := svc.Reconnect(ctx, 1, a1.ID, "still-wrong"); !errors.Is(err, mail.ErrAuth) {
		t.Errorf("reconnect with wrong password: want ErrAuth, got %v", err)
	}
	if err := svc.Reconnect(ctx, 1, a1.ID, "pw1"); err != nil {
		t.Fatal(err)
	}
	if acct, _ := store.Account(ctx, 1, a1.ID); acct.NeedsReconnect {
		t.Error("flag should clear after reconnect")
	}
	if msgs, _, _ = svc.Messages(ctx, 1, 0); len(msgs) != 2 {
		t.Errorf("after reconnect = %+v", msgs)
	}

	// Disconnect removes the account's mail from the list.
	if err := store.Delete(ctx, 1, a2.ID); err != nil {
		t.Fatal(err)
	}
	if msgs, _, _ = svc.Messages(ctx, 1, 0); len(msgs) != 1 || msgs[0].AccountID != a1.ID {
		t.Errorf("after disconnect = %+v", msgs)
	}
}

type fakeIMAP struct {
	probeErr error
	probed   []mail.Credentials
}

func (f *fakeIMAP) Probe(_ context.Context, c mail.Credentials) error {
	f.probed = append(f.probed, c)
	return f.probeErr
}
func (f *fakeIMAP) Unseen(context.Context, mail.Credentials, int) ([]mail.Message, error) {
	return nil, nil
}

func TestService_Connect(t *testing.T) {
	ctx := context.Background()
	store := newStore(t)
	fake := &fakeIMAP{}
	svc := &mail.Service{Store: store, IMAP: fake}

	acct, err := svc.Connect(ctx, 1, "  Me@Gmail.com ", "pw", "")
	if err != nil {
		t.Fatal(err)
	}
	if acct.Email != "me@gmail.com" || acct.IMAPHost != "imap.gmail.com" || acct.IMAPPort != 993 {
		t.Errorf("preset auto-fill: %+v", acct)
	}
	if _, err := svc.Connect(ctx, 1, "me@corp.example", "pw", ""); !errors.Is(err, mail.ErrHostRequired) {
		t.Errorf("unknown domain without host: want ErrHostRequired, got %v", err)
	}
	for _, bad := range [][2]string{{"", "pw"}, {"no-at", "pw"}, {"@x.com", "pw"}, {"a@", "pw"}, {"a@x.com", ""}} {
		if _, err := svc.Connect(ctx, 1, bad[0], bad[1], "h"); !errors.Is(err, mail.ErrInvalidAccount) {
			t.Errorf("Connect(%q,%q): want ErrInvalidAccount, got %v", bad[0], bad[1], err)
		}
	}

	// A failed test login must store nothing.
	fake.probeErr = mail.ErrAuth
	before, _ := store.List(ctx, 1)
	if _, err := svc.Connect(ctx, 1, "other@gmail.com", "bad", ""); !errors.Is(err, mail.ErrAuth) {
		t.Fatalf("want ErrAuth, got %v", err)
	}
	after, _ := store.List(ctx, 1)
	if len(after) != len(before) {
		t.Error("a rejected login must not be saved")
	}
}
