package githubwebhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// verifyHMAC checks the GitHub `X-Hub-Signature-256` header against an
// HMAC-SHA256 of the raw payload, using the shared secret. The header looks
// like `sha256=<hex>`. Constant-time comparison via hmac.Equal.
func verifyHMAC(secret, payload []byte, sigHeader string) error {
	if len(secret) == 0 {
		return errors.New("github webhook secret is empty")
	}
	if sigHeader == "" {
		return errors.New("missing X-Hub-Signature-256 header")
	}
	if !strings.HasPrefix(sigHeader, "sha256=") {
		return errors.New("signature must start with sha256=")
	}
	want, err := hex.DecodeString(strings.TrimPrefix(sigHeader, "sha256="))
	if err != nil {
		return errors.New("signature is not valid hex")
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	got := mac.Sum(nil)
	if !hmac.Equal(got, want) {
		return errors.New("signature mismatch")
	}
	return nil
}
