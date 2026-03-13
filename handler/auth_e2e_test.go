// Package handler_test contains end-to-end style tests for the authentication
// API endpoints (POST /auth/register, POST /auth/login, GET /me).
//
// DLD-717: 2-1: 사용자 인증 — e2e 테스트
package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/handler"
	appmiddleware "github.com/dlddu/pocket-aide/middleware"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// Happy Path
// ---------------------------------------------------------------------------

// TestAuthHandler_Register_ReturnsCreated verifies that a valid registration
// request returns HTTP 201 and a JSON body containing the new user's ID and
// email.
//
// Scenario:
//
//	POST /auth/register  {"email":"new@example.com","password":"Secret1!"}
//	→ 201 Created        {"id":1,"email":"new@example.com"}
func TestAuthHandler_Register_ReturnsCreated(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/register", authHandler.Register)

	body := strings.NewReader(`{"email":"new@example.com","password":"Secret1!"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set(echo.MIMEApplicationJSON, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v", err)
	}
	if _, ok := resp["id"]; !ok {
		t.Error("expected response to contain 'id' field")
	}
	if resp["email"] != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got %v", resp["email"])
	}
}

// TestAuthHandler_Register_DuplicateEmail_ReturnsConflict verifies that
// attempting to register with an already-used email returns HTTP 409.
//
// Scenario:
//
//	Seed: user with email "exists@example.com" already in the database.
//	POST /auth/register  {"email":"exists@example.com","password":"Secret1!"}
//	→ 409 Conflict
func TestAuthHandler_Register_DuplicateEmail_ReturnsConflict(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"exists@example.com", "$2a$10$hashedpassword",
	)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/register", authHandler.Register)

	body := strings.NewReader(`{"email":"exists@example.com","password":"Secret1!"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set(echo.MIMEApplicationJSON, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rec.Code)
	}
}

// TestAuthHandler_Login_ReturnsOKWithToken verifies that valid credentials
// produce HTTP 200 and a response body containing a non-empty JWT token.
//
// Scenario:
//
//	Seed: registered user with email "user@example.com".
//	POST /auth/login  {"email":"user@example.com","password":"Secret1!"}
//	→ 200 OK          {"token":"<jwt>"}
func TestAuthHandler_Login_ReturnsOKWithToken(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("Secret1!"), bcrypt.DefaultCost)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"user@example.com", string(hash),
	)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	body := strings.NewReader(`{"email":"user@example.com","password":"Secret1!"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set(echo.MIMEApplicationJSON, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v", err)
	}
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Error("expected non-empty 'token' field in response")
	}
}

// TestAuthHandler_Login_WrongPassword_ReturnsUnauthorized verifies that an
// incorrect password produces HTTP 401.
//
// Scenario:
//
//	Seed: registered user with email "user@example.com".
//	POST /auth/login  {"email":"user@example.com","password":"wrongpass"}
//	→ 401 Unauthorized
func TestAuthHandler_Login_WrongPassword_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("Secret1!"), bcrypt.DefaultCost)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"user@example.com", string(hash),
	)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	body := strings.NewReader(`{"email":"user@example.com","password":"wrongpass"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set(echo.MIMEApplicationJSON, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Protected Endpoint (GET /me)
// ---------------------------------------------------------------------------

// TestAuthHandler_Me_WithValidToken_ReturnsOK verifies that a request to
// GET /me with a valid Bearer token returns HTTP 200 and the current user's
// information.
//
// Scenario:
//
//	Seed: registered user.
//	Obtain token via POST /auth/login.
//	GET /me  Authorization: Bearer <token>
//	→ 200 OK  {"id":1,"email":"user@example.com"}
func TestAuthHandler_Me_WithValidToken_ReturnsOK(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("Secret1!"), bcrypt.DefaultCost)
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		"user@example.com", string(hash),
	)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)
	e.GET("/me", authHandler.Me, appmiddleware.JWT("test-jwt-secret"))

	// Obtain token
	loginBody := strings.NewReader(`{"email":"user@example.com","password":"Secret1!"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", loginBody)
	loginReq.Header.Set(echo.MIMEApplicationJSON, echo.MIMEApplicationJSON)
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginReq)
	var loginResp map[string]interface{}
	json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got: %v", err)
	}
	if resp["email"] != "user@example.com" {
		t.Errorf("expected email 'user@example.com', got %v", resp["email"])
	}
}

// TestAuthHandler_Me_WithoutToken_ReturnsUnauthorized verifies that a request
// to GET /me without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	GET /me  (no Authorization header)
//	→ 401 Unauthorized
func TestAuthHandler_Me_WithoutToken_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	e := echo.New()
	authHandler := handler.NewAuthHandler(nil, "test-jwt-secret")
	e.GET("/me", authHandler.Me, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// TestAuthHandler_Me_WithInvalidToken_ReturnsUnauthorized verifies that a
// request to GET /me with a malformed or tampered token returns HTTP 401.
//
// Scenario:
//
//	GET /me  Authorization: Bearer invalid.token.value
//	→ 401 Unauthorized
func TestAuthHandler_Me_WithInvalidToken_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	e := echo.New()
	authHandler := handler.NewAuthHandler(nil, "test-jwt-secret")
	e.GET("/me", authHandler.Me, appmiddleware.JWT("test-jwt-secret"))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer invalid.token.value")
	rec := httptest.NewRecorder()

	// Act
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}
