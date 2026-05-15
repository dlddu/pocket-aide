package githubwebhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(t *testing.T, secret, body []byte) string {
	t.Helper()
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyHMAC_Accepts(t *testing.T) {
	secret := []byte("super-secret")
	body := []byte(`{"action":"completed"}`)
	if err := verifyHMAC(secret, body, sign(t, secret, body)); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestVerifyHMAC_RejectsWrongSecret(t *testing.T) {
	secret := []byte("super-secret")
	body := []byte(`{"action":"completed"}`)
	bad := sign(t, []byte("other-secret"), body)
	if err := verifyHMAC(secret, body, bad); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyHMAC_RejectsTamperedPayload(t *testing.T) {
	secret := []byte("super-secret")
	body := []byte(`{"action":"completed"}`)
	sig := sign(t, secret, body)
	if err := verifyHMAC(secret, []byte(`{"action":"queued"}`), sig); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyHMAC_RejectsMissingHeader(t *testing.T) {
	if err := verifyHMAC([]byte("s"), []byte(`{}`), ""); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyHMAC_RejectsBadPrefix(t *testing.T) {
	if err := verifyHMAC([]byte("s"), []byte(`{}`), "sha1=deadbeef"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyHMAC_RejectsNonHex(t *testing.T) {
	if err := verifyHMAC([]byte("s"), []byte(`{}`), "sha256=zzzz"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestVerifyHMAC_RejectsEmptySecret(t *testing.T) {
	if err := verifyHMAC(nil, []byte(`{}`), "sha256=00"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
