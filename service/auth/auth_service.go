// Package auth provides authentication business logic.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/dlddu/pocket-aide/repository"
)

// ErrInvalidCredentials is returned when authentication fails due to an
// unknown email or incorrect password. A single sentinel error is used for
// both cases to prevent user-enumeration attacks.
var ErrInvalidCredentials = errors.New("invalid credentials")

// UserRepository is the interface that AuthService uses to read and write
// user records. Defined here so that the service layer does not depend
// directly on the concrete repository type.
type UserRepository interface {
	CreateUser(email, passwordHash string) (*repository.User, error)
	FindByEmail(email string) (*repository.User, error)
}

// AuthService handles user registration and authentication.
type AuthService struct {
	repo      UserRepository
	jwtSecret string
}

// NewAuthService creates a new AuthService with the provided repository and
// JWT signing secret.
func NewAuthService(repo UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user account. The password is hashed with bcrypt
// before being stored. It returns the newly created user.
// If the email already exists, the repository's error is propagated.
func (s *AuthService) Register(email, password string) (*repository.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.repo.CreateUser(email, string(hash))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user by email and password. On success it returns a
// signed JWT token. If the email is not found or the password does not match,
// ErrInvalidCredentials is returned (never ErrUserNotFound) to prevent
// user-enumeration attacks.
func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		// Map any repository error (including ErrUserNotFound) to
		// ErrInvalidCredentials to prevent leaking whether the email exists.
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// generateJWT creates and signs a JWT for the given user.
// Claims: sub (user ID), email, exp (24 h from now).
func (s *AuthService) generateJWT(user *repository.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return signed, nil
}
