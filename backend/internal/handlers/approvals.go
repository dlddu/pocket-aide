package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/approvals"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
)

const maxApprovalRequestBytes = 1 << 20

type callerCtxKey struct{}

type externalCreatePayload struct {
	ExternalID     *string `json:"externalId"`
	Context        *string `json:"context"`
	RequesterName  *string `json:"requesterName"`
	TimeoutSeconds *int64  `json:"timeoutSeconds"`
}

type externalCreateResponse struct {
	ID        string           `json:"id"`
	Status    approvals.Status `json:"status"`
	CreatedAt int64            `json:"createdAt"`
	ExpiresAt *int64           `json:"expiresAt"`
}

type externalStatusResponse struct {
	ID          string           `json:"id"`
	ExternalID  string           `json:"externalId"`
	Status      approvals.Status `json:"status"`
	CreatedAt   int64            `json:"createdAt"`
	ExpiresAt   *int64           `json:"expiresAt"`
	ProcessedAt *int64           `json:"processedAt"`
}

type approvalResponse struct {
	ID            string           `json:"id"`
	Status        approvals.Status `json:"status"`
	ExternalID    string           `json:"external_id"`
	Context       string           `json:"context"`
	RequesterName string           `json:"requester_name"`
	KeyName       string           `json:"key_name"`
	KeyRevoked    bool             `json:"key_revoked"`
	DecisionMode  string           `json:"decision_mode,omitempty"`
	CreatedAt     int64            `json:"created_at"`
	ExpiresAt     *int64           `json:"expires_at"`
	ProcessedAt   *int64           `json:"processed_at"`
}

type issueKeyPayload struct {
	Name string `json:"name"`
}

func toApprovalResponse(r approvals.Request) approvalResponse {
	return approvalResponse{
		ID:            r.ID,
		Status:        r.Status,
		ExternalID:    r.ExternalID,
		Context:       r.Context,
		RequesterName: r.RequesterName,
		KeyName:       r.CallerKeyName,
		KeyRevoked:    r.KeyRevoked,
		DecisionMode:  r.DecisionMode,
		CreatedAt:     r.CreatedAt,
		ExpiresAt:     r.ExpiresAt,
		ProcessedAt:   r.ProcessedAt,
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeApprovalBody(w http.ResponseWriter, r *http.Request, v any) error {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxApprovalRequestBytes))
	if err != nil {
		return errors.New("request body too large")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errors.New("invalid json body")
	}
	if dec.More() {
		return errors.New("invalid json body")
	}
	return nil
}

func CallerKeyMiddleware(store *approvals.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			raw := strings.TrimPrefix(h, "Bearer ")
			if h == "" || raw == h || raw == "" {
				writeError(w, http.StatusUnauthorized, "caller key required")
				return
			}
			caller, err := store.Authenticate(r.Context(), raw)
			if errors.Is(err, approvals.ErrInvalidKey) {
				writeError(w, http.StatusUnauthorized, "invalid caller key")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "caller key lookup failed")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), callerCtxKey{}, caller)))
		})
	}
}

func callerFromContext(ctx context.Context) (approvals.Caller, bool) {
	c, ok := ctx.Value(callerCtxKey{}).(approvals.Caller)
	return c, ok
}

func CreateExternalApproval(store *approvals.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := callerFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no caller in context")
			return
		}
		var p externalCreatePayload
		if err := decodeApprovalBody(w, r, &p); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if p.ExternalID == nil || strings.TrimSpace(*p.ExternalID) == "" {
			writeError(w, http.StatusBadRequest, "externalId is required")
			return
		}
		if p.Context == nil || strings.TrimSpace(*p.Context) == "" {
			writeError(w, http.StatusBadRequest, "context is required")
			return
		}
		if p.TimeoutSeconds != nil && *p.TimeoutSeconds <= 0 {
			writeError(w, http.StatusBadRequest, "timeoutSeconds must be positive")
			return
		}
		in := approvals.NewRequest{
			ExternalID:     *p.ExternalID,
			Context:        *p.Context,
			TimeoutSeconds: p.TimeoutSeconds,
		}
		if p.RequesterName != nil {
			in.RequesterName = strings.TrimSpace(*p.RequesterName)
		}
		created, err := store.Create(r.Context(), caller, in)
		if errors.Is(err, approvals.ErrDuplicateExternal) {
			writeError(w, http.StatusConflict, "externalId already used by this caller key")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create failed")
			return
		}
		writeJSON(w, http.StatusCreated, externalCreateResponse{
			ID:        created.ID,
			Status:    created.Status,
			CreatedAt: created.CreatedAt,
			ExpiresAt: created.ExpiresAt,
		})
	}
}

func GetExternalApproval(store *approvals.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := callerFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no caller in context")
			return
		}
		got, err := store.GetForCaller(r.Context(), caller, chi.URLParam(r, "id"))
		if errors.Is(err, approvals.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, externalStatusResponse{
			ID:          got.ID,
			ExternalID:  got.ExternalID,
			Status:      got.Status,
			CreatedAt:   got.CreatedAt,
			ExpiresAt:   got.ExpiresAt,
			ProcessedAt: got.ProcessedAt,
		})
	}
}

func ListApprovalKeys(store *approvals.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		keys, err := store.ListKeys(r.Context(), u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"keys": keys})
	}
}

func IssueApprovalKey(store *approvals.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		var p issueKeyPayload
		if err := decodeApprovalBody(w, r, &p); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "name is required")
			return
		}
		issued, err := store.IssueKey(r.Context(), u.ID, name)
		if errors.Is(err, approvals.ErrDuplicateName) {
			writeError(w, http.StatusConflict, "caller name already in use")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "issue failed")
			return
		}
		writeJSON(w, http.StatusCreated, issued)
	}
}

func RevokeApprovalKey(store *approvals.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		id, err := pathID(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		key, err := store.RevokeKey(r.Context(), u.ID, id)
		switch {
		case errors.Is(err, approvals.ErrNotFound):
			writeError(w, http.StatusNotFound, "not found")
		case errors.Is(err, approvals.ErrAlreadyRevoked):
			writeJSON(w, http.StatusConflict, map[string]any{"error": "caller key already revoked", "key": key})
		case err != nil:
			writeError(w, http.StatusInternalServerError, "revoke failed")
		default:
			writeJSON(w, http.StatusOK, key)
		}
	}
}

func ListPendingApprovals(store *approvals.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		pending, err := store.ListPending(r.Context(), u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list failed")
			return
		}
		items := make([]approvalResponse, 0, len(pending))
		for _, p := range pending {
			items = append(items, toApprovalResponse(p))
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func GetApproval(store *approvals.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		got, err := store.Get(r.Context(), u.ID, chi.URLParam(r, "id"))
		if errors.Is(err, approvals.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, toApprovalResponse(got))
	}
}

func DecideApproval(store *approvals.Store, d approvals.Decision) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		got, err := store.Decide(r.Context(), u.ID, chi.URLParam(r, "id"), d)
		switch {
		case errors.Is(err, approvals.ErrNotFound):
			writeError(w, http.StatusNotFound, "not found")
		case errors.Is(err, approvals.ErrNotPending):
			msg := "already processed"
			if got.Status == approvals.StatusExpired {
				msg = "expired"
			}
			writeJSON(w, http.StatusConflict, map[string]any{"error": msg, "approval": toApprovalResponse(got)})
		case err != nil:
			writeError(w, http.StatusInternalServerError, "decide failed")
		default:
			writeJSON(w, http.StatusOK, toApprovalResponse(got))
		}
	}
}
