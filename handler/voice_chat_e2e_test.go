// Package handler_test contains end-to-end style tests for the AI voice chat
// flow: speech recognition result → POST /chat/send → SSE stream response,
// and voice + text mixed history scenarios.
//
// DLD-721: 4-1: AI 채팅 (음성) — e2e 테스트 작성 (skipped)
//
// NOTE: All tests are skipped (t.Skip). Activate after DLD-721:
//   - ChatViewModel integrates SpeechRecognizerProtocol
//   - Voice-transcribed text is forwarded to POST /chat/send
//   - Backend validates empty-message payloads (400)
package handler_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/handler"
	appmiddleware "github.com/dlddu/pocket-aide/middleware"
	"github.com/dlddu/pocket-aide/service/llm"
	"github.com/dlddu/pocket-aide/testutil"
)

// ---------------------------------------------------------------------------
// Happy Path — Voice transcription → POST /chat/send
// ---------------------------------------------------------------------------

// TestVoiceChatHandler_Send_TranscribedText_ReturnsSSEStream verifies that a
// message produced by speech recognition (plain text after transcription) can
// be forwarded to POST /chat/send and receives a valid SSE stream response.
//
// Scenario:
//
//	Client-side MockSpeechRecognizer transcribes "오늘 날씨 어때?" and places
//	the result in the message input field.
//	POST /chat/send  {"message":"오늘 날씨 어때?"} + Bearer token
//	→ 200 OK  Content-Type: text/event-stream, body contains "data:" line
func TestVoiceChatHandler_Send_TranscribedText_ReturnsSSEStream(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "오늘은 맑고 기온은 22도입니다.", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "voice@example.com", "Secret1!")

	// Act — transcribed speech text is sent as a normal chat message
	body := strings.NewReader(`{"message":"오늘 날씨 어때?"}`)
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

