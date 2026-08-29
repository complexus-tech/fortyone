package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestLoggerAddsRequestAndTraceCorrelation(t *testing.T) {
	t.Parallel()

	traceID := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
	ctx = WithRequestID(ctx, "request-test-id")

	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Info(ctx, "correlated")

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if record["request_id"] != "request-test-id" {
		t.Fatalf("request_id = %v", record["request_id"])
	}
	if record["trace_id"] != traceID.String() || record["span_id"] != spanID.String() {
		t.Fatalf("trace correlation = trace %v span %v", record["trace_id"], record["span_id"])
	}
}

func TestLoggerOmitsAbsentCorrelation(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Info(context.Background(), "uncorrelated")

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if _, exists := record["request_id"]; exists {
		t.Fatal("uncorrelated log contains request_id")
	}
	if _, exists := record["trace_id"]; exists {
		t.Fatal("uncorrelated log contains trace_id")
	}
}

func TestLoggerRedactsSensitiveAttributesAndErrorContents(t *testing.T) {
	t.Parallel()

	const (
		email  = "private@example.com"
		token  = "sensitive-bearer-token"
		detail = "provider response contained sensitive material"
	)
	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Error(
		context.Background(),
		"provider operation failed",
		"email", email,
		"invitee_email", email,
		"access_token", token,
		"callback_url", "https://example.com/callback?code="+token,
		"raw_message", detail,
		"error", errors.New(detail),
		"provider", "github",
	)

	if strings.Contains(output.String(), email) || strings.Contains(output.String(), token) || strings.Contains(output.String(), detail) {
		t.Fatalf("sensitive value reached log output: %s", output.String())
	}
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if record["email"] != redactedValue || record["invitee_email"] != redactedValue || record["access_token"] != redactedValue || record["callback_url"] != redactedValue || record["raw_message"] != redactedValue {
		t.Fatalf("sensitive fields were not redacted: %#v", record)
	}
	if record["error"] != "*errors.errorString" {
		t.Fatalf("error = %v, want safe concrete type", record["error"])
	}
	if record["provider"] != "github" {
		t.Fatalf("safe provider field = %v", record["provider"])
	}
}

func TestLoggerReportsHandlerFailureWithoutLeakingErrorText(t *testing.T) {
	t.Parallel()

	const sensitiveFailure = "failed destination included sensitive-token"
	var fallback bytes.Buffer
	log := &Logger{
		handler:        slog.NewJSONHandler(errorWriter{err: errors.New(sensitiveFailure)}, nil),
		fallbackWriter: &fallback,
	}

	log.Info(context.Background(), "will fail")

	if !strings.Contains(fallback.String(), "structured log handler failed") {
		t.Fatalf("fallback output = %q, want observable handler failure", fallback.String())
	}
	if strings.Contains(fallback.String(), sensitiveFailure) || strings.Contains(fallback.String(), "sensitive-token") {
		t.Fatalf("fallback output leaked handler error: %q", fallback.String())
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
