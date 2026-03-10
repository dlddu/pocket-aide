// Package middleware provides Echo middleware components.
package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// JWT returns an Echo middleware that validates JWT tokens using the provided secret.
func JWT(secret string) echo.MiddlewareFunc {
	config := echojwt.Config{
		SigningKey: []byte(secret),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return jwt.MapClaims{}
		},
	}
	return echojwt.WithConfig(config)
}
