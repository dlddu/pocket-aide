// Package handler_test contains end-to-end style tests for the personal todo
// API endpoints (POST/GET/PUT/DELETE /todos, POST /todos/:id/toggle).
//
// DLD-725: 6-1: 개인 투두 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (t.Skip). Activate after DLD-725:
//   - handler.NewTodoHandler is implemented (handler/todo.go)
//   - db/migrations/000004_todos.up.sql migration is applied
//   - Routes /todos, /todos/:id, /todos/:id/toggle are registered
//   - Todo type filtering (?type=personal) is in place
//   - POST /todos/:id/toggle toggles completed_at between null/now
//
// When activating:
//  1. Remove all t.Skip calls.
//  2. Uncomment the todo handler lines inside setupTodoEcho (marked TODO).
//  3. Ensure 000004_todos.up.sql exists so testutil.NewTestDB picks it up.
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

// setupTodoEcho builds an Echo instance with the auth route and — once
// DLD-725 is implemented — the todo routes.
//
// Activation checklist:
//
//	TODO(DLD-725): uncomment the lines below after handler.NewTodoHandler exists.
//	  todoHandler := handler.NewTodoHandler(tdb.DB)
//	  g := e.Group("/todos", appmiddleware.JWT("test-jwt-secret"))
//	  g.POST("",              todoHandler.Create)
//	  g.GET("",               todoHandler.List)
//	  g.GET("/:id",           todoHandler.Get)
//	  g.PUT("/:id",           todoHandler.Update)
//	  g.DELETE("/:id",        todoHandler.Delete)
//	  g.POST("/:id/toggle",   todoHandler.Toggle)
func setupTodoEcho(t *testing.T, tdb *testutil.TestDB) *echo.Echo {
	t.Helper()
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)
	// TODO(DLD-725): uncomment after handler.NewTodoHandler exists
	//   todoHandler := handler.NewTodoHandler(tdb.DB)
	//   g := e.Group("/todos", appmiddleware.JWT("test-jwt-secret"))
	//   g.POST("",            todoHandler.Create)
	//   g.GET("",             todoHandler.List)
	//   g.GET("/:id",         todoHandler.Get)
	//   g.PUT("/:id",         todoHandler.Update)
	//   g.DELETE("/:id",      todoHandler.Delete)
	//   g.POST("/:id/toggle", todoHandler.Toggle)
	return e
}

// todoToken seeds a user with the given email and returns a JWT Bearer token
// string (including the "Bearer " prefix) by reusing seedUserAndLogin defined
// in chat_e2e_test.go (same package).
func todoToken(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email string) string {
	t.Helper()
	return "Bearer " + seedUserAndLogin(t, tdb, e, email, "Secret1!")
}

// ---------------------------------------------------------------------------
// CRUD — POST /todos
// ---------------------------------------------------------------------------

