package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/handler"
	appmiddleware "github.com/dlddu/pocket-aide/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(appmiddleware.RequestLogger())
	e.HTTPErrorHandler = appmiddleware.CustomErrorHandler

	// Routes
	health := handler.NewHealthHandler()
	e.GET("/health", health.Health)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(e.Start(":" + port))
}
