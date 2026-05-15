package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
)

type deviceTokenPayload struct {
	Token string `json:"token"`
}

// RegisterDeviceToken handles POST /api/device-tokens. The body carries the
// hex-encoded APNs device token captured by the iOS client after the user
// granted notification permission.
func RegisterDeviceToken(store *devicetokens.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		body, err := decodeDeviceTokenPayload(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := store.Upsert(r.Context(), u.ID, body.Token); err != nil {
			http.Error(w, "upsert failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
	}
}

func decodeDeviceTokenPayload(r *http.Request) (deviceTokenPayload, error) {
	var p deviceTokenPayload
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		return p, errors.New("invalid json body")
	}
	p.Token = strings.TrimSpace(p.Token)
	if p.Token == "" {
		return p, errors.New("token is required")
	}
	// APNs device tokens are 32 raw bytes → 64 hex chars. Reject anything else
	// before it reaches the DB or APNs.
	if len(p.Token) != 64 || !isHex(p.Token) {
		return p, errors.New("token must be 64 hex characters")
	}
	return p, nil
}

func isHex(s string) bool {
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