// TestTodoHandler_Create_ReturnsCreated verifies that POST /todos with a valid
// title and type=personal returns HTTP 201 and a JSON body that includes id,
// title, type, and completed_at (null for a new todo).
//
// Scenario:
//
//	POST /todos  {"title":"장보기","type":"personal"} + Bearer token
//	→ 201 Created   {"id":1,"title":"장보기","type":"personal","completed_at":null}
func TestTodoHandler_Create_ReturnsCreated(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-create@example.com")

	body := strings.NewReader(`{"title":"장보기","type":"personal"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
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
	if resp["title"] != "장보기" {
		t.Errorf("expected title '장보기', got %v", resp["title"])
	}
	if resp["type"] != "personal" {
		t.Errorf("expected type 'personal', got %v", resp["type"])
	}
	// completed_at must be null for a newly created todo
	if completedAt, ok := resp["completed_at"]; ok && completedAt != nil {
		t.Errorf("expected completed_at to be null, got %v", completedAt)
	}
}

// TestTodoHandler_Create_MissingTitle_ReturnsBadRequest verifies that
// POST /todos without a title field returns HTTP 400 Bad Request.
//
// Scenario:
//
//	POST /todos  {"type":"personal"} + Bearer token
//	→ 400 Bad Request
func TestTodoHandler_Create_MissingTitle_ReturnsBadRequest(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-notitle@example.com")

	body := strings.NewReader(`{"type":"personal"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestTodoHandler_Create_WithoutAuth_ReturnsUnauthorized verifies that
// POST /todos without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /todos  {"title":"운동","type":"personal"}  (no Authorization header)
//	→ 401 Unauthorized
func TestTodoHandler_Create_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	e := echo.New()
	// Register a guarded placeholder so the JWT middleware can reject the
	// request before any real handler runs.  This will be replaced by the
	// real todo handler once implemented.
	e.POST("/todos", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, nil)
	}, appmiddleware.JWT("test-jwt-secret"))

	body := strings.NewReader(`{"title":"운동","type":"personal"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — GET /todos?type=personal
// ---------------------------------------------------------------------------

// TestTodoHandler_List_PersonalTodos_ReturnsList verifies that
// GET /todos?type=personal returns HTTP 200 and a JSON array containing only
// the authenticated user's personal todos.
//
// Scenario:
//
//	Seed: two personal todos created via POST /todos.
//	GET /todos?type=personal  + Bearer token
//	→ 200 OK  JSON array with 2 elements
func TestTodoHandler_List_PersonalTodos_ReturnsList(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-list@example.com")

	// Seed two personal todos
	for _, title := range []string{"장보기", "독서"} {
		payload := fmt.Sprintf(`{"title":"%s","type":"personal"}`, title)
		createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed todo %q failed: status %d, body: %s", title, createRec.Code, createRec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/todos?type=personal", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(todos) != 2 {
		t.Errorf("expected 2 todos, got %d", len(todos))
	}
	for i, todo := range todos {
		if todo["type"] != "personal" {
			t.Errorf("todos[%d]: expected type 'personal', got %v", i, todo["type"])
		}
	}
}

// TestTodoHandler_List_EmptyList_ReturnsEmptyArray verifies that
// GET /todos?type=personal for a user with no todos returns HTTP 200 and an
// empty JSON array (not null).
//
// Scenario:
//
//	Seed: authenticated user with no todos.
//	GET /todos?type=personal  + Bearer token
//	→ 200 OK  []
func TestTodoHandler_List_EmptyList_ReturnsEmptyArray(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-empty@example.com")

	req := httptest.NewRequest(http.MethodGet, "/todos?type=personal", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(todos) != 0 {
		t.Errorf("expected empty array [], got %d items", len(todos))
	}
}

// TestTodoHandler_List_WithoutAuth_ReturnsUnauthorized verifies that
// GET /todos without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	GET /todos  (no Authorization header)
//	→ 401 Unauthorized
func TestTodoHandler_List_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	e := echo.New()
	e.GET("/todos", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []interface{}{})
	}, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// CRUD — GET /todos/:id
// ---------------------------------------------------------------------------

// TestTodoHandler_Get_ReturnsSingleTodo verifies that GET /todos/:id returns
// HTTP 200 and the JSON representation of the requested todo.
//
// Scenario:
//
//	Seed: one todo created via POST /todos.
//	GET /todos/<id>  + Bearer token
//	→ 200 OK  {"id":<id>,"title":"운동하기","type":"personal","completed_at":null}
func TestTodoHandler_Get_ReturnsSingleTodo(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-get@example.com")

	// Create a todo to retrieve
	createBody := strings.NewReader(`{"title":"운동하기","type":"personal"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", id), nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var todo map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &todo); err != nil {
		t.Fatalf("expected valid JSON object, got: %v (body: %s)", err, rec.Body.String())
	}
	if todo["title"] != "운동하기" {
		t.Errorf("expected title '운동하기', got %v", todo["title"])
	}
	if todo["type"] != "personal" {
		t.Errorf("expected type 'personal', got %v", todo["type"])
	}
}

// TestTodoHandler_Get_NotFound_ReturnsNotFound verifies that GET /todos/:id
// for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	GET /todos/99999  + Bearer token
//	→ 404 Not Found
func TestTodoHandler_Get_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-get-notfound@example.com")

	req := httptest.NewRequest(http.MethodGet, "/todos/99999", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — PUT /todos/:id
// ---------------------------------------------------------------------------

// TestTodoHandler_Update_ReturnsOK verifies that PUT /todos/:id with a valid
// updated title returns HTTP 200 and the updated JSON representation.
//
// Scenario:
//
//	Seed: one todo with title "운동하기".
//	PUT /todos/<id>  {"title":"운동하기 (30분)"} + Bearer token
//	→ 200 OK  {"id":<id>,"title":"운동하기 (30분)","type":"personal", ...}
func TestTodoHandler_Update_ReturnsOK(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-update@example.com")

	// Create a todo
	createBody := strings.NewReader(`{"title":"운동하기","type":"personal"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	updateBody := strings.NewReader(`{"title":"운동하기 (30분)"}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/todos/%d", id), updateBody)
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
	if updated["title"] != "운동하기 (30분)" {
		t.Errorf("expected updated title '운동하기 (30분)', got %v", updated["title"])
	}
}

// TestTodoHandler_Update_NotFound_ReturnsNotFound verifies that
// PUT /todos/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	PUT /todos/99999  {"title":"없는 투두"} + Bearer token
//	→ 404 Not Found
func TestTodoHandler_Update_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-update-notfound@example.com")

	body := strings.NewReader(`{"title":"없는 투두"}`)
	req := httptest.NewRequest(http.MethodPut, "/todos/99999", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — DELETE /todos/:id
// ---------------------------------------------------------------------------

// TestTodoHandler_Delete_ReturnsNoContent verifies that DELETE /todos/:id
// returns HTTP 204 and that a subsequent GET returns 404.
//
// Scenario:
//
//	Seed: one todo.
//	DELETE /todos/<id>  + Bearer token → 204 No Content
//	GET    /todos/<id>  + Bearer token → 404 Not Found
func TestTodoHandler_Delete_ReturnsNoContent(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-delete@example.com")

	// Create a todo to delete
	createBody := strings.NewReader(`{"title":"삭제할 투두","type":"personal"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — delete the todo
	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/todos/%d", id), nil)
	delReq.Header.Set(echo.HeaderAuthorization, token)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d (body: %s)", delRec.Code, delRec.Body.String())
	}

	// Assert — subsequent GET must return 404
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", id), nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 after delete, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
}

// TestTodoHandler_Delete_NotFound_ReturnsNotFound verifies that
// DELETE /todos/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	DELETE /todos/99999  + Bearer token
//	→ 404 Not Found
func TestTodoHandler_Delete_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-delete-notfound@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/todos/99999", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Toggle — POST /todos/:id/toggle
// ---------------------------------------------------------------------------

// TestTodoHandler_Toggle_CompletesTodo verifies that POST /todos/:id/toggle
// on a pending todo sets completed_at to a non-null value.
//
// Scenario:
//
//	Seed: one pending todo (completed_at == null).
//	POST /todos/<id>/toggle  + Bearer token → 200 OK
//	GET  /todos/<id>         + Bearer token → completed_at is non-null
func TestTodoHandler_Toggle_CompletesTodo(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-toggle-complete@example.com")

	// Create a pending todo
	createBody := strings.NewReader(`{"title":"책 읽기","type":"personal"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Precondition: completed_at must be null
	if completedAt := created["completed_at"]; completedAt != nil {
		t.Fatalf("precondition: expected completed_at null on creation, got %v", completedAt)
	}

	// Act — toggle (pending → completed)
	toggleReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/todos/%d/toggle", id), nil)
	toggleReq.Header.Set(echo.HeaderAuthorization, token)
	toggleRec := httptest.NewRecorder()
	e.ServeHTTP(toggleRec, toggleReq)

	if toggleRec.Code != http.StatusOK {
		t.Errorf("expected status 200 on toggle, got %d (body: %s)", toggleRec.Code, toggleRec.Body.String())
	}

	// Assert — GET must return non-null completed_at
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", id), nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on GET after toggle, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var todo map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &todo); err != nil {
		t.Fatalf("expected valid JSON object, got: %v (body: %s)", err, getRec.Body.String())
	}
	if todo["completed_at"] == nil {
		t.Error("expected completed_at to be non-null after toggle (pending → completed)")
	}
}

// TestTodoHandler_Toggle_UncompletesTodo verifies that POST /todos/:id/toggle
// called twice on the same todo returns completed_at back to null.
//
// Scenario:
//
//	Seed: one pending todo.
//	POST /todos/<id>/toggle  (1st call) → completed_at is non-null
//	POST /todos/<id>/toggle  (2nd call) → completed_at is null again
func TestTodoHandler_Toggle_UncompletesTodo(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-toggle-uncomplete@example.com")

	// Create a pending todo
	createBody := strings.NewReader(`{"title":"일기 쓰기","type":"personal"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	toggleURL := fmt.Sprintf("/todos/%d/toggle", id)

	// 1st toggle: pending → completed
	toggle1Req := httptest.NewRequest(http.MethodPost, toggleURL, nil)
	toggle1Req.Header.Set(echo.HeaderAuthorization, token)
	toggle1Rec := httptest.NewRecorder()
	e.ServeHTTP(toggle1Rec, toggle1Req)
	if toggle1Rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on 1st toggle, got %d (body: %s)", toggle1Rec.Code, toggle1Rec.Body.String())
	}

	// 2nd toggle: completed → pending
	toggle2Req := httptest.NewRequest(http.MethodPost, toggleURL, nil)
	toggle2Req.Header.Set(echo.HeaderAuthorization, token)
	toggle2Rec := httptest.NewRecorder()
	e.ServeHTTP(toggle2Rec, toggle2Req)
	if toggle2Rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on 2nd toggle, got %d (body: %s)", toggle2Rec.Code, toggle2Rec.Body.String())
	}

	// Assert — GET must return null completed_at
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", id), nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on GET after 2nd toggle, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var todo map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &todo); err != nil {
		t.Fatalf("expected valid JSON object, got: %v (body: %s)", err, getRec.Body.String())
	}
	if todo["completed_at"] != nil {
		t.Errorf("expected completed_at to be null after 2nd toggle (completed → pending), got %v", todo["completed_at"])
	}
}

// ---------------------------------------------------------------------------
// List — Pending / Completed separation
// ---------------------------------------------------------------------------

// TestTodoHandler_List_SeparatesCompletedAndPending verifies that
// GET /todos?type=personal returns todos with a completed_at field that is
// null for pending todos and non-null for completed todos, allowing the client
// to render separate sections.
//
// Scenario:
//
//	Seed: two todos; toggle one to completed.
//	GET /todos?type=personal  + Bearer token
//	→ 200 OK  array contains both todos; completed one has non-null completed_at,
//	           pending one has null completed_at
func TestTodoHandler_List_SeparatesCompletedAndPending(t *testing.T) {
	t.Skip("DLD-725 구현 완료 후 활성화: setupTodoEcho의 TODO 주석 해제 필요")

	tdb := testutil.NewTestDB(t)
	e := setupTodoEcho(t, tdb)
	token := todoToken(t, tdb, e, "todo-list-separated@example.com")

	// Create two todos
	var todoIDs []int
	for _, title := range []string{"진행중 투두", "완료할 투두"} {
		payload := fmt.Sprintf(`{"title":"%s","type":"personal"}`, title)
		createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed todo %q failed: status %d, body: %s", title, createRec.Code, createRec.Body.String())
		}
		var created map[string]interface{}
		if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to unmarshal create response: %v", err)
		}
		todoIDs = append(todoIDs, int(created["id"].(float64)))
	}

	// Toggle the second todo to completed
	toggleReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/todos/%d/toggle", todoIDs[1]), nil)
	toggleReq.Header.Set(echo.HeaderAuthorization, token)
	toggleRec := httptest.NewRecorder()
	e.ServeHTTP(toggleRec, toggleReq)
	if toggleRec.Code != http.StatusOK {
		t.Fatalf("toggle failed: status %d, body: %s", toggleRec.Code, toggleRec.Body.String())
	}

	// Act
	req := httptest.NewRequest(http.MethodGet, "/todos?type=personal", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(todos))
	}

	// Both todos must expose the completed_at field
	pendingCount := 0
	completedCount := 0
	for _, todo := range todos {
		if _, ok := todo["completed_at"]; !ok {
			t.Error("expected every todo to expose 'completed_at' field")
			continue
		}
		if todo["completed_at"] == nil {
			pendingCount++
		} else {
			completedCount++
		}
	}
	if pendingCount != 1 {
		t.Errorf("expected 1 pending todo (completed_at == null), got %d", pendingCount)
	}
	if completedCount != 1 {
		t.Errorf("expected 1 completed todo (completed_at != null), got %d", completedCount)
	}
}
