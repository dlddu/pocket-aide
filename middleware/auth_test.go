package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	appmiddleware "github.com/dlddu/pocket-aide/middleware"
)

const testJWTSecret = "test-secret-key"

func generateTestToken(t *testing.T, secret string, expiry time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": "user123",
		"exp": time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	return signed
}

func TestJWTMiddleware_AllowsValidToken(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	token := generateTestToken(t, testJWTSecret, time.Hour)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := appmiddleware.JWT(testJWTSecret)

	// Act
	err := mw(handler)(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error for valid token, got: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestJWTMiddleware_RejectsRequestWithoutToken(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := appmiddleware.JWT(testJWTSecret)

	// Act
	err := mw(handler)(c)

	// Assert
	if err == nil {
		t.Error("expected error for missing token, got nil")
		return
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", httpErr.Code)
	}
}

func TestJWTMiddleware_RejectsExpiredToken(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	token := generateTestToken(t, testJWTSecret, -time.Hour)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := appmiddleware.JWT(testJWTSecret)

	// Act
	err := mw(handler)(c)

	// Assert
	if err == nil {
		t.Error("expected error for expired token, got nil")
		return
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for expired token, got %d", httpErr.Code)
	}
}

func TestJWTMiddleware_RejectsTokenWithWrongSecret(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	token := generateTestToken(t, "wrong-secret", time.Hour)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	mw := appmiddleware.JWT(testJWTSecret)

	// Act
	err := mw(handler)(c)

	// Assert
	if err == nil {
		t.Error("expected error for token signed with wrong secret, got nil")
		return
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for wrong secret, got %d", httpErr.Code)
	}
}

func TestJWTMiddleware_SetsUserInContext(t *testing.T) {
	// Arrange
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	token := generateTestToken(t, testJWTSecret, time.Hour)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var capturedUser interface{}
	handler := func(c echo.Context) error {
		capturedUser = c.Get("user")
		return c.String(http.StatusOK, "ok")
	}

	mw := appmiddleware.JWT(testJWTSecret)

	// Act
	err := mw(handler)(c)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if capturedUser == nil {
		t.Error("expected 'user' to be set in context, got nil")
	}
}
