// Package handler_test contains end-to-end style tests for the routine
// management API endpoints (POST/GET/PUT/DELETE /routines, POST /routines/:id/complete).
//
// DLD-723: 5-1: 루틴 관리 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (t.Skip). Activate after DLD-723:
//   - handler.NewRoutineHandler is implemented (handler/routine.go)
//   - db/migrations/000003_routines.up.sql migration is applied
//   - Routes /routines, /routines/:id, /routines/:id/complete are registered
//   - D-day calculation logic (next_due_date = last_done_at + interval_days) is in place
//   - POST /routines/:id/complete updates last_done_at and recalculates next_due_date
//
// When activating:
//  1. Remove all t.Skip calls.
//  2. Uncomment the routine handler lines inside setupRoutineEcho (marked TODO).
//  3. Ensure 000003_routines.up.sql exists so testutil.NewTestDB picks it up.
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

// setupRoutineEcho builds an Echo instance with the auth route and — once
// DLD-723 is implemented — the routine routes.
//
// Activation checklist:
//
//	TODO(DLD-723): uncomment the lines below after handler.NewRoutineHandler exists.
//	  routineHandler := handler.NewRoutineHandler(tdb.DB)
//	  g := e.Group("/routines", appmiddleware.JWT("test-jwt-secret"))
//	  g.POST("",              routineHandler.Create)
//	  g.GET("",               routineHandler.List)
//	  g.GET("/:id",           routineHandler.Get)
//	  g.PUT("/:id",           routineHandler.Update)
//	  g.DELETE("/:id",        routineHandler.Delete)
//	  g.POST("/:id/complete", routineHandler.Complete)
func setupRoutineEcho(t *testing.T, tdb *testutil.TestDB) *echo.Echo {
	t.Helper()
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)
	// TODO(DLD-723): register routine routes here once implemented.
	return e
}

// routineToken seeds a user with the given email and returns a JWT Bearer
// token string (including the "Bearer " prefix) by reusing seedUserAndLogin
// defined in chat_e2e_test.go (same package).
func routineToken(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email string) string {
	t.Helper()
	return "Bearer " + seedUserAndLogin(t, tdb, e, email, "Secret1!")
}

// ---------------------------------------------------------------------------
// CRUD — POST /routines
// ---------------------------------------------------------------------------

// TestRoutineHandler_Create_ReturnsCreated verifies that POST /routines with
// a valid name, interval_days, and last_done_at returns HTTP 201 and a JSON
// body that includes id, name, interval_days, last_done_at, next_due_date,
// and d_day.
//
// Scenario:
//
//	POST /routines  {"name":"샤워","interval_days":1,"last_done_at":"2026-03-13"} + Bearer token
//	→ 201 Created   {"id":1,"name":"샤워","interval_days":1,"last_done_at":"2026-03-13",
//	                 "next_due_date":"2026-03-14","d_day":0}
func TestRoutineHandler_Create_ReturnsCreated(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-create@example.com")

	body := strings.NewReader(`{"name":"샤워","interval_days":1,"last_done_at":"2026-03-13"}`)
	req := httptest.NewRequest(http.MethodPost, "/routines", body)
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
	if resp["name"] != "샤워" {
		t.Errorf("expected name '샤워', got %v", resp["name"])
	}
	if resp["interval_days"] != float64(1) {
		t.Errorf("expected interval_days 1, got %v", resp["interval_days"])
	}
	if resp["last_done_at"] != "2026-03-13" {
		t.Errorf("expected last_done_at '2026-03-13', got %v", resp["last_done_at"])
	}
	if resp["next_due_date"] != "2026-03-14" {
		t.Errorf("expected next_due_date '2026-03-14', got %v", resp["next_due_date"])
	}
}

