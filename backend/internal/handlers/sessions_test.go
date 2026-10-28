package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

type fakePlatform struct {
	mu        sync.Mutex
	sessions  map[string]map[string]any
	order     []string
	calls     []string
	creates   []map[string]any
	next      int
	statusFor map[string]int
	config    string
}

const legacyPlatformConfig = `{"claudeCode":{"defaultModel":"~anthropic/claude-opus-latest","models":["sonnet","opus"]}}`

const catalogPlatformConfig = `{"claudeCode":{"defaultModel":"~anthropic/claude-opus-latest","models":["sonnet","opus"]},
 "workloadTypes":[
  {"type":"shell","label":"Shell","description":"bash","takesModel":false,"input":"terminal","defaultModel":null,"models":[]},
  {"type":"claude-code","label":"Claude Code","description":"agent","takesModel":true,"input":"prompt","defaultModel":"~anthropic/claude-opus-latest","models":["sonnet","opus"]},
  {"type":"minimax-code","label":"MiniMax Code","description":"mcode","takesModel":true,"input":"prompt","defaultModel":"platform-default","models":[]},
  {"type":"approval-gated","label":"Approval-gated","description":"gated","takesModel":true,"input":"prompt","defaultModel":"~anthropic/claude-opus-latest","models":["sonnet"]},
  {"type":"future-kind","label":"Future","description":"?","takesModel":false,"input":"hologram","defaultModel":"stray","models":["x"]}
 ]}`

func newFakePlatform() *fakePlatform {
	return &fakePlatform{sessions: map[string]map[string]any{}, statusFor: map[string]int{}, config: legacyPlatformConfig}
}

func (f *fakePlatform) seed(name, workload, lastAccess string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	id := fmt.Sprintf("s-%d", f.next)
	f.sessions[id] = map[string]any{
		"id": id, "name": name, "workloadType": workload, "state": "active",
		"model": "platform-default", "createdAt": "2026-10-04T00:00:00Z", "lastAccess": lastAccess,
		"pod": "session-" + id + "-pod.session-platform.svc.cluster.local", "auxiliaryPods": []string{"helper-" + id},
		"checkpoint": map[string]any{"ref": "s3://checkpoints/" + id, "sizeBytes": 42, "createdAt": "2026-10-04T00:00:00Z", "reclaimed": "pod/session-" + id},
	}
	f.order = append(f.order, id)
	return id
}

func (f *fakePlatform) callCount(prefix string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

func (f *fakePlatform) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	call := r.Method + " " + r.URL.Path
	f.mu.Lock()
	f.calls = append(f.calls, call)
	forced, hasForced := f.statusFor[call]
	f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if hasForced {
		w.WriteHeader(forced)
		_, _ = w.Write([]byte(`{"error":"pod session-x-pod on node k3s-1 failed"}`))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1")
	switch {
	case r.Method == http.MethodGet && path == "/config":
		f.mu.Lock()
		cfg := f.config
		f.mu.Unlock()
		_, _ = w.Write([]byte(cfg))
	case r.Method == http.MethodGet && path == "/sessions":
		f.mu.Lock()
		list := []map[string]any{}
		for _, id := range f.order {
			if s, ok := f.sessions[id]; ok {
				list = append(list, s)
			}
		}
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"sessions": list})
	case r.Method == http.MethodPost && path == "/sessions":
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		f.mu.Lock()
		f.creates = append(f.creates, req)
		f.mu.Unlock()
		id := f.seed(fmt.Sprint(req["name"]), fmt.Sprint(req["workloadType"]), "2026-10-04T01:00:00Z")
		f.mu.Lock()
		s := f.sessions[id]
		f.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(s)
	default:
		parts := strings.Split(strings.TrimPrefix(path, "/sessions/"), "/")
		f.mu.Lock()
		s, ok := f.sessions[parts[0]]
		if ok && r.Method == http.MethodDelete {
			delete(f.sessions, parts[0])
		}
		f.mu.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"session not found"}`))
			return
		}
		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case len(parts) == 1:
			_ = json.NewEncoder(w).Encode(s)
		case parts[1] == "read":
			_ = json.NewEncoder(w).Encode(map[string]any{"session": s, "path": "active", "payload": "hello", "nextOffset": 5})
		case parts[1] == "write":
			_ = json.NewEncoder(w).Encode(map[string]any{"session": s, "path": "active"})
		default:
			_ = json.NewEncoder(w).Encode(s)
		}
	}
}

func sessionRoutes(r chi.Router, client *sessions.Client) {
	r.Get("/api/sessions/config", handlers.SessionConfig(client))
	r.Get("/api/sessions", handlers.ListSessions(client))
	r.Post("/api/sessions", handlers.CreateSession(client))
	r.Get("/api/sessions/{id}", handlers.GetSession(client))
	r.Delete("/api/sessions/{id}", handlers.DeleteSession(client))
	r.Post("/api/sessions/{id}/read", handlers.ReadSession(client))
	r.Post("/api/sessions/{id}/write", handlers.WriteSession(client))
	r.Post("/api/sessions/{id}/switch", handlers.SwitchSession(client))
	r.Post("/api/sessions/{id}/snapshot", handlers.SnapshotSession(client))
}

type sessionFixture struct {
	router   http.Handler
	platform *fakePlatform
}

func newSessionFixture(t *testing.T) sessionFixture {
	t.Helper()
	platform := newFakePlatform()
	srv := httptest.NewServer(platform)
	t.Cleanup(srv.Close)
	r := chi.NewRouter()
	sessionRoutes(r, sessions.NewClient(srv.URL))
	return sessionFixture{router: r, platform: platform}
}

func createSession(t *testing.T, fx sessionFixture, body map[string]any, userID int64) sessions.Session {
	t.Helper()
	rec := doTodo(t, fx.router, http.MethodPost, "/api/sessions", body, userID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d want 201 (body=%s)", rec.Code, rec.Body.String())
	}
	var s sessions.Session
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return s
}

func listSessionIDs(t *testing.T, fx sessionFixture, userID int64) []string {
	t.Helper()
	rec := doTodo(t, fx.router, http.MethodGet, "/api/sessions", nil, userID)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: got %d (body=%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		Sessions []sessions.Session `json:"sessions"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	ids := []string{}
	for _, s := range body.Sessions {
		ids = append(ids, s.ID)
	}
	return ids
}

