package middleware

import (
	"log"
	"time"

	"github.com/labstack/echo/v4"
)

// RequestLogger returns an Echo middleware that logs incoming HTTP requests and their responses.
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status
			log.Printf("%s %s %d %s", req.Method, req.RequestURI, status, duration)

			return err
		}
	}
}
