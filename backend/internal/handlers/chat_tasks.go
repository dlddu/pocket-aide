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

// chatRequestBodyLimit bounds one stored request's body. The slot's own cap is
// smaller still; this only keeps a stray large body from being read.
const chatRequestBodyLimit = 1 << 16

// GetChatTasks handles GET /api/chat/tasks: the work records of the codex
// session answering the caller's own conversation. Read-only — the route never
// restores, writes to or deletes a session, and the session is derived from the
// authenticated user's row rather than taken from the request, so one user
// cannot read another's work.
//
// The three answers the panel has to tell apart are separate statuses: an empty
// list is "no tasks", 501 is "this session keeps no work records", and 409 is
// "the session is not awake, so the list cannot be read" — never an empty list,
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
	// ChoiceTurnID, when set, is the turn whose offered sentences the person
	// picked from. The slot refuses a pick from a turn the task has moved past,
	// so a screen opened before a newer turn arrived cannot act on it. Free
	// text omits it and is never refused for staleness.
	//
	// It replaced the ledger's (choice_id, choice_basis_version) pair: the turn
	// a pick answers is the one fact that decides whether it is current, and a
	// version carried beside a choice id was a second way of saying it that
	// could disagree with the first.
	ChoiceTurnID int64 `json:"choice_turn_id"`
}

// PostChatTaskRequest handles POST /api/chat/tasks/{taskID}/requests: one user
// request left on a task of the caller's own session ledger.
//
// It is not a chat message. Nothing is added to the conversation and no turn of
// the session's own is started. The answer is the turn the request became, and
// that turn being open is what says it is waiting — never that it has been
// dealt with, which only the parent can say by finishing the turn. The session
// is derived from the authenticated user's row, so one user cannot leave a
// request on another's slot.
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
		stored, err := engine.CreateTaskRequest(
			r.Context(), u.ID, chi.URLParam(r, "taskID"), body.ID, body.Text, body.ChoiceTurnID)
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

// GetChatTaskTurns handles GET /api/chat/tasks/{taskID}/turns: one page of a
// task's turns, with the task itself in the agreed shape.
//
// This is what the detail pane reads, and a task's turns are its whole history:
// the ledger's separate history route is gone with the ledger, because a past
// management event was never a turn.
func GetChatTaskTurns(engine *chat.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		var cursor chat.TaskTurnCursor
		limit := 0
		query := r.URL.Query()
		if raw := strings.TrimSpace(query.Get("after_turn_id")); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || parsed < 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "after_turn_id must be a nonnegative number", "code": "task_request_invalid",
				})
				return
			}
			cursor.After = parsed
		}
		// before_turn_id — present even with no value — is the pane's page: the
		// turns before it, newest first, with nothing meaning "from the newest".
		// Naming both ends is refused rather than resolved, because they are
		// opposite ends of the record.
		if query.Has("before_turn_id") {
			if cursor.After != 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "after_turn_id and before_turn_id cannot be used together",
					"code":  "task_request_invalid",
				})
				return
			}
			cursor.Newest = true
			if raw := strings.TrimSpace(query.Get("before_turn_id")); raw != "" {
				parsed, err := strconv.ParseInt(raw, 10, 64)
				if err != nil || parsed < 0 {
					writeJSON(w, http.StatusBadRequest, map[string]string{
						"error": "before_turn_id must be a nonnegative number", "code": "task_request_invalid",
					})
					return
				}
				cursor.Before = parsed
			}
		}
		if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 1 || parsed > chat.TaskTurnPageLimitMax {
				writeJSON(w, http.StatusBadRequest, map[string]string{
					"error": "limit must be a number from 1 to 100", "code": "task_request_invalid",
				})
				return
			}
			limit = parsed
		}
		page, err := engine.TaskTurns(r.Context(), u.ID, chi.URLParam(r, "taskID"), cursor, limit)
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
			case errors.Is(err, chat.ErrTaskTurnsOrderUnsupported):
				// Its own status, because the pane's sentence is its own: the
				// records are readable, just not yet from the end it needs, so
				// trying again later is the move rather than giving up on them.
				writeJSON(w, http.StatusBadGateway, map[string]string{
					"error": err.Error(), "code": "task_turns_order_unsupported",
				})
			default:
				writeChatError(w, err)
			}
			return
		}
		writeJSON(w, http.StatusOK, page)
	}
}
