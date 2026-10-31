package fcm_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"golang.org/x/oauth2"

	"github.com/dlddu/pocket-aide/backend/internal/fcm"
)

type captured struct {
	path, auth string
	body       map[string]any
}

func fcmServer(t *testing.T, status int, respBody string) (*httptest.Server, *captured) {
	t.Helper()
	got := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path = r.URL.Path
		got.auth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &got.body); err != nil {
			t.Errorf("request body is not json: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, respBody)
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

func staticClient(t *testing.T, endpoint string) *fcm.Client {
	t.Helper()
	c, err := fcm.NewWithTokenSource("demo-project", oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-access", TokenType: "Bearer"}))
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	return c.WithEndpoint(endpoint)
}

func TestSendWithCategoryPostsTheV1Message(t *testing.T) {
	srv, got := fcmServer(t, http.StatusOK, `{"name":"projects/demo-project/messages/1"}`)
	c := staticClient(t, srv.URL).WithChannels(map[string]string{"APPROVAL_REQUEST": fcm.ChannelApproval})

	err := c.SendWithCategory(context.Background(), "reg-token", "APPROVAL_REQUEST", "승인 요청 · infra-ci", "prod 배포",
		map[string]any{"approval_id": "req-1", "event_id": int64(42), "flag": true})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if got.path != "/v1/projects/demo-project/messages:send" {
		t.Fatalf("path: %q", got.path)
	}
	if got.auth != "Bearer test-access" {
		t.Fatalf("authorization: %q", got.auth)
	}
	want := map[string]any{
		"message": map[string]any{
			"token":        "reg-token",
			"notification": map[string]any{"title": "승인 요청 · infra-ci", "body": "prod 배포"},
			"data": map[string]any{
				"approval_id": "req-1",
				"event_id":    "42",
				"flag":        "true",
				"category":    "APPROVAL_REQUEST",
			},
			"android": map[string]any{
				"priority":     "high",
				"notification": map[string]any{"channel_id": "approval"},
			},
		},
	}
	if !reflect.DeepEqual(got.body, want) {
		t.Fatalf("body:\n got %v\nwant %v", got.body, want)
	}
}

func TestSendWithDataUsesTheDefaultChannelAndNoCategory(t *testing.T) {
	srv, got := fcmServer(t, http.StatusOK, `{}`)
	c := staticClient(t, srv.URL).WithChannels(map[string]string{"APPROVAL_REQUEST": fcm.ChannelApproval})

	if err := c.SendWithData(context.Background(), "reg-token", "t", "b", map[string]any{"kind": "external_notification", "notification_id": int64(7)}); err != nil {
		t.Fatalf("send: %v", err)
	}
	msg := got.body["message"].(map[string]any)
	if !reflect.DeepEqual(msg["data"], map[string]any{"kind": "external_notification", "notification_id": "7"}) {
		t.Fatalf("data: %v", msg["data"])
	}
	if ch := msg["android"].(map[string]any)["notification"].(map[string]any)["channel_id"]; ch != "default" {
		t.Fatalf("channel: %v", ch)
	}
}

func TestRejectionsCarryTheFCMErrorCode(t *testing.T) {
	for name, tc := range map[string]struct {
		status  int
		body    string
		code    string
		invalid bool
	}{
		"unregistered":     {404, `{"error":{"code":404,"message":"Requested entity was not found.","status":"NOT_FOUND","details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`, "UNREGISTERED", true},
		"sender mismatch":  {403, `{"error":{"code":403,"status":"PERMISSION_DENIED","details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"SENDER_ID_MISMATCH"}]}}`, "SENDER_ID_MISMATCH", true},
		"bare not found":   {404, `not json`, "", true},
		"invalid argument": {400, `{"error":{"code":400,"status":"INVALID_ARGUMENT","details":[{"@type":"type.googleapis.com/google.rpc.BadRequest"},{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"INVALID_ARGUMENT"}]}}`, "INVALID_ARGUMENT", false},
		"quota":            {429, `{"error":{"code":429,"status":"RESOURCE_EXHAUSTED","details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"QUOTA_EXCEEDED"}]}}`, "QUOTA_EXCEEDED", false},
		"unavailable":      {503, `{"error":{"code":503,"status":"UNAVAILABLE"}}`, "", false},
	} {
		t.Run(name, func(t *testing.T) {
			srv, _ := fcmServer(t, tc.status, tc.body)
			err := staticClient(t, srv.URL).SendWithData(context.Background(), "reg-token", "t", "b", nil)
			var rej *fcm.RejectedError
			if !errors.As(err, &rej) {
				t.Fatalf("want *fcm.RejectedError, got %T %v", err, err)
			}
			if rej.StatusCode != tc.status || rej.ErrorCode != tc.code {
				t.Fatalf("rejected = %+v, want status %d code %q", rej, tc.status, tc.code)
			}
			if rej.TokenInvalid() != tc.invalid {
				t.Fatalf("TokenInvalid() = %v, want %v", rej.TokenInvalid(), tc.invalid)
			}
		})
	}
}

func TestNewReadsTheProjectAndSignsATokenRequest(t *testing.T) {
	var tokenForm url.Values
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		tokenForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"minted","token_type":"Bearer","expires_in":3600}`)
	}))
	t.Cleanup(tokenSrv.Close)
	srv, got := fcmServer(t, http.StatusOK, `{}`)

	c, err := fcm.New(serviceAccountJSON(t, tokenSrv.URL))
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if c.ProjectID() != "sa-project" {
		t.Fatalf("project: %q", c.ProjectID())
	}
	if err := c.WithEndpoint(srv.URL).SendWithData(context.Background(), "reg-token", "t", "b", nil); err != nil {
		t.Fatalf("send: %v", err)
	}
	if got.path != "/v1/projects/sa-project/messages:send" || got.auth != "Bearer minted" {
		t.Fatalf("path=%q auth=%q", got.path, got.auth)
	}
	if tokenForm.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || tokenForm.Get("assertion") == "" {
		t.Fatalf("token request form: %v", tokenForm)
	}
}

func TestNewRejectsIncompleteCredentials(t *testing.T) {
	for name, raw := range map[string]string{
		"not json":   `{`,
		"no project": `{"type":"service_account","client_email":"a@b","private_key":"k"}`,
		"no key":     `{"type":"service_account","project_id":"p","client_email":"a@b"}`,
		"user creds": `{"type":"authorized_user","project_id":"p","client_email":"a@b","private_key":"k"}`,
	} {
		if _, err := fcm.New([]byte(raw)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func serviceAccountJSON(t *testing.T, tokenURI string) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("genkey: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	raw, err := json.Marshal(map[string]string{
		"type":           "service_account",
		"project_id":     "sa-project",
		"private_key_id": "kid",
		"private_key":    string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
		"client_email":   "pusher@sa-project.iam.gserviceaccount.com",
		"token_uri":      tokenURI,
	})
	if err != nil {
		t.Fatalf("json: %v", err)
	}
	return raw
}
