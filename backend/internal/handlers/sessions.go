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

func fetchRuntimeConfig(w http.ResponseWriter, r *http.Request, client *sessions.Client) (sessions.RuntimeConfig, bool) {
	var cfg sessions.RuntimeConfig
	status, raw, err := client.Do(r.Context(), http.MethodGet, "/config", nil)
	if err != nil {
		writeUnreachable(w)
		return cfg, false
	}
	if status != http.StatusOK {
		writeSessionError(w, status)
		return cfg, false
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		writeSessionError(w, http.StatusBadGateway)
		return cfg, false
	}
	cfg.Normalize()
	return cfg, true
}

func SessionConfig(client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, ok := fetchRuntimeConfig(w, r, client)
		if !ok {
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, cfg)
	}
}

func ListSessions(client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		out := body.Sessions
		if out == nil {
			out = []sessions.Session{}
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].LastAccess > out[j].LastAccess })
		writeJSON(w, http.StatusOK, map[string]any{"sessions": out})
	}
}

func CreateSession(client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		workload := sessions.WorkloadClaudeCode
		if p.WorkloadType != nil {
			workload = *p.WorkloadType
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
			return
		}
		cfg, ok := fetchRuntimeConfig(w, r, client)
		if !ok {
			return
		}
		info, known := cfg.Workload(workload)
		if !known {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "workloadType is not offered by the session platform"})
			return
		}
		model := strings.TrimSpace(p.Model)
		if model != "" && !info.TakesModel {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "this workloadType takes no model"})
			return
		}
		upstream, err := json.Marshal(platformCreateRequest{
			Name:         name,
			WorkloadType: workload,
			Model:        model,
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
		writeJSON(w, http.StatusCreated, created)
	}
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

func GetSession(client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		status, raw, err := client.Do(r.Context(), http.MethodGet, "/sessions/"+url.PathEscape(id), nil)
		if err != nil {
			writeUnreachable(w)
			return
		}
		relaySession[sessions.Session](w, status, raw)
	}
}

func DeleteSession(client *sessions.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		status, _, err := client.Do(r.Context(), http.MethodDelete, "/sessions/"+url.PathEscape(id), nil)
		if err != nil {
			writeUnreachable(w)
			return
		}
		if status != http.StatusNoContent {
			writeSessionError(w, status)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func sessionAction[T any](client *sessions.Client, action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
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

func ReadSession(client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.ReadResult](client, "read")
}

func WriteSession(client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.WriteResult](client, "write")
}

func SwitchSession(client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.Session](client, "switch")
}

func SnapshotSession(client *sessions.Client) http.HandlerFunc {
	return sessionAction[sessions.Session](client, "snapshot")
}
