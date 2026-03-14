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

// TodoHandler handles todo management HTTP requests.
type TodoHandler struct {
	repo *repository.TodoRepository
}

// NewTodoHandler constructs a TodoHandler wired to the given database.
func NewTodoHandler(db *sql.DB) *TodoHandler {
	return &TodoHandler{
		repo: repository.NewTodoRepository(db),
	}
}

// todoResponse is the JSON shape returned for a single todo.
type todoResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Type        string  `json:"type"`
	CompletedAt *string `json:"completed_at"`
}

// toTodoResponse converts a repository.Todo to todoResponse.
func toTodoResponse(td *repository.Todo) todoResponse {
	return todoResponse{
		ID:          td.ID,
		Title:       td.Title,
		Type:        td.Type,
		CompletedAt: td.CompletedAt,
	}
}

// createTodoRequest is the expected JSON body for POST /todos.
type createTodoRequest struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}

// updateTodoRequest is the expected JSON body for PUT /todos/:id.
type updateTodoRequest struct {
	Title string `json:"title"`
}

// Create handles POST /todos.
// Returns 201 Created with the new todo, or 400 on validation failure.
func (h *TodoHandler) Create(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req createTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title is required"})
	}

	td, err := h.repo.Create(userID, req.Title, req.Type)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, toTodoResponse(td))
}

// List handles GET /todos.
// Supports ?type=personal query parameter filtering.
// Returns 200 OK with a JSON array of the authenticated user's todos.
func (h *TodoHandler) List(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	todoType := c.QueryParam("type")

	todos, err := h.repo.ListByUserIDAndType(userID, todoType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve todos"})
	}

	resp := make([]todoResponse, 0, len(todos))
	for _, td := range todos {
		resp = append(resp, toTodoResponse(td))
	}
	return c.JSON(http.StatusOK, resp)
}

// Get handles GET /todos/:id.
// Returns 200 OK with the todo, 404 if not found, 400 for invalid ID format.
func (h *TodoHandler) Get(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseTodoID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	td, err := h.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTodoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve todo"})
	}

	return c.JSON(http.StatusOK, toTodoResponse(td))
}

// Update handles PUT /todos/:id.
// Returns 200 OK with updated todo, 404 if not found, 400 for invalid ID.
func (h *TodoHandler) Update(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseTodoID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var req updateTodoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	updates := repository.TodoUpdates{
		Title: req.Title,
	}

	td, err := h.repo.Update(id, userID, updates)
	if err != nil {
		if errors.Is(err, repository.ErrTodoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update todo"})
	}

	return c.JSON(http.StatusOK, toTodoResponse(td))
}

// Delete handles DELETE /todos/:id.
// Returns 204 No Content on success, 404 if not found, 400 for invalid ID.
func (h *TodoHandler) Delete(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseTodoID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	if err := h.repo.Delete(id, userID); err != nil {
		if errors.Is(err, repository.ErrTodoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete todo"})
	}

	return c.NoContent(http.StatusNoContent)
}

// Toggle handles POST /todos/:id/toggle.
// Switches completed_at between null and current timestamp.
// Returns 200 OK with updated todo, 404 if not found, 400 for invalid ID.
func (h *TodoHandler) Toggle(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseTodoID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	td, err := h.repo.Toggle(id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrTodoNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to toggle todo"})
	}

	return c.JSON(http.StatusOK, toTodoResponse(td))
}

// parseTodoID parses the :id URL parameter as int64.
func parseTodoID(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
