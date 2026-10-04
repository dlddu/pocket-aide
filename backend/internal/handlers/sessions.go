package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/sessions"
)

const maxSessionRequestBytes = 8 << 20

type createSessionPayload struct {
	Name         string  `json:"name"`
	Model        string  `json:"model,omitempty"`
	WorkloadType *string `json:"workloadType,omitempty"`
}

type platformCreateRequest struct {
	Name         string `json:"name"`
	WorkloadType string `json:"workloadType"`
	Model        string `json:"model,omitempty"`
}

var platformErrorMessages = map[int]string{
	http.StatusBadRequest:            "invalid request",
	http.StatusNotFound:              "session not found",
	http.StatusConflict:              "session is busy",
	http.StatusRequestEntityTooLarge: "request too large",
	http.StatusUnprocessableEntity:   "session is snapshotted",
	http.StatusTooManyRequests:       "prompt queue is full",
	http.StatusServiceUnavailable:    "feature disabled on the session platform",
	http.StatusInsufficientStorage:   "session output quota is full",
}

func writeSessionError(w http.ResponseWriter, status int) {
	msg, ok := platformErrorMessages[status]
	if !ok {
		msg = "session platform error"
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeUnreachable(w http.ResponseWriter) {
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "session platform unreachable"})
}

func readSessionBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSessionRequestBytes))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeSessionError(w, http.StatusRequestEntityTooLarge)
			return nil, false
		}
		writeSessionError(w, http.StatusBadRequest)
		return nil, false
	}
	return raw, true
}

func SessionConfig(client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, raw, err := client.Do(r.Context(), http.MethodGet, "/config", nil)
		if err != nil {
			writeUnreachable(w)
			return
		}
		if status != http.StatusOK {
			writeSessionError(w, status)
			return
		}
		var cfg sessions.RuntimeConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			writeSessionError(w, http.StatusBadGateway)
			return
		}
		if cfg.ClaudeCode.Models == nil {
			cfg.ClaudeCode.Models = []string{}
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, cfg)
	}
}

func ListSessions(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		owned, err := store.Owned(r.Context(), u.ID)
		if err != nil {
			http.Error(w, "list failed", http.StatusInternalServerError)
			return
		}
		out := []sessions.Session{}
		if len(owned) == 0 {
			writeJSON(w, http.StatusOK, map[string]any{"sessions": out})
			return
		}
		status, raw, err := client.Do(r.Context(), http.MethodGet, "/sessions", nil)
		if err != nil {
			writeUnreachable(w)
			return
		}
		if status != http.StatusOK {
			writeSessionError(w, status)
			return
		}
		var body struct {
			Sessions []sessions.Session `json:"sessions"`
		}
		if err := json.Unmarshal(raw, &body); err != nil {
			writeSessionError(w, http.StatusBadGateway)
			return
		}
		for _, s := range body.Sessions {
			if owned[s.ID] {
				out = append(out, s)
			}
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].LastAccess > out[j].LastAccess })
		writeJSON(w, http.StatusOK, map[string]any{"sessions": out})
	}
}

func CreateSession(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		raw, ok := readSessionBody(w, r)
		if !ok {
			return
		}
		var p createSessionPayload
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&p); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
			return
		}
		if p.WorkloadType != nil && *p.WorkloadType != sessions.WorkloadClaudeCode {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "workloadType must be claude-code"})
			return
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
			return
		}
		upstream, err := json.Marshal(platformCreateRequest{
			Name:         name,
			WorkloadType: sessions.WorkloadClaudeCode,
			Model:        strings.TrimSpace(p.Model),
		})
		if err != nil {
			http.Error(w, "encode failed", http.StatusInternalServerError)
			return
		}
		status, resp, err := client.Do(r.Context(), http.MethodPost, "/sessions", upstream)
		if err != nil {
			writeUnreachable(w)
			return
		}
		if status != http.StatusCreated {
			writeSessionError(w, status)
			return
		}
		var created sessions.Session
		if err := json.Unmarshal(resp, &created); err != nil || created.ID == "" {
			writeSessionError(w, http.StatusBadGateway)
			return
		}
		if err := store.Add(r.Context(), u.ID, created.ID); err != nil {
			_, _, _ = client.Do(r.Context(), http.MethodDelete, "/sessions/"+url.PathEscape(created.ID), nil)
			http.Error(w, "create failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

func ownedSessionID(w http.ResponseWriter, r *http.Request, store *sessions.Store) (int64, string, bool) {
	u, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "no user in context", http.StatusInternalServerError)
		return 0, "", false
	}
	id := chi.URLParam(r, "id")
	owns, err := store.Owns(r.Context(), u.ID, id)
	if err != nil {
		http.Error(w, "lookup failed", http.StatusInternalServerError)
		return 0, "", false
	}
	if !owns {
		writeSessionError(w, http.StatusNotFound)
		return 0, "", false
	}
	return u.ID, id, true
}

func relaySession[T any](w http.ResponseWriter, status int, raw []byte) {
	if status != http.StatusOK {
		writeSessionError(w, status)
		return
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		writeSessionError(w, http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func GetSession(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, id, ok := ownedSessionID(w, r, store)
		if !ok {
			return
		}
		status, raw, err := client.Do(r.Context(), http.MethodGet, "/sessions/"+url.PathEscape(id), nil)
		if err != nil {
			writeUnreachable(w)
			return
		}
		relaySession[sessions.Session](w, status, raw)
	}
}

func DeleteSession(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, id, ok := ownedSessionID(w, r, store)
		if !ok {
			return
		}
		status, _, err := client.Do(r.Context(), http.MethodDelete, "/sessions/"+url.PathEscape(id), nil)
		if err != nil {
			writeUnreachable(w)
			return
		}
		if status != http.StatusNoContent && status != http.StatusNotFound {
			writeSessionError(w, status)
			return
		}
		if err := store.Remove(r.Context(), userID, id); err != nil {
			http.Error(w, "delete failed", http.StatusInternalServerError)
			return
		}
		if status == http.StatusNotFound {
			writeSessionError(w, status)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func sessionAction[T any](store *sessions.Store, client *sessions.Client, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, id, ok := ownedSessionID(w, r, store)
		if !ok {
			return
		}
		raw, ok := readSessionBody(w, r)
		if !ok {
			return
		}
		var body []byte
		if len(bytes.TrimSpace(raw)) > 0 {
			body = raw
		}
		status, resp, err := client.Do(r.Context(), http.MethodPost, "/sessions/"+url.PathEscape(id)+"/"+action, body)
		if err != nil {
			writeUnreachable(w)
			return
		}
		relaySession[T](w, status, resp)
	}
}

func ReadSession(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.ReadResult](store, client, "read")
}

func WriteSession(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.WriteResult](store, client, "write")
}

func SwitchSession(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.Session](store, client, "switch")
}

func SnapshotSession(store *sessions.Store, client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.Session](store, client, "snapshot")
}
