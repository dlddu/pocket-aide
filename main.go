package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/dlddu/pocket-aide/db"
	"github.com/dlddu/pocket-aide/handler"
	appmiddleware "github.com/dlddu/pocket-aide/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(appmiddleware.RequestLogger())
	e.HTTPErrorHandler = appmiddleware.CustomErrorHandler

	// Database
	database, err := db.New("pocket-aide.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := db.RunMigrations(database, "db/migrations"); err != nil {
		log.Fatal(err)
	}

	// JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}

	// Routes
	health := handler.NewHealthHandler()
	e.GET("/health", health.Health)

	// Auth routes
	authHandler := handler.NewAuthHandler(database, jwtSecret)
	e.POST("/auth/register", authHandler.Register)
	e.POST("/auth/login", authHandler.Login)
	e.GET("/me", authHandler.Me, appmiddleware.JWT(jwtSecret))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(e.Start(":" + port))
}
