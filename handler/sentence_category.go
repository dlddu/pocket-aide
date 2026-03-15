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

// SentenceCategoryHandler handles sentence category management HTTP requests.
type SentenceCategoryHandler struct {
	repo *repository.SentenceCategoryRepository
}

// NewSentenceCategoryHandler constructs a SentenceCategoryHandler wired to the given database.
func NewSentenceCategoryHandler(db *sql.DB) *SentenceCategoryHandler {
	return &SentenceCategoryHandler{
		repo: repository.NewSentenceCategoryRepository(db),
	}
}

// sentenceCategoryResponse is the JSON shape returned for a single sentence category.
type sentenceCategoryResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// createSentenceCategoryRequest is the expected JSON body for POST /sentences/categories.
type createSentenceCategoryRequest struct {
	Name string `json:"name"`
}

// updateSentenceCategoryRequest is the expected JSON body for PUT /sentences/categories/:id.
type updateSentenceCategoryRequest struct {
	Name string `json:"name"`
}

// Create handles POST /sentences/categories.
// Returns 201 Created with the new category, or 400 on validation failure.
func (h *SentenceCategoryHandler) Create(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req createSentenceCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	cat, err := h.repo.Create(userID, req.Name)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, sentenceCategoryResponse{
		ID:   cat.ID,
		Name: cat.Name,
	})
}

// List handles GET /sentences/categories.
// Returns 200 OK with a JSON array of the authenticated user's sentence categories.
func (h *SentenceCategoryHandler) List(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	categories, err := h.repo.ListByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve sentence categories"})
	}

	resp := make([]sentenceCategoryResponse, 0, len(categories))
	for _, cat := range categories {
		resp = append(resp, sentenceCategoryResponse{
			ID:   cat.ID,
			Name: cat.Name,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// Update handles PUT /sentences/categories/:id.
// Returns 200 OK with updated category, 404 if not found, 400 for invalid ID.
func (h *SentenceCategoryHandler) Update(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var req updateSentenceCategoryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	cat, err := h.repo.Update(id, userID, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrSentenceCategoryNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update sentence category"})
	}

	return c.JSON(http.StatusOK, sentenceCategoryResponse{
		ID:   cat.ID,
		Name: cat.Name,
	})
}

// Delete handles DELETE /sentences/categories/:id.
// Returns 204 No Content on success, 404 if not found, 400 for invalid ID.
func (h *SentenceCategoryHandler) Delete(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	if err := h.repo.Delete(id, userID); err != nil {
		if errors.Is(err, repository.ErrSentenceCategoryNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete sentence category"})
	}

	return c.NoContent(http.StatusNoContent)
}
