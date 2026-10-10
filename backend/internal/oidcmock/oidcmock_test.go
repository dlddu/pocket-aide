package oidcmock_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dlddu/pocket-aide/backend/internal/oidcmock"
)

const (
	testClientID    = "pocket-aide-ios-dev"
	testRedirectURI = "pocketaide-dev://callback"
	testVerifier    = "0123456789abcdef0123456789abcdef0123456789abcdef"
)

func newMock(t *testing.T) *httptest.Server {
	t.Helper()
	mock, err := oidcmock.New(oidcmock.Options{ClientID: testClientID, Subject: "default-user"})
	if err != nil {
		t.Fatalf("oidcmock new: %v", err)
	}
	ts := httptest.NewServer(mock.Handler())
	t.Cleanup(ts.Close)
	mock.SetIssuer(ts.URL)
	return ts
}

func authorize(t *testing.T, ts *httptest.Server, loginHint string) string {
	t.Helper()
	sum := sha256.Sum256([]byte(testVerifier))
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {testClientID},
		"redirect_uri":          {testRedirectURI},
		"scope":                 {"openid profile email"},
		"state":                 {"s"},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(sum[:])},
		"code_challenge_method": {"S256"},
	}
	if loginHint != "" {
		q.Set("login_hint", loginHint)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Get(ts.URL + "/authorize?" + q.Encode())
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("authorize status: got %d want %d", resp.StatusCode, http.StatusFound)
	}
	loc, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	code := loc.Query().Get("code")
	if code == "" {
		t.Fatalf("authorize redirect has no code: %s", loc)
	}
	return code
}

func tokenSubjects(t *testing.T, ts *httptest.Server, code string) (string, string) {
	t.Helper()
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {testVerifier},
		"client_id":     {testClientID},
		"redirect_uri":  {testRedirectURI},
	}
	resp, err := http.Post(ts.URL+"/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("token status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	var body struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	return subjectOf(t, body.AccessToken), subjectOf(t, body.IDToken)
}

func subjectOf(t *testing.T, raw string) string {
	t.Helper()
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(raw, claims); err != nil {
		t.Fatalf("parse token: %v", err)
	}
	sub, _ := claims["sub"].(string)
	return sub
}

func TestAuthorizationCodeUsesDefaultSubjectWithoutLoginHint(t *testing.T) {
	ts := newMock(t)
	access, id := tokenSubjects(t, ts, authorize(t, ts, ""))
	if access != "default-user" || id != "default-user" {
		t.Fatalf("subjects: got access=%q id=%q want default-user", access, id)
	}
}

func TestAuthorizationCodeUsesLoginHintAsSubject(t *testing.T) {
	ts := newMock(t)
	access, id := tokenSubjects(t, ts, authorize(t, ts, "second-user"))
	if access != "second-user" || id != "second-user" {
		t.Fatalf("subjects: got access=%q id=%q want second-user", access, id)
	}
	access, _ = tokenSubjects(t, ts, authorize(t, ts, ""))
	if access != "default-user" {
		t.Fatalf("a later request without login_hint should fall back to the default subject, got %q", access)
	}
}

func setNextLoginSubject(t *testing.T, ts *httptest.Server, sub string) {
	t.Helper()
	resp, err := http.PostForm(ts.URL+"/e2e/next-login-subject", url.Values{"sub": {sub}})
	if err != nil {
		t.Fatalf("next-login-subject: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("next-login-subject status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestNextLoginSubjectAppliesToOneAuthorizationWithoutLoginHint(t *testing.T) {
	ts := newMock(t)
	setNextLoginSubject(t, ts, "app-second-user")
	access, id := tokenSubjects(t, ts, authorize(t, ts, ""))
	if access != "app-second-user" || id != "app-second-user" {
		t.Fatalf("subjects: got access=%q id=%q want app-second-user", access, id)
	}
	access, _ = tokenSubjects(t, ts, authorize(t, ts, ""))
	if access != "default-user" {
		t.Fatalf("the next-login subject should be spent by one authorization, got %q", access)
	}
}

func TestLoginHintWinsOverNextLoginSubjectAndLeavesItPending(t *testing.T) {
	ts := newMock(t)
	setNextLoginSubject(t, ts, "app-second-user")
	access, _ := tokenSubjects(t, ts, authorize(t, ts, "runner-user"))
	if access != "runner-user" {
		t.Fatalf("login_hint should decide the subject, got %q", access)
	}
	access, _ = tokenSubjects(t, ts, authorize(t, ts, ""))
	if access != "app-second-user" {
		t.Fatalf("the pending next-login subject should survive a login_hint request, got %q", access)
	}
}

func TestEmptyNextLoginSubjectClearsThePendingOne(t *testing.T) {
	ts := newMock(t)
	setNextLoginSubject(t, ts, "app-second-user")
	setNextLoginSubject(t, ts, "")
	access, _ := tokenSubjects(t, ts, authorize(t, ts, ""))
	if access != "default-user" {
		t.Fatalf("clearing should restore the default subject, got %q", access)
	}
}

func TestNextLoginSubjectRejectsGet(t *testing.T) {
	ts := newMock(t)
	resp, err := http.Get(ts.URL + "/e2e/next-login-subject?sub=x")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}
