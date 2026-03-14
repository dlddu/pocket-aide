// Package handler contains Echo HTTP handler implementations.
package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/repository"
)

// MemoHandler handles memo management HTTP requests.
type MemoHandler struct {
	repo     *repository.MemoRepository
	todoRepo *repository.TodoRepository
}

// NewMemoHandler constructs a MemoHandler wired to the given database.
func NewMemoHandler(db *sql.DB) *MemoHandler {
	return &MemoHandler{
		repo:     repository.NewMemoRepository(db),
		todoRepo: repository.NewTodoRepository(db),
	}
}

// memoResponse is the JSON shape returned for a single memo.
type memoResponse struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

// createMemoRequest is the expected JSON body for POST /memos.
type createMemoRequest struct {
	Content string `json:"content"`
	Source  string `json:"source"`
}

// updateMemoRequest is the expected JSON body for PUT /memos/:id.
type updateMemoRequest struct {
	Content string `json:"content"`
}

// moveRequest is the expected JSON body for POST /memos/:id/move.
type moveRequest struct {
	Target string `json:"target"`
}

// Create handles POST /memos.
// Returns 201 Created with the new memo, or 400 on validation failure.
func (h *MemoHandler) Create(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req createMemoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content is required"})
	}

	m, err := h.repo.Create(userID, req.Content, req.Source)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, memoResponse{
		ID:      m.ID,
		Content: m.Content,
		Source:  m.Source,
	})
}

// List handles GET /memos.
// Returns 200 OK with a JSON array of the authenticated user's memos.
func (h *MemoHandler) List(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	memos, err := h.repo.ListByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve memos"})
	}

	resp := make([]memoResponse, 0, len(memos))
	for _, m := range memos {
		resp = append(resp, memoResponse{
			ID:      m.ID,
			Content: m.Content,
			Source:  m.Source,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// Update handles PUT /memos/:id.
// Returns 200 OK with updated memo, 404 if not found, 400 for invalid ID.
func (h *MemoHandler) Update(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseMemoID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var req updateMemoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	m, err := h.repo.Update(id, userID, req.Content)
	if err != nil {
		if errors.Is(err, repository.ErrMemoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update memo"})
	}

	return c.JSON(http.StatusOK, memoResponse{
		ID:      m.ID,
		Content: m.Content,
		Source:  m.Source,
	})
}

// Delete handles DELETE /memos/:id.
// Returns 204 No Content on success, 404 if not found, 400 for invalid ID.
func (h *MemoHandler) Delete(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseMemoID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	if err := h.repo.Delete(id, userID); err != nil {
		if errors.Is(err, repository.ErrMemoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete memo"})
	}

	return c.NoContent(http.StatusNoContent)
}

// Move handles POST /memos/:id/move.
// Converts the memo to a todo by target type and deletes the original memo.
// Returns 200 OK on success, 404 if memo not found, 400 for invalid ID.
func (h *MemoHandler) Move(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseMemoID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var req moveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	memo, err := h.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrMemoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to find memo"})
	}

	// Map target to todo type
	todoType := "personal"
	switch req.Target {
	case "personal_todo":
		todoType = "personal"
	case "work_todo":
		todoType = "work"
	}

	_, err = h.todoRepo.Create(userID, memo.Content, todoType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create todo"})
	}

	if err := h.repo.Delete(id, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete memo"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "memo moved successfully"})
}

// parseMemoID parses the :id URL parameter as int64.
func parseMemoID(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
