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
	"github.com/dlddu/pocket-aide/backend/internal/approvals"
	"github.com/dlddu/pocket-aide/backend/internal/auth"
	"github.com/dlddu/pocket-aide/backend/internal/db"
	"github.com/dlddu/pocket-aide/backend/internal/devicetokens"
	"github.com/dlddu/pocket-aide/backend/internal/excludedrepos"
	"github.com/dlddu/pocket-aide/backend/internal/githubwebhook"
	"github.com/dlddu/pocket-aide/backend/internal/handlers"
	"github.com/dlddu/pocket-aide/backend/internal/notificationhistory"
	"github.com/dlddu/pocket-aide/backend/internal/notificationsettings"
	"github.com/dlddu/pocket-aide/backend/internal/routines"
	"github.com/dlddu/pocket-aide/backend/internal/scratchpad"
	"github.com/dlddu/pocket-aide/backend/internal/sessions"
	"github.com/dlddu/pocket-aide/backend/internal/todos"
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
	excludedStore := excludedrepos.New(conn)
	historyStore := notificationhistory.New(conn)
	settingsStore := notificationsettings.New(conn)
	todoStore := todos.New(conn)
	scratchStore := scratchpad.New(conn)
	routineStore := routines.New(conn)
	sessionStore := sessions.New(conn)
	platform := sessions.NewClient(cfg.SessionPlatformURL)
	approvalStore := approvals.New(conn)

	r.Group(func(p chi.Router) {
		p.Use(auth.Middleware(verifier, conn))
		p.Get("/api/me", handlers.Me())
		p.Get("/api/affirmations", handlers.ListAffirmations(affStore))
		p.Post("/api/affirmations", handlers.CreateAffirmation(affStore))
		p.Patch("/api/affirmations/{id}", handlers.UpdateAffirmation(affStore))
		p.Delete("/api/affirmations/{id}", handlers.DeleteAffirmation(affStore))
		p.Post("/api/device-tokens", handlers.RegisterDeviceToken(deviceStore))
		p.Get("/api/excluded-repos", handlers.ListExcludedRepos(excludedStore))
		p.Post("/api/excluded-repos", handlers.AddExcludedRepo(excludedStore))
		p.Delete("/api/excluded-repos/{id}", handlers.DeleteExcludedRepo(excludedStore))
		p.Get("/api/notification-history", handlers.ListNotificationHistory(historyStore))
		p.Post("/api/notification-history/{id}/ack", handlers.AcknowledgeNotification(historyStore))
		p.Get("/api/notification-settings", handlers.GetNotificationSettings(settingsStore))
		p.Patch("/api/notification-settings", handlers.UpdateNotificationSettings(settingsStore))
		p.Get("/api/todos/{area}", handlers.ListTodos(todoStore))
		p.Post("/api/todos/{area}", handlers.CreateTodo(todoStore))
		p.Patch("/api/todos/{area}/{id}", handlers.UpdateTodo(todoStore))
		p.Delete("/api/todos/{area}/{id}", handlers.DeleteTodo(todoStore))
		p.Get("/api/scratchpad", handlers.ListScratchpad(scratchStore))
		p.Post("/api/scratchpad", handlers.CreateScratchpadItem(scratchStore))
		p.Delete("/api/scratchpad/{id}", handlers.DeleteScratchpadItem(scratchStore))
		p.Post("/api/scratchpad/{id}/move", handlers.MoveScratchpadItem(scratchStore, todoStore, affStore, routineStore))
		p.Get("/api/routines", handlers.ListRoutines(routineStore))
		p.Post("/api/routines", handlers.CreateRoutine(routineStore))
		p.Delete("/api/routines/{id}", handlers.DeleteRoutine(routineStore))
		p.Post("/api/routines/{id}/steps", handlers.AddRoutineStep(routineStore))
		p.Delete("/api/routines/{id}/steps/{stepID}", handlers.DeleteRoutineStep(routineStore))
		p.Get("/api/routines/days/{day}", handlers.ListRoutineDay(routineStore))
		p.Patch("/api/routines/{id}/days/{day}/steps/{stepID}", handlers.SetRoutineStepCheck(routineStore))
		p.Get("/api/routines/{id}/history/{day}", handlers.RoutineHistory(routineStore))
		p.Get("/api/sessions/config", handlers.SessionConfig(platform))
		p.Get("/api/sessions", handlers.ListSessions(sessionStore, platform))
		p.Post("/api/sessions", handlers.CreateSession(sessionStore, platform))
		p.Get("/api/sessions/{id}", handlers.GetSession(sessionStore, platform))
		p.Delete("/api/sessions/{id}", handlers.DeleteSession(sessionStore, platform))
		p.Post("/api/sessions/{id}/read", handlers.ReadSession(sessionStore, platform))
		p.Post("/api/sessions/{id}/write", handlers.WriteSession(sessionStore, platform))
		p.Post("/api/sessions/{id}/switch", handlers.SwitchSession(sessionStore, platform))
		p.Post("/api/sessions/{id}/snapshot", handlers.SnapshotSession(sessionStore, platform))
		p.Get("/api/approval-keys", handlers.ListApprovalKeys(approvalStore))
		p.Post("/api/approval-keys", handlers.IssueApprovalKey(approvalStore))
		p.Post("/api/approval-keys/{id}/revoke", handlers.RevokeApprovalKey(approvalStore))
		p.Get("/api/approvals", handlers.ListPendingApprovals(approvalStore))
		p.Get("/api/approvals/{id}", handlers.GetApproval(approvalStore))
		p.Post("/api/approvals/{id}/approve", handlers.DecideApproval(approvalStore, approvals.DecisionApprove))
		p.Post("/api/approvals/{id}/reject", handlers.DecideApproval(approvalStore, approvals.DecisionReject))
	})

	r.Group(func(x chi.Router) {
		x.Use(handlers.CallerKeyMiddleware(approvalStore))
		x.Post("/api/external/approvals", handlers.CreateExternalApproval(approvalStore))
		x.Get("/api/external/approvals/{id}", handlers.GetExternalApproval(approvalStore))
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
		var apnsClient *apns.Client
		if cfg.APNSDisabled {
			log.Printf("pr-monitor: APNs push disabled (APNS_DISABLED=true); history is still recorded")
		} else {
			apnsClient, err = apns.New(
				cfg.APNSKeyID, cfg.APNSTeamID, cfg.APNSBundleID,
				[]byte(cfg.APNSAuthKeyP8), cfg.APNSUseProduction,
			)
			if err != nil {
				log.Fatalf("apns: %v", err)
			}
		}
		dispatch := func(ctx context.Context, evt githubwebhook.WorkflowRunEvent) error {
			// Looked up outside InsertBatchTx so the transaction holds
			// SQLite's write lock only for the inserts.
			userIDs, err := excludedStore.ListUserIDsExcluding(ctx, evt.Repo)
			if err != nil {
				return fmt.Errorf("list matched users: %w", err)
			}
			if len(userIDs) == 0 {
				log.Printf("githubwebhook: no matched users for repo=%s", evt.Repo)
				return nil
			}

			// PRD-10 AC11: history must be persisted before any push goes out.
			historyEvt := notificationhistory.Event{
				RepoFullName: evt.Repo,
				PRNumber:     evt.PRNumber,
				PRTitle:      evt.PRTitle,
				PRURL:        evt.PRURL,
				CommitURL:    evt.CommitURL,
				RunURL:       evt.HTMLURL,
				WorkflowName: evt.WorkflowName,
				HeadBranch:   evt.HeadBranch,
				HeadSHA:      evt.HeadSHA,
				Conclusion:   evt.Conclusion,
			}
			ids, err := historyStore.InsertBatchTx(ctx, userIDs, historyEvt)
			if err != nil {
				// Returning error makes handleMessage skip DeleteMessage —
				// SQS redelivers after VisibilityTimeout.
				return fmt.Errorf("persist history: %w", err)
			}
			if apnsClient == nil {
				return nil
			}

			if !shouldPush(evt) {
				log.Printf("githubwebhook: history-only (start event) repo=%s workflow=%s", evt.Repo, evt.WorkflowName)
				return nil
			}

			// Push failures are logged, not returned: the history rows are
			// already committed, so an SQS redelivery would insert them again.
			title, body := formatPushText(evt)
			for i, uid := range userIDs {
				settings, err := settingsStore.Get(ctx, uid)
				if err != nil {
					log.Printf("apns: notification settings for user=%d: %v", uid, err)
					continue
				}
				if !settings.AllowsPush(evt.Conclusion) {
					continue
				}
				tokens, err := deviceStore.ListByUserID(ctx, uid)
				if err != nil {
					log.Printf("apns: list tokens for user=%d: %v", uid, err)
					continue
				}
				for _, t := range tokens {
					data := map[string]any{"event_id": ids[i]}
					if err := apnsClient.SendWithData(ctx, t, title, body, data); err != nil {
						log.Printf("apns send user=%d token=%s…: %v", uid, safePrefix(t), err)
					}
				}
			}
			return nil
		}
		consumer, err := githubwebhook.New(consumerCtx, cfg.SQSQueueURL, cfg.AWSRoleARN, dispatch)
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

	SessionPlatformURL string

	PRMonitorEnabled  bool
	SQSQueueURL       string
	AWSRoleARN        string
	APNSKeyID         string
	APNSTeamID        string
	APNSBundleID      string
	APNSAuthKeyP8     string
	APNSUseProduction bool
	// APNSDisabled keeps the consumer and history writes on but skips the
	// push fan-out, so the E2E backend can run the pipeline without Apple
	// credentials (docs/e2e-mocking-policy.md).
	APNSDisabled bool
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

		SessionPlatformURL: envOr("SESSION_PLATFORM_URL", sessions.DefaultPlatformURL),
	}
	if c.SQSQueueURL != "" {
		c.PRMonitorEnabled = true
		c.AWSRoleARN = os.Getenv("AWS_ROLE_ARN")
		c.APNSDisabled = os.Getenv("APNS_DISABLED") == "true"
		if !c.APNSDisabled {
			c.APNSKeyID = mustEnv("APNS_KEY_ID")
			c.APNSTeamID = mustEnv("APNS_TEAM_ID")
			c.APNSBundleID = mustEnv("APNS_BUNDLE_ID")
			c.APNSAuthKeyP8 = mustEnv("APNS_AUTH_KEY_P8")
			c.APNSUseProduction = envOr("APNS_USE_PRODUCTION", "false") == "true"
		}
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

func shouldPush(evt githubwebhook.WorkflowRunEvent) bool {
	return evt.Completed
}

// safePrefix returns the first 8 chars of a token (or fewer) so log lines can
// identify devices without leaking the full token.
func safePrefix(t string) string {
	if len(t) <= 8 {
		return t
	}
	return t[:8]
}

// The PR-less fallback is not an edge case: GitHub leaves
// workflow_run.pull_requests empty for runs triggered by a direct push (e.g. main).
func formatPushText(evt githubwebhook.WorkflowRunEvent) (title, body string) {
	verdict := "CI " + evt.Conclusion
	switch evt.Conclusion {
	case "success":
		verdict = "CI 통과"
	case "failure":
		verdict = "CI 실패"
	case "queued", "requested", "in_progress", "pending", "waiting":
		verdict = "CI 시작"
	}
	if evt.PRNumber > 0 {
		title = fmt.Sprintf("%s — %s #%d", verdict, evt.Repo, evt.PRNumber)
		if evt.PRTitle != "" {
			body = evt.PRTitle
		} else {
			body = fmt.Sprintf("%s on %s", evt.WorkflowName, evt.HeadBranch)
		}
		return
	}
	title = fmt.Sprintf("%s — %s", evt.Repo, evt.Conclusion)
	body = fmt.Sprintf("%s on %s", evt.WorkflowName, evt.HeadBranch)
	return
}

// Registered at the router root, not on a route group, so unmatched (404)
// requests are still logged.
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
