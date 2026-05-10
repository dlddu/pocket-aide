// Package oidcmock provides an OpenID Connect mock server for tests and local
// development. It implements the minimum surface of an OIDC IdP: discovery,
// JWKS, authorization (with PKCE), and token endpoints. Tokens are signed with
// an in-memory RSA key.
package oidcmock

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultClientID = "pocket-aide-ios-dev"
	defaultAudience = "pocket-aide-dev"
	defaultSubject  = "mock-user-123"
	keyID           = "pocket-aide-mock-key"
)

// Server is an in-process OIDC mock. Build it with New, expose the handler with
// Handler, and register issuer overrides via SetIssuer when running on a
// pre-existing host like httptest.Server.
type Server struct {
	mu        sync.Mutex
	issuer    string
	clientID  string
	audience  string
	subject   string
	signKey   *rsa.PrivateKey
	codes     map[string]authCode
	authDelay time.Duration
}

type authCode struct {
	codeChallenge       string
	codeChallengeMethod string
	clientID            string
	redirectURI         string
	scope               string
	createdAt           time.Time
}

// Options tweak the mock's defaults. Zero value is fine for tests.
type Options struct {
	Issuer   string
	ClientID string
	Audience string
	Subject  string
}

// New returns a Server with a freshly generated RSA key.
func New(opts Options) (*Server, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate rsa key: %w", err)
	}
	s := &Server{
		issuer:   opts.Issuer,
		clientID: orDefault(opts.ClientID, defaultClientID),
		audience: orDefault(opts.Audience, defaultAudience),
		subject:  orDefault(opts.Subject, defaultSubject),
		signKey:  priv,
		codes:    make(map[string]authCode),
	}
	return s, nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// SetIssuer updates the issuer URL after the server is bound to a real HTTP
// listener (e.g. httptest.Server.URL).
func (s *Server) SetIssuer(issuer string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.issuer = issuer
}

// Issuer returns the currently configured issuer URL.
func (s *Server) Issuer() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.issuer
}

// ClientID returns the configured client id.
func (s *Server) ClientID() string { return s.clientID }

// Audience returns the configured audience.
func (s *Server) Audience() string { return s.audience }

// Subject returns the subject claim used for minted tokens.
func (s *Server) Subject() string { return s.subject }

// Handler returns the HTTP handler implementing the OIDC endpoints.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", s.handleDiscovery)
	mux.HandleFunc("/jwks.json", s.handleJWKS)
	mux.HandleFunc("/authorize", s.handleAuthorize)
	mux.HandleFunc("/token", s.handleToken)
	return mux
}

func (s *Server) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	issuer := s.Issuer()
	doc := map[string]any{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/authorize",
		"token_endpoint":                        issuer + "/token",
		"jwks_uri":                              issuer + "/jwks.json",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"code_challenge_methods_supported":      []string{"S256"},
	}
	writeJSON(w, http.StatusOK, doc)
}

func (s *Server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	pub := &s.signKey.PublicKey
	jwk := map[string]any{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": keyID,
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
	writeJSON(w, http.StatusOK, map[string]any{"keys": []any{jwk}})
}

func (s *Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	q := r.Form
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")
	scope := q.Get("scope")
	responseType := q.Get("response_type")

	if responseType != "code" {
		http.Error(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	if clientID != s.clientID {
		http.Error(w, "unknown client_id", http.StatusBadRequest)
		return
	}
	if redirectURI == "" {
		http.Error(w, "missing redirect_uri", http.StatusBadRequest)
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		http.Error(w, "PKCE S256 required", http.StatusBadRequest)
		return
	}

	code := randomString(32)
	s.mu.Lock()
	s.codes[code] = authCode{
		codeChallenge:       codeChallenge,
		codeChallengeMethod: codeChallengeMethod,
		clientID:            clientID,
		redirectURI:         redirectURI,
		scope:               scope,
		createdAt:           time.Now(),
	}
	s.mu.Unlock()

	cb, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	qs := cb.Query()
	qs.Set("code", code)
	if state != "" {
		qs.Set("state", state)
	}
	cb.RawQuery = qs.Encode()
	http.Redirect(w, r, cb.String(), http.StatusFound)
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	grant := r.PostForm.Get("grant_type")
	switch grant {
	case "authorization_code":
		s.handleTokenAuthCode(w, r)
	case "refresh_token":
		s.handleTokenRefresh(w, r)
	default:
		http.Error(w, "unsupported grant_type", http.StatusBadRequest)
	}
}

func (s *Server) handleTokenAuthCode(w http.ResponseWriter, r *http.Request) {
	code := r.PostForm.Get("code")
	verifier := r.PostForm.Get("code_verifier")
	clientID := r.PostForm.Get("client_id")
	redirectURI := r.PostForm.Get("redirect_uri")

	s.mu.Lock()
	stored, ok := s.codes[code]
	if ok {
		delete(s.codes, code)
	}
	s.mu.Unlock()
	if !ok {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}
	if clientID != stored.clientID || redirectURI != stored.redirectURI {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}
	if !verifyPKCE(verifier, stored.codeChallenge) {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}

	resp, err := s.mintTokenResponse(stored.scope)
	if err != nil {
		http.Error(w, "token mint failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	rt := r.PostForm.Get("refresh_token")
	if rt == "" {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
		return
	}
	resp, err := s.mintTokenResponse("openid profile email")
	if err != nil {
		http.Error(w, "token mint failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) mintTokenResponse(scope string) (map[string]any, error) {
	access, err := s.SignAccessToken(s.subject, time.Hour)
	if err != nil {
		return nil, err
	}
	idTok, err := s.signIDToken(s.subject, time.Hour)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"access_token":  access,
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": randomString(48),
		"id_token":      idTok,
		"scope":         scope,
	}, nil
}

// SignAccessToken mints a signed access token with the given subject and TTL.
// Exposed so integration tests can build tokens without going through the full
// auth-code flow.
func (s *Server) SignAccessToken(sub string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": s.Issuer(),
		"sub": sub,
		"aud": s.audience,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
		"nbf": now.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = keyID
	return tok.SignedString(s.signKey)
}

func (s *Server) signIDToken(sub string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": s.Issuer(),
		"sub": sub,
		"aud": s.clientID,
		"exp": now.Add(ttl).Unix(),
		"iat": now.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = keyID
	return tok.SignedString(s.signKey)
}

func verifyPKCE(verifier, challenge string) bool {
	if verifier == "" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	enc := base64.RawURLEncoding.EncodeToString(sum[:])
	return enc == challenge
}

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// PEMPublicKey returns the PEM-encoded public key. Useful for debugging.
func (s *Server) PEMPublicKey() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(&s.signKey.PublicKey)
	if err != nil {
		return "", err
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: der}
	var sb strings.Builder
	if err := pem.Encode(&sb, block); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// ErrNoCode is returned when an auth code cannot be found.
var ErrNoCode = errors.New("auth code not found")
