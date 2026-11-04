// Package mail is the backend half of the widget mail area (PRD-8 AC3, AC11):
// connected IMAP accounts, their encrypted app passwords, and a read-only
// unseen-mail fetch.
package mail

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// ErrBadCiphertext is returned by Cipher.Decrypt for anything it cannot open:
// malformed text, an unknown key version, or a failed authentication tag
// (wrong or lost key).
var ErrBadCiphertext = errors.New("mail: cannot decrypt stored password")

// Cipher seals app passwords with AES-256-GCM. The sealed form is
// "<version>:<base64(nonce||ciphertext)>" so a later key rotation or a move to
// an external KMS can tell which key a row was sealed with.
type Cipher struct {
	version string
	aead    cipher.AEAD
}

// NewCipher builds a Cipher from a base64-encoded 32-byte key.
func NewCipher(version, keyB64 string) (*Cipher, error) {
	if version == "" || strings.Contains(version, ":") {
		return nil, errors.New("mail: key version must be non-empty and contain no ':'")
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(keyB64))
	if err != nil {
		return nil, fmt.Errorf("mail: key is not valid base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("mail: key must be 32 bytes (AES-256), got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("mail: aes: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("mail: gcm: %w", err)
	}
	return &Cipher{version: version, aead: aead}, nil
}

// Encrypt seals plaintext with a fresh random nonce.
func (c *Cipher) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("mail: nonce: %w", err)
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(plaintext), []byte(c.version))
	return c.version + ":" + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt opens a value produced by Encrypt.
func (c *Cipher) Decrypt(sealed string) (string, error) {
	version, body, ok := strings.Cut(sealed, ":")
	if !ok || version != c.version {
		return "", ErrBadCiphertext
	}
	raw, err := base64.StdEncoding.DecodeString(body)
	if err != nil || len(raw) < c.aead.NonceSize() {
		return "", ErrBadCiphertext
	}
	nonce, ct := raw[:c.aead.NonceSize()], raw[c.aead.NonceSize():]
	plain, err := c.aead.Open(nil, nonce, ct, []byte(version))
	if err != nil {
		return "", ErrBadCiphertext
	}
	return string(plain), nil
}
