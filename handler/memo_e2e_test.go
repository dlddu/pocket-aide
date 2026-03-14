// Package handler_test contains end-to-end style tests for the memo (scratch
// space) API endpoints (POST/GET/PUT/DELETE /memos, POST /memos/:id/move).
//
// DLD-729: 8-1: 임시 공간 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (t.Skip). Activate after DLD-730:
//   - handler.NewMemoHandler is implemented (handler/memo.go)
//   - db/migrations/000006_memos.up.sql migration is applied
//   - Routes /memos, /memos/:id, /memos/:id/move are registered
//   - POST /memos accepts content and source fields (source: "text" | "voice")
//   - POST /memos/:id/move converts memo to todo by target type
//     (personal_todo, work_todo, routine) and deletes the original memo
//
// When activating:
//  1. Remove all t.Skip calls.
//  2. Uncomment the memo handler lines inside setupMemoEcho (marked TODO).
//  3. Uncomment the todo handler lines inside setupMemoEcho (marked TODO).
//  4. Ensure the memos migration exists so testutil.NewTestDB picks it up.
package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/handler"
	appmiddleware "github.com/dlddu/pocket-aide/middleware"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// setupMemoEcho builds an Echo instance with the auth route and — once
// DLD-730 is implemented — the memo routes and todo routes (needed for move
// verification).
//
// Activation checklist:
//
//	TODO(DLD-730): uncomment the lines below after handler.NewMemoHandler exists.
//	  memoHandler := handler.NewMemoHandler(tdb.DB)
//	  mg := e.Group("/memos", appmiddleware.JWT("test-jwt-secret"))
//	  mg.POST("",        memoHandler.Create)
//	  mg.GET("",         memoHandler.List)
//	  mg.PUT("/:id",     memoHandler.Update)
//	  mg.DELETE("/:id",  memoHandler.Delete)
//	  mg.POST("/:id/move", memoHandler.Move)
//
//	TODO(DLD-730): uncomment the lines below after handler.NewTodoHandler
//	  supports listing by type (needed for POST /memos/:id/move verification).
//	  todoHandler := handler.NewTodoHandler(tdb.DB)
//	  tg := e.Group("/todos", appmiddleware.JWT("test-jwt-secret"))
//	  tg.GET("", todoHandler.List)
func setupMemoEcho(t *testing.T, tdb *testutil.TestDB) *echo.Echo {
	t.Helper()
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)
	memoHandler := handler.NewMemoHandler(tdb.DB)
	mg := e.Group("/memos", appmiddleware.JWT("test-jwt-secret"))
	mg.POST("",          memoHandler.Create)
	mg.GET("",           memoHandler.List)
	mg.PUT("/:id",       memoHandler.Update)
	mg.DELETE("/:id",    memoHandler.Delete)
	mg.POST("/:id/move", memoHandler.Move)

	// Todo routes needed for move verification
	todoHandler := handler.NewTodoHandler(tdb.DB)
	tg := e.Group("/todos", appmiddleware.JWT("test-jwt-secret"))
	tg.GET("", todoHandler.List)
	return e
}

// memoToken seeds a user with the given email and returns a JWT Bearer token
// string (including the "Bearer " prefix) by reusing seedUserAndLogin defined
// in chat_e2e_test.go (same package).
func memoToken(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email string) string {
	t.Helper()
	return "Bearer " + seedUserAndLogin(t, tdb, e, email, "Secret1!")
}

// ---------------------------------------------------------------------------
// CRUD — POST /memos
// ---------------------------------------------------------------------------

// TestMemoHandler_Create_ReturnsCreated verifies that POST /memos with valid
// content and source fields returns HTTP 201 and a JSON body that includes id,
// content, and source.
//
// Scenario:
//
//	POST /memos  {"content":"장보기 목록 작성","source":"text"} + Bearer token
//	→ 201 Created  {"id":1,"content":"장보기 목록 작성","source":"text"}
func TestMemoHandler_Create_ReturnsCreated(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-create@example.com")

	body := strings.NewReader(`{"content":"장보기 목록 작성","source":"text"}`)
	req := httptest.NewRequest(http.MethodPost, "/memos", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v (body: %s)", err, rec.Body.String())
	}
	if _, ok := resp["id"]; !ok {
		t.Error("expected response to contain 'id' field")
	}
	if resp["content"] != "장보기 목록 작성" {
		t.Errorf("expected content '장보기 목록 작성', got %v", resp["content"])
	}
	if resp["source"] != "text" {
		t.Errorf("expected source 'text', got %v", resp["source"])
	}
}

