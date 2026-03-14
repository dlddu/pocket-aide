// Package handler_test contains end-to-end style tests for the work (company)
// todo API endpoints (POST/GET/PUT/DELETE /todos, POST /todos/:id/toggle)
// with type=work and priority support.
//
// DLD-727: 7-1: 회사 투두 — e2e 테스트 작성
//
// These tests require:
//   - handler.NewTodoHandler is implemented (handler/todo.go)
//   - db/migrations/000004_todos.up.sql migration is applied
//   - Routes /todos, /todos/:id, /todos/:id/toggle are registered
//   - Todo type filtering (?type=work) is in place
//   - POST /todos/:id/toggle toggles completed_at between null/now
//   - priority column added to todos table (high/medium/low)
//   - GET /todos?type=work returns todos sorted by priority (high → medium → low)
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

// setupWorkTodoEcho builds an Echo instance with the auth route and the todo
// routes scoped to work-type operations.
func setupWorkTodoEcho(t *testing.T, tdb *testutil.TestDB) *echo.Echo {
	t.Helper()
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)
	todoHandler := handler.NewTodoHandler(tdb.DB)
	g := e.Group("/todos", appmiddleware.JWT("test-jwt-secret"))
	g.POST("", todoHandler.Create)
	g.GET("", todoHandler.List)
	g.GET("/:id", todoHandler.Get)
	g.PUT("/:id", todoHandler.Update)
	g.DELETE("/:id", todoHandler.Delete)
	g.POST("/:id/toggle", todoHandler.Toggle)
	return e
}

// workTodoToken seeds a user with the given email and returns a JWT Bearer
// token string (including the "Bearer " prefix) by reusing seedUserAndLogin
// defined in chat_e2e_test.go (same package).
func workTodoToken(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email string) string {
	t.Helper()
	return "Bearer " + seedUserAndLogin(t, tdb, e, email, "Secret1!")
}

// ---------------------------------------------------------------------------
// CRUD — POST /todos (type=work)
// ---------------------------------------------------------------------------

// TestWorkTodoHandler_Create_ReturnsCreated verifies that POST /todos with a
// valid title and type=work returns HTTP 201 and a JSON body that includes id,
// title, type, and completed_at (null for a new todo).
//
// Scenario:
//
//	POST /todos  {"title":"기획서 작성","type":"work"} + Bearer token
//	→ 201 Created   {"id":1,"title":"기획서 작성","type":"work","completed_at":null}
func TestWorkTodoHandler_Create_ReturnsCreated(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-create@example.com")

	body := strings.NewReader(`{"title":"기획서 작성","type":"work"}`)
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
	if resp["title"] != "기획서 작성" {
		t.Errorf("expected title '기획서 작성', got %v", resp["title"])
	}
	if resp["type"] != "work" {
		t.Errorf("expected type 'work', got %v", resp["type"])
	}
	// completed_at must be null for a newly created todo
	if completedAt, ok := resp["completed_at"]; ok && completedAt != nil {
		t.Errorf("expected completed_at to be null, got %v", completedAt)
	}
}

