package commitstatus

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

type githubCall struct {
	method, path, auth string
	body               map[string]any
}

func newTestApp(t *testing.T, handler func(w http.ResponseWriter, call githubCall)) (*App, *rsa.PrivateKey, func() []githubCall) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	var (
		mu    sync.Mutex
		calls []githubCall
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := githubCall{method: r.Method, path: r.URL.Path, auth: r.Header.Get("Authorization")}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &call.body)
		}
		mu.Lock()
		calls = append(calls, call)
		mu.Unlock()
		handler(w, call)
	}))
	t.Cleanup(srv.Close)
	app, err := NewApp("Iv1.test", pemKey, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	return app, key, func() []githubCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]githubCall(nil), calls...)
	}
}

func happyGitHub(w http.ResponseWriter, call githubCall) {
	switch {
	case call.method == http.MethodGet && call.path == "/repos/dlddu/pocket-aide/installation":
		_, _ = w.Write([]byte(`{"id": 77}`))
	case call.method == http.MethodPost && call.path == "/app/installations/77/access_tokens":
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"token": "ghs_scoped"}`))
	case call.method == http.MethodPost && strings.HasPrefix(call.path, "/repos/dlddu/pocket-aide/statuses/"):
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 123456}`))
	case call.method == http.MethodDelete && call.path == "/installation/token":
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}
}

func TestAppWritesStatusWithARepoScopedStatusesToken(t *testing.T) {
	app, key, calls := newTestApp(t, happyGitHub)

	id, err := app.WriteStatus(t.Context(), StatusWrite{
		Owner: "dlddu", Repo: "pocket-aide", SHA: "0123456789abcdef0123456789abcdef01234567", State: "success",
		Context: "review/manual-approval", Description: "수동 승인", TargetURL: "https://github.com/dlddu/pocket-aide/pull/1",
	})
	if err != nil || id != 123456 {
		t.Fatalf("WriteStatus = %d, %v; want GitHub's id 123456", id, err)
	}
	got := calls()
	if len(got) != 4 {
		t.Fatalf("calls %+v: want installation lookup, token mint, status write, token revoke", got)
	}

	appToken := strings.TrimPrefix(got[0].auth, "Bearer ")
	claims := jwt.RegisteredClaims{}
	if _, err := jwt.ParseWithClaims(appToken, &claims, func(*jwt.Token) (any, error) { return &key.PublicKey, nil },
		jwt.WithValidMethods([]string{"RS256"})); err != nil || claims.Issuer != "Iv1.test" {
		t.Fatalf("installation lookup was not signed as the app: iss=%q err=%v", claims.Issuer, err)
	}
	if got[1].auth != got[0].auth {
		t.Fatalf("token mint must use the app JWT")
	}
	mint, _ := json.Marshal(got[1].body)
	if string(mint) != `{"permissions":{"statuses":"write"},"repositories":["pocket-aide"]}` {
		t.Fatalf("mint body %s: want only this repository and statuses: write", mint)
	}
	if got[2].auth != "Bearer ghs_scoped" || got[2].path != "/repos/dlddu/pocket-aide/statuses/0123456789abcdef0123456789abcdef01234567" {
		t.Fatalf("status write %+v: want the scoped token on the full sha", got[2])
	}
	body, _ := json.Marshal(got[2].body)
	if string(body) != `{"context":"review/manual-approval","description":"수동 승인","state":"success","target_url":"https://github.com/dlddu/pocket-aide/pull/1"}` {
		t.Fatalf("status body %s", body)
	}
	if got[3].method != http.MethodDelete || got[3].auth != "Bearer ghs_scoped" {
		t.Fatalf("token was not revoked: %+v", got[3])
	}
}

func TestAppSurfacesGitHubMessage(t *testing.T) {
	app, _, calls := newTestApp(t, func(w http.ResponseWriter, call githubCall) {
		if strings.Contains(call.path, "/statuses/") {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"message": "No commit found for SHA: 0123"}`))
			return
		}
		happyGitHub(w, call)
	})
	_, err := app.WriteStatus(t.Context(), StatusWrite{Owner: "dlddu", Repo: "pocket-aide", SHA: "0123456789abcdef0123456789abcdef01234567", State: "success", Context: "review/x"})
	if err == nil || !strings.Contains(err.Error(), "422") || !strings.Contains(err.Error(), "No commit found") {
		t.Fatalf("err %v: want GitHub's status code and message", err)
	}
	if strings.Contains(err.Error(), "ghs_scoped") {
		t.Fatalf("error leaks the installation token: %v", err)
	}
	if last := calls()[len(calls())-1]; last.method != http.MethodDelete {
		t.Fatalf("token not revoked after a failed write: %+v", last)
	}
}

func TestAppNotInstalledOnRepo(t *testing.T) {
	app, _, calls := newTestApp(t, happyGitHub)
	_, err := app.WriteStatus(t.Context(), StatusWrite{Owner: "someone", Repo: "elsewhere", SHA: "0123456789abcdef0123456789abcdef01234567", State: "success", Context: "review/x"})
	if err == nil || !strings.Contains(err.Error(), "someone/elsewhere") {
		t.Fatalf("err %v: want the repository named", err)
	}
	if len(calls()) != 1 {
		t.Fatalf("calls %+v: nothing should follow a failed installation lookup", calls())
	}
}

func TestNewAppRejectsBadConfig(t *testing.T) {
	if _, err := NewApp("", []byte("x"), ""); err == nil {
		t.Fatal("empty client id accepted")
	}
	if _, err := NewApp("Iv1.test", []byte("not a pem"), ""); err == nil {
		t.Fatal("bad key accepted")
	}
}
