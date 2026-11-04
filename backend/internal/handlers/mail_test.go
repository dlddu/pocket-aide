package handlers_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/mail"
)

type stubIMAP struct {
	probeErr error
	unseen   map[string][]mail.Message
	authFail map[string]bool
}

func (s *stubIMAP) Probe(context.Context, mail.Credentials) error { return s.probeErr }
func (s *stubIMAP) Unseen(_ context.Context, c mail.Credentials, _ int) ([]mail.Message, error) {
	if s.authFail[c.Email] {
		return nil, mail.ErrAuth
	}
	return s.unseen[c.Email], nil
}

func newMailRouter(t *testing.T, imap *stubIMAP) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "mail.db"))
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO users (id, oidc_sub) VALUES (1, 'u1'), (2, 'u2')`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	c, err := mail.NewCipher("1", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	svc := &mail.Service{Store: mail.NewStore(conn, c), IMAP: imap}
	r := chi.NewRouter()
	r.Get("/api/mail/accounts", handlers.ListMailAccounts(svc))
	r.Post("/api/mail/accounts", handlers.ConnectMailAccount(svc))
	r.Put("/api/mail/accounts/{id}/password", handlers.ReconnectMailAccount(svc))
	r.Delete("/api/mail/accounts/{id}", handlers.DisconnectMailAccount(svc))
	r.Get("/api/mail/messages", handlers.ListMailMessages(svc))
	return r
}

func do(h http.Handler, method, target string, body any, user int64) *httptest.ResponseRecorder {
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, authedRequest(method, target, raw, user))
	return rec
}

func TestMail_ConnectListMessagesDisconnect(t *testing.T) {
	stub := &stubIMAP{unseen: map[string][]mail.Message{
		"me@gmail.com": {{UID: 1, Subject: "hi", ReceivedAt: time.Unix(100, 0)}},
	}}
	h := newMailRouter(t, stub)

	rec := do(h, http.MethodPost, "/api/mail/accounts", map[string]any{"email": "me@gmail.com", "app_password": "pw"}, 1)
	if rec.Code != http.StatusCreated {
		t.Fatalf("connect: %d %s", rec.Code, rec.Body)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("pw")) && bytes.Contains(rec.Body.Bytes(), []byte("password")) {
		t.Errorf("response must not echo the app password: %s", rec.Body)
	}
	var created mail.Account
	_ = json.NewDecoder(rec.Body).Decode(&created)
	if created.IMAPHost != "imap.gmail.com" {
		t.Errorf("host auto-fill: %+v", created)
	}

	if rec = do(h, http.MethodPost, "/api/mail/accounts", map[string]any{"email": "me@gmail.com", "app_password": "pw"}, 1); rec.Code != http.StatusConflict {
		t.Errorf("duplicate connect: %d, want 409", rec.Code)
	}

	rec = do(h, http.MethodGet, "/api/mail/messages", nil, 1)
	var list struct {
		Items    []mail.ListedMessage `json:"items"`
		Accounts []mail.AccountStatus `json:"accounts"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if rec.Code != http.StatusOK || len(list.Items) != 1 || list.Items[0].AccountEmail != "me@gmail.com" {
		t.Fatalf("messages: %d %+v", rec.Code, list)
	}

	// Another user sees nothing and cannot delete it.
	rec = do(h, http.MethodGet, "/api/mail/accounts", nil, 2)
	var accts struct{ Items []mail.Account }
	_ = json.NewDecoder(rec.Body).Decode(&accts)
	if len(accts.Items) != 0 {
		t.Errorf("user 2 sees accounts: %+v", accts.Items)
	}
	id := strconv.FormatInt(created.ID, 10)
	if rec = do(h, http.MethodDelete, "/api/mail/accounts/"+id, nil, 2); rec.Code != http.StatusNotFound {
		t.Errorf("foreign delete: %d, want 404", rec.Code)
	}
	if rec = do(h, http.MethodGet, "/api/mail/messages?account_id="+id, nil, 2); rec.Code != http.StatusNotFound {
		t.Errorf("foreign account filter: %d, want 404", rec.Code)
	}
	if rec = do(h, http.MethodDelete, "/api/mail/accounts/"+id, nil, 1); rec.Code != http.StatusNoContent {
		t.Errorf("delete: %d, want 204", rec.Code)
	}
}

