// Package auth_test contains unit tests for AuthService.
//
// AuthService depends on a UserRepository. These tests use a mock repository
// so they run without a real database, keeping them fast and isolated.
//
// TDD Red Phase: the AuthService type and its methods do not yet exist.
// All tests are expected to fail until service/auth/auth_service.go is implemented.
package auth_test

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/dlddu/pocket-aide/repository"
	"github.com/dlddu/pocket-aide/service/auth"
)

// ---------------------------------------------------------------------------
// Mock UserRepository
// ---------------------------------------------------------------------------

// mockUserRepository is a test double for the UserRepository interface that
// AuthService will depend on.
// Each field holds an optional stub function; if nil the method returns an error.
type mockUserRepository struct {
	createUserFunc  func(email, passwordHash string) (*repository.User, error)
	findByEmailFunc func(email string) (*repository.User, error)
}

func (m *mockUserRepository) CreateUser(email, passwordHash string) (*repository.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(email, passwordHash)
	}
	return nil, errors.New("mockUserRepository.CreateUser not configured")
}

func (m *mockUserRepository) FindByEmail(email string) (*repository.User, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(email)
	}
	return nil, errors.New("mockUserRepository.FindByEmail not configured")
}

// bcryptHashForTest generates a bcrypt hash using minimum cost so tests stay fast.
func bcryptHashForTest(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword failed in test setup: %v", err)
	}
	return string(hash)
}

// ---------------------------------------------------------------------------
// Register
// ---------------------------------------------------------------------------

