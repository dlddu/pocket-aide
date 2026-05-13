// Command server is the pocket-aide backend HTTP server.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dlddu/pocket-aide/backend/internal/affirmations"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/oidcmock"
)

func main() {
	cfg, mock, mockSrv := loadConfig()
	if mockSrv != nil {
		defer func() { _ = mockSrv.Close() }()
	}

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
	if mock != nil {
		authCfg.DevAuthTokenPath = "/dev/auth-token"
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", handlers.Health(conn))
	r.Get("/api/auth/config", handlers.AuthConfigHandler(authCfg))

	if mock != nil {
		r.Post("/dev/auth-token", devAuthTokenHandler(mock))
		log.Println("dev mode: /dev/auth-token enabled — DO NOT use in production")
	}

	affStore := affirmations.NewStore(conn)

	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(verifier, conn))
		p.Get("/api/me", handlers.Me())

		p.Get("/api/affirmations", handlers.ListAffirmations(affStore))
		p.Post("/api/affirmations", handlers.CreateAffirmation(affStore))
		p.Patch("/api/affirmations/{id}", handlers.UpdateAffirmation(affStore))
		p.Delete("/api/affirmations/{id}", handlers.DeleteAffirmation(affStore))
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

// loadConfig reads configuration from the environment. When POCKET_AIDE_DEV=1
// it spins up an in-process oidcmock IdP on a side listener and routes the
// main server's OIDC config to it so dev/UI-test clients can mint real
// RS256-signed tokens that the auth middleware will verify.
func loadConfig() (config, *oidcmock.Server, *http.Server) {
	if os.Getenv("POCKET_AIDE_DEV") == "1" {
		return loadDevConfig()
	}
	cfg := config{
		Port:            envOr("PORT", "8080"),
		DatabasePath:    envOr("DATABASE_PATH", "/data/pocket-aide.db"),
		OIDCIssuer:      mustEnv("OIDC_ISSUER"),
		OIDCAudience:    mustEnv("OIDC_AUDIENCE"),
		OIDCClientID:    mustEnv("OIDC_CLIENT_ID"),
		OIDCRedirectURI: mustEnv("OIDC_REDIRECT_URI"),
	}
	return cfg, nil, nil
}

func loadDevConfig() (config, *oidcmock.Server, *http.Server) {
	mock, err := oidcmock.New(oidcmock.Options{})
	if err != nil {
		log.Fatalf("dev: oidcmock new: %v", err)
	}
	listener, err := net.Listen("tcp", envOr("OIDC_MOCK_ADDR", "127.0.0.1:0"))
	if err != nil {
		log.Fatalf("dev: oidcmock listen: %v", err)
	}
	issuer := "http://" + listener.Addr().String()
	mock.SetIssuer(issuer)

	srv := &http.Server{
		Handler:           mock.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("dev: oidcmock serve: %v", err)
		}
	}()
	log.Printf("dev: oidcmock listening on %s", issuer)

	cfg := config{
		Port:            envOr("PORT", "8080"),
		DatabasePath:    envOr("DATABASE_PATH", "/tmp/pocket-aide-dev.db"),
		OIDCIssuer:      issuer,
		OIDCAudience:    mock.Audience(),
		OIDCClientID:    mock.ClientID(),
		OIDCRedirectURI: envOr("OIDC_REDIRECT_URI", "pocketaide://oauth-callback"),
	}
	return cfg, mock, srv
}

// devAuthTokenHandler issues an oidcmock-signed access token without going
// through the full PKCE auth-code flow. Wired only when dev mode is on, so it
// cannot be reached in production builds.
func devAuthTokenHandler(mock *oidcmock.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ttl := time.Hour
		access, err := mock.SignAccessToken(mock.Subject(), ttl)
		if err != nil {
			http.Error(w, "token mint failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": access,
			"token_type":   "Bearer",
			"expires_in":   int(ttl.Seconds()),
		})
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
