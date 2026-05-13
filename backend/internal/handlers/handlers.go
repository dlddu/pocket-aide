// Package handlers wires the HTTP routes (health, auth/config, me).
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
)

const pingTimeout = 2 * time.Second

// AuthConfig is what /api/auth/config returns to the iOS client. Values come
// from the backend's environment (ConfigMap in K8s) so the IdP can be swapped
// without rebuilding the app. DevAuthTokenPath is non-empty only when the
// backend is running with POCKET_AIDE_DEV=1, signalling that the client can
// skip the interactive OIDC flow and POST there to mint an oidcmock-signed
// access token.
type AuthConfig struct {
	Issuer           string `json:"issuer"`
	ClientID         string `json:"client_id"`
	RedirectURI      string `json:"redirect_uri"`
	Audience         string `json:"audience"`
	DevAuthTokenPath string `json:"dev_auth_token_path,omitempty"`
}

// Health returns an /healthz handler that pings the database.
func Health(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "db unhealthy: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// AuthConfigHandler echoes the OIDC settings the iOS client should use.
func AuthConfigHandler(cfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, cfg)
	}
}

// Me returns the authenticated user (set by auth.Middleware).
func Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":  u.ID,
			"sub": u.Sub,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
