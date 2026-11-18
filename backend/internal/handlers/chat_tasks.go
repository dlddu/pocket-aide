package handlers

import (
	"errors"
	"net/http"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
)

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
