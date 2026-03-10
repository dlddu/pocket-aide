package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	appmiddleware "github.com/dlddu/pocket-aide/middleware"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func TestCustomErrorHandler_HandlesHTTPError(t *testing.T) {
	// Arrange
	e := echo.New()
	e.HTTPErrorHandler = appmiddleware.CustomErrorHandler

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	httpErr := echo.NewHTTPError(http.StatusNotFound, "resource not found")

	// Act
	appmiddleware.CustomErrorHandler(httpErr, c)

	// Assert
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}
	if resp.Code != http.StatusNotFound {
		t.Errorf("expected response code 404, got %d", resp.Code)
	}
}

func TestCustomErrorHandler_HandlesGenericError(t *testing.T) {
	// Arrange
	e := echo.New()
	e.HTTPErrorHandler = appmiddleware.CustomErrorHandler

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	genericErr := errors.New("something went wrong internally")

	// Act
	appmiddleware.CustomErrorHandler(genericErr, c)

	// Assert
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 for generic error, got %d", rec.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid JSON response, got error: %v", err)
	}
	if resp.Code != http.StatusInternalServerError {
		t.Errorf("expected response code 500, got %d", resp.Code)
	}
}

func TestCustomErrorHandler_Returns422ForValidationError(t *testing.T) {
	// Arrange
	e := echo.New()
	e.HTTPErrorHandler = appmiddleware.CustomErrorHandler

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	validationErr := echo.NewHTTPError(http.StatusUnprocessableEntity, "validation failed")

	// Act
	appmiddleware.CustomErrorHandler(validationErr, c)

	// Assert
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", rec.Code)
	}
}

func TestCustomErrorHandler_ResponseContentTypeIsJSON(t *testing.T) {
	// Arrange
	e := echo.New()
	e.HTTPErrorHandler = appmiddleware.CustomErrorHandler

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	httpErr := echo.NewHTTPError(http.StatusBadRequest, "bad request")

	// Act
	appmiddleware.CustomErrorHandler(httpErr, c)

	// Assert
	contentType := rec.Header().Get("Content-Type")
	if contentType == "" {
		t.Error("expected Content-Type header to be set")
	}
}