type sessionVerifier struct{}

func (sessionVerifier) Verify(_ context.Context, raw string) (*oidc.IDToken, error) {
	if raw == "good" {
		return &oidc.IDToken{Subject: "u1"}, nil
	}
	return nil, errors.New("expired")
}

func TestSessionsRequireAuthAndNeverReachPlatform(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	platform := newFakePlatform()
	srv := httptest.NewServer(platform)
	t.Cleanup(srv.Close)
	r := chi.NewRouter()
	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(sessionVerifier{}, conn))
		sessionRoutes(p, sessions.NewClient(srv.URL))
	})
	for _, tc := range []struct{ method, target, token string }{
		{http.MethodGet, "/api/sessions", ""},
		{http.MethodGet, "/api/sessions", "expired"},
		{http.MethodPost, "/api/sessions", ""},
		{http.MethodGet, "/api/sessions/config", ""},
		{http.MethodPost, "/api/sessions/s-1/write", "expired"},
		{http.MethodDelete, "/api/sessions/s-1", ""},
	} {
		req := httptest.NewRequest(tc.method, tc.target, strings.NewReader(`{"name":"x"}`))
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s token=%q: got %d want 401", tc.method, tc.target, tc.token, rec.Code)
		}
	}
	if n := platform.callCount(""); n != 0 {
		t.Fatalf("platform received %d calls from unauthenticated requests, want 0", n)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer good")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid token list: got %d want 200 (body=%s)", rec.Code, rec.Body.String())
	}
}

