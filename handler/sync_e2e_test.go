// Package handler_test contains end-to-end style tests for the data sync
// API endpoint (POST /sync).
//
// DLD-739: 13-1: 데이터 동기화 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (t.Skip). Activate after DLD-739:
//   - handler.NewSyncHandler is implemented (handler/sync.go)
//   - db/migrations/000010_sync.up.sql migration is applied (or equivalent)
//   - Route POST /sync is registered in main.go
//   - POST /sync accepts a JSON body with an "changes" array of client-side
//     change objects (each with: entity, id, operation, payload, updated_at)
//   - POST /sync returns HTTP 200 with a "server_data" field containing the
//     latest server-side state for the authenticated user
//   - Conflict resolution uses last-write-wins based on updated_at timestamps
//
// When activating:
//  1. Remove all t.Skip calls.
//  2. Uncomment the sync handler lines inside setupSyncEcho (marked TODO).
//  3. Ensure the sync migration exists so testutil.NewTestDB picks it up.
package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/handler"
	appmiddleware "github.com/dlddu/pocket-aide/middleware"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// setupSyncEcho builds an Echo instance with the auth route and — once
// DLD-739 is implemented — the sync route and todo routes (needed for
// verifying server state after sync).
//
// Activation checklist:
//
//	TODO(DLD-739): uncomment the lines below after handler.NewSyncHandler exists.
//	  syncHandler := handler.NewSyncHandler(tdb.DB)
//	  e.POST("/sync", syncHandler.Sync, appmiddleware.JWT("test-jwt-secret"))
//
//	TODO(DLD-739): uncomment the lines below to verify todo state after sync.
//	  todoHandler := handler.NewTodoHandler(tdb.DB)
//	  tg := e.Group("/todos", appmiddleware.JWT("test-jwt-secret"))
//	  tg.GET("", todoHandler.List)
func setupSyncEcho(t *testing.T, tdb *testutil.TestDB) *echo.Echo {
	t.Helper()
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	// TODO(DLD-739): uncomment after handler.NewSyncHandler is implemented.
	// syncHandler := handler.NewSyncHandler(tdb.DB)
	// e.POST("/sync", syncHandler.Sync, appmiddleware.JWT("test-jwt-secret"))

	// TODO(DLD-739): uncomment for server-state verification after sync.
	// todoHandler := handler.NewTodoHandler(tdb.DB)
	// tg := e.Group("/todos", appmiddleware.JWT("test-jwt-secret"))
	// tg.GET("", todoHandler.List)

	return e
}

// syncToken seeds a user with the given email and returns a JWT Bearer token
// string (including the "Bearer " prefix) by reusing seedUserAndLogin defined
// in chat_e2e_test.go (same package).
func syncToken(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email string) string {
	t.Helper()
	return "Bearer " + seedUserAndLogin(t, tdb, e, email, "Secret1!")
}

// ---------------------------------------------------------------------------
// Happy Path — POST /sync (오프라인 변경사항 업로드)
// ---------------------------------------------------------------------------

