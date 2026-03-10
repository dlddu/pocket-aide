// Package handler contains Echo HTTP handler implementations.
package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check requests.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health responds with a JSON payload indicating the service is healthy.
func (h *HealthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
