package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

func newDeviceStore(t *testing.T) *devicetokens.Store {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "dt.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return devicetokens.New(conn)
}

func TestRegisterDeviceToken_Created(t *testing.T) {
	store := newDeviceStore(t)
	hex64 := strings.Repeat("a", 64)
	body := []byte(`{"token":"` + hex64 + `"}`)

	rec := httptest.NewRecorder()
	handlers.RegisterDeviceToken(store).ServeHTTP(
		rec,
		authedRequest(http.MethodPost, "/api/device-tokens", body, 1),
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status: got %d want 201 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestRegisterDeviceToken_RejectsShortToken(t *testing.T) {
	store := newDeviceStore(t)
	rec := httptest.NewRecorder()
	handlers.RegisterDeviceToken(store).ServeHTTP(
		rec,
		authedRequest(http.MethodPost, "/api/device-tokens", []byte(`{"token":"abcd"}`), 1),
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestRegisterDeviceToken_RejectsNonHex(t *testing.T) {
	store := newDeviceStore(t)
	rec := httptest.NewRecorder()
	body := []byte(`{"token":"` + strings.Repeat("z", 64) + `"}`)
	handlers.RegisterDeviceToken(store).ServeHTTP(
		rec,
		authedRequest(http.MethodPost, "/api/device-tokens", body, 1),
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestRegisterDeviceToken_UpsertRebindsToNewUser(t *testing.T) {
	store := newDeviceStore(t)
	hex64 := strings.Repeat("b", 64)
	body := []byte(`{"token":"` + hex64 + `"}`)

	for _, uid := range []int64{1, 2} {
		rec := httptest.NewRecorder()
		handlers.RegisterDeviceToken(store).ServeHTTP(
			rec,
			authedRequest(http.MethodPost, "/api/device-tokens", body, uid),
		)
		if rec.Code != http.StatusCreated {
			t.Fatalf("uid=%d status: got %d want 201", uid, rec.Code)
		}
	}

	for uid, want := range map[int64]int{1: 0, 2: 1} {
		tokens, err := store.ListByUserID(context.Background(), uid)
		if err != nil {
			t.Fatalf("uid=%d list: %v", uid, err)
		}
		if len(tokens) != want {
			t.Errorf("uid=%d: expected %d token(s) after upsert rebind, got %d", uid, want, len(tokens))
		}
	}
}

func registerDevice(t *testing.T, store *devicetokens.Store, uid int64, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	handlers.RegisterDeviceToken(store).ServeHTTP(
		rec,
		authedRequest(http.MethodPost, "/api/device-tokens", []byte(body), uid),
	)
	return rec
}

func TestRegisterDeviceToken_StoresPlatform(t *testing.T) {
	store := newDeviceStore(t)
	hex64 := strings.Repeat("c", 64)
	fcm := "dGVzdA:APA91b-" + strings.Repeat("x", 140)
	for _, body := range []string{
		`{"token":"` + hex64 + `"}`,
		`{"token":"` + fcm + `","platform":"android"}`,
	} {
		if rec := registerDevice(t, store, 1, body); rec.Code != http.StatusCreated {
			t.Fatalf("body=%s status: got %d want 201 (%s)", body, rec.Code, rec.Body.String())
		}
	}

	got, err := store.ListByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	sort.Slice(got, func(i, j int) bool { return got[i].Platform < got[j].Platform })
	want := []devicetokens.Device{
		{Token: fcm, Platform: devicetokens.PlatformAndroid},
		{Token: hex64, Platform: devicetokens.PlatformIOS},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("devices: got %+v want %+v", got, want)
	}
}

func TestRegisterDeviceToken_ExplicitIOSKeepsHexRule(t *testing.T) {
	store := newDeviceStore(t)
	if rec := registerDevice(t, store, 1, `{"token":"not-hex","platform":"ios"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestRegisterDeviceToken_RejectsBadAndroidTokens(t *testing.T) {
	store := newDeviceStore(t)
	for name, body := range map[string]string{
		"empty":            `{"token":"  ","platform":"android"}`,
		"inner whitespace": `{"token":"abc def","platform":"android"}`,
		"too long":         `{"token":"` + strings.Repeat("a", 4097) + `","platform":"android"}`,
	} {
		if rec := registerDevice(t, store, 1, body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status got %d want 400", name, rec.Code)
		}
	}
	if rec := registerDevice(t, store, 1, `{"token":"`+strings.Repeat("a", 4096)+`","platform":"android"}`); rec.Code != http.StatusCreated {
		t.Errorf("4096 chars: status got %d want 201", rec.Code)
	}
}

func TestRegisterDeviceToken_RejectsUnknownPlatform(t *testing.T) {
	store := newDeviceStore(t)
	if rec := registerDevice(t, store, 1, `{"token":"`+strings.Repeat("a", 64)+`","platform":"web"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rec.Code)
	}
}

func TestRegisterDeviceToken_ReRegisterUpdatesPlatform(t *testing.T) {
	store := newDeviceStore(t)
	tok := strings.Repeat("d", 64)
	if rec := registerDevice(t, store, 1, `{"token":"`+tok+`"}`); rec.Code != http.StatusCreated {
		t.Fatalf("ios: %d", rec.Code)
	}
	if rec := registerDevice(t, store, 1, `{"token":"`+tok+`","platform":"android"}`); rec.Code != http.StatusCreated {
		t.Fatalf("android: %d", rec.Code)
	}
	got, err := store.ListByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].Platform != devicetokens.PlatformAndroid {
		t.Fatalf("devices: %+v", got)
	}
}