// TestSyncHandler_Sync_UploadOfflineChanges_ReturnsOK verifies that
// POST /sync with a valid changes payload returns HTTP 200 and a JSON body
// that includes a "server_data" field.
//
// Scenario:
//
//	오프라인 상태에서 생성된 투두 1건을 변경사항 배열에 담아 POST /sync.
//	POST /sync  {"changes":[{"entity":"todo","id":"client-1","operation":"create",
//	             "payload":{"title":"오프라인 투두","type":"personal"},
//	             "updated_at":"<RFC3339>"}]} + Bearer token
//	→ 200 OK  {"server_data": {...}}
func TestSyncHandler_Sync_UploadOfflineChanges_ReturnsOK(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)
	token := syncToken(t, tdb, e, "sync-upload@example.com")

	updatedAt := time.Now().UTC().Format(time.RFC3339)
	payload := `{
		"changes": [
			{
				"entity": "todo",
				"id": "client-1",
				"operation": "create",
				"payload": {"title": "오프라인 투두", "type": "personal"},
				"updated_at": "` + updatedAt + `"
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v (body: %s)", err, rec.Body.String())
	}
	if _, ok := resp["server_data"]; !ok {
		t.Error("expected response to contain 'server_data' field")
	}
}

// TestSyncHandler_Sync_DownloadServerData_ReturnsTodos verifies that
// POST /sync with an empty changes array returns the server-side todos for
// the authenticated user in "server_data".
//
// Scenario:
//
//	서버에 투두 1건 미리 생성 (시드).
//	POST /sync  {"changes":[]} + Bearer token
//	→ 200 OK  {"server_data": {"todos": [{"title":"서버 투두", ...}]}}
func TestSyncHandler_Sync_DownloadServerData_ReturnsTodos(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)
	token := syncToken(t, tdb, e, "sync-download@example.com")

	// Seed a todo directly in the database to simulate pre-existing server data.
	// NOTE: Adjust the INSERT query once the todos table schema is confirmed.
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, created_at, updated_at)
		 VALUES (
			(SELECT id FROM users WHERE email = 'sync-download@example.com'),
			'서버 투두', 'personal', datetime('now'), datetime('now')
		 )`,
	)

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(`{"changes":[]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v (body: %s)", err, rec.Body.String())
	}
	serverData, ok := resp["server_data"]
	if !ok {
		t.Fatal("expected 'server_data' field in response")
	}
	serverDataMap, ok := serverData.(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'server_data' to be an object, got %T", serverData)
	}
	todos, ok := serverDataMap["todos"].([]interface{})
	if !ok {
		t.Fatalf("expected 'server_data.todos' to be an array, got %T", serverDataMap["todos"])
	}
	if len(todos) < 1 {
		t.Errorf("expected at least 1 todo in server_data.todos, got %d", len(todos))
	}
}

// ---------------------------------------------------------------------------
// Conflict Resolution — POST /sync (last-write-wins)
// ---------------------------------------------------------------------------

// TestSyncHandler_Sync_ConflictResolution_LastWriteWins verifies that when
// a client change and a server record have the same entity ID but different
// updated_at timestamps, the one with the more recent updated_at wins.
//
// Scenario:
//
//	서버에 투두 1건 존재 (updated_at: 1분 전).
//	클라이언트가 같은 투두를 더 최신 updated_at으로 수정하여 POST /sync.
//	→ 200 OK  서버 데이터가 클라이언트 변경사항으로 덮어써짐 (last-write-wins).
//	   GET /todos → 클라이언트가 보낸 title이 반영되어 있음.
func TestSyncHandler_Sync_ConflictResolution_LastWriteWins(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)
	token := syncToken(t, tdb, e, "sync-conflict@example.com")

	// Seed: server has an older version of the todo.
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, created_at, updated_at)
		 VALUES (
			(SELECT id FROM users WHERE email = 'sync-conflict@example.com'),
			'서버 원본 제목', 'personal',
			datetime('now', '-2 minutes'), datetime('now', '-1 minute')
		 )`,
	)

	// Client sends a newer update (updated_at is now, which is later than server's).
	clientUpdatedAt := time.Now().UTC().Format(time.RFC3339)
	payload := `{
		"changes": [
			{
				"entity": "todo",
				"id": "server-todo-id-1",
				"operation": "update",
				"payload": {"title": "클라이언트 수정 제목", "type": "personal"},
				"updated_at": "` + clientUpdatedAt + `"
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 on sync, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	// Verify via GET /todos that the client's newer title won.
	getReq := httptest.NewRequest(http.MethodGet, "/todos?type=personal", nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on GET /todos after sync, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("expected valid JSON array from GET /todos, got: %v (body: %s)", err, getRec.Body.String())
	}
	found := false
	for _, todo := range todos {
		if todo["title"] == "클라이언트 수정 제목" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected client's newer title '클라이언트 수정 제목' to win the conflict (last-write-wins)")
	}
}

// TestSyncHandler_Sync_ConflictResolution_ServerWins_WhenServerIsNewer verifies
// that when the server record has a more recent updated_at than the client
// change, the server data is preserved (last-write-wins).
//
// Scenario:
//
//	서버에 투두 1건 존재 (updated_at: 현재, 즉 클라이언트보다 최신).
//	클라이언트가 같은 투두를 더 오래된 updated_at으로 수정하여 POST /sync.
//	→ 200 OK  서버 데이터가 유지됨 (서버가 더 최신이므로).
//	   GET /todos → 서버 원본 title이 그대로 유지됨.
func TestSyncHandler_Sync_ConflictResolution_ServerWins_WhenServerIsNewer(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)
	token := syncToken(t, tdb, e, "sync-server-wins@example.com")

	// Seed: server has a very recent record (1 second ago — newer than client's change).
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, created_at, updated_at)
		 VALUES (
			(SELECT id FROM users WHERE email = 'sync-server-wins@example.com'),
			'서버 최신 제목', 'personal',
			datetime('now', '-10 minutes'), datetime('now', '-1 second')
		 )`,
	)

	// Client sends an older update (updated_at is 5 minutes ago).
	staleClientUpdatedAt := time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339)
	payload := `{
		"changes": [
			{
				"entity": "todo",
				"id": "server-todo-id-1",
				"operation": "update",
				"payload": {"title": "클라이언트 오래된 제목", "type": "personal"},
				"updated_at": "` + staleClientUpdatedAt + `"
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 on sync, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	// Verify via GET /todos that the server's title was NOT overwritten.
	getReq := httptest.NewRequest(http.MethodGet, "/todos?type=personal", nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on GET /todos after sync, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("expected valid JSON array from GET /todos, got: %v (body: %s)", err, getRec.Body.String())
	}
	for _, todo := range todos {
		if todo["title"] == "클라이언트 오래된 제목" {
			t.Error("stale client change should NOT overwrite a newer server record (last-write-wins)")
		}
	}
}

// ---------------------------------------------------------------------------
// Edge Case — POST /sync (빈 변경사항)
// ---------------------------------------------------------------------------

// TestSyncHandler_Sync_EmptyChanges_ReturnsOK verifies that POST /sync with
// an empty "changes" array is valid and returns HTTP 200 with "server_data".
//
// Scenario:
//
//	POST /sync  {"changes":[]} + Bearer token
//	→ 200 OK  {"server_data": {...}}
func TestSyncHandler_Sync_EmptyChanges_ReturnsOK(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)
	token := syncToken(t, tdb, e, "sync-empty@example.com")

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(`{"changes":[]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for empty changes, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v (body: %s)", err, rec.Body.String())
	}
	if _, ok := resp["server_data"]; !ok {
		t.Error("expected 'server_data' field even when changes is empty")
	}
}