// TestVoiceChatHandler_Send_LongTranscribedText_StreamsMultipleTokens verifies
// that a long voice-recognition result is streamed back as multiple SSE tokens.
//
// Scenario:
//
//	MockSpeechRecognizer transcribes a long sentence.
//	POST /chat/send  {"message":"내일 오전에 회의 일정을 잡아줘"} + Bearer token
//	→ 200 OK  SSE stream with ≥ 3 "data:" lines
func TestVoiceChatHandler_Send_LongTranscribedText_StreamsMultipleTokens(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "내일 오전 10시에 회의 일정을 등록했습니다.", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "voice-long@example.com", "Secret1!")

	// Act
	body := strings.NewReader(`{"message":"내일 오전에 회의 일정을 잡아줘"}`)
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
		t.Errorf("expected >= 3 'data:' lines for long transcription SSE stream, got %d (body: %s)", dataLineCount, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Error Case — Empty voice recognition result
// ---------------------------------------------------------------------------

// TestVoiceChatHandler_Send_EmptyTranscription_ReturnsBadRequest verifies
// that forwarding an empty speech recognition result (empty message string)
// to POST /chat/send returns HTTP 400 Bad Request.
//
// Scenario:
//
//	MockSpeechRecognizer produces an empty transcript (e.g. background noise
//	only, or user stopped before speaking).
//	POST /chat/send  {"message":""} + Bearer token
//	→ 400 Bad Request
func TestVoiceChatHandler_Send_EmptyTranscription_ReturnsBadRequest(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "response", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "voice-empty@example.com", "Secret1!")

	// Act — empty transcription forwarded as empty message
	body := strings.NewReader(`{"message":""}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/send", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty transcription, got %d (body: %s)", rec.Code, rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Model Selection — Voice transcription + model field
// ---------------------------------------------------------------------------

// TestVoiceChatHandler_Send_WithModelSelection_ReturnsSSEStream verifies that
// a voice-transcribed message sent with an explicit model field is processed
// correctly and returns a valid SSE stream.
//
// Scenario:
//
//	User selects "gpt-4o" model, then activates voice input.
//	MockSpeechRecognizer transcribes "파이썬으로 피보나치 수열 짜줘".
//	POST /chat/send  {"message":"파이썬으로 피보나치 수열 짜줘","model":"gpt-4o"} + Bearer token
//	→ 200 OK  SSE stream; MockProvider.CallCount == 1
func TestVoiceChatHandler_Send_WithModelSelection_ReturnsSSEStream(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "def fibonacci(n): return n if n <= 1 else fibonacci(n-1) + fibonacci(n-2)", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "voice-model@example.com", "Secret1!")

	// Act
	body := strings.NewReader(`{"message":"파이썬으로 피보나치 수열 짜줘","model":"gpt-4o"}`)
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
		t.Errorf("expected SSE body to contain 'data:' line, got: %s", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// History — Voice message appears in chat history
// ---------------------------------------------------------------------------

// TestVoiceChatHandler_History_ContainsVoiceMessage verifies that a message
// sent via voice input (after transcription) appears in the chat history just
// like a regular text message.
//
// Scenario:
//
//	POST /chat/send  {"message":"음성으로 보낸 메시지"} + Bearer token  → 200 OK
//	GET  /chat/history  + Bearer token
//	→ 200 OK  JSON array contains {"role":"user","content":"음성으로 보낸 메시지"}
func TestVoiceChatHandler_History_ContainsVoiceMessage(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "네, 알겠습니다.", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))
	e.GET("/chat/history", chatHandler.History, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "voice-history@example.com", "Secret1!")

	// Send a voice-transcribed message
	sendBody := strings.NewReader(`{"message":"음성으로 보낸 메시지"}`)
	sendReq := httptest.NewRequest(http.MethodPost, "/chat/send", sendBody)
	sendReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	sendReq.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	sendRec := httptest.NewRecorder()
	e.ServeHTTP(sendRec, sendReq)
	if sendRec.Code != http.StatusOK {
		t.Fatalf("send voice message failed: status %d, body: %s", sendRec.Code, sendRec.Body.String())
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

	found := false
	for _, msg := range history {
		if msg["role"] == "user" && msg["content"] == "음성으로 보낸 메시지" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected history to contain voice message, got: %v", history)
	}
}

// ---------------------------------------------------------------------------
// Error Case — Unauthenticated voice message
// ---------------------------------------------------------------------------

// TestVoiceChatHandler_Send_WithoutAuth_ReturnsUnauthorized verifies that
// forwarding a voice-transcribed message without an Authorization header
// returns HTTP 401.
//
// Scenario:
//
//	POST /chat/send  {"message":"인증 없이 보낸 음성 메시지"}  (no Authorization header)
//	→ 401 Unauthorized
func TestVoiceChatHandler_Send_WithoutAuth_ReturnsUnauthorized(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	e := echo.New()
	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "response", nil
	}}
	chatHandler := handler.NewChatHandler(nil, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	// Act
	body := strings.NewReader(`{"message":"인증 없이 보낸 음성 메시지"}`)
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
// Intermediate transcription — only final result is sent
// ---------------------------------------------------------------------------

// TestVoiceChatHandler_Send_FinalTranscriptOnly_ReturnsSSEStream verifies
// that when the client receives multiple intermediate transcription updates
// (simulatedPhrases stream), only the final consolidated result is forwarded
// to POST /chat/send as a single request.
//
// Scenario:
//
//	MockSpeechRecognizer emits interim updates: "회의", "회의 일정", "회의 일정 추가해줘"
//	Client sends only the final transcript once recording stops.
//	POST /chat/send  {"message":"회의 일정 추가해줘"} + Bearer token
//	→ 200 OK  SSE stream; MockProvider.CallCount == 1 (only one request sent)
func TestVoiceChatHandler_Send_FinalTranscriptOnly_ReturnsSSEStream(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "회의 일정을 추가했습니다.", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "voice-interim@example.com", "Secret1!")

	// Act — client sends only the final result of the transcription stream
	// (intermediate results "회의", "회의 일정" are discarded client-side)
	body := strings.NewReader(`{"message":"회의 일정 추가해줘"}`)
	req := httptest.NewRequest(http.MethodPost, "/chat/send", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Assert — exactly one LLM call (no duplicate calls from intermediate results)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if mockLLM.CallCount != 1 {
		t.Errorf("expected exactly 1 LLM call for final transcript, got %d", mockLLM.CallCount)
	}
	if !strings.Contains(rec.Body.String(), "data:") {
		t.Errorf("expected SSE body to contain 'data:', got: %s", rec.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Mixed usage — Voice and text messages appear in order in history
// ---------------------------------------------------------------------------

// TestVoiceChatHandler_History_MixedVoiceAndText_InOrder verifies that when a
// user alternates between voice and text input, the chat history reflects all
// messages in the correct chronological order.
//
// Scenario:
//
//	1. POST /chat/send  {"message":"안녕하세요"} (text)      + Bearer token → 200 OK
//	2. POST /chat/send  {"message":"오늘 일정 알려줘"} (voice) + Bearer token → 200 OK
//	3. POST /chat/send  {"message":"고마워"} (text)          + Bearer token → 200 OK
//	GET  /chat/history  + Bearer token
//	→ 200 OK  JSON array with user messages in order:
//	  [user:"안녕하세요", assistant:…, user:"오늘 일정 알려줘", assistant:…, user:"고마워", assistant:…]
func TestVoiceChatHandler_History_MixedVoiceAndText_InOrder(t *testing.T) {
	t.Skip("DLD-721: 미구현 — 음성 채팅 e2e 테스트")

	// Arrange
	tdb := testutil.NewTestDB(t)
	e := echo.New()
	authHandler := handler.NewAuthHandler(tdb.DB, "test-jwt-secret")
	e.POST("/auth/login", authHandler.Login)

	mockLLM := &llm.MockProvider{CompleteFunc: func(ctx context.Context, prompt string) (string, error) {
		return "AI response", nil
	}}
	chatHandler := handler.NewChatHandler(tdb.DB, mockLLM, "mock")
	e.POST("/chat/send", chatHandler.Send, appmiddleware.JWT("test-jwt-secret"))
	e.GET("/chat/history", chatHandler.History, appmiddleware.JWT("test-jwt-secret"))

	token := seedUserAndLogin(t, tdb, e, "voice-mixed@example.com", "Secret1!")

	sendMessage := func(message string) {
		t.Helper()
		b := strings.NewReader(`{"message":"` + message + `"}`)
		r := httptest.NewRequest(http.MethodPost, "/chat/send", b)
		r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		r.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
		rr := httptest.NewRecorder()
		e.ServeHTTP(rr, r)
		if rr.Code != http.StatusOK {
			t.Fatalf("send message %q failed: status %d, body: %s", message, rr.Code, rr.Body.String())
		}
	}

	// 1. Text message
	sendMessage("안녕하세요")
	// 2. Voice-transcribed message
	sendMessage("오늘 일정 알려줘")
	// 3. Text message
	sendMessage("고마워")

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

	// Expect at least 6 entries: 3 user messages + 3 AI responses
	if len(history) < 6 {
		t.Errorf("expected >= 6 messages in mixed history, got %d", len(history))
	}

	// Verify user messages appear in correct order
	expectedUserMessages := []string{"안녕하세요", "오늘 일정 알려줘", "고마워"}
	userMsgIdx := 0
	for _, msg := range history {
		if msg["role"] == "user" {
			if userMsgIdx >= len(expectedUserMessages) {
				t.Errorf("more user messages than expected")
				break
			}
			if msg["content"] != expectedUserMessages[userMsgIdx] {
				t.Errorf("expected user message[%d] == %q, got %q", userMsgIdx, expectedUserMessages[userMsgIdx], msg["content"])
			}
			userMsgIdx++
		}
	}
	if userMsgIdx != len(expectedUserMessages) {
		t.Errorf("expected %d user messages in history, found %d", len(expectedUserMessages), userMsgIdx)
	}
}
