// Package handler contains Echo HTTP handler implementations.
package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/repository"
	authsvc "github.com/dlddu/pocket-aide/service/auth"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	service   *authsvc.AuthService
	repo      *repository.UserRepository
	jwtSecret string
}

// NewAuthHandler constructs an AuthHandler wired to the given database and
// JWT signing secret.
func NewAuthHandler(db *sql.DB, jwtSecret string) *AuthHandler {
	repo := repository.NewUserRepository(db)
	svc := authsvc.NewAuthService(repo, jwtSecret)
	return &AuthHandler{
		service:   svc,
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// registerRequest is the expected JSON body for POST /auth/register.
type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRequest is the expected JSON body for POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles POST /auth/register.
// On success it responds with 201 and the new user's id and email.
// On duplicate email it responds with 409.
// On missing or malformed request body it responds with 400.
func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
	}

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return c.JSON(http.StatusConflict, map[string]string{"error": "email already exists"})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

// Login handles POST /auth/login.
// On success it responds with 200 and a signed JWT token.
// On invalid credentials it responds with 401.
func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "email and password are required"})
	}

	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, authsvc.ErrInvalidCredentials) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// Me handles GET /me.
// It reads the user ID from the JWT claims injected by the JWT middleware,
// fetches the user from the database, and responds with the user's id and email.
func (h *AuthHandler) Me(c echo.Context) error {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// The "sub" claim is set as int64 (user.ID) during JWT generation.
	// After JWT parsing MapClaims decodes numbers as float64.
	subRaw, exists := claims["sub"]
	if !exists {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var userID int64
	switch v := subRaw.(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	default:
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// Fetch user from DB by ID via a direct query through the repository's db.
	// We use FindByEmail is not applicable here so query directly.
	user, err := h.repo.FindByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}
