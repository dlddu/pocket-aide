package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/chat"
	"github.com/go-chi/chi/v5"
)

// PRTitles is where a pull request's recorded title comes from: this user's own
// CI notification history, which already stores a repository, a number and the
// title that came with them.
//
// It is an interface so this route depends on "something that can name a pull
// request this user has already been told about" rather than on the history
// store itself — and so a deployment without one answers titles as unknown
// instead of failing the review answer.
type PRTitles interface {
	PRTitle(ctx context.Context, userID int64, repoFullName string, prNumber int) (string, error)
}

// GetChatTaskMaterials handles GET /api/chat/tasks/{taskID}/materials: what a
// person needs in order to review one task (PRD-AI채팅 AC10, 검토 자료).
//
// Read-only, one task, and asked for when somebody opens that task. It is
// deliberately not folded into the task list: the list is polled while the
// panel is on screen, so gathering this per row would be one history read per
// task per poll.
//
// Titles are added here rather than in the projection because they come from
// this server's own tables while the materials come from the session's records.
// A title is only ever attached to a pull request whose repository *and* number
// a row already names; a lookup that found nothing leaves the identifier the
// records gave, which is the honest answer and still openable.
func GetChatTaskMaterials(engine *chat.Engine, titles PRTitles) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.FromContext(r.Context())
		if !ok {
			http.Error(w, "no user in context", http.StatusInternalServerError)
			return
		}
		materials, err := engine.TaskMaterials(r.Context(), u.ID, chi.URLParam(r, "taskID"))
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
		addPRTitles(r.Context(), titles, u.ID, materials.Artifacts)
		writeJSON(w, http.StatusOK, materials)
	}
}

// addPRTitles fills in the titles that are recorded, leaving the rest as they
// are. A lookup that fails is not an error for the review: the materials are
// the session's records, and a missing title costs a nicer line on screen
// rather than the pull request itself.
func addPRTitles(ctx context.Context, titles PRTitles, userID int64, artifacts []chat.TaskArtifact) {
	if titles == nil {
		return
	}
	// One task names few artifacts (the slot caps a turn at 20), and the same
	// pull request on several turns is already one entry by the time it gets
	// here, so this is a handful of indexed single-row reads.
	known := map[string]string{}
	for i := range artifacts {
		a := &artifacts[i]
		if a.Repo == "" || a.Number <= 0 {
			continue
		}
		key := a.Repo + "#" + strconv.Itoa(a.Number)
		if held, ok := known[key]; ok {
			a.Title = held
			continue
		}
		title, err := titles.PRTitle(ctx, userID, a.Repo, a.Number)
		if err != nil {
			continue
		}
		known[key] = title
		a.Title = title
	}
}
