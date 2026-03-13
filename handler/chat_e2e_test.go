// Package handler_test contains end-to-end style tests for the AI chat
// API endpoints (POST /chat/send, GET /chat/history).
//
// DLD-719: 3-1: AI 채팅 (텍스트) — e2e 테스트 작성 (skipped)
//
// NOTE: All tests in this file are skipped with t.Skip() because the
// ChatHandler and its supporting infrastructure (chat repository, SSE
// streaming, model-selection logic) have not yet been implemented.
// Remove the t.Skip() call in each test once the implementation is ready.
//
// When activating these tests, add the following imports:
//
//	"bufio"
//	"encoding/json"
//	"net/http"
//	"net/http/httptest"
//	"strings"
//	"golang.org/x/crypto/bcrypt"
//	"github.com/labstack/echo/v4"
//	"github.com/dlddu/pocket-aide/handler"
//	appmiddleware "github.com/dlddu/pocket-aide/middleware"
//	"github.com/dlddu/pocket-aide/service/llm"
//	"github.com/dlddu/pocket-aide/testutil"
//
// Also implement the seedUserAndLogin helper:
//
//	func seedUserAndLogin(t *testing.T, tdb *testutil.TestDB, e *echo.Echo, email, password string) string
//	  → Creates a user via tdb.Seed, calls POST /auth/login, extracts and returns the JWT token.
package handler_test

import (
	"testing"
)

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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
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
	t.Skip("chat handler not yet implemented — activate after DLD-719")
}