// TestAuthService_Register_ReturnsUser verifies that Register hashes the
// password and stores a new user, returning the persisted user with a non-zero ID.
//
// Scenario:
//
//	Register(email="new@example.com", password="Secret1!")
//	→ User{ID: >0, Email: "new@example.com"}, nil
func TestAuthService_Register_ReturnsUser(t *testing.T) {
	// Arrange
	wantEmail := "new@example.com"
	mock := &mockUserRepository{
		createUserFunc: func(email, passwordHash string) (*repository.User, error) {
			return &repository.User{ID: 1, Email: email, PasswordHash: passwordHash}, nil
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	user, err := svc.Register(wantEmail, "Secret1!")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user, got nil")
	}
	if user.ID == 0 {
		t.Error("expected non-zero user ID")
	}
	if user.Email != wantEmail {
		t.Errorf("expected email '%s', got '%s'", wantEmail, user.Email)
	}
}

// TestAuthService_Register_HashesPassword verifies that Register never stores
// the plaintext password — the repository receives a bcrypt hash, not the
// original password string.
//
// Scenario:
//
//	Register(email="hash@example.com", password="MyPass99!")
//	→ repository.CreateUser is called with a bcrypt hash, not "MyPass99!"
func TestAuthService_Register_HashesPassword(t *testing.T) {
	// Arrange
	var capturedHash string
	mock := &mockUserRepository{
		createUserFunc: func(email, passwordHash string) (*repository.User, error) {
			capturedHash = passwordHash
			return &repository.User{ID: 1, Email: email, PasswordHash: passwordHash}, nil
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	_, err := svc.Register("hash@example.com", "MyPass99!")
	if err != nil {
		t.Fatalf("unexpected error from Register: %v", err)
	}

	// Assert: captured value must not be the plaintext password
	if capturedHash == "MyPass99!" {
		t.Error("Register stored plaintext password; expected bcrypt hash")
	}
	if len(capturedHash) == 0 {
		t.Error("expected a non-empty password hash to be stored")
	}
}

// TestAuthService_Register_DuplicateEmail_ReturnsError verifies that when the
// repository signals a duplicate email, Register propagates an error.
//
// Scenario:
//
//	Repository returns ErrDuplicateEmail for CreateUser.
//	Register(email="dup@example.com", password="Secret1!")
//	→ non-nil error, nil user
func TestAuthService_Register_DuplicateEmail_ReturnsError(t *testing.T) {
	// Arrange
	mock := &mockUserRepository{
		createUserFunc: func(email, passwordHash string) (*repository.User, error) {
			return nil, repository.ErrDuplicateEmail
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	user, err := svc.Register("dup@example.com", "Secret1!")

	// Assert
	if err == nil {
		t.Error("expected error for duplicate email, got nil")
	}
	if user != nil {
		t.Errorf("expected nil user on error, got %+v", user)
	}
}

// TestAuthService_Register_DuplicateEmail_WrapsErrDuplicateEmail verifies that
// the error returned by Register can be unwrapped to repository.ErrDuplicateEmail
// so that the HTTP handler can produce a 409 Conflict response.
//
// Scenario:
//
//	Repository returns ErrDuplicateEmail.
//	Register(email="dup@example.com", ...)
//	→ errors.Is(err, repository.ErrDuplicateEmail) == true
func TestAuthService_Register_DuplicateEmail_WrapsErrDuplicateEmail(t *testing.T) {
	// Arrange
	mock := &mockUserRepository{
		createUserFunc: func(email, passwordHash string) (*repository.User, error) {
			return nil, repository.ErrDuplicateEmail
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	_, err := svc.Register("dup@example.com", "Secret1!")

	// Assert
	if !errors.Is(err, repository.ErrDuplicateEmail) {
		t.Errorf("expected error chain to contain ErrDuplicateEmail, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Login
// ---------------------------------------------------------------------------

// TestAuthService_Login_ReturnsJWTToken verifies that Login returns a non-empty
// JWT token string when the email and password are correct.
//
// Scenario:
//
//	Repository returns user with bcrypt-hashed "Secret1!".
//	Login(email="user@example.com", password="Secret1!")
//	→ non-empty JWT token string, nil error
func TestAuthService_Login_ReturnsJWTToken(t *testing.T) {
	// Arrange
	const plainPassword = "Secret1!"
	hashedPassword := bcryptHashForTest(t, plainPassword)

	mock := &mockUserRepository{
		findByEmailFunc: func(email string) (*repository.User, error) {
			return &repository.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashedPassword,
			}, nil
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	token, err := svc.Login("user@example.com", plainPassword)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty JWT token, got empty string")
	}
}

// TestAuthService_Login_TokenIsValidJWT verifies that the token returned by
// Login has the three-segment dot-separated structure of a JWT
// (header.payload.signature).
//
// Scenario:
//
//	Login with valid credentials
//	→ returned token has exactly two '.' characters
func TestAuthService_Login_TokenIsValidJWT(t *testing.T) {
	// Arrange
	const plainPassword = "Secret1!"
	hashedPassword := bcryptHashForTest(t, plainPassword)

	mock := &mockUserRepository{
		findByEmailFunc: func(email string) (*repository.User, error) {
			return &repository.User{ID: 1, Email: email, PasswordHash: hashedPassword}, nil
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	token, err := svc.Login("user@example.com", plainPassword)
	if err != nil {
		t.Fatalf("Login returned unexpected error: %v", err)
	}

	// Assert: a compact-serialized JWT has exactly two dots
	dotCount := 0
	for _, ch := range token {
		if ch == '.' {
			dotCount++
		}
	}
	if dotCount != 2 {
		t.Errorf("expected JWT with 2 dots (3 segments), got token: %s", token)
	}
}

// TestAuthService_Login_WrongPassword_ReturnsError verifies that Login returns
// an error and an empty token when the password does not match the stored hash.
//
// Scenario:
//
//	Repository returns user with bcrypt-hashed "Secret1!".
//	Login(email="user@example.com", password="wrongpass")
//	→ non-nil error, empty token
func TestAuthService_Login_WrongPassword_ReturnsError(t *testing.T) {
	// Arrange
	hashedPassword := bcryptHashForTest(t, "Secret1!")

	mock := &mockUserRepository{
		findByEmailFunc: func(email string) (*repository.User, error) {
			return &repository.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashedPassword,
			}, nil
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	token, err := svc.Login("user@example.com", "wrongpass")

	// Assert
	if err == nil {
		t.Error("expected error for wrong password, got nil")
	}
	if token != "" {
		t.Errorf("expected empty token on auth failure, got: %s", token)
	}
}

// TestAuthService_Login_WrongPassword_ReturnsErrInvalidCredentials verifies
// that the error returned for a wrong password can be unwrapped to
// auth.ErrInvalidCredentials, allowing the HTTP handler to return 401.
//
// Scenario:
//
//	Login with wrong password
//	→ errors.Is(err, auth.ErrInvalidCredentials) == true
func TestAuthService_Login_WrongPassword_ReturnsErrInvalidCredentials(t *testing.T) {
	// Arrange
	hashedPassword := bcryptHashForTest(t, "Secret1!")

	mock := &mockUserRepository{
		findByEmailFunc: func(email string) (*repository.User, error) {
			return &repository.User{
				ID:           1,
				Email:        email,
				PasswordHash: hashedPassword,
			}, nil
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	_, err := svc.Login("user@example.com", "wrongpass")

	// Assert
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

// TestAuthService_Login_UnknownEmail_ReturnsError verifies that Login returns
// an error when the repository cannot find a user with the given email.
//
// Scenario:
//
//	Repository returns ErrUserNotFound for FindByEmail.
//	Login(email="ghost@example.com", password="Secret1!")
//	→ non-nil error, empty token
func TestAuthService_Login_UnknownEmail_ReturnsError(t *testing.T) {
	// Arrange
	mock := &mockUserRepository{
		findByEmailFunc: func(email string) (*repository.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	token, err := svc.Login("ghost@example.com", "Secret1!")

	// Assert
	if err == nil {
		t.Error("expected error for unknown email, got nil")
	}
	if token != "" {
		t.Errorf("expected empty token when user not found, got: %s", token)
	}
}

// TestAuthService_Login_UnknownEmail_ReturnsErrInvalidCredentials verifies
// that Login returns the same ErrInvalidCredentials sentinel for an unknown
// email as for a wrong password, preventing user-enumeration attacks.
//
// Scenario:
//
//	Repository returns ErrUserNotFound.
//	Login(email="ghost@example.com", ...)
//	→ errors.Is(err, auth.ErrInvalidCredentials) == true
func TestAuthService_Login_UnknownEmail_ReturnsErrInvalidCredentials(t *testing.T) {
	// Arrange
	mock := &mockUserRepository{
		findByEmailFunc: func(email string) (*repository.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}
	svc := auth.NewAuthService(mock, "test-jwt-secret")

	// Act
	_, err := svc.Login("ghost@example.com", "Secret1!")

	// Assert: must be ErrInvalidCredentials, not ErrUserNotFound, to avoid
	// leaking whether the email address is registered in the system.
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials (not ErrUserNotFound) to prevent user enumeration, got: %v", err)
	}
}