func TestSessionResponsesHideInternalTopology(t *testing.T) {
	fx := newSessionFixture(t)
	s := createSession(t, fx, map[string]any{"name": "fix tests"}, 1)
	for _, tc := range []struct{ method, target string }{
		{http.MethodPost, "/api/sessions"},
		{http.MethodGet, "/api/sessions"},
		{http.MethodGet, "/api/sessions/" + s.ID},
		{http.MethodPost, "/api/sessions/" + s.ID + "/read"},
		{http.MethodPost, "/api/sessions/" + s.ID + "/write"},
		{http.MethodPost, "/api/sessions/" + s.ID + "/switch"},
		{http.MethodPost, "/api/sessions/" + s.ID + "/snapshot"},
	} {
		var body any
		if tc.method == http.MethodPost {
			body = map[string]any{"name": "again"}
			if tc.target != "/api/sessions" {
				body = nil
			}
		}
		rec := doTodo(t, fx.router, tc.method, tc.target, body, 1)
		if rec.Code/100 != 2 {
			t.Fatalf("%s %s: got %d (body=%s)", tc.method, tc.target, rec.Code, rec.Body.String())
		}
		got := rec.Body.String()
		for _, leak := range []string{"pod", "auxiliaryPods", "svc.cluster.local", "s3://", "reclaimed", `"ref"`} {
			if strings.Contains(got, leak) {
				t.Errorf("%s %s leaks %q: %s", tc.method, tc.target, leak, got)
			}
		}
	}
	rec := doTodo(t, fx.router, http.MethodGet, "/api/sessions/"+s.ID, nil, 1)
	var got sessions.Session
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Checkpoint == nil || got.Checkpoint.SizeBytes == nil || *got.Checkpoint.SizeBytes != 42 || got.Checkpoint.CreatedAt == "" {
		t.Fatalf("checkpoint size/createdAt must survive sanitising, got %+v", got.Checkpoint)
	}
}

func TestSessionUpstreamErrorKeepsStatusButNotBody(t *testing.T) {
	fx := newSessionFixture(t)
	s := createSession(t, fx, map[string]any{"name": "snap"}, 1)
	for _, status := range []int{http.StatusConflict, http.StatusServiceUnavailable, http.StatusInternalServerError} {
		fx.platform.mu.Lock()
		fx.platform.statusFor["POST /api/v1/sessions/"+s.ID+"/snapshot"] = status
		fx.platform.mu.Unlock()
		rec := doTodo(t, fx.router, http.MethodPost, "/api/sessions/"+s.ID+"/snapshot", nil, 1)
		if rec.Code != status {
			t.Errorf("upstream %d: got %d", status, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "k3s-1") || strings.Contains(rec.Body.String(), "pod") {
			t.Errorf("upstream %d: body leaks upstream detail: %s", status, rec.Body.String())
		}
	}
}

func TestSessionPlatformUnreachableIsBadGateway(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	r := chi.NewRouter()
	sessionRoutes(r, sessions.NewClient(url))
	rec := doTodo(t, r, http.MethodPost, "/api/sessions", map[string]any{"name": "x"}, 1)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("got %d want 502 (body=%s)", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "127.0.0.1") {
		t.Fatalf("body leaks upstream address: %s", rec.Body.String())
	}
}

func TestSessionListShowsEveryPlatformSessionToEveryUser(t *testing.T) {
	fx := newSessionFixture(t)
	mine := createSession(t, fx, map[string]any{"name": "mine"}, 1)
	spa := fx.platform.seed("spa shell", "shell", "2026-10-04T09:00:00Z")
	want := strings.Join([]string{spa, mine.ID}, ",")
	for _, userID := range []int64{1, 2} {
		if ids := listSessionIDs(t, fx, userID); strings.Join(ids, ",") != want {
			t.Fatalf("user %d list: got %v want [%s] (app- and SPA-created sessions alike)", userID, ids, want)
		}
	}
	for _, id := range []string{mine.ID, spa} {
		for _, tc := range []struct{ method, suffix string }{
			{http.MethodGet, ""},
			{http.MethodPost, "/read"},
			{http.MethodPost, "/write"},
			{http.MethodPost, "/switch"},
			{http.MethodPost, "/snapshot"},
		} {
			rec := doTodo(t, fx.router, tc.method, "/api/sessions/"+id+tc.suffix, nil, 2)
			if rec.Code != http.StatusOK {
				t.Errorf("user 2 %s %s%s: got %d want 200 (body=%s)", tc.method, id, tc.suffix, rec.Code, rec.Body.String())
			}
		}
	}
	if rec := doTodo(t, fx.router, http.MethodDelete, "/api/sessions/"+spa, nil, 2); rec.Code != http.StatusNoContent {
		t.Fatalf("user 2 delete of a SPA session: got %d want 204", rec.Code)
	}
	if ids := listSessionIDs(t, fx, 1); len(ids) != 1 || ids[0] != mine.ID {
		t.Fatalf("list after delete: got %v want [%s]", ids, mine.ID)
	}
}

func TestSessionUnknownIDRelaysPlatformNotFound(t *testing.T) {
	fx := newSessionFixture(t)
	for _, tc := range []struct{ method, suffix string }{
		{http.MethodGet, ""},
		{http.MethodDelete, ""},
		{http.MethodPost, "/read"},
		{http.MethodPost, "/write"},
		{http.MethodPost, "/switch"},
		{http.MethodPost, "/snapshot"},
	} {
		rec := doTodo(t, fx.router, tc.method, "/api/sessions/s-404"+tc.suffix, nil, 1)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s s-404%s: got %d want 404", tc.method, tc.suffix, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "pod") {
			t.Errorf("%s s-404%s: body leaks upstream detail: %s", tc.method, tc.suffix, rec.Body.String())
		}
	}
}