// TestSyncHandler_Sync_MultipleEntities_ReturnsOK verifies that POST /sync
// correctly handles a changes array that contains multiple entity types
// (e.g., both todos and memos) in a single request.
//
// Scenario:
//
//	클라이언트가 오프라인에서 투두 1건 생성 + 메모 1건 생성.
//	POST /sync  {"changes":[todo_change, memo_change]} + Bearer token
//	→ 200 OK  두 변경사항 모두 서버에 반영됨.
func TestSyncHandler_Sync_MultipleEntities_ReturnsOK(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)
	token := syncToken(t, tdb, e, "sync-multi@example.com")

	updatedAt := time.Now().UTC().Format(time.RFC3339)
	payload := `{
		"changes": [
			{
				"entity": "todo",
				"id": "client-todo-1",
				"operation": "create",
				"payload": {"title": "오프라인 투두", "type": "personal"},
				"updated_at": "` + updatedAt + `"
			},
			{
				"entity": "memo",
				"id": "client-memo-1",
				"operation": "create",
				"payload": {"content": "오프라인 메모", "source": "text"},
				"updated_at": "` + updatedAt + `"
			}
		]
	}`

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for multi-entity sync, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v (body: %s)", err, rec.Body.String())
	}
	if _, ok := resp["server_data"]; !ok {
		t.Error("expected 'server_data' field in multi-entity sync response")
	}
}

// ---------------------------------------------------------------------------
// Error Cases
// ---------------------------------------------------------------------------

// TestSyncHandler_Sync_WithoutAuth_ReturnsUnauthorized verifies that
// POST /sync without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /sync  {"changes":[]}  (no Authorization header)
//	→ 401 Unauthorized
func TestSyncHandler_Sync_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	e := echo.New()
	// Register a guarded placeholder so the JWT middleware can reject the
	// request before any real handler runs. Replace with the real sync handler
	// once handler.NewSyncHandler is implemented.
	e.POST("/sync", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"server_data": nil})
	}, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(`{"changes":[]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestSyncHandler_Sync_InvalidBody_ReturnsBadRequest verifies that
// POST /sync with a malformed JSON body returns HTTP 400 Bad Request.
//
// Scenario:
//
//	POST /sync  (malformed JSON: "not-json") + Bearer token
//	→ 400 Bad Request
func TestSyncHandler_Sync_InvalidBody_ReturnsBadRequest(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)
	token := syncToken(t, tdb, e, "sync-badreq@example.com")

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader("not-json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for malformed body, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestSyncHandler_Sync_IsolatesUserData_ReturnsOnlyOwnData verifies that
// POST /sync returns only the authenticated user's data and does not leak
// another user's records.
//
// Scenario:
//
//	사용자 A의 투두 1건이 서버에 존재.
//	사용자 B로 로그인하여 POST /sync (changes: []) 요청.
//	→ 200 OK  server_data.todos에 사용자 A의 투두가 포함되지 않음.
func TestSyncHandler_Sync_IsolatesUserData_ReturnsOnlyOwnData(t *testing.T) {
	t.Skip("DLD-739: 구현 전")

	tdb := testutil.NewTestDB(t)
	e := setupSyncEcho(t, tdb)

	// Seed user A and their todo.
	seedUserAndLogin(t, tdb, e, "sync-user-a@example.com", "Secret1!")
	tdb.Seed(t,
		`INSERT INTO todos (user_id, title, type, created_at, updated_at)
		 VALUES (
			(SELECT id FROM users WHERE email = 'sync-user-a@example.com'),
			'사용자A 개인 투두', 'personal', datetime('now'), datetime('now')
		 )`,
	)

	// User B logs in and syncs.
	tokenB := syncToken(t, tdb, e, "sync-user-b@example.com")

	req := httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(`{"changes":[]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, tokenB)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v (body: %s)", err, rec.Body.String())
	}
	serverData, ok := resp["server_data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'server_data' to be an object")
	}
	todos, _ := serverData["todos"].([]interface{})
	for _, item := range todos {
		todo, _ := item.(map[string]interface{})
		if todo["title"] == "사용자A 개인 투두" {
			t.Error("user B's sync response must not contain user A's private todo (data isolation failure)")
		}
	}
}
