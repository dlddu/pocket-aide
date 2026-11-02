// Package telemetry wires OpenTelemetry tracing: an OTLP/HTTP exporter, W3C
// trace-context propagation, a server middleware, an instrumented
// http.DefaultTransport and database/sql spans.
//
// Tracing is opt-in. With OTEL_EXPORTER_OTLP_ENDPOINT (or
// OTEL_EXPORTER_OTLP_TRACES_ENDPOINT) unset, Setup leaves the global no-op
// tracer in place, so tests, the e2e stacks and local runs export nothing.
//
// Why an SDK at all when OBI (eBPF) already traces this process: OBI's Go
// server probe looks for the traceparent header only in the first 1024 bytes
// of the request buffer. Traefik writes headers in alphabetical order, so the
// app's long Authorization (OIDC bearer) pushes Traceparent past that window
// and those requests started a new trace instead of continuing Traefik's. The
// SDK reads the parsed headers, so the edge-to-app link holds. OBI skips
// processes that export OTLP themselves
// (discovery.exclude_otel_instrumented_services).
package telemetry

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/XSAM/otelsql"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

// Enabled reports whether an OTLP endpoint is configured.
func Enabled() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""
}

// Setup installs the global tracer provider and propagator and wraps
// http.DefaultTransport so outgoing requests from clients that use it carry
// traceparent and get client spans. The returned function flushes and stops
// the exporter.
func Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }
	if !Enabled() {
		return noop, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return noop, err
	}
	// Later options win: the compiled-in name is the default and
	// OTEL_SERVICE_NAME / OTEL_RESOURCE_ATTRIBUTES override it.
	res, err := resource.New(ctx,
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceName(serviceName)),
		resource.WithFromEnv(),
	)
	if err != nil && !errors.Is(err, resource.ErrPartialResource) {
		return noop, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// The OTLP exporter keeps its own transport, so its export requests are
	// not traced by this wrapper.
	http.DefaultTransport = otelhttp.NewTransport(http.DefaultTransport)

	return tp.Shutdown, nil
}

// Handler wraps h with a server span per request. Requests to any of
// skipPaths (probes) get no span. The span starts as "METHOD path";
// ChiRouteName renames it to the matched route once chi has routed.
func Handler(h http.Handler, skipPaths ...string) http.Handler {
	skip := make(map[string]bool, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = true
	}
	return otelhttp.NewHandler(h, "http.server",
		otelhttp.WithFilter(func(r *http.Request) bool { return !skip[r.URL.Path] }),
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)
}

// ChiRouteName is chi middleware that names the server span after the
// matched route pattern ("GET /api/approvals/{id}") so IDs in the path do not
// make every span name unique.
func ChiRouteName(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		rctx := chi.RouteContext(r.Context())
		if rctx == nil {
			return
		}
		pattern := rctx.RoutePattern()
		if pattern == "" {
			return
		}
		span := trace.SpanFromContext(r.Context())
		span.SetName(r.Method + " " + pattern)
		span.SetAttributes(semconv.HTTPRoute(pattern))
	})
}

// OpenDB opens driverName like sql.Open and records a span per query and
// statement. Only calls inside an existing trace are recorded, so background
// loops and health checks do not start traces of their own, and the per-row,
// prepare, connect and reset spans are left out.
func OpenDB(driverName, dsn, dbSystem string) (*sql.DB, error) {
	return otelsql.Open(driverName, dsn,
		otelsql.WithAttributes(semconv.DBSystemNameKey.String(dbSystem)),
		otelsql.WithSpanNameFormatter(func(_ context.Context, method otelsql.Method, query string) string {
			if verb, _, ok := strings.Cut(strings.TrimSpace(query), " "); ok && verb != "" {
				return strings.ToUpper(verb)
			}
			return string(method)
		}),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			OmitConnPrepare:      true,
			OmitRows:             true,
			OmitConnectorConnect: true,
			SpanFilter: func(ctx context.Context, _ otelsql.Method, _ string, _ []driver.NamedValue) bool {
				return trace.SpanContextFromContext(ctx).IsValid()
			},
		}),
	)
}

// StartSQSMessage starts the root span for handling one SQS message, so the
// SQL and outgoing calls made while processing it share a trace. The long
// poll itself (ReceiveMessage every 20s) is deliberately not traced: it is a
// background loop with no request behind it and would only add root spans.
func StartSQSMessage(ctx context.Context, queueURL, messageID, githubEvent string) (context.Context, trace.Span) {
	queue := queueURL
	if i := strings.LastIndex(queueURL, "/"); i >= 0 {
		queue = queueURL[i+1:]
	}
	return otel.Tracer("github.com/dlddu/pocket-aide/backend/internal/telemetry").Start(ctx, "process "+queue,
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			semconv.MessagingSystemAWSSQS,
			semconv.MessagingOperationTypeProcess,
			semconv.MessagingDestinationName(queue),
			semconv.MessagingMessageID(messageID),
			attribute.String("github.event", githubEvent),
		),
	)
}

// RecordError marks span as failed with err.
func RecordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}