func TestSessionCreateValidatesAgainstPlatformCatalog(t *testing.T) {
	fx := newSessionFixture(t)
	fx.platform.mu.Lock()
	fx.platform.config = catalogPlatformConfig
	fx.platform.mu.Unlock()
	for _, body := range []map[string]any{
		{"name": "x", "workloadType": ""},
		{"name": "x", "workloadType": "python"},
		{"name": "x", "workloadType": 3},
		{"name": "x", "workloadType": "shell", "model": "opus"},
	} {
		rec := doTodo(t, fx.router, http.MethodPost, "/api/sessions", body, 1)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("body=%v: got %d want 400", body, rec.Code)
		}
	}
	for _, name := range []string{"", "   ", "\t\n"} {
		rec := doTodo(t, fx.router, http.MethodPost, "/api/sessions", map[string]any{"name": name}, 1)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("name=%q: got %d want 400", name, rec.Code)
		}
	}
	if n := fx.platform.callCount("POST /api/v1/sessions"); n != 0 {
		t.Fatalf("rejected creates reached the platform: %d", n)
	}
	created := []sessions.Session{
		createSession(t, fx, map[string]any{"name": "  trimmed  "}, 1),
		createSession(t, fx, map[string]any{"name": "explicit", "workloadType": "claude-code", "model": "opus"}, 1),
		createSession(t, fx, map[string]any{"name": "mm", "workloadType": "minimax-code", "model": "MiniMax-M2"}, 1),
		createSession(t, fx, map[string]any{"name": "sh", "workloadType": "shell", "model": "  "}, 1),
		createSession(t, fx, map[string]any{"name": "new", "workloadType": "future-kind"}, 1),
	}
	wantTypes := []string{"claude-code", "claude-code", "minimax-code", "shell", "future-kind"}
	for i, s := range created {
		if s.WorkloadType != wantTypes[i] {
			t.Errorf("create %d workloadType: got %q want %q", i, s.WorkloadType, wantTypes[i])
		}
	}
	if ids := listSessionIDs(t, fx, 1); len(ids) != len(created) {
		t.Errorf("list: got %v, want all %d created sessions of every type", ids, len(created))
	}
	fx.platform.mu.Lock()
	creates := fx.platform.creates
	fx.platform.mu.Unlock()
	if len(creates) != len(created) {
		t.Fatalf("platform creates: got %d want %d", len(creates), len(created))
	}
	if creates[0]["workloadType"] != "claude-code" || creates[0]["name"] != "trimmed" {
		t.Errorf("an omitted workloadType must still create claude-code, sent %v", creates[0])
	}
	if _, ok := creates[0]["model"]; ok {
		t.Errorf("omitted model must stay omitted so the platform records platform-default, sent %v", creates[0])
	}
	if creates[2]["model"] != "MiniMax-M2" {
		t.Errorf("a type with an empty catalog still takes a free-text model, sent %v", creates[2])
	}
	if _, ok := creates[3]["model"]; ok || creates[3]["workloadType"] != "shell" {
		t.Errorf("a blank model on a shell create must be dropped, sent %v", creates[3])
	}
}

func TestSessionCreateFallsBackToClaudeCodeOnLegacyPlatform(t *testing.T) {
	fx := newSessionFixture(t)
	for _, wt := range []string{"shell", "minimax-code"} {
		rec := doTodo(t, fx.router, http.MethodPost, "/api/sessions", map[string]any{"name": "x", "workloadType": wt}, 1)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("legacy platform, workloadType=%q: got %d want 400", wt, rec.Code)
		}
	}
	if s := createSession(t, fx, map[string]any{"name": "ok", "model": "opus"}, 1); s.WorkloadType != "claude-code" {
		t.Fatalf("legacy platform create: got %q", s.WorkloadType)
	}
}

