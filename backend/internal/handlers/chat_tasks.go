package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/go-chi/chi/v5"
)

// chatRequestBodyLimit bounds one stored request's body. The ledger's own cap
// is smaller still; this only keeps a stray large body from being read.
const chatRequestBodyLimit = 1 << 16

// GetChatTasks handles GET /api/chat/tasks: the task ledger of the codex
// session answering the caller's own conversation. Read-only — the route never
// restores, writes to or deletes a session, and the session is derived from the
// authenticated user's row rather than taken from the request, so one user
// cannot read another's ledger.
//
// The three answers the panel has to tell apart are separate statuses: an empty
// list is "no tasks", 501 is "this session keeps no ledger", and 409 is "the
// session is not awake, so the list cannot be read" — never an empty list,
// which would claim there is no work.
func GetChatTasks(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		tasks, err := engine.Tasks(r.Context(), u.ID)
		if err != nil {
			switch {
			case errors.Is(err, chat.ErrTasksUnsupported):
				writeJSON(w, http.StatusNotImplemented, map[string]string{
					"error": chat.ErrTasksUnsupported.Error(),
					"code":  "task_ledger_unsupported",
				})
			case errors.Is(err, chat.ErrTasksUnavailable):
				writeJSON(w, http.StatusConflict, map[string]string{
					"error": chat.ErrTasksUnavailable.Error(),
					"code":  "task_ledger_unavailable",
				})
			default:
				writeChatError(w, err)
			}
			return
		}
		writeJSON(w, http.StatusOK, tasks)
	}
}

// postChatTaskRequestBody is what the app sends: the words the person typed and
// the id that makes a retry one request. The task is in the path, so the body
// cannot aim the request at another task, and there is no author field — the
// server derives who asked from the authenticated user.
type postChatTaskRequestBody struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	// ChoiceID, when set, is the session's own offered choice the person
	// picked; empty means they typed their own words.
	ChoiceID string `json:"choice_id"`
	// ChoiceBasisVersion is the task version the screen saw that choice on.
	// The slot checks it, so a pick from a screen opened before the task moved
	// is refused rather than stored.
	ChoiceBasisVersion int64 `json:"choice_basis_version"`
}

// PostChatTaskRequest handles POST /api/chat/tasks/{taskID}/requests: one user
// request left on a task of the caller's own session ledger.
//
// It is not a chat message. Nothing is added to the conversation, no turn is
// started, and the answer says the request is *stored and waiting* — never that
// it has been dealt with, which only the session can say by settling it. The
// session is derived from the authenticated user's row, so one user cannot
// leave a request on another's ledger.
func PostChatTaskRequest(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var body postChatTaskRequestBody
		if err := json.NewDecoder(io.LimitReader(r.Body, chatRequestBodyLimit)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "body must be {id, text}", "code": "task_request_invalid",
			})
			return
		}
		stored, err := engine.CreateTaskRequest(r.Context(), u.ID, chi.URLParam(r, "taskID"), body.ID, body.Text, body.ChoiceID, body.ChoiceBasisVersion)
		if err != nil {
			switch {
			case errors.Is(err, chat.ErrTaskRequestInvalid):
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": err.Error(), "code": "task_request_invalid",
				})
			case errors.Is(err, chat.ErrTaskNotFound), errors.Is(err, chat.ErrNoConversation):
				writeJSON(w, http.StatusNotFound, map[string]string{
					"error": chat.ErrTaskNotFound.Error(), "code": "task_not_found",
				})
			case errors.Is(err, chat.ErrTaskRequestStaleChoice):
				// Its own code so the app can keep the draft and re-read the
				// task instead of telling the person their words clashed.
				writeJSON(w, http.StatusConflict, map[string]string{
					"error": err.Error(), "code": "task_request_stale_choice",
				})
			case errors.Is(err, chat.ErrTaskRequestConflict):
				writeJSON(w, http.StatusConflict, map[string]string{
					"error": err.Error(), "code": "task_request_conflict",
				})
			case errors.Is(err, chat.ErrTaskRequestUnavailable):
				writeJSON(w, http.StatusConflict, map[string]string{
					"error": err.Error(), "code": "task_request_unavailable",
				})
			case errors.Is(err, chat.ErrTaskRequestUnsupported):
				writeJSON(w, http.StatusNotImplemented, map[string]string{
					"error": err.Error(), "code": "task_request_unsupported",
				})
			default:
				writeChatError(w, err)
			}
			return
		}
		// 200 for a request this id already stored, 201 for a first store: the
		// app retrying an answer it never saw learns which happened without
		// comparing rows.
		status := http.StatusCreated
		if stored.Replayed {
			status = http.StatusOK
		}
		writeJSON(w, status, stored)
	}
}

// GetChatTaskEvents handles GET /api/chat/tasks/{taskID}/events: one page of
// that task's history, oldest first.
//
// Read-only and scoped to the caller's own session, like the task list. A task
// with nothing recorded answers an empty page; a session that keeps no history
// says so with 501 rather than pretending the task has none.
func GetChatTaskEvents(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		after, limit := int64(0), 0
		if raw := strings.TrimSpace(r.URL.Query().Get("after_seq")); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || parsed < 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "after_seq must be a nonnegative number", "code": "task_request_invalid",
				})
				return
			}
			after = parsed
		}
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > chat.TaskEventPageLimitMax {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "limit must be a number from 1 to 100", "code": "task_request_invalid",
				})
				return
			}
			limit = parsed
		}
		page, err := engine.TaskEvents(r.Context(), u.ID, chi.URLParam(r, "taskID"), after, limit)
		if err != nil {
			switch {
			case errors.Is(err, chat.ErrTaskRequestInvalid):
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error(), "code": "task_request_invalid"})
			case errors.Is(err, chat.ErrTaskNotFound), errors.Is(err, chat.ErrNoConversation):
				writeJSON(w, http.StatusNotFound, map[string]string{"error": chat.ErrTaskNotFound.Error(), "code": "task_not_found"})
			case errors.Is(err, chat.ErrTaskRequestUnavailable):
				writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error(), "code": "task_request_unavailable"})
			case errors.Is(err, chat.ErrTaskRequestUnsupported):
				writeJSON(w, http.StatusNotImplemented, map[string]string{"error": err.Error(), "code": "task_request_unsupported"})
			default:
				writeChatError(w, err)
			}
			return
		}
		writeJSON(w, http.StatusOK, page)
	}
}
