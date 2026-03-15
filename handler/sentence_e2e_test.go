// Package handler_test contains end-to-end style tests for the sentence
// collection API endpoints (POST/GET/PUT/DELETE /sentences/categories,
// POST/GET/PUT/DELETE /sentences).
//
// DLD-733: 10-1: 문장 모음 — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (t.Skip). Activate after DLD-734:
//   - handler.NewSentenceCategoryHandler is implemented (handler/sentence_category.go)
//   - handler.NewSentenceHandler is implemented (handler/sentence.go)
//   - db/migrations/000007_sentence_categories.up.sql migration is applied
//   - db/migrations/000008_sentences.up.sql migration is applied
//   - Routes /sentences/categories, /sentences/categories/:id are registered
//   - Routes /sentences, /sentences/:id are registered
//   - GET /sentences supports ?category_id=:id query parameter for filtering
//
// When activating:
//  1. Remove all t.Skip calls.
//  2. Uncomment the sentence handler lines inside setupSentenceEcho (marked TODO).
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

// setupSentenceEcho builds an Echo instance with the auth route and — once
// DLD-734 is implemented — the sentence category routes and sentence routes.
//
// Activation checklist:
//
//	TODO(DLD-734): uncomment the lines below after handler.NewSentenceCategoryHandler exists.
//	  scHandler := handler.NewSentenceCategoryHandler(tdb.DB)
//	  scg := e.Group("/sentences/categories", appmiddleware.JWT("test-jwt-secret"))
//	  scg.POST("",      scHandler.Create)
//	  scg.GET("",       scHandler.List)
//	  scg.PUT("/:id",   scHandler.Update)
//	  scg.DELETE("/:id", scHandler.Delete)
//
//	TODO(DLD-734): uncomment the lines below after handler.NewSentenceHandler exists.
//	  sHandler := handler.NewSentenceHandler(tdb.DB)
//	  sg := e.Group("/sentences", appmiddleware.JWT("test-jwt-secret"))
//	  sg.POST("",      sHandler.Create)
//	  sg.GET("",       sHandler.List)
//	  sg.PUT("/:id",   sHandler.Update)
//	  sg.DELETE("/:id", sHandler.Delete)
func setupSentenceEcho(t *testing.T, tdb *testutil.TestDB) *echo.Echo {
	t.Helper()
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)
	return e
}

// sentenceToken seeds a user with the given email and returns a JWT Bearer
// token string (including the "Bearer " prefix) by reusing seedUserAndLogin
// defined in chat_e2e_test.go (same package).
func sentenceToken(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email string) string {
	t.Helper()
	return "Bearer " + seedUserAndLogin(t, tdb, e, email, "Secret1!")
}

// ---------------------------------------------------------------------------
// Category CRUD — POST /sentences/categories
// ---------------------------------------------------------------------------

// TestSentenceCategoryHandler_Create_ReturnsCreated verifies that
// POST /sentences/categories with a valid name field returns HTTP 201 and
// a JSON body that includes id and name.
//
// Scenario:
//
//	POST /sentences/categories  {"name":"인사말"} + Bearer token
//	→ 201 Created  {"id":1,"name":"인사말"}
func TestSentenceCategoryHandler_Create_ReturnsCreated(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-create@example.com")

	body := strings.NewReader(`{"name":"인사말"}`)
	req := httptest.NewRequest(http.MethodPost, "/sentences/categories", body)
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
	if resp["name"] != "인사말" {
		t.Errorf("expected name '인사말', got %v", resp["name"])
	}
}

