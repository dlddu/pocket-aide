// Command server is the pocket-aide backend HTTP server.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/dlddu/pocket-aide/backend/internal/affirmations"
	"github.com/dlddu/pocket-aide/backend/internal/apns"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/githubwebhook"
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
	r.Use(loggerSkipping("/healthz"))

	r.Get("/healthz", handlers.Health(conn))
	r.Get("/api/auth/config", handlers.AuthConfigHandler(authCfg))

	affStore := affirmations.New(conn)
	deviceStore := devicetokens.New(conn)

	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(verifier, conn))
		p.Get("/api/me", handlers.Me())
		p.Get("/api/affirmations", handlers.ListAffirmations(affStore))
		p.Post("/api/affirmations", handlers.CreateAffirmation(affStore))
		p.Patch("/api/affirmations/{id}", handlers.UpdateAffirmation(affStore))
		p.Delete("/api/affirmations/{id}", handlers.DeleteAffirmation(affStore))
		p.Post("/api/device-tokens", handlers.RegisterDeviceToken(deviceStore))
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

	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	defer cancelConsumer()
	if cfg.PRMonitorEnabled {
		apnsClient, err := apns.New(
			cfg.APNSKeyID, cfg.APNSTeamID, cfg.APNSBundleID,
			[]byte(cfg.APNSAuthKeyP8), cfg.APNSUseProduction,
		)
		if err != nil {
			log.Fatalf("apns: %v", err)
		}
		dispatch := func(ctx context.Context, evt githubwebhook.WorkflowRunEvent) error {
			tokens, err := deviceStore.ListAll(ctx)
			if err != nil {
				return err
			}
			title := fmt.Sprintf("%s — %s", evt.Repo, evt.Conclusion)
			body := fmt.Sprintf("%s on %s", evt.WorkflowName, evt.HeadBranch)
			for _, t := range tokens {
				if err := apnsClient.Send(ctx, t, title, body); err != nil {
					log.Printf("apns send token=%s…: %v", safePrefix(t), err)
				}
			}
			return nil
		}
		consumer, err := githubwebhook.New(consumerCtx, cfg.SQSQueueURL, cfg.GitHubWebhookSecret, dispatch)
		if err != nil {
			log.Fatalf("sqs consumer: %v", err)
		}
		go consumer.Run(consumerCtx)
	} else {
		log.Printf("pr-monitor: disabled (SQS_QUEUE_URL not set)")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
	cancelConsumer()
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

	// PR monitor pipeline. Disabled when SQS_QUEUE_URL is empty so local /
	// test environments don't need APNs or SQS configured.
	PRMonitorEnabled    bool
	SQSQueueURL         string
	GitHubWebhookSecret string
	APNSKeyID           string
	APNSTeamID          string
	APNSBundleID        string
	APNSAuthKeyP8       string
	APNSUseProduction   bool
}

func loadConfig() config {
	c := config{
		Port:            envOr("PORT", "8080"),
		DatabasePath:    envOr("DATABASE_PATH", "/data/pocket-aide.db"),
		OIDCIssuer:      mustEnv("OIDC_ISSUER"),
		OIDCAudience:    mustEnv("OIDC_AUDIENCE"),
		OIDCClientID:    mustEnv("OIDC_CLIENT_ID"),
		OIDCRedirectURI: mustEnv("OIDC_REDIRECT_URI"),
		SQSQueueURL:     os.Getenv("SQS_QUEUE_URL"),
	}
	if c.SQSQueueURL != "" {
		c.PRMonitorEnabled = true
		c.GitHubWebhookSecret = mustEnv("GITHUB_WEBHOOK_SECRET")
		c.APNSKeyID = mustEnv("APNS_KEY_ID")
		c.APNSTeamID = mustEnv("APNS_TEAM_ID")
		c.APNSBundleID = mustEnv("APNS_BUNDLE_ID")
		c.APNSAuthKeyP8 = mustEnv("APNS_AUTH_KEY_P8")
		c.APNSUseProduction = envOr("APNS_USE_PRODUCTION", "false") == "true"
	}
	return c
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

// safePrefix returns the first 8 chars of a token (or fewer) so log lines can
// identify devices without leaking the full token.
func safePrefix(t string) string {
	if len(t) <= 8 {
		return t
	}
	return t[:8]
}

// loggerSkipping wraps middleware.Logger so that requests to the given paths
// bypass access logging. Registered at the router root so unmatched routes
// (404) are logged too.
func loggerSkipping(paths ...string) func(http.Handler) http.Handler {
	skip := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		skip[p] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		logged := middleware.Logger(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			logged.ServeHTTP(w, r)
		})
	}
}