// TestRoutineHandler_Create_MissingName_ReturnsBadRequest verifies that
// POST /routines without a name field returns HTTP 400 Bad Request.
//
// Scenario:
//
//	POST /routines  {"interval_days":7} + Bearer token
//	→ 400 Bad Request
func TestRoutineHandler_Create_MissingName_ReturnsBadRequest(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-noname@example.com")

	body := strings.NewReader(`{"interval_days":7}`)
	req := httptest.NewRequest(http.MethodPost, "/routines", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestRoutineHandler_Create_WithoutAuth_ReturnsUnauthorized verifies that
// POST /routines without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /routines  {"name":"운동","interval_days":2}  (no Authorization header)
//	→ 401 Unauthorized
func TestRoutineHandler_Create_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	e := echo.New()
	// Register a guarded placeholder so the JWT middleware can reject the request
	// before any real handler runs.  This will be replaced by the real routine
	// handler once implemented.
	e.POST("/routines", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, nil)
	}, appmiddleware.JWT("test-jwt-secret"))

	body := strings.NewReader(`{"name":"운동","interval_days":2}`)
	req := httptest.NewRequest(http.MethodPost, "/routines", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — GET /routines
// ---------------------------------------------------------------------------

// TestRoutineHandler_List_ReturnsRoutines verifies that GET /routines returns
// HTTP 200 and a JSON array containing the authenticated user's routines, each
// with a d_day field.
//
// Scenario:
//
//	Seed: two routines created via POST /routines.
//	GET /routines  + Bearer token
//	→ 200 OK  JSON array with 2 elements; each element includes d_day
func TestRoutineHandler_List_ReturnsRoutines(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-list@example.com")

	// Seed two routines
	for _, name := range []string{"양치", "세수"} {
		payload := fmt.Sprintf(`{"name":"%s","interval_days":1,"last_done_at":"2026-03-13"}`, name)
		createReq := httptest.NewRequest(http.MethodPost, "/routines", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed routine %q failed: status %d, body: %s", name, createRec.Code, createRec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/routines", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var routines []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &routines); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(routines) != 2 {
		t.Errorf("expected 2 routines, got %d", len(routines))
	}
	for i, r := range routines {
		if _, ok := r["d_day"]; !ok {
			t.Errorf("routine[%d] missing 'd_day' field", i)
		}
	}
}

// TestRoutineHandler_List_EmptyList_ReturnsEmptyArray verifies that
// GET /routines for a user with no routines returns HTTP 200 and an empty
// JSON array (not null).
//
// Scenario:
//
//	Seed: authenticated user with no routines.
//	GET /routines  + Bearer token
//	→ 200 OK  []
func TestRoutineHandler_List_EmptyList_ReturnsEmptyArray(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-empty@example.com")

	req := httptest.NewRequest(http.MethodGet, "/routines", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var routines []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &routines); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(routines) != 0 {
		t.Errorf("expected empty array [], got %d items", len(routines))
	}
}

// TestRoutineHandler_List_WithoutAuth_ReturnsUnauthorized verifies that
// GET /routines without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	GET /routines  (no Authorization header)
//	→ 401 Unauthorized
func TestRoutineHandler_List_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	e := echo.New()
	e.GET("/routines", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []interface{}{})
	}, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodGet, "/routines", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// CRUD — GET /routines/:id
// ---------------------------------------------------------------------------

// TestRoutineHandler_Get_ReturnsSingleRoutine verifies that
// GET /routines/:id returns HTTP 200 and the JSON representation of the
// requested routine.
//
// Scenario:
//
//	Seed: one routine created via POST /routines.
//	GET /routines/<id>  + Bearer token
//	→ 200 OK  {"id":<id>,"name":"스트레칭","interval_days":3, ...}
func TestRoutineHandler_Get_ReturnsSingleRoutine(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-get@example.com")

	// Create a routine to retrieve
	createBody := strings.NewReader(`{"name":"스트레칭","interval_days":3,"last_done_at":"2026-03-11"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/routines", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create routine failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/routines/%d", id), nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var routine map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &routine); err != nil {
		t.Fatalf("expected valid JSON object, got: %v (body: %s)", err, rec.Body.String())
	}
	if routine["name"] != "스트레칭" {
		t.Errorf("expected name '스트레칭', got %v", routine["name"])
	}
	if routine["interval_days"] != float64(3) {
		t.Errorf("expected interval_days 3, got %v", routine["interval_days"])
	}
}

// TestRoutineHandler_Get_NotFound_ReturnsNotFound verifies that
// GET /routines/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	GET /routines/99999  + Bearer token
//	→ 404 Not Found
func TestRoutineHandler_Get_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-get-notfound@example.com")

	req := httptest.NewRequest(http.MethodGet, "/routines/99999", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — PUT /routines/:id
// ---------------------------------------------------------------------------

// TestRoutineHandler_Update_ReturnsOK verifies that PUT /routines/:id with
// valid updated fields returns HTTP 200 and the updated JSON representation.
//
// Scenario:
//
//	Seed: one routine with name "스트레칭", interval_days=3.
//	PUT /routines/<id>  {"name":"스트레칭 (강화)","interval_days":2} + Bearer token
//	→ 200 OK  {"id":<id>,"name":"스트레칭 (강화)","interval_days":2, ...}
func TestRoutineHandler_Update_ReturnsOK(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-update@example.com")

	// Create a routine
	createBody := strings.NewReader(`{"name":"스트레칭","interval_days":3,"last_done_at":"2026-03-11"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/routines", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create routine failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	updateBody := strings.NewReader(`{"name":"스트레칭 (강화)","interval_days":2}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/routines/%d", id), updateBody)
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
	if updated["name"] != "스트레칭 (강화)" {
		t.Errorf("expected updated name '스트레칭 (강화)', got %v", updated["name"])
	}
	if updated["interval_days"] != float64(2) {
		t.Errorf("expected updated interval_days 2, got %v", updated["interval_days"])
	}
}

// TestRoutineHandler_Update_NotFound_ReturnsNotFound verifies that
// PUT /routines/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	PUT /routines/99999  {"name":"없는 루틴","interval_days":1} + Bearer token
//	→ 404 Not Found
func TestRoutineHandler_Update_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-update-notfound@example.com")

	body := strings.NewReader(`{"name":"없는 루틴","interval_days":1}`)
	req := httptest.NewRequest(http.MethodPut, "/routines/99999", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// CRUD — DELETE /routines/:id
// ---------------------------------------------------------------------------

// TestRoutineHandler_Delete_ReturnsNoContent verifies that
// DELETE /routines/:id returns HTTP 204 and that a subsequent GET returns 404.
//
// Scenario:
//
//	Seed: one routine.
//	DELETE /routines/<id>  + Bearer token → 204 No Content
//	GET    /routines/<id>  + Bearer token → 404 Not Found
func TestRoutineHandler_Delete_ReturnsNoContent(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-delete@example.com")

	// Create a routine to delete
	createBody := strings.NewReader(`{"name":"삭제할 루틴","interval_days":7,"last_done_at":"2026-03-07"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/routines", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create routine failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — delete the routine
	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/routines/%d", id), nil)
	delReq.Header.Set(echo.HeaderAuthorization, token)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d (body: %s)", delRec.Code, delRec.Body.String())
	}

	// Assert — subsequent GET must return 404
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/routines/%d", id), nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 after delete, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
}

// TestRoutineHandler_Delete_NotFound_ReturnsNotFound verifies that
// DELETE /routines/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	DELETE /routines/99999  + Bearer token
//	→ 404 Not Found
func TestRoutineHandler_Delete_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-delete-notfound@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/routines/99999", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// D-day Calculation
// ---------------------------------------------------------------------------

// TestRoutineHandler_Create_DDayCalculation_NextDueDateIsCorrect verifies
// that when a routine is created, next_due_date equals last_done_at + interval_days.
//
// Scenario:
//
//	POST /routines  {"name":"세탁","interval_days":7,"last_done_at":"2026-03-07"} + Bearer token
//	→ 201 Created   next_due_date == "2026-03-14"
func TestRoutineHandler_Create_DDayCalculation_NextDueDateIsCorrect(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-dday-create@example.com")

	body := strings.NewReader(`{"name":"세탁","interval_days":7,"last_done_at":"2026-03-07"}`)
	req := httptest.NewRequest(http.MethodPost, "/routines", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v (body: %s)", err, rec.Body.String())
	}
	// next_due_date = "2026-03-07" + 7 days = "2026-03-14"
	if resp["next_due_date"] != "2026-03-14" {
		t.Errorf("expected next_due_date '2026-03-14' (last_done_at + interval_days), got %v", resp["next_due_date"])
	}
}

// TestRoutineHandler_List_DDayCalculation_IncludesDaysRemaining verifies that
// GET /routines returns each routine with a d_day field calculated as
// next_due_date - today (negative when overdue).
//
// Scenario (today = 2026-03-14):
//
//	POST /routines  {"name":"청소","interval_days":7,"last_done_at":"2026-03-10"}
//	  → next_due_date = "2026-03-17"
//	GET /routines  + Bearer token
//	  → routines[0].d_day == 3   (2026-03-17 minus 2026-03-14)
func TestRoutineHandler_List_DDayCalculation_IncludesDaysRemaining(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-dday-list@example.com")

	// Create a routine whose next_due_date will be 3 days ahead of today
	createBody := strings.NewReader(`{"name":"청소","interval_days":7,"last_done_at":"2026-03-10"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/routines", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create routine failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/routines", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var routines []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &routines); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(routines) == 0 {
		t.Fatal("expected at least 1 routine in list")
	}
	dDay, ok := routines[0]["d_day"]
	if !ok {
		t.Fatal("expected 'd_day' field in routine list item")
	}
	// today = 2026-03-14, next_due_date = 2026-03-17 → d_day = 3
	if dDay != float64(3) {
		t.Errorf("expected d_day 3 (next_due_date 2026-03-17 minus today 2026-03-14), got %v", dDay)
	}
}

// ---------------------------------------------------------------------------
// Completion — POST /routines/:id/complete
// ---------------------------------------------------------------------------

// TestRoutineHandler_Complete_UpdatesLastDoneAndNextDue verifies that
// POST /routines/:id/complete updates last_done_at to today and recalculates
// next_due_date as today + interval_days.
//
// Scenario (today = 2026-03-14):
//
//	Seed: routine with interval_days=7, last_done_at="2026-03-01"
//	POST /routines/<id>/complete  + Bearer token  → 200 OK
//	GET  /routines/<id>           + Bearer token
//	  → last_done_at  == "2026-03-14"
//	  → next_due_date == "2026-03-21"
//	  → d_day         == 7
func TestRoutineHandler_Complete_UpdatesLastDoneAndNextDue(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-complete@example.com")

	// Create a routine with an old last_done_at
	createBody := strings.NewReader(`{"name":"독서","interval_days":7,"last_done_at":"2026-03-01"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/routines", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create routine failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — complete the routine (stamps today as last_done_at)
	completeReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/routines/%d/complete", id), nil)
	completeReq.Header.Set(echo.HeaderAuthorization, token)
	completeRec := httptest.NewRecorder()
	e.ServeHTTP(completeRec, completeReq)

	if completeRec.Code != http.StatusOK {
		t.Errorf("expected status 200 on complete, got %d (body: %s)", completeRec.Code, completeRec.Body.String())
	}

	// Assert — retrieve the routine and verify updated fields
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/routines/%d", id), nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on GET after complete, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var routine map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &routine); err != nil {
		t.Fatalf("expected valid JSON object, got: %v (body: %s)", err, getRec.Body.String())
	}
	// last_done_at must be updated to today (2026-03-14)
	if routine["last_done_at"] != "2026-03-14" {
		t.Errorf("expected last_done_at '2026-03-14' after complete, got %v", routine["last_done_at"])
	}
	// next_due_date = today(2026-03-14) + interval(7) = "2026-03-21"
	if routine["next_due_date"] != "2026-03-21" {
		t.Errorf("expected next_due_date '2026-03-21' after complete, got %v", routine["next_due_date"])
	}
	// d_day = 7
	if routine["d_day"] != float64(7) {
		t.Errorf("expected d_day 7 after complete, got %v", routine["d_day"])
	}
}

// TestRoutineHandler_Complete_NotFound_ReturnsNotFound verifies that
// POST /routines/:id/complete for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	POST /routines/99999/complete  + Bearer token
//	→ 404 Not Found
func TestRoutineHandler_Complete_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-complete-notfound@example.com")

	req := httptest.NewRequest(http.MethodPost, "/routines/99999/complete", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestRoutineHandler_Complete_WithoutAuth_ReturnsUnauthorized verifies that
// POST /routines/:id/complete without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /routines/1/complete  (no Authorization header)
//	→ 401 Unauthorized
func TestRoutineHandler_Complete_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	e := echo.New()
	// Register a guarded placeholder so the JWT middleware rejects the request.
	e.POST("/routines/:id/complete", func(c echo.Context) error {
		return c.JSON(http.StatusOK, nil)
	}, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodPost, "/routines/1/complete", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Interval Update — next_due_date recalculation
// ---------------------------------------------------------------------------

// TestRoutineHandler_Update_IntervalChange_RecalculatesNextDue verifies that
// when interval_days is changed via PUT /routines/:id, next_due_date is
// recalculated as last_done_at + new_interval_days.
//
// Scenario:
//
//	Seed: routine with interval_days=7, last_done_at="2026-03-07"
//	  → initial next_due_date = "2026-03-14"
//	PUT /routines/<id>  {"interval_days":3}  + Bearer token
//	  → next_due_date recalculated: "2026-03-07" + 3 = "2026-03-10"
func TestRoutineHandler_Update_IntervalChange_RecalculatesNextDue(t *testing.T) {
	t.Skip("DLD-723: routine handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupRoutineEcho(t, tdb)
	token := routineToken(t, tdb, e, "routine-interval@example.com")

	// Create a routine
	createBody := strings.NewReader(`{"name":"조깅","interval_days":7,"last_done_at":"2026-03-07"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/routines", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create routine failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Verify precondition: initial next_due_date is "2026-03-14"
	if created["next_due_date"] != "2026-03-14" {
		t.Fatalf("precondition: expected initial next_due_date '2026-03-14', got %v", created["next_due_date"])
	}

	// Act — change interval from 7 to 3 days
	updateBody := strings.NewReader(`{"interval_days":3}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/routines/%d", id), updateBody)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var updated map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("expected valid JSON object, got: %v (body: %s)", err, rec.Body.String())
	}
	// next_due_date = last_done_at("2026-03-07") + new interval(3) = "2026-03-10"
	if updated["next_due_date"] != "2026-03-10" {
		t.Errorf("expected next_due_date '2026-03-10' after interval change to 3, got %v", updated["next_due_date"])
	}
}
