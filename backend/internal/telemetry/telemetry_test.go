package telemetry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	_ "modernc.org/sqlite"
)

func withRecorder(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prevTP, prevProp := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})
	return rec
}

func router() http.Handler {
	r := chi.NewRouter()
	r.Use(ChiRouteName)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {})
	r.Get("/api/approvals/{id}", func(w http.ResponseWriter, _ *http.Request) {})
	return Handler(r, "/healthz")
}

// The reason this package exists: a traceparent behind a long Authorization
// header (as Traefik writes it, alphabetically) must still be continued.
func TestHandlerContinuesTraceBehindLongAuthorization(t *testing.T) {
	rec := withRecorder(t)

	const traceID = "4bf92f3577b34da6a3ce929d0e0e4736"
	req := httptest.NewRequest(http.MethodGet, "/api/approvals/42", nil)
	req.Header.Set("Authorization", "Bearer "+strings.Repeat("a", 2000))
	req.Header.Set("Traceparent", "00-"+traceID+"-00f067aa0ba902b7-01")
	router().ServeHTTP(httptest.NewRecorder(), req)

	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("got %d spans, want 1", len(spans))
	}
	s := spans[0]
	if got := s.SpanContext().TraceID().String(); got != traceID {
		t.Errorf("trace id = %s, want %s", got, traceID)
	}
	if got := s.Name(); got != "GET /api/approvals/{id}" {
		t.Errorf("span name = %q, want route pattern", got)
	}
}

func TestHandlerSkipsHealthz(t *testing.T) {
	rec := withRecorder(t)
	router().ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if n := len(rec.Ended()); n != 0 {
		t.Errorf("got %d spans for /healthz, want 0", n)
	}
}

func TestOpenDBRecordsOnlyInsideATrace(t *testing.T) {
	rec := withRecorder(t)
	db, err := OpenDB("sqlite", "file:"+filepath.Join(t.TempDir(), "t.db"), "sqlite")
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), "CREATE TABLE t (id INTEGER)"); err != nil {
		t.Fatalf("exec without trace: %v", err)
	}
	if n := len(rec.Ended()); n != 0 {
		t.Fatalf("got %d spans outside a trace, want 0", n)
	}

	ctx, parent := otel.Tracer("test").Start(context.Background(), "request")
	if _, err := db.ExecContext(ctx, "INSERT INTO t (id) VALUES (1)"); err != nil {
		t.Fatalf("exec inside trace: %v", err)
	}
	parent.End()

	var names []string
	for _, s := range rec.Ended() {
		names = append(names, s.Name())
	}
	if len(names) != 2 || names[0] != "INSERT" {
		t.Errorf("span names = %v, want [INSERT request]", names)
	}
}

func TestSetupIsNoopWithoutEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	before := http.DefaultTransport
	shutdown, err := Setup(t.Context(), "svc")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if http.DefaultTransport != before {
		t.Error("Setup replaced http.DefaultTransport while tracing is disabled")
	}
}

func TestStartSQSMessageRootsProcessingWork(t *testing.T) {
	rec := withRecorder(t)
	db, err := OpenDB("sqlite", "file:"+filepath.Join(t.TempDir(), "q.db"), "sqlite")
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, span := StartSQSMessage(context.Background(),
		"https://sqs.ap-northeast-2.amazonaws.com/123456789012/github-events", "m-1", "workflow_run")
	if _, err := db.ExecContext(ctx, "CREATE TABLE t (id INTEGER)"); err != nil {
		t.Fatalf("exec: %v", err)
	}
	RecordError(span, errors.New("dispatch failed"))
	span.End()

	spans := rec.Ended()
	if len(spans) != 2 {
		t.Fatalf("got %d spans, want 2 (CREATE + process)", len(spans))
	}
	root := spans[1]
	if got := root.Name(); got != "process github-events" {
		t.Errorf("root name = %q, want %q", got, "process github-events")
	}
	if root.SpanKind() != trace.SpanKindConsumer {
		t.Errorf("root kind = %v, want consumer", root.SpanKind())
	}
	if root.Parent().IsValid() {
		t.Error("message span should be a trace root")
	}
	if root.Status().Code != codes.Error {
		t.Errorf("root status = %v, want error", root.Status().Code)
	}
	if spans[0].Parent().SpanID() != root.SpanContext().SpanID() {
		t.Error("SQL span is not a child of the message span")
	}
}
