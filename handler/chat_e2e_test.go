// Package handler_test contains end-to-end style tests for the AI chat
// API endpoints (POST /chat/send, GET /chat/history).
//
// DLD-719: 3-1: AI 채팅 (텍스트) — e2e 테스트 작성
//
// NOTE: Tests activated after DLD-719. ChatHandler and its supporting
// infrastructure (chat repository, SSE streaming, model-selection logic)
// are now implemented.
package handler_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/handler"
	appmiddleware "github.com/dlddu/pocket-aide/middleware"
	"github.com/dlddu/pocket-aide/service/llm"
	"github.com/dlddu/pocket-aide/testutil"
)

// seedUserAndLogin creates a user in the database via tdb.Seed, calls POST /auth/login,
// and returns the JWT token from the response.
func seedUserAndLogin(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	tdb.Seed(t,
		"INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email, string(hash),
	)

	loginBody := strings.NewReader(`{"email":"` + email + `","password":"` + password + `"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", loginBody)
	loginReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	loginRec := httptest.NewRecorder()
	e.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("login failed: status %d, body: %s", loginRec.Code, loginRec.Body.String())
	}

	var loginResp map[string]interface{}
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to unmarshal login response: %v", err)
	}
	token, ok := loginResp["token"].(string)
	if !ok || token == "" {
		t.Fatal("expected non-empty 'token' field in login response")
	}
	return token
}

// ---------------------------------------------------------------------------
// Happy Path — POST /chat/send
// ---------------------------------------------------------------------------

// TestChatHandler_Send_ReturnsSSEStream verifies that POST /chat/send with a
// valid prompt and authentication returns HTTP 200 with Content-Type
// text/event-stream and at least one SSE data line.
//
// Scenario:
//
//	Seed: authenticated user.
//	POST /chat/send  {"message":"Hello"} + Bearer token
//	→ 200 OK  Content-Type: text/event-stream, body contains "data:" line
//
// Setup:
//
//	tdb := testutil.NewTestDB(t)
//	mockLLM := &llm.MockProvider{CompleteFunc: returns "Hello from AI"}
//	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM)
//	token := seedUserAndLogin(t, tdb, e, "chat@example.com", "Secret1!")
//	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))
func TestChatHandler_Send_ReturnsSSEStream(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "Hello from AI", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM)
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "chat@example.com", "Secret1!")

	// Act
	body := strings.NewReader(`{"message":"Hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/send", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get(echo.HeaderContentType)
	if !strings.Contains(ct, "text/event-stream") {
		t.Errorf("expected Content-Type text/event-stream, got %q", ct)
	}
	if !strings.Contains(rec.Body.String(), "data:") {
		t.Errorf("expected SSE body to contain 'data:' line, got: %s", rec.Body.String())
	}
}

// TestChatHandler_Send_StreamsMultipleTokens verifies that the SSE stream
// emits multiple token events when the LLM produces a multi-word response.
//
// Scenario:
//
//	MockProvider returns "one two three" (three tokens).
//	POST /chat/send  {"message":"count"} + Bearer token
//	→ 200 OK  SSE stream with ≥ 3 "data:" lines (verified via bufio.Scanner)
func TestChatHandler_Send_StreamsMultipleTokens(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "one two three", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM)
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "multi@example.com", "Secret1!")

	// Act
	body := strings.NewReader(`{"message":"count"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/send", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	dataLineCount := 0
	scanner := bufio.NewScanner(strings.NewReader(rec.Body.String()))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			dataLineCount++
		}
	}
	if dataLineCount < 3 {
		t.Errorf("expected >= 3 'data:' lines in SSE stream, got %d (body: %s)", dataLineCount, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Error Case — POST /chat/send (unauthenticated)
// ---------------------------------------------------------------------------

// TestChatHandler_Send_WithoutAuth_ReturnsUnauthorized verifies that
// POST /chat/send without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	POST /chat/send  {"message":"Hello"}  (no Authorization header)
//	→ 401 Unauthorized
func TestChatHandler_Send_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	e := echo.New()
	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "response", nil
	}}
	chatHandler := handler.NewChatHandler(nil, mockLLM)
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	// Act
	body := strings.NewReader(`{"message":"Hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/send", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Happy Path — GET /chat/history
// ---------------------------------------------------------------------------

// TestChatHandler_History_ReturnsConversation verifies that GET /chat/history
// returns HTTP 200 and a JSON array containing previously sent messages and
// their corresponding AI responses.
//
// Scenario:
//
//	Seed: authenticated user + one chat exchange via POST /chat/send.
//	GET /chat/history  Authorization: Bearer <token>
//	→ 200 OK  [{"role":"user","content":"Hi"},{"role":"assistant","content":"Hello from AI"}]
//	Assert len(history) >= 2
func TestChatHandler_History_ReturnsConversation(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "Hello from AI", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM)
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))
	e.GET("/chat/history", chatHandler.History, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "history@example.com", "Secret1!")

	// Send a message first to create conversation history
	sendBody := strings.NewReader(`{"message":"Hi"}`)
	sendReq := httptest.NewRequest(http.MethodPost, "/chat/send", sendBody)
	sendReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	sendReq.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	sendRec := httptest.NewRecorder()
	e.ServeHTTP(sendRec, sendReq)
	if sendRec.Code != http.StatusOK {
		t.Fatalf("send message failed: status %d, body: %s", sendRec.Code, sendRec.Body.String())
	}

	// Act
	req := httptest.NewRequest(http.MethodGet, "/chat/history", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var history []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &history); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(history) < 2 {
		t.Errorf("expected >= 2 messages in history, got %d", len(history))
	}
}

