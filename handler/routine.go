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

// RoutineHandler handles routine management HTTP requests.
type RoutineHandler struct {
	repo *repository.RoutineRepository
}

// NewRoutineHandler constructs a RoutineHandler wired to the given database.
func NewRoutineHandler(db *sql.DB) *RoutineHandler {
	return &RoutineHandler{
		repo: repository.NewRoutineRepository(db),
	}
}

// routineResponse is the JSON shape returned for a single routine.
type routineResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	IntervalDays int    `json:"interval_days"`
	LastDoneAt   string `json:"last_done_at"`
	NextDueDate  string `json:"next_due_date"`
	DDay         int    `json:"d_day"`
}

// toRoutineResponse converts a repository.Routine to routineResponse.
func toRoutineResponse(rt *repository.Routine) routineResponse {
	return routineResponse{
		ID:           rt.ID,
		Name:         rt.Name,
		IntervalDays: rt.IntervalDays,
		LastDoneAt:   rt.LastDoneAt,
		NextDueDate:  rt.NextDueDate,
		DDay:         rt.DDay,
	}
}

// createRoutineRequest is the expected JSON body for POST /routines.
type createRoutineRequest struct {
	Name         string `json:"name"`
	IntervalDays int    `json:"interval_days"`
	LastDoneAt   string `json:"last_done_at"`
}

// updateRoutineRequest is the expected JSON body for PUT /routines/:id.
type updateRoutineRequest struct {
	Name          string `json:"name"`
	IntervalDays  int    `json:"interval_days"`
	Note          string `json:"note"`
	NotifyEnabled *bool  `json:"notify_enabled"`
}

// Create handles POST /routines.
// Returns 201 Created with the new routine, or 400 on validation failure.
func (h *RoutineHandler) Create(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req createRoutineRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}
	if req.IntervalDays <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "interval_days must be positive"})
	}

	rt, err := h.repo.Create(userID, req.Name, req.IntervalDays, req.LastDoneAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, toRoutineResponse(rt))
}

// List handles GET /routines.
// Returns 200 OK with a JSON array of the authenticated user's routines.
func (h *RoutineHandler) List(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	routines, err := h.repo.ListByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve routines"})
	}

	resp := make([]routineResponse, 0, len(routines))
	for _, rt := range routines {
		resp = append(resp, toRoutineResponse(rt))
	}
	return c.JSON(http.StatusOK, resp)
}

// Get handles GET /routines/:id.
// Returns 200 OK with the routine, 404 if not found, 400 for invalid ID format.
func (h *RoutineHandler) Get(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseRoutineID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	rt, err := h.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRoutineNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to retrieve routine"})
	}

	return c.JSON(http.StatusOK, toRoutineResponse(rt))
}

// Update handles PUT /routines/:id.
// Returns 200 OK with updated routine, 404 if not found, 400 for invalid ID.
func (h *RoutineHandler) Update(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseRoutineID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var req updateRoutineRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	updates := repository.RoutineUpdates{
		Name:          req.Name,
		IntervalDays:  req.IntervalDays,
		Note:          req.Note,
		NotifyEnabled: req.NotifyEnabled,
	}

	rt, err := h.repo.Update(id, userID, updates)
	if err != nil {
		if errors.Is(err, repository.ErrRoutineNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update routine"})
	}

	return c.JSON(http.StatusOK, toRoutineResponse(rt))
}

// Delete handles DELETE /routines/:id.
// Returns 204 No Content on success, 404 if not found, 400 for invalid ID.
func (h *RoutineHandler) Delete(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseRoutineID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	if err := h.repo.Delete(id, userID); err != nil {
		if errors.Is(err, repository.ErrRoutineNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete routine"})
	}

	return c.NoContent(http.StatusNoContent)
}

// Complete handles POST /routines/:id/complete.
// Sets last_done_at to today and recalculates next_due_date.
// Returns 200 OK with updated routine, 404 if not found, 400 for invalid ID.
func (h *RoutineHandler) Complete(c echo.Context) error {
	userID, err := extractUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	id, err := parseRoutineID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	rt, err := h.repo.Complete(id, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRoutineNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to complete routine"})
	}

	return c.JSON(http.StatusOK, toRoutineResponse(rt))
}

// parseRoutineID parses the :id URL parameter as int64.
func parseRoutineID(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
