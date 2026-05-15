package apns

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"
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
