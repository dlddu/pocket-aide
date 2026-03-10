package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// CustomErrorHandler is an Echo HTTPErrorHandler that returns structured JSON error responses.
func CustomErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	if httpErr, ok := err.(*echo.HTTPError); ok {
		code = httpErr.Code
		if msg, ok := httpErr.Message.(string); ok {
			message = msg
		} else {
			message = http.StatusText(code)
		}
	}

	resp := errorResponse{
		Message: message,
		Code:    code,
	}

	if !c.Response().Committed {
		c.JSON(code, resp) //nolint:errcheck
	}
}
