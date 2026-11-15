package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/commitstatus"
	"github.com/dlddu/pocket-aide/backend/internal/commitstatuspush"
)

const commitStatusHistoryLimit = 30

type externalCommitStatusPayload struct {
	Owner       *string `json:"owner"`
	Repo        *string `json:"repo"`
	SHA         *string `json:"sha"`
	State       *string `json:"state"`
	Context     *string `json:"context"`
	Description *string `json:"description"`
	TargetURL   *string `json:"targetUrl"`
	Requester   *string `json:"requester"`
	Reason      *string `json:"reason"`
	SessionID   *string `json:"sessionId"`
}

type externalCommitStatusResponse struct {
	ID        string              `json:"id"`
	Status    commitstatus.Status `json:"status"`
	CreatedAt int64               `json:"createdAt"`
}

type CommitStatusPusher interface {
	Notify(req commitstatuspush.Request)
}

// CreateExternalCommitStatusRequest handles POST /api/external/commit-status-requests: it stores the request and returns at once; the status is written only when the owner approves (PRD-12 AC18).
func CreateExternalCommitStatusRequest(store *commitstatus.Store, push CommitStatusPusher, names SessionNamer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		caller, ok := callerFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no caller in context")
			return
		}
		var p externalCommitStatusPayload
		if err := decodeApprovalBody(w, r, &p); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		in := commitstatus.NewRequest{
			Owner:       trimmedOptional(p.Owner),
			Repo:        trimmedOptional(p.Repo),
			SHA:         trimmedOptional(p.SHA),
			State:       trimmedOptional(p.State),
			Context:     trimmedOptional(p.Context),
			Description: trimmedOptional(p.Description),
			TargetURL:   trimmedOptional(p.TargetURL),
			Requester:   trimmedOptional(p.Requester),
			Reason:      trimmedOptional(p.Reason),
			SessionID:   trimmedOptional(p.SessionID),
		}
		if err := commitstatus.Validate(in); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		in.SessionName = lookupSessionName(r.Context(), names, in.SessionID)
		created, err := store.Create(r.Context(), caller, in)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "create failed")
			return
		}
		writeJSON(w, http.StatusCreated, externalCommitStatusResponse{ID: created.ID, Status: created.Status, CreatedAt: created.CreatedAt})
		if push != nil {
			push.Notify(commitstatuspush.Request{
				UserID:    caller.UserID,
				KeyName:   caller.Name,
				RequestID: created.ID,
				Owner:     created.Owner,
				Repo:      created.Repo,
				SHA:       created.SHA,
				State:     created.State,
				Context:   created.Context,
				Requester: created.Requester,
			})
		}
	}
}

// ListCommitStatusRequests handles GET /api/commit-status-requests: pending requests plus the most recent decisions (PRD-12 AC19).
func ListCommitStatusRequests(store *commitstatus.Store) http.HandlerFunc {
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
		history, err := store.ListHistory(r.Context(), u.ID, commitStatusHistoryLimit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": pending, "history": history})
	}
}

// GetCommitStatusRequest handles GET /api/commit-status-requests/{id}: the detail a push tap opens.
func GetCommitStatusRequest(store *commitstatus.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		got, err := store.Get(r.Context(), u.ID, chi.URLParam(r, "id"))
		if errors.Is(err, commitstatus.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, got)
	}
}

type commitStatusDecide func(ctx context.Context, userID int64, id string) (commitstatus.Request, error)

// ApproveCommitStatusRequest handles POST /api/commit-status-requests/{id}/approve: it writes the status to GitHub, then records the approval (PRD-12 AC20).
func ApproveCommitStatusRequest(d *commitstatus.Decider) http.HandlerFunc {
	return decideCommitStatus(d.Approve)
}

// RejectCommitStatusRequest handles POST /api/commit-status-requests/{id}/reject.
func RejectCommitStatusRequest(d *commitstatus.Decider) http.HandlerFunc {
	return decideCommitStatus(d.Reject)
}

func decideCommitStatus(decide commitStatusDecide) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			writeError(w, http.StatusInternalServerError, "no user in context")
			return
		}
		got, err := decide(r.Context(), u.ID, chi.URLParam(r, "id"))
		var writeErr *commitstatus.WriteError
		switch {
		case err == nil:
			writeJSON(w, http.StatusOK, got)
		case errors.Is(err, commitstatus.ErrNotFound):
			writeError(w, http.StatusNotFound, "not found")
		case errors.Is(err, commitstatus.ErrNotPending):
			writeJSON(w, http.StatusConflict, map[string]any{"error": "already processed", "request": got})
		case errors.Is(err, commitstatus.ErrWriterUnavailable):
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error(), "request": got})
		case errors.As(err, &writeErr):
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": strings.TrimSpace(writeErr.Message), "request": writeErr.Request})
		default:
			writeError(w, http.StatusInternalServerError, "decide failed")
		}
	}
}
