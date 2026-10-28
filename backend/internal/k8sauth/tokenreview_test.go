package k8sauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

const issuer = "https://kubernetes.default.svc.cluster.local"

func jwt(t *testing.T, claims map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	return "eyJhbGciOiJSUzI1NiJ9." + base64.RawURLEncoding.EncodeToString(raw) + ".sig"
}

type apiserver struct {
	calls    atomic.Int32
	status   int
	username string
	auds     []string
	authed   bool
	lastBody tokenReview
	lastAuth string
}

func (a *apiserver) start(t *testing.T) *Reviewer {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.calls.Add(1)
		if r.Method != http.MethodPost || r.URL.Path != "/apis/authentication.k8s.io/v1/tokenreviews" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		a.lastAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&a.lastBody)
		if a.status != 0 {
			w.WriteHeader(a.status)
			return
		}
		out := a.lastBody
		out.Status.Authenticated = a.authed
		out.Status.User.Username = a.username
		out.Status.Audiences = a.auds
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(out)
	}))
	t.Cleanup(srv.Close)
	return New(Config{APIURL: srv.URL, Audience: "pocket-aide", Issuer: issuer, Token: func() (string, error) { return "own-token", nil }})
}

func TestReviewReturnsServiceAccountAndSendsAudience(t *testing.T) {
	api := &apiserver{authed: true, username: "system:serviceaccount:homelab-k3s-mcp:homelab-k3s-mcp", auds: []string{"pocket-aide"}}
	r := api.start(t)
	tok := jwt(t, map[string]any{"iss": issuer, "aud": []string{"pocket-aide"}})
	sa, err := r.Review(context.Background(), tok)
	if err != nil || sa != "homelab-k3s-mcp/homelab-k3s-mcp" {
		t.Fatalf("Review = %q, %v", sa, err)
	}
	if api.lastAuth != "Bearer own-token" || api.lastBody.Spec.Token != tok || len(api.lastBody.Spec.Audiences) != 1 || api.lastBody.Spec.Audiences[0] != "pocket-aide" {
		t.Fatalf("request auth=%q spec=%+v", api.lastAuth, api.lastBody.Spec)
	}
}

func TestReviewSkipsApiserverForTokensThatCannotPass(t *testing.T) {
	api := &apiserver{authed: true, username: "system:serviceaccount:a:b", auds: []string{"pocket-aide"}}
	r := api.start(t)
	for name, tok := range map[string]string{
		"not a jwt":       "pak_abc",
		"other issuer":    jwt(t, map[string]any{"iss": "https://evil.example", "aud": "pocket-aide"}),
		"apiserver token": jwt(t, map[string]any{"iss": issuer, "aud": []string{"https://kubernetes.default.svc.cluster.local"}}),
		"bad payload":     "a.!!!.c",
	} {
		if _, err := r.Review(context.Background(), tok); !errors.Is(err, ErrUnauthenticated) {
			t.Errorf("%s: err = %v, want ErrUnauthenticated", name, err)
		}
	}
	if n := api.calls.Load(); n != 0 {
		t.Fatalf("apiserver calls = %d, want 0", n)
	}
}

func TestReviewRejectsWhatTheApiserverDoesNotVouchFor(t *testing.T) {
	tok := jwt(t, map[string]any{"iss": issuer, "aud": "pocket-aide"})
	for name, api := range map[string]*apiserver{
		"not authenticated": {authed: false},
		"audience dropped":  {authed: true, username: "system:serviceaccount:a:b", auds: []string{"other"}},
		"not an SA":         {authed: true, username: "alice", auds: []string{"pocket-aide"}},
		"malformed SA":      {authed: true, username: "system:serviceaccount:a", auds: []string{"pocket-aide"}},
	} {
		if _, err := api.start(t).Review(context.Background(), tok); !errors.Is(err, ErrUnauthenticated) {
			t.Errorf("%s: err = %v, want ErrUnauthenticated", name, err)
		}
	}
	if _, err := (&apiserver{status: http.StatusForbidden}).start(t).Review(context.Background(), tok); !errors.Is(err, ErrUnavailable) {
		t.Errorf("403 from apiserver: err = %v, want ErrUnavailable", err)
	}
}

func TestReviewCachesVerdictsButNotOutages(t *testing.T) {
	api := &apiserver{authed: true, username: "system:serviceaccount:a:b", auds: []string{"pocket-aide"}}
	r := api.start(t)
	now := time.Unix(1_800_000_000, 0)
	r.now = func() time.Time { return now }
	tok := jwt(t, map[string]any{"iss": issuer, "aud": "pocket-aide", "exp": now.Add(time.Hour).Unix()})

	for i := 0; i < 3; i++ {
		if _, err := r.Review(context.Background(), tok); err != nil {
			t.Fatal(err)
		}
	}
	if n := api.calls.Load(); n != 1 {
		t.Fatalf("calls within TTL = %d, want 1", n)
	}
	now = now.Add(cacheTTL)
	if _, err := r.Review(context.Background(), tok); err != nil || api.calls.Load() != 2 {
		t.Fatalf("after TTL: err=%v calls=%d, want a fresh review", err, api.calls.Load())
	}

	api.status = http.StatusInternalServerError
	other := jwt(t, map[string]any{"iss": issuer, "aud": "pocket-aide", "sub": "x"})
	for i := 0; i < 2; i++ {
		if _, err := r.Review(context.Background(), other); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("outage: %v", err)
		}
	}
	if n := api.calls.Load(); n != 4 {
		t.Fatalf("calls during outage = %d, want every attempt to reach the apiserver", n)
	}

	api.status = 0
	expiring := jwt(t, map[string]any{"iss": issuer, "aud": "pocket-aide", "exp": now.Add(10 * time.Second).Unix()})
	_, _ = r.Review(context.Background(), expiring)
	now = now.Add(11 * time.Second)
	_, _ = r.Review(context.Background(), expiring)
	if n := api.calls.Load(); n != 6 {
		t.Fatalf("cache outlived token exp: calls = %d, want 6", n)
	}
}
