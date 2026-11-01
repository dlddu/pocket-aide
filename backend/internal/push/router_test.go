package push_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"golang.org/x/oauth2"

	"github.com/dlddu/pocket-aide/backend/internal/apns"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/fcm"
	"github.com/dlddu/pocket-aide/backend/internal/push"
)

type hits struct {
	mu    sync.Mutex
	paths []string
}

func (h *hits) add(p string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paths = append(h.paths, p)
}

func (h *hits) all() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]string(nil), h.paths...)
}

func receiver(t *testing.T, status int, body string) (*httptest.Server, *hits) {
	t.Helper()
	h := &hits{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.add(r.URL.Path)
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	return srv, h
}

func apnsClient(t *testing.T, host string) *apns.Client {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	c, err := apns.New("KEYID", "TEAMID", "com.example.app", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), false)
	if err != nil {
		t.Fatal(err)
	}
	return c.WithHost(host)
}

func fcmClient(t *testing.T, endpoint string) *fcm.Client {
	t.Helper()
	c, err := fcm.NewWithTokenSource("demo", oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "x"}))
	if err != nil {
		t.Fatal(err)
	}
	return c.WithEndpoint(endpoint)
}

var (
	iosDevice     = devicetokens.Device{Token: "ios-token", Platform: devicetokens.PlatformIOS}
	androidDevice = devicetokens.Device{Token: "android-token", Platform: devicetokens.PlatformAndroid}
)

func TestRouterDispatchesByPlatform(t *testing.T) {
	apnsSrv, apnsHits := receiver(t, http.StatusOK, "")
	fcmSrv, fcmHits := receiver(t, http.StatusOK, `{}`)
	r := push.New(apnsClient(t, apnsSrv.URL), fcmClient(t, fcmSrv.URL))

	if err := r.SendWithData(context.Background(), iosDevice, "t", "b", map[string]any{"event_id": 1}); err != nil {
		t.Fatalf("ios: %v", err)
	}
	if err := r.SendWithCategory(context.Background(), androidDevice, "APPROVAL_REQUEST", "t", "b", nil); err != nil {
		t.Fatalf("android: %v", err)
	}
	if got := apnsHits.all(); len(got) != 1 || got[0] != "/3/device/ios-token" {
		t.Fatalf("apns hits: %v", got)
	}
	if got := fcmHits.all(); len(got) != 1 || got[0] != "/v1/projects/demo/messages:send" {
		t.Fatalf("fcm hits: %v", got)
	}
}

func TestRouterWithoutAClientForThePlatformSkips(t *testing.T) {
	fcmSrv, fcmHits := receiver(t, http.StatusOK, `{}`)
	onlyFCM := push.New(nil, fcmClient(t, fcmSrv.URL))
	if err := onlyFCM.SendWithData(context.Background(), iosDevice, "t", "b", nil); !errors.Is(err, push.ErrPlatformDisabled) {
		t.Fatalf("ios on fcm-only router: %v", err)
	}
	apnsSrv, apnsHits := receiver(t, http.StatusOK, "")
	onlyAPNs := push.New(apnsClient(t, apnsSrv.URL), nil)
	if err := onlyAPNs.SendWithData(context.Background(), androidDevice, "t", "b", nil); !errors.Is(err, push.ErrPlatformDisabled) {
		t.Fatalf("android on apns-only router: %v", err)
	}
	if len(fcmHits.all())+len(apnsHits.all()) != 0 {
		t.Fatalf("nothing should be sent: fcm=%v apns=%v", fcmHits.all(), apnsHits.all())
	}
}

func TestRouterRejectsUnknownPlatform(t *testing.T) {
	r := push.New(nil, nil)
	err := r.SendWithData(context.Background(), devicetokens.Device{Token: "x", Platform: "web"}, "t", "b", nil)
	if err == nil || errors.Is(err, push.ErrPlatformDisabled) || !strings.Contains(err.Error(), "web") {
		t.Fatalf("err = %v", err)
	}
}

func TestIsTokenInvalidSeesBothServicesThroughWrapping(t *testing.T) {
	for name, tc := range map[string]struct {
		err  error
		want bool
	}{
		"apns unregistered": {&apns.RejectedError{StatusCode: 410, Reason: "Unregistered"}, true},
		"apns throttled":    {&apns.RejectedError{StatusCode: 429, Reason: "TooManyRequests"}, false},
		"fcm unregistered":  {fmt.Errorf("wrapped: %w", &fcm.RejectedError{StatusCode: 404, ErrorCode: "UNREGISTERED"}), true},
		"fcm quota":         {&fcm.RejectedError{StatusCode: 429, ErrorCode: "QUOTA_EXCEEDED"}, false},
		"disabled":          {push.ErrPlatformDisabled, false},
		"network":           {errors.New("network down"), false},
	} {
		if got := push.IsTokenInvalid(tc.err); got != tc.want {
			t.Errorf("%s: IsTokenInvalid = %v, want %v", name, got, tc.want)
		}
	}
}
