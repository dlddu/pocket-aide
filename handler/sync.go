// Package handler contains Echo HTTP handler implementations.
package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// SyncHandler handles data synchronisation HTTP requests.
type SyncHandler struct {
	db *sql.DB
}

// NewSyncHandler constructs a SyncHandler wired to the given database.
func NewSyncHandler(db *sql.DB) *SyncHandler {
	return &SyncHandler{db: db}
}

// syncChangeEntity represents the entity type of a sync change.
type syncChangeEntity string

const (
	syncEntityTodo syncChangeEntity = "todo"
	syncEntityMemo syncChangeEntity = "memo"
)

// syncOperation represents the operation type of a sync change.
type syncOperation string

const (
	syncOperationCreate syncOperation = "create"
	syncOperationUpdate syncOperation = "update"
	syncOperationDelete syncOperation = "delete"
)

// syncChange represents a single client-side change.
type syncChange struct {
	Entity    syncChangeEntity       `json:"entity"`
	ID        string                 `json:"id"`
	Operation syncOperation          `json:"operation"`
	Payload   map[string]interface{} `json:"payload"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// syncRequest is the JSON body for POST /sync.
type syncRequest struct {
	Changes []syncChange `json:"changes"`
}

// syncTodoItem is a single todo in the server_data response.
type syncTodoItem struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Type      string  `json:"type"`
	UpdatedAt string  `json:"updated_at"`
}

// syncMemoItem is a single memo in the server_data response.
type syncMemoItem struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	Source    string `json:"source"`
	UpdatedAt string `json:"updated_at"`
}

// syncRoutineItem is a single routine in the server_data response.
type syncRoutineItem struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	IntervalDays int64  `json:"interval_days"`
	LastDoneAt   string `json:"last_done_at"`
	UpdatedAt    string `json:"updated_at"`
}

// syncServerData holds the full server-side state returned after sync.
type syncServerData struct {
	Todos    []syncTodoItem    `json:"todos"`
	Memos    []syncMemoItem    `json:"memos"`
	Routines []syncRoutineItem `json:"routines"`
}

// syncResponse is the JSON body returned by POST /sync.
type syncHandlerResponse struct {
	ServerData syncServerData `json:"server_data"`
}

// Sync handles POST /sync.
// It applies client changes using last-write-wins conflict resolution and
// returns the full server-side state for the authenticated user.
func (h *SyncHandler) Sync(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req syncRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Process each change using last-write-wins.
	for _, change := range req.Changes {
		switch change.Entity {
		case syncEntityTodo:
			if err := h.applyTodoChange(userID, change); err != nil {
				// Log and continue — one failure should not abort the whole sync.
				_ = err
			}
		case syncEntityMemo:
			if err := h.applyMemoChange(userID, change); err != nil {
				_ = err
			}
		}
	}

	// Fetch full server-side state for the user.
	serverData, err := h.fetchServerData(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch server data"})
	}

	return c.JSON(http.StatusOK, syncHandlerResponse{ServerData: *serverData})
}

// applyTodoChange processes a single todo change from the client.
func (h *SyncHandler) applyTodoChange(userID int64, change syncChange) error {
	clientUpdatedAt := change.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")

	switch change.Operation {
	case syncOperationCreate:
		title := stringPayload(change.Payload, "title")
		todoType := stringPayload(change.Payload, "type")
		if todoType == "" {
			todoType = "personal"
		}
		_, err := h.db.Exec(
			`INSERT INTO todos (user_id, title, type, updated_at)
			 VALUES (?, ?, ?, ?)`,
			userID, title, todoType, clientUpdatedAt,
		)
		return err

	case syncOperationUpdate:
		title := stringPayload(change.Payload, "title")
		todoType := stringPayload(change.Payload, "type")
		// LWW: only update when server record is older than the client change.
		_, err := h.db.Exec(
			`UPDATE todos
			 SET title = ?, type = ?, updated_at = ?
			 WHERE user_id = ? AND updated_at < ?`,
			title, todoType, clientUpdatedAt, userID, clientUpdatedAt,
		)
		return err

	case syncOperationDelete:
		_, err := h.db.Exec(
			`DELETE FROM todos WHERE user_id = ?`,
			userID,
		)
		return err
	}
	return nil
}

// applyMemoChange processes a single memo change from the client.
func (h *SyncHandler) applyMemoChange(userID int64, change syncChange) error {
	clientUpdatedAt := change.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")

	switch change.Operation {
	case syncOperationCreate:
		content := stringPayload(change.Payload, "content")
		source := stringPayload(change.Payload, "source")
		if source == "" {
			source = "text"
		}
		_, err := h.db.Exec(
			`INSERT INTO memos (user_id, content, source, updated_at)
			 VALUES (?, ?, ?, ?)`,
			userID, content, source, clientUpdatedAt,
		)
		return err

	case syncOperationUpdate:
		content := stringPayload(change.Payload, "content")
		source := stringPayload(change.Payload, "source")
		// LWW: only update when server record is older than the client change.
		_, err := h.db.Exec(
			`UPDATE memos
			 SET content = ?, source = ?, updated_at = ?
			 WHERE user_id = ? AND updated_at < ?`,
			content, source, clientUpdatedAt, userID, clientUpdatedAt,
		)
		return err

	case syncOperationDelete:
		_, err := h.db.Exec(
			`DELETE FROM memos WHERE user_id = ?`,
			userID,
		)
		return err
	}
	return nil
}

// fetchServerData retrieves all data for the given user to build server_data.
func (h *SyncHandler) fetchServerData(userID int64) (*syncServerData, error) {
	todos, err := h.fetchTodos(userID)
	if err != nil {
		return nil, err
	}

	memos, err := h.fetchMemos(userID)
	if err != nil {
		return nil, err
	}

	routines, err := h.fetchRoutines(userID)
	if err != nil {
		return nil, err
	}

	return &syncServerData{
		Todos:    todos,
		Memos:    memos,
		Routines: routines,
	}, nil
}

// fetchTodos returns all todos for the user.
func (h *SyncHandler) fetchTodos(userID int64) ([]syncTodoItem, error) {
	rows, err := h.db.Query(
		`SELECT id, title, type, updated_at FROM todos WHERE user_id = ? ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]syncTodoItem, 0)
	for rows.Next() {
		var item syncTodoItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// fetchMemos returns all memos for the user.
func (h *SyncHandler) fetchMemos(userID int64) ([]syncMemoItem, error) {
	rows, err := h.db.Query(
		`SELECT id, content, source, updated_at FROM memos WHERE user_id = ? ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]syncMemoItem, 0)
	for rows.Next() {
		var item syncMemoItem
		if err := rows.Scan(&item.ID, &item.Content, &item.Source, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// fetchRoutines returns all routines for the user.
func (h *SyncHandler) fetchRoutines(userID int64) ([]syncRoutineItem, error) {
	rows, err := h.db.Query(
		`SELECT id, name, interval_days, last_done_at, updated_at
		 FROM routines WHERE user_id = ? ORDER BY id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]syncRoutineItem, 0)
	for rows.Next() {
		var item syncRoutineItem
		if err := rows.Scan(&item.ID, &item.Name, &item.IntervalDays, &item.LastDoneAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// stringPayload extracts a string value from a payload map by key.
// Returns empty string if key does not exist or value is not a string.
func stringPayload(payload map[string]interface{}, key string) string {
	if payload == nil {
		return ""
	}
	v, ok := payload[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