func TestMail_RejectedLoginIsNotSavedAndBadInputIs400(t *testing.T) {
	stub := &stubIMAP{probeErr: mail.ErrAuth}
	h := newMailRouter(t, stub)

	if rec := do(h, http.MethodPost, "/api/mail/accounts", map[string]any{"email": "me@gmail.com", "app_password": "bad"}, 1); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("rejected login: %d, want 422", rec.Code)
	}
	rec := do(h, http.MethodGet, "/api/mail/accounts", nil, 1)
	var accts struct{ Items []mail.Account }
	_ = json.NewDecoder(rec.Body).Decode(&accts)
	if len(accts.Items) != 0 {
		t.Errorf("rejected login was saved: %+v", accts.Items)
	}
	for name, body := range map[string]any{
		"missing password": map[string]any{"email": "me@gmail.com"},
		"unknown host":     map[string]any{"email": "me@corp.example", "app_password": "pw"},
		"unknown field":    map[string]any{"email": "me@gmail.com", "app_password": "pw", "x": 1},
	} {
		if rec := do(h, http.MethodPost, "/api/mail/accounts", body, 1); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: %d, want 400", name, rec.Code)
		}
	}
	if rec := do(h, http.MethodGet, "/api/mail/messages?account_id=abc", nil, 1); rec.Code != http.StatusBadRequest {
		t.Errorf("bad account_id: %d, want 400", rec.Code)
	}
}

func TestMail_ReauthFlagAndReconnect(t *testing.T) {
	stub := &stubIMAP{authFail: map[string]bool{}, unseen: map[string][]mail.Message{"me@gmail.com": {{UID: 1, Subject: "s"}}}}
	h := newMailRouter(t, stub)
	rec := do(h, http.MethodPost, "/api/mail/accounts", map[string]any{"email": "me@gmail.com", "app_password": "pw"}, 1)
	var created mail.Account
	_ = json.NewDecoder(rec.Body).Decode(&created)
	id := strconv.FormatInt(created.ID, 10)

	stub.authFail["me@gmail.com"] = true
	rec = do(h, http.MethodGet, "/api/mail/messages", nil, 1)
	var list struct {
		Items    []mail.ListedMessage `json:"items"`
		Accounts []mail.AccountStatus `json:"accounts"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if rec.Code != http.StatusOK || len(list.Items) != 0 || list.Accounts[0].Status != mail.StatusReauth {
		t.Fatalf("revoked password: %d %+v", rec.Code, list)
	}
	rec = do(h, http.MethodGet, "/api/mail/accounts", nil, 1)
	var accts struct{ Items []mail.Account }
	_ = json.NewDecoder(rec.Body).Decode(&accts)
	if len(accts.Items) != 1 || !accts.Items[0].NeedsReconnect {
		t.Errorf("account should be listed with needs_reconnect: %+v", accts.Items)
	}

	stub.authFail["me@gmail.com"] = false
	if rec = do(h, http.MethodPut, "/api/mail/accounts/"+id+"/password", map[string]any{"app_password": "new"}, 1); rec.Code != http.StatusNoContent {
		t.Fatalf("reconnect: %d %s", rec.Code, rec.Body)
	}
	rec = do(h, http.MethodGet, "/api/mail/accounts", nil, 1)
	accts.Items = nil
	_ = json.NewDecoder(rec.Body).Decode(&accts)
	if accts.Items[0].NeedsReconnect {
		t.Error("needs_reconnect should clear after reconnect")
	}
}

func TestMailDisabled_Is503(t *testing.T) {
	rec := httptest.NewRecorder()
	handlers.MailDisabled().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/mail/accounts", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("got %d, want 503", rec.Code)
	}
}
