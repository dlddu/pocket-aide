// Command server is the pocket-aide backend HTTP server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
)

func main() {
	cfg := loadConfig()

	conn, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer func() { _ = conn.Close() }()

	bootstrapCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	verifier, err := auth.NewVerifier(bootstrapCtx, cfg.OIDCIssuer, cfg.OIDCAudience)
	if err != nil {
		log.Fatalf("oidc verifier: %v", err)
	}

	authCfg := handlers.AuthConfig{
		Issuer:      cfg.OIDCIssuer,
		ClientID:    cfg.OIDCClientID,
		RedirectURI: cfg.OIDCRedirectURI,
		Audience:    cfg.OIDCAudience,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", handlers.Health(conn))

	r.Group(func(g chi.Router) {
		g.Use(middleware.Logger)
		g.Get("/api/auth/config", handlers.AuthConfigHandler(authCfg))

		g.Group(func(p chi.Router) {
			p.Use(auth.Middleware(verifier, conn))
			p.Get("/api/me", handlers.Me())
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("pocket-aide listening on :%s (issuer=%s)", cfg.Port, cfg.OIDCIssuer)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
	shutdownCtx, sc := context.WithTimeout(context.Background(), 10*time.Second)
	defer sc()
	_ = srv.Shutdown(shutdownCtx)
}

type config struct {
	Port            string
	DatabasePath    string
	OIDCIssuer      string
	OIDCAudience    string
	OIDCClientID    string
	OIDCRedirectURI string
}

func loadConfig() config {
	return config{
		Port:            envOr("PORT", "8080"),
		DatabasePath:    envOr("DATABASE_PATH", "/data/pocket-aide.db"),
		OIDCIssuer:      mustEnv("OIDC_ISSUER"),
		OIDCAudience:    mustEnv("OIDC_AUDIENCE"),
		OIDCClientID:    mustEnv("OIDC_CLIENT_ID"),
		OIDCRedirectURI: mustEnv("OIDC_REDIRECT_URI"),
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("required env var %s is not set", k)
	}
	return v
}
