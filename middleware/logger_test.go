package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	appmiddleware "github.com/dlddu/pocket-aide/middleware"
)

func TestRequestLogger_PassesRequestToNextHandler(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return c.String(http.StatusOK, "ok")
	}

	mw := appmiddleware.RequestLogger()

	// Act
	err := mw(handler)(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !handlerCalled {
		t.Error("expected next handler to be called, but it was not")
	}
}

func TestRequestLogger_PreservesResponseStatus(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/resource", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusCreated, "created")
	}

	mw := appmiddleware.RequestLogger()

	// Act
	err := mw(handler)(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
}

func TestRequestLogger_PropagatesHandlerError(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedErr := echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	handler := func(c echo.Context) error {
		return expectedErr
	}

	mw := appmiddleware.RequestLogger()

	// Act
	err := mw(handler)(c)

	// Assert
	if err == nil {
		t.Error("expected error to be propagated, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected propagated error to match, got: %v", err)
	}
}

func TestRequestLogger_ReturnsMiddlewareFunc(t *testing.T) {
	// Arrange & Act
	mw := appmiddleware.RequestLogger()

	// Assert
	if mw == nil {
		t.Error("expected RequestLogger to return a non-nil MiddlewareFunc")
	}
}