// TestWorkTodoHandler_Create_MissingTitle_ReturnsBadRequest verifies that
// POST /todos without a title field returns HTTP 400 Bad Request.
//
// Scenario:
//
//	POST /todos  {"type":"work"} + Bearer token
//	→ 400 Bad Request
func TestWorkTodoHandler_Create_MissingTitle_ReturnsBadRequest(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-notitle@example.com")

	body := strings.NewReader(`{"type":"work"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestWorkTodoHandler_Create_WithoutAuth_ReturnsUnauthorized verifies that
// POST /todos without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /todos  {"title":"회의 준비","type":"work"}  (no Authorization header)
//	→ 401 Unauthorized
func TestWorkTodoHandler_Create_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	e := echo.New()
	// Register a guarded placeholder so the JWT middleware can reject the
	// request before any real handler runs.
	e.POST("/todos", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, nil)
	}, appmiddleware.JWT("test-jwt-secret"))

	body := strings.NewReader(`{"title":"회의 준비","type":"work"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — GET /todos?type=work
// ---------------------------------------------------------------------------

// TestWorkTodoHandler_List_WorkTodos_ReturnsList verifies that
// GET /todos?type=work returns HTTP 200 and a JSON array containing only
// the authenticated user's work todos.
//
// Scenario:
//
//	Seed: two work todos created via POST /todos.
//	GET /todos?type=work  + Bearer token
//	→ 200 OK  JSON array with 2 elements; each element has type="work"
func TestWorkTodoHandler_List_WorkTodos_ReturnsList(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-list@example.com")

	// Seed two work todos
	for _, title := range []string{"기획서 작성", "코드 리뷰"} {
		payload := fmt.Sprintf(`{"title":"%s","type":"work"}`, title)
		createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed work todo %q failed: status %d, body: %s", title, createRec.Code, createRec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/todos?type=work", nil)
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
		if todo["type"] != "work" {
			t.Errorf("todos[%d]: expected type 'work', got %v", i, todo["type"])
		}
	}
}

// TestWorkTodoHandler_List_ExcludesPersonalTodos_ReturnsWorkOnly verifies that
// GET /todos?type=work does not return todos with type=personal even when both
// types exist for the same user.
//
// Scenario:
//
//	Seed: one work todo and one personal todo.
//	GET /todos?type=work  + Bearer token
//	→ 200 OK  JSON array with 1 element; element has type="work"
func TestWorkTodoHandler_List_ExcludesPersonalTodos_ReturnsWorkOnly(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-filter@example.com")

	// Seed one work todo and one personal todo
	for _, item := range []struct{ title, todoType string }{
		{"업무 보고서", "work"},
		{"장보기", "personal"},
	} {
		payload := fmt.Sprintf(`{"title":"%s","type":"%s"}`, item.title, item.todoType)
		createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed todo %q (%s) failed: status %d, body: %s", item.title, item.todoType, createRec.Code, createRec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/todos?type=work", nil)
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
	if len(todos) != 1 {
		t.Errorf("expected 1 work todo (personal excluded), got %d", len(todos))
	}
	if len(todos) > 0 && todos[0]["type"] != "work" {
		t.Errorf("expected type 'work', got %v", todos[0]["type"])
	}
}

// TestWorkTodoHandler_List_EmptyList_ReturnsEmptyArray verifies that
// GET /todos?type=work for a user with no work todos returns HTTP 200 and an
// empty JSON array (not null).
//
// Scenario:
//
//	Seed: authenticated user with no todos.
//	GET /todos?type=work  + Bearer token
//	→ 200 OK  []
func TestWorkTodoHandler_List_EmptyList_ReturnsEmptyArray(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-empty@example.com")

	req := httptest.NewRequest(http.MethodGet, "/todos?type=work", nil)
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

// ---------------------------------------------------------------------------
// CRUD — GET /todos/:id
// ---------------------------------------------------------------------------

// TestWorkTodoHandler_Get_ReturnsSingleTodo verifies that GET /todos/:id
// returns HTTP 200 and the JSON representation of the requested work todo.
//
// Scenario:
//
//	Seed: one work todo created via POST /todos.
//	GET /todos/<id>  + Bearer token
//	→ 200 OK  {"id":<id>,"title":"회의록 정리","type":"work","completed_at":null}
func TestWorkTodoHandler_Get_ReturnsSingleTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-get@example.com")

	// Create a work todo to retrieve
	createBody := strings.NewReader(`{"title":"회의록 정리","type":"work"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create work todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
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
	if todo["title"] != "회의록 정리" {
		t.Errorf("expected title '회의록 정리', got %v", todo["title"])
	}
	if todo["type"] != "work" {
		t.Errorf("expected type 'work', got %v", todo["type"])
	}
}

// ---------------------------------------------------------------------------
// CRUD — PUT /todos/:id
// ---------------------------------------------------------------------------

// TestWorkTodoHandler_Update_ReturnsOK verifies that PUT /todos/:id with a
// valid updated title returns HTTP 200 and the updated JSON representation.
//
// Scenario:
//
//	Seed: one work todo with title "주간 보고서".
//	PUT /todos/<id>  {"title":"주간 보고서 (수정)"} + Bearer token
//	→ 200 OK  {"id":<id>,"title":"주간 보고서 (수정)","type":"work", ...}
func TestWorkTodoHandler_Update_ReturnsOK(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-update@example.com")

	// Create a work todo
	createBody := strings.NewReader(`{"title":"주간 보고서","type":"work"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create work todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	updateBody := strings.NewReader(`{"title":"주간 보고서 (수정)"}`)
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
	if updated["title"] != "주간 보고서 (수정)" {
		t.Errorf("expected updated title '주간 보고서 (수정)', got %v", updated["title"])
	}
}

// ---------------------------------------------------------------------------
// CRUD — DELETE /todos/:id
// ---------------------------------------------------------------------------

// TestWorkTodoHandler_Delete_ReturnsNoContent verifies that
// DELETE /todos/:id returns HTTP 204 and that a subsequent GET returns 404.
//
// Scenario:
//
//	Seed: one work todo.
//	DELETE /todos/<id>  + Bearer token → 204 No Content
//	GET    /todos/<id>  + Bearer token → 404 Not Found
func TestWorkTodoHandler_Delete_ReturnsNoContent(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-delete@example.com")

	// Create a work todo to delete
	createBody := strings.NewReader(`{"title":"삭제할 업무","type":"work"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create work todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — delete the work todo
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

// ---------------------------------------------------------------------------
// Toggle — POST /todos/:id/toggle
// ---------------------------------------------------------------------------

// TestWorkTodoHandler_Toggle_CompletesTodo verifies that
// POST /todos/:id/toggle on a pending work todo sets completed_at to a
// non-null value.
//
// Scenario:
//
//	Seed: one pending work todo (completed_at == null).
//	POST /todos/<id>/toggle  + Bearer token → 200 OK
//	GET  /todos/<id>         + Bearer token → completed_at is non-null
func TestWorkTodoHandler_Toggle_CompletesTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-toggle-complete@example.com")

	// Create a pending work todo
	createBody := strings.NewReader(`{"title":"스프린트 회의","type":"work"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create work todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
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

// TestWorkTodoHandler_Toggle_UncompletesTodo verifies that
// POST /todos/:id/toggle called twice on the same work todo returns
// completed_at back to null.
//
// Scenario:
//
//	Seed: one pending work todo.
//	POST /todos/<id>/toggle  (1st call) → completed_at is non-null
//	POST /todos/<id>/toggle  (2nd call) → completed_at is null again
func TestWorkTodoHandler_Toggle_UncompletesTodo(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-toggle-uncomplete@example.com")

	// Create a pending work todo
	createBody := strings.NewReader(`{"title":"PR 리뷰","type":"work"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create work todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
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
// Priority — POST /todos (type=work + priority)
// ---------------------------------------------------------------------------

// TestWorkTodoHandler_Create_WithPriorityHigh_ReturnsPriority verifies that
// POST /todos with type=work and priority=high returns HTTP 201 and a JSON
// body that includes priority="high".
//
// NOTE: Requires priority column in todos table (not yet implemented).
//
// Scenario:
//
//	POST /todos  {"title":"긴급 버그 수정","type":"work","priority":"high"} + Bearer token
//	→ 201 Created   {"id":1,"title":"긴급 버그 수정","type":"work","priority":"high","completed_at":null}
func TestWorkTodoHandler_Create_WithPriorityHigh_ReturnsPriority(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-priority-create@example.com")

	body := strings.NewReader(`{"title":"긴급 버그 수정","type":"work","priority":"high"}`)
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
	if resp["priority"] != "high" {
		t.Errorf("expected priority 'high', got %v", resp["priority"])
	}
	if resp["type"] != "work" {
		t.Errorf("expected type 'work', got %v", resp["type"])
	}
}

// TestWorkTodoHandler_Update_Priority_ReturnsUpdatedPriority verifies that
// PUT /todos/:id with a priority field updates the todo's priority and returns
// HTTP 200 with the updated JSON representation.
//
// NOTE: Requires priority column in todos table (not yet implemented).
//
// Scenario:
//
//	Seed: one work todo with no priority set.
//	PUT /todos/<id>  {"priority":"high"} + Bearer token
//	→ 200 OK  {"id":<id>,"priority":"high", ...}
func TestWorkTodoHandler_Update_Priority_ReturnsUpdatedPriority(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-priority-update@example.com")

	// Create a work todo without priority
	createBody := strings.NewReader(`{"title":"배포 작업","type":"work"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/todos", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create work todo failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — update priority to "high"
	updateBody := strings.NewReader(`{"priority":"high"}`)
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
	if updated["priority"] != "high" {
		t.Errorf("expected updated priority 'high', got %v", updated["priority"])
	}
}

// TestWorkTodoHandler_List_SortedByPriority_ReturnsHighMediumLow verifies that
// GET /todos?type=work returns work todos sorted by priority in descending
// order: high → medium → low.
//
// NOTE: Requires priority column in todos table and sort-by-priority logic
// in handler/todo.go (not yet implemented).
//
// Scenario:
//
//	Seed: three work todos with priorities low, high, medium (in that creation order).
//	GET /todos?type=work  + Bearer token
//	→ 200 OK  [{"priority":"high",...}, {"priority":"medium",...}, {"priority":"low",...}]
func TestWorkTodoHandler_List_SortedByPriority_ReturnsHighMediumLow(t *testing.T) {
	tdb := testutil.NewTestDB(t)
	e := setupWorkTodoEcho(t, tdb)
	token := workTodoToken(t, tdb, e, "work-todo-priority-sort@example.com")

	// Seed three work todos in low → high → medium creation order
	for _, item := range []struct{ title, priority string }{
		{"낮은 우선순위 업무", "low"},
		{"높은 우선순위 업무", "high"},
		{"중간 우선순위 업무", "medium"},
	} {
		payload := fmt.Sprintf(`{"title":"%s","type":"work","priority":"%s"}`, item.title, item.priority)
		createReq := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed work todo %q (priority=%s) failed: status %d, body: %s",
				item.title, item.priority, createRec.Code, createRec.Body.String())
		}
	}

	// Act
	req := httptest.NewRequest(http.MethodGet, "/todos?type=work", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &todos); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(todos) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(todos))
	}

	// Verify sort order: high → medium → low
	expectedOrder := []string{"high", "medium", "low"}
	for i, expected := range expectedOrder {
		got := todos[i]["priority"]
		if got != expected {
			t.Errorf("todos[%d]: expected priority %q, got %v (order should be high → medium → low)", i, expected, got)
		}
	}
}
