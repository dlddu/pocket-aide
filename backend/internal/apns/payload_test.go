package apns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFitAlertBodyKeepsBodiesThatFit(t *testing.T) {
	data := map[string]any{"approval_id": "r"}
	if got := FitAlertBody("", "승인 요청 · infra-ci", "prod 배포", data); got != "prod 배포" {
		t.Fatalf("short body changed: %q", got)
	}
}

func TestFitAlertBodyCutsToThePayloadLimit(t *testing.T) {
	data := map[string]any{"approval_id": "0f8c1d2e-1111-2222-3333-444455556666"}
	title := "승인 요청 · infra-ci"
	for name, body := range map[string]string{
		"korean 10KB":  strings.Repeat("배포를 승인해 주세요 ", 700),
		"ascii 10KB":   strings.Repeat("x", 10*1024),
		"escape heavy": strings.Repeat("<&\"\n", 3000),
	} {
		t.Run(name, func(t *testing.T) {
			got := FitAlertBody("", title, body, data)
			if !strings.HasSuffix(got, "…") {
				t.Fatalf("cut body should end with …")
			}
			kept := strings.TrimSuffix(got, "…")
			if !strings.HasPrefix(body, kept) {
				t.Fatalf("cut body should be a prefix of the original")
			}
			if size := PayloadSize("", title, got, data); size > MaxPayloadBytes {
				t.Fatalf("payload %d bytes > %d", size, MaxPayloadBytes)
			}
			longer := string([]rune(body)[:len([]rune(kept))+1]) + "…"
			if PayloadSize("", title, longer, data) <= MaxPayloadBytes {
				t.Fatalf("one more character would still fit — the cut is not tight")
			}
		})
	}
}

func TestSendWithDataReportsRejections(t *testing.T) {
	var received map[string]any
	status, reason := http.StatusOK, ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &received)
		w.Header().Set("apns-id", "ID-1")
		w.WriteHeader(status)
		if reason != "" {
			_, _ = w.Write([]byte(`{"reason":"` + reason + `"}`))
		}
	}))
	defer srv.Close()
	c, err := New("KEYID", "TEAMID", "com.example.app", generateP8(t), false)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	c = c.WithHost(srv.URL)

	if err := c.SendWithData(context.Background(), "tok", "T", "B", map[string]any{"approval_id": "r-1"}); err != nil {
		t.Fatalf("accepted push: %v", err)
	}
	alert := received["aps"].(map[string]any)["alert"].(map[string]any)
	if alert["title"] != "T" || alert["body"] != "B" || received["approval_id"] != "r-1" {
		t.Fatalf("payload: %v", received)
	}
	if _, ok := received["aps"].(map[string]any)["category"]; ok {
		t.Fatalf("plain push should carry no category: %v", received)
	}

	if err := c.SendWithCategory(context.Background(), "tok", "APPROVAL_REQUEST", "T", "B", map[string]any{"approval_id": "r-1"}); err != nil {
		t.Fatalf("accepted push: %v", err)
	}
	if got := received["aps"].(map[string]any)["category"]; got != "APPROVAL_REQUEST" {
		t.Fatalf("category: %v", received)
	}

	cases := []struct {
		status  int
		reason  string
		invalid bool
	}{
		{http.StatusGone, "Unregistered", true},
		{http.StatusBadRequest, "BadDeviceToken", true},
		{http.StatusBadRequest, "DeviceTokenNotForTopic", true},
		{http.StatusTooManyRequests, "TooManyRequests", false},
		{http.StatusInternalServerError, "InternalServerError", false},
	}
	for _, tc := range cases {
		status, reason = tc.status, tc.reason
		err := c.SendWithData(context.Background(), "tok", "T", "B", nil)
		var rejected *RejectedError
		if !errors.As(err, &rejected) {
			t.Fatalf("%s: want *RejectedError, got %v", tc.reason, err)
		}
		if rejected.TokenInvalid() != tc.invalid {
			t.Fatalf("%s: TokenInvalid=%v want %v", tc.reason, rejected.TokenInvalid(), tc.invalid)
		}
		if want := fmt.Sprintf("apns rejected (status=%d reason=%q apns-id=ID-1)", tc.status, tc.reason); err.Error() != want {
			t.Fatalf("error text changed: %q", err.Error())
		}
	}
}