// TestMemoHandler_Create_MissingContent_ReturnsBadRequest verifies that
// POST /memos without a content field returns HTTP 400 Bad Request.
//
// Scenario:
//
//	POST /memos  {"source":"text"} + Bearer token
//	→ 400 Bad Request
func TestMemoHandler_Create_MissingContent_ReturnsBadRequest(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-nocontent@example.com")

	body := strings.NewReader(`{"source":"text"}`)
	req := httptest.NewRequest(http.MethodPost, "/memos", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestMemoHandler_Create_WithoutAuth_ReturnsUnauthorized verifies that
// POST /memos without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /memos  {"content":"메모","source":"text"}  (no Authorization header)
//	→ 401 Unauthorized
func TestMemoHandler_Create_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	e := echo.New()
	// Register a guarded placeholder so the JWT middleware can reject the
	// request before any real handler runs. This will be replaced by the
	// real memo handler once implemented.
	e.POST("/memos", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, nil)
	}, appmiddleware.JWT("test-jwt-secret"))

	body := strings.NewReader(`{"content":"메모","source":"text"}`)
	req := httptest.NewRequest(http.MethodPost, "/memos", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — GET /memos
// ---------------------------------------------------------------------------

// TestMemoHandler_List_ReturnsList verifies that GET /memos returns HTTP 200
// and a JSON array containing the authenticated user's memos.
//
// Scenario:
//
//	Seed: two memos created via POST /memos.
//	GET /memos  + Bearer token
//	→ 200 OK  JSON array with 2 elements, each having id/content/source
func TestMemoHandler_List_ReturnsList(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-list@example.com")

	// Seed two memos
	for _, content := range []string{"첫 번째 메모", "두 번째 메모"} {
		payload := fmt.Sprintf(`{"content":"%s","source":"text"}`, content)
		createReq := httptest.NewRequest(http.MethodPost, "/memos", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed memo %q failed: status %d, body: %s", content, createRec.Code, createRec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/memos", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var memos []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &memos); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(memos) != 2 {
		t.Errorf("expected 2 memos, got %d", len(memos))
	}
}

// TestMemoHandler_List_EmptyList_ReturnsEmptyArray verifies that GET /memos
// for a user with no memos returns HTTP 200 and an empty JSON array (not null).
//
// Scenario:
//
//	Seed: authenticated user with no memos.
//	GET /memos  + Bearer token
//	→ 200 OK  []
func TestMemoHandler_List_EmptyList_ReturnsEmptyArray(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-empty@example.com")

	req := httptest.NewRequest(http.MethodGet, "/memos", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var memos []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &memos); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(memos) != 0 {
		t.Errorf("expected empty array [], got %d items", len(memos))
	}
}

// TestMemoHandler_List_WithoutAuth_ReturnsUnauthorized verifies that
// GET /memos without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	GET /memos  (no Authorization header)
//	→ 401 Unauthorized
func TestMemoHandler_List_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	e := echo.New()
	e.GET("/memos", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []interface{}{})
	}, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodGet, "/memos", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// CRUD — PUT /memos/:id
// ---------------------------------------------------------------------------

// TestMemoHandler_Update_ReturnsOK verifies that PUT /memos/:id with a valid
// updated content returns HTTP 200 and the updated JSON representation.
//
// Scenario:
//
//	Seed: one memo with content "원래 메모".
//	PUT /memos/<id>  {"content":"수정된 메모"} + Bearer token
//	→ 200 OK  {"id":<id>,"content":"수정된 메모","source":"text"}
func TestMemoHandler_Update_ReturnsOK(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-update@example.com")

	// Create a memo
	createBody := strings.NewReader(`{"content":"원래 메모","source":"text"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/memos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create memo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	updateBody := strings.NewReader(`{"content":"수정된 메모"}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/memos/%d", id), updateBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var updated map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("expected valid JSON object, got: %v (body: %s)", err, rec.Body.String())
	}
	if updated["content"] != "수정된 메모" {
		t.Errorf("expected updated content '수정된 메모', got %v", updated["content"])
	}
}

// TestMemoHandler_Update_NotFound_ReturnsNotFound verifies that
// PUT /memos/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	PUT /memos/99999  {"content":"없는 메모"} + Bearer token
//	→ 404 Not Found
func TestMemoHandler_Update_NotFound_ReturnsNotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-update-notfound@example.com")

	body := strings.NewReader(`{"content":"없는 메모"}`)
	req := httptest.NewRequest(http.MethodPut, "/memos/99999", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — DELETE /memos/:id
// ---------------------------------------------------------------------------

// TestMemoHandler_Delete_ReturnsNoContent verifies that DELETE /memos/:id
// returns HTTP 204 and that a subsequent GET /memos no longer includes the
// deleted memo.
//
// Scenario:
//
//	Seed: one memo.
//	DELETE /memos/<id>  + Bearer token → 204 No Content
//	GET    /memos       + Bearer token → array does not contain the deleted id
func TestMemoHandler_Delete_ReturnsNoContent(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-delete@example.com")

	// Create a memo to delete
	createBody := strings.NewReader(`{"content":"삭제할 메모","source":"text"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/memos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create memo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — delete the memo
	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/memos/%d", id), nil)
	delReq.Header.Set(echo.HeaderAuthorization, token)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d (body: %s)", delRec.Code, delRec.Body.String())
	}

	// Assert — subsequent GET /memos must not contain the deleted memo
	getReq := httptest.NewRequest(http.MethodGet, "/memos", nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on list after delete, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var memos []map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &memos); err != nil {
		t.Fatalf("expected valid JSON array after delete, got: %v (body: %s)", err, getRec.Body.String())
	}
	for _, m := range memos {
		if int(m["id"].(float64)) == id {
			t.Errorf("deleted memo (id=%d) still appears in GET /memos", id)
		}
	}
}

// TestMemoHandler_Delete_NotFound_ReturnsNotFound verifies that
// DELETE /memos/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	DELETE /memos/99999  + Bearer token
//	→ 404 Not Found
func TestMemoHandler_Delete_NotFound_ReturnsNotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-delete-notfound@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/memos/99999", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Move — POST /memos/:id/move
// ---------------------------------------------------------------------------

// TestMemoHandler_Move_ToPersonalTodo_ReturnsOK verifies the core scratch
// space flow: moving a memo to personal_todo converts it to a todo and removes
// it from the memo list.
//
// Scenario:
//
//	Seed: one memo with content "운동 계획 세우기".
//	POST /memos/<id>/move  {"target":"personal_todo"} + Bearer token
//	→ 200 OK
//	GET /memos              + Bearer token → memo no longer in list
//	GET /todos?type=personal + Bearer token → new todo with title matching memo content
func TestMemoHandler_Move_ToPersonalTodo_ReturnsOK(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-move@example.com")

	// Create a memo to move
	createBody := strings.NewReader(`{"content":"운동 계획 세우기","source":"text"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/memos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create memo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — move memo to personal_todo
	moveBody := strings.NewReader(`{"target":"personal_todo"}`)
	moveReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/memos/%d/move", id), moveBody)
	moveReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	moveReq.Header.Set(echo.HeaderAuthorization, token)
	moveRec := httptest.NewRecorder()
	e.ServeHTTP(moveRec, moveReq)

	if moveRec.Code != http.StatusOK {
		t.Errorf("expected status 200 on move, got %d (body: %s)", moveRec.Code, moveRec.Body.String())
	}

	// Assert — GET /memos must no longer contain the moved memo
	listReq := httptest.NewRequest(http.MethodGet, "/memos", nil)
	listReq.Header.Set(echo.HeaderAuthorization, token)
	listRec := httptest.NewRecorder()
	e.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on GET /memos after move, got %d (body: %s)", listRec.Code, listRec.Body.String())
	}
	var memos []map[string]interface{}
	if err := json.Unmarshal(listRec.Body.Bytes(), &memos); err != nil {
		t.Fatalf("expected valid JSON array from GET /memos, got: %v (body: %s)", err, listRec.Body.String())
	}
	for _, m := range memos {
		if int(m["id"].(float64)) == id {
			t.Errorf("moved memo (id=%d) still appears in GET /memos after move", id)
		}
	}

	// Assert — GET /todos?type=personal must contain a new todo derived from the memo
	todoListReq := httptest.NewRequest(http.MethodGet, "/todos?type=personal", nil)
	todoListReq.Header.Set(echo.HeaderAuthorization, token)
	todoListRec := httptest.NewRecorder()
	e.ServeHTTP(todoListRec, todoListReq)

	if todoListRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on GET /todos?type=personal, got %d (body: %s)", todoListRec.Code, todoListRec.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(todoListRec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("expected valid JSON array from GET /todos, got: %v (body: %s)", err, todoListRec.Body.String())
	}
	found := false
	for _, todo := range todos {
		if todo["title"] == "운동 계획 세우기" && todo["type"] == "personal" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a personal todo with title '운동 계획 세우기' after move, but not found in GET /todos?type=personal")
	}
}

// TestMemoHandler_Move_NotFound_ReturnsNotFound verifies that
// POST /memos/:id/move for a non-existent memo ID returns HTTP 404.
//
// Scenario:
//
//	POST /memos/99999/move  {"target":"personal_todo"} + Bearer token
//	→ 404 Not Found
func TestMemoHandler_Move_NotFound_ReturnsNotFound(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupMemoEcho(t, tdb)
	token := memoToken(t, tdb, e, "memo-move-notfound@example.com")

	body := strings.NewReader(`{"target":"personal_todo"}`)
	req := httptest.NewRequest(http.MethodPost, "/memos/99999/move", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
