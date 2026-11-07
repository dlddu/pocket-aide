package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/githubwebhook"
	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

func TestFormatPushText(t *testing.T) {
	cases := []struct {
		name      string
		evt       githubwebhook.WorkflowRunEvent
		wantTitle string
		wantBody  string
	}{
		{
			name: "start event with PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "feature/x",
				Conclusion: "queued", PRNumber: 42, PRTitle: "feat(x): hello",
			},
			wantTitle: "CI 시작 — dlddu/pocket-aide #42",
			wantBody:  "feat(x): hello",
		},
		{
			name: "start event in_progress without PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "main",
				Conclusion: "in_progress",
			},
			wantTitle: "dlddu/pocket-aide — in_progress",
			wantBody:  "CI on main",
		},
		{
			name: "completed success with PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "feature/x",
				Conclusion: "success", PRNumber: 7, PRTitle: "fix: thing",
			},
			wantTitle: "CI 통과 — dlddu/pocket-aide #7",
			wantBody:  "fix: thing",
		},
		{
			name: "completed failure without PR",
			evt: githubwebhook.WorkflowRunEvent{
				Repo: "dlddu/pocket-aide", WorkflowName: "CI", HeadBranch: "main",
				Conclusion: "failure",
			},
			wantTitle: "dlddu/pocket-aide — failure",
			wantBody:  "CI on main",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			title, body := formatPushText(tc.evt)
			if title != tc.wantTitle {
				t.Errorf("title: got %q want %q", title, tc.wantTitle)
			}
			if body != tc.wantBody {
				t.Errorf("body: got %q want %q", body, tc.wantBody)
			}
		})
	}
}

func TestShouldPush(t *testing.T) {
	cases := []struct {
		name string
		evt  githubwebhook.WorkflowRunEvent
		want bool
	}{
		{"start queued is history only", githubwebhook.WorkflowRunEvent{Conclusion: "queued"}, false},
		{"start in_progress is history only", githubwebhook.WorkflowRunEvent{Conclusion: "in_progress"}, false},
		{"completed success", githubwebhook.WorkflowRunEvent{Conclusion: "success", Completed: true}, true},
		{"completed failure", githubwebhook.WorkflowRunEvent{Conclusion: "failure", Completed: true}, true},
		{"completed cancelled", githubwebhook.WorkflowRunEvent{Conclusion: "cancelled", Completed: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldPush(tc.evt); got != tc.want {
				t.Errorf("shouldPush: got %v want %v", got, tc.want)
			}
		})
	}
}

func TestLoadConfigAPNSDisabled(t *testing.T) {
	t.Setenv("OIDC_ISSUER", "http://issuer")
	t.Setenv("OIDC_AUDIENCE", "aud")
	t.Setenv("OIDC_CLIENT_ID", "client")
	t.Setenv("OIDC_REDIRECT_URI", "app://cb")
	t.Setenv("SQS_QUEUE_URL", "http://localhost:4566/000000000000/q")
	t.Setenv("APNS_DISABLED", "true")

	c := loadConfig()
	if !c.PRMonitorEnabled {
		t.Fatalf("PRMonitorEnabled = false, want true")
	}
	if !c.APNSDisabled {
		t.Fatalf("APNSDisabled = false, want true")
	}
	if c.APNSKeyID != "" || c.APNSAuthKeyP8 != "" {
		t.Fatalf("APNs credentials should stay empty when disabled, got key=%q", c.APNSKeyID)
	}
}

type staticTokens string

func (s staticTokens) Token(context.Context) (string, error) { return string(s), nil }

func TestTenantPlatformClient(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"sessions":[]}`))
	}))
	defer srv.Close()

	var asked []string
	tokensFor := func(account string) (sessions.TokenSource, error) {
		asked = append(asked, account)
		return staticTokens("tok-" + account), nil
	}

	tenant, err := tenantPlatformClient(srv.URL, "pocket-aide-sessions", tokensFor)
	if err != nil {
		t.Fatal(err)
	}
	operator, err := tenantPlatformClient(srv.URL, "", tokensFor)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []*sessions.Client{tenant, operator} {
		if _, _, err := c.Do(context.Background(), http.MethodGet, "/sessions", nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(asked) != 1 || asked[0] != "pocket-aide-sessions" {
		t.Fatalf("token sources requested for %v, want only pocket-aide-sessions", asked)
	}
	if len(got) != 2 || got[0] != "Bearer tok-pocket-aide-sessions" || got[1] != "" {
		t.Fatalf("Authorization headers = %q, want tenant bearer then none", got)
	}
}

func TestTenantPlatformClientTokenError(t *testing.T) {
	_, err := tenantPlatformClient("http://unused", "pocket-aide-sessions", func(string) (sessions.TokenSource, error) {
		return nil, errors.New("not in cluster")
	})
	if err == nil {
		t.Fatal("token source failure must stop startup rather than fall back to the operator view")
	}
}