func TestSessionCreateUnreachableConfigDoesNotCreate(t *testing.T) {
	fx := newSessionFixture(t)
	fx.platform.mu.Lock()
	fx.platform.statusFor["GET /api/v1/config"] = http.StatusInternalServerError
	fx.platform.mu.Unlock()
	rec := doTodo(t, fx.router, http.MethodPost, "/api/sessions", map[string]any{"name": "x"}, 1)
	if rec.Code == http.StatusCreated {
		t.Fatalf("create must fail when the catalog cannot be read, got 201")
	}
	if n := fx.platform.callCount("POST /api/v1/sessions"); n != 0 {
		t.Fatalf("create reached the platform without a catalog: %d", n)
	}
}

func TestSessionDeleteRemovesPlatformSession(t *testing.T) {
	fx := newSessionFixture(t)
	s := createSession(t, fx, map[string]any{"name": "done"}, 1)
	rec := doTodo(t, fx.router, http.MethodDelete, "/api/sessions/"+s.ID, nil, 1)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d want 204 (body=%s)", rec.Code, rec.Body.String())
	}
	if rec := doTodo(t, fx.router, http.MethodGet, "/api/sessions/"+s.ID, nil, 1); rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete: got %d want 404", rec.Code)
	}
}

func TestSessionListSortedByLastAccess(t *testing.T) {
	fx := newSessionFixture(t)
	ids := []string{}
	for _, last := range []string{"2026-10-01T00:00:00Z", "2026-10-03T00:00:00Z", "2026-10-02T00:00:00Z"} {
		ids = append(ids, fx.platform.seed("s", "claude-code", last))
	}
	got := listSessionIDs(t, fx, 1)
	want := []string{ids[1], ids[2], ids[0]}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("order: got %v want %v", got, want)
	}
	if n := fx.platform.callCount("POST"); n != 0 {
		t.Fatalf("listing must not promote sessions (no read/write/switch), got %d POSTs", n)
	}
}

func TestSessionConfigRelaysCatalog(t *testing.T) {
	fx := newSessionFixture(t)
	read := func() sessions.RuntimeConfig {
		t.Helper()
		rec := doTodo(t, fx.router, http.MethodGet, "/api/sessions/config", nil, 1)
		if rec.Code != http.StatusOK {
			t.Fatalf("config: got %d", rec.Code)
		}
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("Cache-Control: got %q", rec.Header().Get("Cache-Control"))
		}
		var cfg sessions.RuntimeConfig
		if err := json.NewDecoder(rec.Body).Decode(&cfg); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return cfg
	}
	legacy := read()
	if legacy.ClaudeCode.DefaultModel != "~anthropic/claude-opus-latest" || len(legacy.ClaudeCode.Models) != 2 {
		t.Fatalf("config: got %+v", legacy)
	}
	if lw := legacy.WorkloadTypes; len(lw) != 1 || lw[0].Type != "claude-code" || !lw[0].TakesModel || lw[0].Input != "prompt" || lw[0].DefaultModel == nil || *lw[0].DefaultModel != "~anthropic/claude-opus-latest" || len(lw[0].Models) != 2 {
		t.Fatalf("a platform without workloadTypes must surface claude-code alone with its catalog, got %+v", legacy.WorkloadTypes)
	}

	fx.platform.mu.Lock()
	fx.platform.config = catalogPlatformConfig
	fx.platform.mu.Unlock()
	cat := read()
	var types []string
	for _, w := range cat.WorkloadTypes {
		types = append(types, w.Type)
	}
	if strings.Join(types, ",") != "shell,claude-code,minimax-code,approval-gated,future-kind" {
		t.Fatalf("catalog order must be the platform's, got %v", types)
	}
	mm, _ := cat.Workload("minimax-code")
	if !mm.TakesModel || mm.Models == nil || len(mm.Models) != 0 || mm.DefaultModel == nil || *mm.DefaultModel != "platform-default" {
		t.Errorf("minimax-code: got %+v", mm)
	}
	if sh, _ := cat.Workload("shell"); sh.TakesModel || sh.Input != "terminal" || sh.DefaultModel != nil {
		t.Errorf("shell: got %+v", sh)
	}
	fk, _ := cat.Workload("future-kind")
	if fk.Input != "prompt" {
		t.Errorf("an unknown input must fall back to prompt, got %q", fk.Input)
	}
	if fk.DefaultModel != nil || len(fk.Models) != 0 {
		t.Errorf("a type that takes no model must relay no model catalog, got %+v", fk)
	}
}
