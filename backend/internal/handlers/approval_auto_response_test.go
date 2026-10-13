package handlers_test

import (
	"net/http"
	"testing"
	"time"
)

type approvalSettingsBody struct {
	AutoResponse string `json:"auto_response"`
}

func (e *approvalEnv) settings(sub string) approvalSettingsBody {
	e.t.Helper()
	rec := e.do(http.MethodGet, "/api/approval-settings", "oidc-"+sub, nil)
	expectCode(e.t, rec, http.StatusOK, "get approval settings")
	return decodeApproval[approvalSettingsBody](e.t, rec)
}

func (e *approvalEnv) setMode(sub, mode string) approvalSettingsBody {
	e.t.Helper()
	rec := e.do(http.MethodPatch, "/api/approval-settings", "oidc-"+sub, map[string]any{"auto_response": mode})
	expectCode(e.t, rec, http.StatusOK, "set auto_response "+mode)
	return decodeApproval[approvalSettingsBody](e.t, rec)
}

func TestApprovalSettingsDefaultUpdateAndValidation(t *testing.T) {
	env := newApprovalEnv(t)
	if got := env.settings("a").AutoResponse; got != "NONE" {
		t.Fatalf("default auto_response: got %q want NONE", got)
	}
	for _, mode := range []string{"AUTO_APPROVE", "AUTO_REJECT", "NONE", "AUTO_REJECT"} {
		if got := env.setMode("a", mode).AutoResponse; got != mode {
			t.Fatalf("PATCH %s: got %q", mode, got)
		}
		if got := env.settings("a").AutoResponse; got != mode {
			t.Fatalf("GET after PATCH %s: got %q", mode, got)
		}
	}
	if got := env.settings("b").AutoResponse; got != "NONE" {
		t.Fatalf("another user's mode leaked: got %q want NONE", got)
	}
	for _, body := range []any{
		map[string]any{"auto_response": "ALWAYS"},
		map[string]any{"auto_response": "auto_approve"},
		map[string]any{},
		map[string]any{"auto_response": "NONE", "extra": true},
		"not json",
	} {
		rec := env.do(http.MethodPatch, "/api/approval-settings", "oidc-a", body)
		expectCode(t, rec, http.StatusBadRequest, "invalid settings body")
	}
	if got := env.settings("a").AutoResponse; got != "AUTO_REJECT" {
		t.Fatalf("rejected PATCH changed the mode: got %q", got)
	}
	expectCode(t, env.do(http.MethodGet, "/api/approval-settings", "", nil), http.StatusUnauthorized, "settings without a user")
}

func TestApprovalAutoResponseProcessesNewRequestsWithoutPushOrRetroaction(t *testing.T) {
	pusher := &recordingPusher{}
	env := newApprovalEnvWithPush(t, pusher)
	key := env.issue("a", "infra-ci")

	pending := env.create(key.Key, map[string]any{"externalId": "p", "context": "기존 대기 요청"})
	if pending.Status != "PENDING" {
		t.Fatalf("NONE mode create: got %s want PENDING", pending.Status)
	}
	if n := len(pusher.all()); n != 1 {
		t.Fatalf("NONE mode pushes: got %d want 1", n)
	}

	for _, tc := range []struct {
		mode, externalID, status, label string
	}{
		{"AUTO_APPROVE", "q-approve", "APPROVED", "auto approve"},
		{"AUTO_REJECT", "q-reject", "REJECTED", "auto reject"},
	} {
		env.setMode("a", tc.mode)
		env.advance(time.Second)
		created := env.create(key.Key, map[string]any{"externalId": tc.externalID, "context": tc.label + " 요청", "timeoutSeconds": 600})
		if created.Status != tc.status {
			t.Fatalf("%s: create response status got %s want %s", tc.label, created.Status, tc.status)
		}
		polled := env.poll(key.Key, created.ID)
		if polled.Status != tc.status || polled.ProcessedAt == nil || *polled.ProcessedAt != created.CreatedAt {
			t.Fatalf("%s: caller poll got status %s processedAt %v want %s at %d", tc.label, polled.Status, polled.ProcessedAt, tc.status, created.CreatedAt)
		}
		var found bool
		for _, item := range env.history("a", "").Items {
			if item.ExternalID == tc.externalID {
				found = true
				if item.Status != tc.status || item.DecisionMode != "auto" {
					t.Fatalf("%s: history item got status %s mode %q want %s auto", tc.label, item.Status, item.DecisionMode, tc.status)
				}
			}
		}
		if !found {
			t.Fatalf("%s: %s is not in the history", tc.label, tc.externalID)
		}
		env.advance(time.Hour)
		if got := env.poll(key.Key, created.ID).Status; got != tc.status {
			t.Fatalf("%s: auto-processed request with a timeout read %s after it passed", tc.label, got)
		}
	}
	if n := len(pusher.all()); n != 1 {
		t.Fatalf("auto modes must not push: got %d pushes want 1", n)
	}

	items := env.pending("a")
	if len(items) != 1 || items[0].ExternalID != "p" || items[0].Status != "PENDING" {
		t.Fatalf("mode change must not touch the earlier PENDING request: %+v", items)
	}
	env.decide("a", pending.ID, "approve")

	env.setMode("a", "NONE")
	again := env.create(key.Key, map[string]any{"externalId": "r", "context": "다시 수동"})
	if again.Status != "PENDING" {
		t.Fatalf("back to NONE: got %s want PENDING", again.Status)
	}
	if n := len(pusher.all()); n != 2 {
		t.Fatalf("NONE mode pushes again: got %d want 2", n)
	}

	env.setMode("b", "AUTO_APPROVE")
	other := env.create(key.Key, map[string]any{"externalId": "s", "context": "다른 사용자의 모드는 무관"})
	if other.Status != "PENDING" {
		t.Fatalf("another user's mode applied: got %s want PENDING", other.Status)
	}
}
