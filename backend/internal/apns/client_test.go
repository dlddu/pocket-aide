package apns

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sideshow/apns2"
)

// generateP8 produces an ECDSA P-256 PKCS#8 PEM resembling an Apple-issued
// .p8 key. AuthKeyFromBytes accepts any P-256 PKCS#8 key, so this is enough
// to exercise the parsing path without hitting Apple.
func generateP8(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("genkey: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func TestNew_ParsesValidKey(t *testing.T) {
	c, err := New("KEYID", "TEAMID", "com.example.app", generateP8(t), false)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if c == nil || c.cli == nil {
		t.Fatal("expected non-nil client")
	}
	if c.bundleID != "com.example.app" {
		t.Errorf("bundleID mismatch: %s", c.bundleID)
	}
}

func TestNew_RejectsEmptyPEM(t *testing.T) {
	if _, err := New("k", "t", "b", nil, false); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_RejectsGarbagePEM(t *testing.T) {
	if _, err := New("k", "t", "b", []byte("not a pem"), false); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNew_RejectsMissingMetadata(t *testing.T) {
	pem := generateP8(t)
	cases := []struct {
		name             string
		k, team, bundle  string
		wantNonNilClient bool
	}{
		{"no key id", "", "t", "b", false},
		{"no team id", "k", "", "b", false},
		{"no bundle id", "k", "t", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New(tc.k, tc.team, tc.bundle, pem, false); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestNew_KeepsAppleHostWithoutOverride(t *testing.T) {
	dev, err := New("KEYID", "TEAMID", "com.example.app", generateP8(t), false)
	if err != nil {
		t.Fatal(err)
	}
	prod, err := New("KEYID", "TEAMID", "com.example.app", generateP8(t), true)
	if err != nil {
		t.Fatal(err)
	}
	if dev.cli.Host != apns2.HostDevelopment || prod.cli.Host != apns2.HostProduction {
		t.Fatalf("hosts = %q, %q", dev.cli.Host, prod.cli.Host)
	}
}

func TestWithHost_SendsToPlainHTTPReceiver(t *testing.T) {
	type received struct {
		path, topic, auth string
		body              map[string]any
	}
	got := make(chan received, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		got <- received{r.URL.Path, r.Header.Get("apns-topic"), r.Header.Get("authorization"), body}
		w.Header().Set("apns-id", "fake-id")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := New("KEYID", "TEAMID", "com.example.app", generateP8(t), false)
	if err != nil {
		t.Fatal(err)
	}
	c = c.WithHost(srv.URL + "/")
	if err := c.SendWithData(context.Background(), "abc123", "CI 실패", "body", map[string]any{"event_id": 7}); err != nil {
		t.Fatalf("send: %v", err)
	}
	r := <-got
	if r.path != "/3/device/abc123" || r.topic != "com.example.app" || !strings.HasPrefix(r.auth, "bearer ") {
		t.Fatalf("request = %+v", r)
	}
	alert, _ := r.body["aps"].(map[string]any)["alert"].(map[string]any)
	if alert["title"] != "CI 실패" || r.body["event_id"] != float64(7) {
		t.Fatalf("payload = %v", r.body)
	}
}

func TestWithHost_ReportsRejection(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"reason":"BadDeviceToken"}`)
	}))
	defer srv.Close()

	c, err := New("KEYID", "TEAMID", "com.example.app", generateP8(t), false)
	if err != nil {
		t.Fatal(err)
	}
	err = c.WithHost(srv.URL).Send(context.Background(), "bad", "t", "b")
	if err == nil || !strings.Contains(err.Error(), "BadDeviceToken") {
		t.Fatalf("err = %v", err)
	}
}
