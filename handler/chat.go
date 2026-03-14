// Package handler contains Echo HTTP handler implementations.
package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/service/chat"
	"github.com/dlddu/pocket-aide/service/llm"
)

// ChatHandler handles AI chat HTTP requests.
type ChatHandler struct {
	service *chat.ChatService
}

// NewChatHandler constructs a ChatHandler wired to the given database,
// LLM provider, and default model name.
func NewChatHandler(db *sql.DB, llmProvider llm.Provider, defaultModel string) *ChatHandler {
	router := llm.NewRouter(defaultModel)
	router.RegisterProvider(defaultModel, llmProvider)

	repo := repository.NewChatRepository(db)
	svc := chat.NewChatService(repo, router)

	return &ChatHandler{
		service: svc,
	}
}

// sendRequest is the expected JSON body for POST /chat/send.
type sendRequest struct {
	Message string `json:"message"`
	Model   string `json:"model"`
}

// historyItem is a single message in the history response.
type historyItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Send handles POST /chat/send.
// It sends the user's message to the LLM and streams the response via SSE.
func (h *ChatHandler) Send(c echo.Context) error {
	// Extract user ID from JWT claims
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req sendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Message == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "message is required"})
	}

	// Get LLM response
	response, err := h.service.SendMessage(c.Request().Context(), userID, req.Message, req.Model)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to process message"})
	}

	// Set SSE headers
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	// Stream response as SSE tokens (split by whitespace)
	tokens := strings.Fields(response)
	flusher, hasFlusher := c.Response().Writer.(http.Flusher)

	for i, token := range tokens {
		var sseData string
		if i == 0 {
			sseData = fmt.Sprintf("data: %s\n\n", token)
		} else {
			sseData = fmt.Sprintf("data:  %s\n\n", token)
		}
		if _, err := fmt.Fprint(c.Response().Writer, sseData); err != nil {
			break
		}
		if hasFlusher {
			flusher.Flush()
		}
	}

	// Send done event
	fmt.Fprint(c.Response().Writer, "data: [DONE]\n\n")
	if hasFlusher {
		flusher.Flush()
	}

	return nil
}

// History handles GET /chat/history.
// It returns the chat history for the authenticated user as a JSON array.
func (h *ChatHandler) History(c echo.Context) error {
	// Extract user ID from JWT claims
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	messages, err := h.service.GetHistory(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve history"})
	}

	items := make([]historyItem, 0, len(messages))
	for _, msg := range messages {
		items = append(items, historyItem{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return c.JSON(http.StatusOK, items)
}

// extractUserID extracts the user ID from the JWT token stored in the echo context.
func extractUserID(c echo.Context) (int64, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return 0, fmt.Errorf("no token in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims type")
	}

	subRaw, exists := claims["sub"]
	if !exists {
		return 0, fmt.Errorf("missing sub claim")
	}

	switch v := subRaw.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("unexpected sub claim type: %T", subRaw)
	}
}
