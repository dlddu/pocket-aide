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

// SentenceHandler handles sentence management HTTP requests.
type SentenceHandler struct {
	repo *repository.SentenceRepository
}

// NewSentenceHandler constructs a SentenceHandler wired to the given database.
func NewSentenceHandler(db *sql.DB) *SentenceHandler {
	return &SentenceHandler{
		repo: repository.NewSentenceRepository(db),
	}
}

// sentenceResponse is the JSON shape returned for a single sentence.
type sentenceResponse struct {
	ID         int64  `json:"id"`
	Content    string `json:"content"`
	CategoryID int64  `json:"category_id"`
}

// createSentenceRequest is the expected JSON body for POST /sentences.
type createSentenceRequest struct {
	Content    string `json:"content"`
	CategoryID int64  `json:"category_id"`
}

// updateSentenceRequest is the expected JSON body for PUT /sentences/:id.
type updateSentenceRequest struct {
	Content string `json:"content"`
}

// Create handles POST /sentences.
// Returns 201 Created with the new sentence, or 400 on validation failure.
func (h *SentenceHandler) Create(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req createSentenceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content is required"})
	}

	s, err := h.repo.Create(userID, req.CategoryID, req.Content)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, sentenceResponse{
		ID:         s.ID,
		Content:    s.Content,
		CategoryID: s.CategoryID,
	})
}

// List handles GET /sentences.
// Supports optional ?category_id=:id query parameter for filtering.
// Returns 200 OK with a JSON array of the authenticated user's sentences.
func (h *SentenceHandler) List(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var sentences []*repository.Sentence

	categoryIDStr := c.QueryParam("category_id")
	if categoryIDStr != "" {
		categoryID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid category_id"})
		}
		sentences, err = h.repo.ListByCategoryID(userID, categoryID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve sentences"})
		}
	} else {
		sentences, err = h.repo.ListByUserID(userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve sentences"})
		}
	}

	resp := make([]sentenceResponse, 0, len(sentences))
	for _, s := range sentences {
		resp = append(resp, sentenceResponse{
			ID:         s.ID,
			Content:    s.Content,
			CategoryID: s.CategoryID,
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// Update handles PUT /sentences/:id.
// Returns 200 OK with updated sentence, 404 if not found, 400 for invalid ID.
func (h *SentenceHandler) Update(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var req updateSentenceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "content is required"})
	}

	s, err := h.repo.Update(id, userID, req.Content)
	if err != nil {
		if errors.Is(err, repository.ErrSentenceNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update sentence"})
	}

	return c.JSON(http.StatusOK, sentenceResponse{
		ID:         s.ID,
		Content:    s.Content,
		CategoryID: s.CategoryID,
	})
}

// Delete handles DELETE /sentences/:id.
// Returns 204 No Content on success, 404 if not found, 400 for invalid ID.
func (h *SentenceHandler) Delete(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	if err := h.repo.Delete(id, userID); err != nil {
		if errors.Is(err, repository.ErrSentenceNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete sentence"})
	}

	return c.NoContent(http.StatusNoContent)
}