// TestChatHandler_History_EmptyHistory_ReturnsEmptyList verifies that
// GET /chat/history for a user who has not sent any messages returns
// HTTP 200 with an empty JSON array.
//
// Scenario:
//
//	Seed: authenticated user with no chat messages.
//	GET /chat/history  Authorization: Bearer <token>
//	→ 200 OK  []
func TestChatHandler_History_EmptyHistory_ReturnsEmptyList(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "response", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM)
	e.GET("/chat/history", chatHandler.History, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "empty@example.com", "Secret1!")

	// Act
	req := httptest.NewRequest(http.MethodGet, "/chat/history", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	var history []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &history); err != nil {
		t.Fatalf("expected valid JSON array, got: %v (body: %s)", err, rec.Body.String())
	}
	if len(history) != 0 {
		t.Errorf("expected empty history [], got %d items", len(history))
	}
}

// ---------------------------------------------------------------------------
// Error Case — GET /chat/history (unauthenticated)
// ---------------------------------------------------------------------------

// TestChatHandler_History_WithoutAuth_ReturnsUnauthorized verifies that
// GET /chat/history without an Authorization header returns HTTP 401.
//
// Scenario:
//
//	GET /chat/history  (no Authorization header)
//	→ 401 Unauthorized
func TestChatHandler_History_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	// Arrange
	e := echo.New()
	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "response", nil
	}}
	chatHandler := handler.NewChatHandler(nil, mockLLM)
	e.GET("/chat/history", chatHandler.History, appmiddleware.JWT("test-jwt-secret"))

	// Act
	req := httptest.NewRequest(http.MethodGet, "/chat/history", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Model Selection — POST /chat/send
// ---------------------------------------------------------------------------

// TestChatHandler_Send_WithModelSelection_UsesSpecifiedModel verifies that
// when a "model" field is included in the request body, the handler routes
// the prompt to the specified model and the response reflects that choice.
//
// Scenario:
//
//	POST /chat/send  {"message":"Hello","model":"gpt-4o"} + Bearer token
//	→ 200 OK  SSE stream; MockProvider.CallCount == 1; body contains "gpt-4o response"
func TestChatHandler_Send_WithModelSelection_UsesSpecifiedModel(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "gpt-4o response", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM)
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "model@example.com", "Secret1!")

	// Act
	// model:"gpt-4o" is unknown → Router falls back to default "mock" → MockProvider called
	body := strings.NewReader(`{"message":"Hello","model":"gpt-4o"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/send", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if mockLLM.CallCount != 1 {
		t.Errorf("expected MockProvider.CallCount == 1, got %d", mockLLM.CallCount)
	}
	// SSE streaming splits the response by whitespace into tokens.
	// "gpt-4o response" becomes two separate data: lines.
	// Verify both tokens appear in the SSE body.
	sseBody := rec.Body.String()
	if !strings.Contains(sseBody, "data: gpt-4o") {
		t.Errorf("expected SSE body to contain 'data: gpt-4o' token, got: %s", sseBody)
	}
	if !strings.Contains(sseBody, "response") {
		t.Errorf("expected SSE body to contain 'response' token, got: %s", sseBody)
	}
}

// TestChatHandler_Send_DefaultModel_UsesFallback verifies that when no
// "model" field is provided, the handler uses the configured default/fallback
// model and still returns a valid SSE stream.
//
// Scenario:
//
//	POST /chat/send  {"message":"Hello"}  (no "model" field) + Bearer token
//	→ 200 OK  SSE stream using default model; MockProvider.CallCount == 1; body contains "data:"
func TestChatHandler_Send_DefaultModel_UsesFallback(t *testing.T) {
	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "default model response", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM)
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "fallback@example.com", "Secret1!")

	// Act
	body := strings.NewReader(`{"message":"Hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/send", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if mockLLM.CallCount != 1 {
		t.Errorf("expected MockProvider.CallCount == 1, got %d", mockLLM.CallCount)
	}
	if !strings.Contains(rec.Body.String(), "data:") {
		t.Errorf("expected SSE body to contain 'data:', got: %s", rec.Body.String())
	}
}