// TestSentenceCategoryHandler_Create_MissingName_ReturnsBadRequest verifies
// that POST /sentences/categories without a name field returns HTTP 400.
//
// Scenario:
//
//	POST /sentences/categories  {} + Bearer token
//	→ 400 Bad Request
func TestSentenceCategoryHandler_Create_MissingName_ReturnsBadRequest(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-noname@example.com")

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/sentences/categories", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestSentenceCategoryHandler_Create_WithoutAuth_ReturnsUnauthorized verifies
// that POST /sentences/categories without an Authorization header returns
// HTTP 401.
//
// Scenario:
//
//	POST /sentences/categories  {"name":"인사말"}  (no Authorization header)
//	→ 401 Unauthorized
func TestSentenceCategoryHandler_Create_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	e := echo.New()
	// Register a guarded placeholder so the JWT middleware can reject the
	// request before any real handler runs. This will be replaced by the
	// real sentence category handler once implemented.
	e.POST("/sentences/categories", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, nil)
	}, appmiddleware.JWT("test-jwt-secret"))

	body := strings.NewReader(`{"name":"인사말"}`)
	req := httptest.NewRequest(http.MethodPost, "/sentences/categories", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Category CRUD — GET /sentences/categories
// ---------------------------------------------------------------------------

// TestSentenceCategoryHandler_List_ReturnsList verifies that
// GET /sentences/categories returns HTTP 200 and a JSON array containing
// the authenticated user's categories.
//
// Scenario:
//
//	Seed: two categories created via POST /sentences/categories.
//	GET /sentences/categories  + Bearer token
//	→ 200 OK  JSON array with 2 elements, each having id/name
func TestSentenceCategoryHandler_List_ReturnsList(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-list@example.com")

	// Seed two categories
	for _, name := range []string{"인사말", "감사 표현"} {
		payload := fmt.Sprintf(`{"name":"%s"}`, name)
		createReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed category %q failed: status %d, body: %s", name, createRec.Code, createRec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/sentences/categories", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var categories []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &categories); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
}

// TestSentenceCategoryHandler_List_EmptyList_ReturnsEmptyArray verifies that
// GET /sentences/categories for a user with no categories returns HTTP 200
// and an empty JSON array (not null).
//
// Scenario:
//
//	Seed: authenticated user with no categories.
//	GET /sentences/categories  + Bearer token
//	→ 200 OK  []
func TestSentenceCategoryHandler_List_EmptyList_ReturnsEmptyArray(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-empty@example.com")

	req := httptest.NewRequest(http.MethodGet, "/sentences/categories", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var categories []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &categories); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(categories) != 0 {
		t.Errorf("expected empty array [], got %d items", len(categories))
	}
}

// ---------------------------------------------------------------------------
// Category CRUD — PUT /sentences/categories/:id
// ---------------------------------------------------------------------------

// TestSentenceCategoryHandler_Update_ReturnsOK verifies that
// PUT /sentences/categories/:id with a valid updated name returns HTTP 200
// and the updated JSON representation.
//
// Scenario:
//
//	Seed: one category with name "인사말".
//	PUT /sentences/categories/<id>  {"name":"감사 표현"} + Bearer token
//	→ 200 OK  {"id":<id>,"name":"감사 표현"}
func TestSentenceCategoryHandler_Update_ReturnsOK(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-update@example.com")

	// Create a category
	createBody := strings.NewReader(`{"name":"인사말"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create category failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	updateBody := strings.NewReader(`{"name":"감사 표현"}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/sentences/categories/%d", id), updateBody)
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
	if updated["name"] != "감사 표현" {
		t.Errorf("expected updated name '감사 표현', got %v", updated["name"])
	}
}

// TestSentenceCategoryHandler_Update_NotFound_ReturnsNotFound verifies that
// PUT /sentences/categories/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	PUT /sentences/categories/99999  {"name":"없는 카테고리"} + Bearer token
//	→ 404 Not Found
func TestSentenceCategoryHandler_Update_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-update-notfound@example.com")

	body := strings.NewReader(`{"name":"없는 카테고리"}`)
	req := httptest.NewRequest(http.MethodPut, "/sentences/categories/99999", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Category CRUD — DELETE /sentences/categories/:id
// ---------------------------------------------------------------------------

// TestSentenceCategoryHandler_Delete_ReturnsNoContent verifies that
// DELETE /sentences/categories/:id returns HTTP 204 and that a subsequent
// GET /sentences/categories no longer includes the deleted category.
//
// Scenario:
//
//	Seed: one category.
//	DELETE /sentences/categories/<id>  + Bearer token → 204 No Content
//	GET    /sentences/categories       + Bearer token → array does not contain the deleted id
func TestSentenceCategoryHandler_Delete_ReturnsNoContent(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-delete@example.com")

	// Create a category to delete
	createBody := strings.NewReader(`{"name":"삭제할 카테고리"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", createBody)
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create category failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — delete the category
	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/sentences/categories/%d", id), nil)
	delReq.Header.Set(echo.HeaderAuthorization, token)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d (body: %s)", delRec.Code, delRec.Body.String())
	}

	// Assert — subsequent GET must not contain the deleted category
	getReq := httptest.NewRequest(http.MethodGet, "/sentences/categories", nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on list after delete, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var categories []map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &categories); err != nil {
		t.Fatalf("expected valid JSON array after delete, got: %v (body: %s)", err, getRec.Body.String())
	}
	for _, c := range categories {
		if int(c["id"].(float64)) == id {
			t.Errorf("deleted category (id=%d) still appears in GET /sentences/categories", id)
		}
	}
}

// TestSentenceCategoryHandler_Delete_NotFound_ReturnsNotFound verifies that
// DELETE /sentences/categories/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	DELETE /sentences/categories/99999  + Bearer token
//	→ 404 Not Found
func TestSentenceCategoryHandler_Delete_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-cat-delete-notfound@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/sentences/categories/99999", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Sentence CRUD — POST /sentences
// ---------------------------------------------------------------------------

// TestSentenceHandler_Create_ReturnsCreated verifies that POST /sentences
// with valid content and category_id fields returns HTTP 201 and a JSON body
// that includes id, content, and category_id.
//
// Scenario:
//
//	Seed: one category with name "인사말".
//	POST /sentences  {"content":"안녕하세요","category_id":1} + Bearer token
//	→ 201 Created  {"id":1,"content":"안녕하세요","category_id":1}
func TestSentenceHandler_Create_ReturnsCreated(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-create@example.com")

	// Seed a category first
	catBody := strings.NewReader(`{"name":"인사말"}`)
	catReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", catBody)
	catReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	catReq.Header.Set(echo.HeaderAuthorization, token)
	catRec := httptest.NewRecorder()
	e.ServeHTTP(catRec, catReq)
	if catRec.Code != http.StatusCreated {
		t.Fatalf("setup: create category failed: status %d, body: %s", catRec.Code, catRec.Body.String())
	}

	var cat map[string]interface{}
	if err := json.Unmarshal(catRec.Body.Bytes(), &cat); err != nil {
		t.Fatalf("failed to unmarshal category response: %v", err)
	}
	categoryID := int(cat["id"].(float64))

	// Act
	payload := fmt.Sprintf(`{"content":"안녕하세요","category_id":%d}`, categoryID)
	body := strings.NewReader(payload)
	req := httptest.NewRequest(http.MethodPost, "/sentences", body)
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
	if resp["content"] != "안녕하세요" {
		t.Errorf("expected content '안녕하세요', got %v", resp["content"])
	}
	if int(resp["category_id"].(float64)) != categoryID {
		t.Errorf("expected category_id %d, got %v", categoryID, resp["category_id"])
	}
}

// TestSentenceHandler_Create_MissingContent_ReturnsBadRequest verifies that
// POST /sentences without a content field returns HTTP 400 Bad Request.
//
// Scenario:
//
//	POST /sentences  {"category_id":1} + Bearer token
//	→ 400 Bad Request
func TestSentenceHandler_Create_MissingContent_ReturnsBadRequest(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-nocontent@example.com")

	body := strings.NewReader(`{"category_id":1}`)
	req := httptest.NewRequest(http.MethodPost, "/sentences", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// TestSentenceHandler_Create_WithoutAuth_ReturnsUnauthorized verifies that
// POST /sentences without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /sentences  {"content":"안녕하세요","category_id":1}  (no Authorization header)
//	→ 401 Unauthorized
func TestSentenceHandler_Create_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	e := echo.New()
	// Register a guarded placeholder so the JWT middleware can reject the
	// request before any real handler runs. This will be replaced by the
	// real sentence handler once implemented.
	e.POST("/sentences", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, nil)
	}, appmiddleware.JWT("test-jwt-secret"))

	body := strings.NewReader(`{"content":"안녕하세요","category_id":1}`)
	req := httptest.NewRequest(http.MethodPost, "/sentences", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Sentence CRUD — GET /sentences
// ---------------------------------------------------------------------------

// TestSentenceHandler_List_ReturnsList verifies that GET /sentences returns
// HTTP 200 and a JSON array containing the authenticated user's sentences.
//
// Scenario:
//
//	Seed: two sentences created via POST /sentences.
//	GET /sentences  + Bearer token
//	→ 200 OK  JSON array with 2 elements, each having id/content/category_id
func TestSentenceHandler_List_ReturnsList(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-list@example.com")

	// Seed a category
	catBody := strings.NewReader(`{"name":"인사말"}`)
	catReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", catBody)
	catReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	catReq.Header.Set(echo.HeaderAuthorization, token)
	catRec := httptest.NewRecorder()
	e.ServeHTTP(catRec, catReq)
	if catRec.Code != http.StatusCreated {
		t.Fatalf("setup: create category failed: status %d, body: %s", catRec.Code, catRec.Body.String())
	}

	var cat map[string]interface{}
	if err := json.Unmarshal(catRec.Body.Bytes(), &cat); err != nil {
		t.Fatalf("failed to unmarshal category response: %v", err)
	}
	categoryID := int(cat["id"].(float64))

	// Seed two sentences
	for _, content := range []string{"안녕하세요", "반갑습니다"} {
		payload := fmt.Sprintf(`{"content":"%s","category_id":%d}`, content, categoryID)
		createReq := httptest.NewRequest(http.MethodPost, "/sentences", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed sentence %q failed: status %d, body: %s", content, createRec.Code, createRec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/sentences", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var sentences []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &sentences); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(sentences) != 2 {
		t.Errorf("expected 2 sentences, got %d", len(sentences))
	}
}

// TestSentenceHandler_List_ByCategoryID_ReturnsFiltered verifies that
// GET /sentences?category_id=:id returns HTTP 200 and only the sentences
// belonging to the specified category.
//
// Scenario:
//
//	Seed: two categories ("인사말", "감사 표현"), one sentence in each.
//	GET /sentences?category_id=<id of "인사말">  + Bearer token
//	→ 200 OK  JSON array with 1 element whose category_id matches the filter
func TestSentenceHandler_List_ByCategoryID_ReturnsFiltered(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-list-filter@example.com")

	// Seed two categories
	var categoryIDs []int
	for _, name := range []string{"인사말", "감사 표현"} {
		payload := fmt.Sprintf(`{"name":"%s"}`, name)
		catReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", strings.NewReader(payload))
		catReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		catReq.Header.Set(echo.HeaderAuthorization, token)
		catRec := httptest.NewRecorder()
		e.ServeHTTP(catRec, catReq)
		if catRec.Code != http.StatusCreated {
			t.Fatalf("seed category %q failed: status %d, body: %s", name, catRec.Code, catRec.Body.String())
		}
		var cat map[string]interface{}
		if err := json.Unmarshal(catRec.Body.Bytes(), &cat); err != nil {
			t.Fatalf("failed to unmarshal category response: %v", err)
		}
		categoryIDs = append(categoryIDs, int(cat["id"].(float64)))
	}

	// Seed one sentence per category
	sentences := []struct {
		content    string
		categoryID int
	}{
		{"안녕하세요", categoryIDs[0]},
		{"감사합니다", categoryIDs[1]},
	}
	for _, s := range sentences {
		payload := fmt.Sprintf(`{"content":"%s","category_id":%d}`, s.content, s.categoryID)
		createReq := httptest.NewRequest(http.MethodPost, "/sentences", strings.NewReader(payload))
		createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createReq.Header.Set(echo.HeaderAuthorization, token)
		createRec := httptest.NewRecorder()
		e.ServeHTTP(createRec, createReq)
		if createRec.Code != http.StatusCreated {
			t.Fatalf("seed sentence %q failed: status %d, body: %s", s.content, createRec.Code, createRec.Body.String())
		}
	}

	// Act — filter by the first category
	url := fmt.Sprintf("/sentences?category_id=%d", categoryIDs[0])
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var filtered []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(filtered) != 1 {
		t.Errorf("expected 1 sentence for category_id=%d, got %d", categoryIDs[0], len(filtered))
	}
	if len(filtered) > 0 {
		if int(filtered[0]["category_id"].(float64)) != categoryIDs[0] {
			t.Errorf("expected category_id %d, got %v", categoryIDs[0], filtered[0]["category_id"])
		}
		if filtered[0]["content"] != "안녕하세요" {
			t.Errorf("expected content '안녕하세요', got %v", filtered[0]["content"])
		}
	}
}

// ---------------------------------------------------------------------------
// Sentence CRUD — PUT /sentences/:id
// ---------------------------------------------------------------------------

// TestSentenceHandler_Update_ReturnsOK verifies that PUT /sentences/:id with
// a valid updated content returns HTTP 200 and the updated JSON representation.
//
// Scenario:
//
//	Seed: one sentence with content "안녕하세요".
//	PUT /sentences/<id>  {"content":"안녕히 가세요"} + Bearer token
//	→ 200 OK  {"id":<id>,"content":"안녕히 가세요","category_id":<id>}
func TestSentenceHandler_Update_ReturnsOK(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-update@example.com")

	// Seed a category
	catBody := strings.NewReader(`{"name":"인사말"}`)
	catReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", catBody)
	catReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	catReq.Header.Set(echo.HeaderAuthorization, token)
	catRec := httptest.NewRecorder()
	e.ServeHTTP(catRec, catReq)
	if catRec.Code != http.StatusCreated {
		t.Fatalf("setup: create category failed: status %d, body: %s", catRec.Code, catRec.Body.String())
	}

	var cat map[string]interface{}
	if err := json.Unmarshal(catRec.Body.Bytes(), &cat); err != nil {
		t.Fatalf("failed to unmarshal category response: %v", err)
	}
	categoryID := int(cat["id"].(float64))

	// Create a sentence
	createPayload := fmt.Sprintf(`{"content":"안녕하세요","category_id":%d}`, categoryID)
	createReq := httptest.NewRequest(http.MethodPost, "/sentences", strings.NewReader(createPayload))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create sentence failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act
	updateBody := strings.NewReader(`{"content":"안녕히 가세요"}`)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/sentences/%d", id), updateBody)
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
	if updated["content"] != "안녕히 가세요" {
		t.Errorf("expected updated content '안녕히 가세요', got %v", updated["content"])
	}
}

// TestSentenceHandler_Update_NotFound_ReturnsNotFound verifies that
// PUT /sentences/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	PUT /sentences/99999  {"content":"없는 문장"} + Bearer token
//	→ 404 Not Found
func TestSentenceHandler_Update_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-update-notfound@example.com")

	body := strings.NewReader(`{"content":"없는 문장"}`)
	req := httptest.NewRequest(http.MethodPut, "/sentences/99999", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Sentence CRUD — DELETE /sentences/:id
// ---------------------------------------------------------------------------

// TestSentenceHandler_Delete_ReturnsNoContent verifies that
// DELETE /sentences/:id returns HTTP 204 and that a subsequent GET /sentences
// no longer includes the deleted sentence.
//
// Scenario:
//
//	Seed: one sentence with content "안녕하세요".
//	DELETE /sentences/<id>  + Bearer token → 204 No Content
//	GET    /sentences       + Bearer token → array does not contain the deleted id
func TestSentenceHandler_Delete_ReturnsNoContent(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-delete@example.com")

	// Seed a category
	catBody := strings.NewReader(`{"name":"인사말"}`)
	catReq := httptest.NewRequest(http.MethodPost, "/sentences/categories", catBody)
	catReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	catReq.Header.Set(echo.HeaderAuthorization, token)
	catRec := httptest.NewRecorder()
	e.ServeHTTP(catRec, catReq)
	if catRec.Code != http.StatusCreated {
		t.Fatalf("setup: create category failed: status %d, body: %s", catRec.Code, catRec.Body.String())
	}

	var cat map[string]interface{}
	if err := json.Unmarshal(catRec.Body.Bytes(), &cat); err != nil {
		t.Fatalf("failed to unmarshal category response: %v", err)
	}
	categoryID := int(cat["id"].(float64))

	// Create a sentence to delete
	createPayload := fmt.Sprintf(`{"content":"삭제할 문장","category_id":%d}`, categoryID)
	createReq := httptest.NewRequest(http.MethodPost, "/sentences", strings.NewReader(createPayload))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createReq.Header.Set(echo.HeaderAuthorization, token)
	createRec := httptest.NewRecorder()
	e.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: create sentence failed: status %d, body: %s", createRec.Code, createRec.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}
	id := int(created["id"].(float64))

	// Act — delete the sentence
	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/sentences/%d", id), nil)
	delReq.Header.Set(echo.HeaderAuthorization, token)
	delRec := httptest.NewRecorder()
	e.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d (body: %s)", delRec.Code, delRec.Body.String())
	}

	// Assert — subsequent GET /sentences must not contain the deleted sentence
	getReq := httptest.NewRequest(http.MethodGet, "/sentences", nil)
	getReq.Header.Set(echo.HeaderAuthorization, token)
	getRec := httptest.NewRecorder()
	e.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 on list after delete, got %d (body: %s)", getRec.Code, getRec.Body.String())
	}
	var sentences []map[string]interface{}
	if err := json.Unmarshal(getRec.Body.Bytes(), &sentences); err != nil {
		t.Fatalf("expected valid JSON array after delete, got: %v (body: %s)", err, getRec.Body.String())
	}
	for _, s := range sentences {
		if int(s["id"].(float64)) == id {
			t.Errorf("deleted sentence (id=%d) still appears in GET /sentences", id)
		}
	}
}

// TestSentenceHandler_Delete_NotFound_ReturnsNotFound verifies that
// DELETE /sentences/:id for a non-existent ID returns HTTP 404.
//
// Scenario:
//
//	DELETE /sentences/99999  + Bearer token
//	→ 404 Not Found
func TestSentenceHandler_Delete_NotFound_ReturnsNotFound(t *testing.T) {
	t.Skip("DLD-734: sentence handler not yet implemented")

	tdb := testutil.NewTestDB(t)
	e := setupSentenceEcho(t, tdb)
	token := sentenceToken(t, tdb, e, "sentence-delete-notfound@example.com")

	req := httptest.NewRequest(http.MethodDelete, "/sentences/99999", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}
