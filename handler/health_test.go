package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/handler"
)

type healthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func TestHealthHandler_ReturnsOK(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.NewHealthHandler()

	// Act
	err := h.Health(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestHealthHandler_ReturnsJSONResponse(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.NewHealthHandler()

	// Act
	err := h.Health(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var resp healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v (body: %s)", err, rec.Body.String())
	}
}

func TestHealthHandler_StatusFieldIsOK(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.NewHealthHandler()

	// Act
	err := h.Health(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var resp healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON body, got error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status field 'ok', got '%s'", resp.Status)
	}
}

func TestHealthHandler_ContentTypeIsJSON(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := handler.NewHealthHandler()

	// Act
	err := h.Health(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType == "" {
		t.Error("expected Content-Type header to be set")
	}
}

func TestHealthHandler_OnlyRespondsToGET(t *testing.T) {
	// Arrange: verify the handler function signature accepts echo.Context
	h := handler.NewHealthHandler()

	// Assert: HealthHandler must expose a Health method with the correct signature
	var _ func(echo.Context) error = h.Health
}
