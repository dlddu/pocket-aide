// Package auth verifies OIDC bearer tokens, JIT-provisions users, and exposes
// the authenticated user via request context.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type ctxKey string

const userCtxKey ctxKey = "auth.user"

// User is the authenticated principal injected into request context.
type User struct {
	ID  int64
	Sub string
}

// Verifier wraps the go-oidc IDTokenVerifier so it can be swapped in tests.
type Verifier interface {
	Verify(ctx context.Context, rawToken string) (*oidc.IDToken, error)
}

// Middleware returns a chi-compatible middleware that verifies the bearer
// token, JIT-provisions a user row keyed by `sub`, and stores the User in
// context. Requests without a token or with an invalid token get 401.
func Middleware(v Verifier, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, err := bearerToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			tok, err := v.Verify(r.Context(), raw)
			if err != nil {
				http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}
			user, err := upsertUser(r.Context(), db, tok.Subject)
			if err != nil {
				http.Error(w, "user provisioning failed", http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userCtxKey, user)))
		})
	}
}

func bearerToken(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("missing Authorization header")
	}
	if !strings.HasPrefix(h, "Bearer ") {
		return "", errors.New("authorization header must be Bearer")
	}
	tok := strings.TrimPrefix(h, "Bearer ")
	if tok == "" {
		return "", errors.New("empty bearer token")
	}
	return tok, nil
}

func upsertUser(ctx context.Context, db *sql.DB, sub string) (User, error) {
	var u User
	err := db.QueryRowContext(ctx, `SELECT id, oidc_sub FROM users WHERE oidc_sub = ?`, sub).Scan(&u.ID, &u.Sub)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return User{}, fmt.Errorf("select user: %w", err)
	}
	res, err := db.ExecContext(ctx, `INSERT INTO users (oidc_sub) VALUES (?)`, sub)
	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("last insert id: %w", err)
	}
	return User{ID: id, Sub: sub}, nil
}

// FromContext returns the authenticated user, if any.
func FromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userCtxKey).(User)
	return u, ok
}

// WithUser injects a User into context. Exported so tests in other packages
// can construct an authenticated request without going through the middleware.
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

// NewVerifier builds a real OIDC verifier against the given issuer + audience.
// The provider is fetched once at startup; JWKs rotation is handled internally
// by go-oidc via background refresh.
func NewVerifier(ctx context.Context, issuer, audience string) (Verifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}
	cfg := &oidc.Config{
		ClientID:          audience,
		SkipClientIDCheck: false,
	}
	return provider.Verifier(cfg), nil
}
