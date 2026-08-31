package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	errorDetails, ok := record["error"].(map[string]any)
	if !ok {
		t.Fatalf("error = %#v, want structured safe diagnostic", record["error"])
	}
	if errorDetails["type"] != "*errors.errorString" {
		t.Fatalf("error type = %v, want safe concrete type", errorDetails["type"])
	}
	if _, exists := errorDetails["safe_message"]; exists {
		t.Fatalf("unclassified error exposed a message: %#v", errorDetails)
	}
	if record["provider"] != "github" {
		t.Fatalf("safe provider field = %v", record["provider"])
	}
}

func TestLoggerWritesReviewedDiagnosticWithoutRawCause(t *testing.T) {
	t.Parallel()

	const rawCause = "postgres://admin:sensitive-password@private.example/customer"
	cause := errors.New(rawCause)
	definition := MustDefineError("worker.database.connection_failed", "Worker could not connect to PostgreSQL")
	classified := definition.Wrap(fmt.Errorf("open application database: %w", cause))

	if !errors.Is(classified, cause) {
		t.Fatal("classified error does not preserve errors.Is traversal")
	}
	for _, format := range []string{"%s", "%v", "%+v", "%#v", "%q"} {
		if rendered := fmt.Sprintf(format, classified); strings.Contains(rendered, rawCause) {
			t.Fatalf("format %q exposed raw cause: %q", format, rendered)
		}
	}

	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Error(context.Background(), "worker bootstrap failed", "phase", "bootstrap", "error", classified)

	if strings.Contains(output.String(), rawCause) || strings.Contains(output.String(), "sensitive-password") {
		t.Fatalf("classified error exposed its cause: %s", output.String())
	}
	record := decodeLogRecord(t, output.Bytes())
	errorDetails := requireErrorDetails(t, record, "error")
	if errorDetails["type"] != "*fmt.wrapError" {
		t.Fatalf("error type = %v, want *fmt.wrapError", errorDetails["type"])
	}
	if errorDetails["code"] != "worker.database.connection_failed" {
		t.Fatalf("error code = %v", errorDetails["code"])
	}
	if errorDetails["safe_message"] != "Worker could not connect to PostgreSQL" {
		t.Fatalf("safe message = %v", errorDetails["safe_message"])
	}
	typeChain, ok := errorDetails["type_chain"].([]any)
	if !ok || len(typeChain) < 2 || typeChain[0] != "*fmt.wrapError" || typeChain[1] != "*errors.errorString" {
		t.Fatalf("type chain = %#v", errorDetails["type_chain"])
	}
}

func TestErrorDefinitionFallbackPreservesSpecificDiagnostic(t *testing.T) {
	t.Parallel()

	cause := errors.New("sensitive cause")
	specific := MustDefineError("worker.slack.signing_secret_missing", "SLACK_SIGNING_SECRET is not configured")
	fallback := MustDefineError("worker.bootstrap.failed", "Worker failed during startup")
	classified := specific.Wrap(cause)

	if got := fallback.WrapIfUnclassified(classified); got != classified {
		t.Fatal("fallback replaced a more specific diagnostic")
	}
	if got := fallback.WrapIfUnclassified(cause); got == cause {
		t.Fatal("fallback did not classify an unknown error")
	}
}

func TestLoggerSanitizesErrorsRegardlessOfAttributeKey(t *testing.T) {
	t.Parallel()

	const rawCause = "provider response included private@example.com and bearer-token"
	for _, testCase := range []struct {
		name    string
		logArgs []any
		wantKey string
	}{
		{name: "reason", logArgs: []any{"reason", errors.New(rawCause)}, wantKey: "reason"},
		{name: "orphan", logArgs: []any{errors.New(rawCause)}, wantKey: "error"},
		{name: "log valuer", logArgs: []any{"reason", logValuerError{message: rawCause}}, wantKey: "reason"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
			log.Error(context.Background(), "operation failed", testCase.logArgs...)

			if strings.Contains(output.String(), rawCause) || strings.Contains(output.String(), "private@example.com") {
				t.Fatalf("raw error reached output: %s", output.String())
			}
			record := decodeLogRecord(t, output.Bytes())
			requireErrorDetails(t, record, testCase.wantKey)
			if _, exists := record["!BADKEY"]; exists {
				t.Fatalf("malformed error retained !BADKEY: %#v", record)
			}
		})
	}
}

func TestLoggerTraversesJoinedErrorsWithinBounds(t *testing.T) {
	t.Parallel()

	const rawCause = "secret provider body"
	definition := MustDefineError("worker.runtime.failed", "Worker runtime stopped unexpectedly")
	joined := errors.Join(
		fmt.Errorf("first branch: %w", errors.New(rawCause)),
		definition.Wrap(errors.New(rawCause)),
	)

	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Error(context.Background(), "worker failed", "error", joined)

	if strings.Contains(output.String(), rawCause) {
		t.Fatalf("joined error exposed raw cause: %s", output.String())
	}
	errorDetails := requireErrorDetails(t, decodeLogRecord(t, output.Bytes()), "error")
	if errorDetails["code"] != "worker.runtime.failed" {
		t.Fatalf("joined error code = %v", errorDetails["code"])
	}
	typeChain, ok := errorDetails["type_chain"].([]any)
	if !ok || len(typeChain) == 0 || len(typeChain) > maxErrorChainTypes {
		t.Fatalf("joined type chain = %#v", errorDetails["type_chain"])
	}
}

func TestLoggerBoundsCyclicErrorsAndHandlesTypedNil(t *testing.T) {
	t.Parallel()

	var typedNil *typedNilError
	for _, testCase := range []struct {
		name string
		err  error
	}{
		{name: "cycle", err: &cyclicError{}},
		{name: "typed nil", err: typedNil},
		{name: "panicking unwrap", err: panickingUnwrapError{}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
			log.Error(context.Background(), "operation failed", "error", testCase.err)

			if strings.Contains(output.String(), "sensitive panic detail") {
				t.Fatalf("defensive error inspection exposed text: %s", output.String())
			}
			errorDetails := requireErrorDetails(t, decodeLogRecord(t, output.Bytes()), "error")
			typeChain, ok := errorDetails["type_chain"].([]any)
			if !ok || len(typeChain) == 0 || len(typeChain) > maxErrorChainTypes {
				t.Fatalf("bounded type chain = %#v", errorDetails["type_chain"])
			}
		})
	}
}

func TestLoggerSanitizesErrorsInsideGroupsBeforeResolvingLogValuer(t *testing.T) {
	t.Parallel()

	const rawCause = "provider body included sensitive-token"
	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Error(
		context.Background(),
		"operation failed",
		slog.Group("failure", "reason", logValuerError{message: rawCause}),
	)

	if strings.Contains(output.String(), rawCause) || strings.Contains(output.String(), "sensitive-token") {
		t.Fatalf("grouped LogValuer error exposed raw text: %s", output.String())
	}
	record := decodeLogRecord(t, output.Bytes())
	failure, ok := record["failure"].(map[string]any)
	if !ok {
		t.Fatalf("failure = %#v, want structured group", record["failure"])
	}
	requireErrorDetails(t, failure, "reason")
}

func TestLoggerRedactsErrorMessageAliases(t *testing.T) {
	t.Parallel()

	const rawCause = "sensitive provider detail"
	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Error(
		context.Background(),
		"operation failed",
		"error_message", rawCause,
		"error_detail", rawCause,
		"error_reason", rawCause,
		"cause", rawCause,
	)

	if strings.Contains(output.String(), rawCause) {
		t.Fatalf("error alias exposed raw text: %s", output.String())
	}
	record := decodeLogRecord(t, output.Bytes())
	for _, key := range []string{"error_message", "error_detail", "error_reason", "cause"} {
		if record[key] != redactedValue {
			t.Fatalf("%s = %v, want redacted", key, record[key])
		}
	}
}

func TestLoggerRedactsSensitiveGroupsBeforeResolvingChildren(t *testing.T) {
	t.Parallel()

	const sensitiveValue = "private group value with bearer-token"
	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Error(
		context.Background(),
		"operation failed",
		slog.Group("payload", "value", sensitiveValue),
		slog.Group("authorization", "reason", logValuerError{message: sensitiveValue}),
		"cause", errors.New(sensitiveValue),
	)

	if strings.Contains(output.String(), sensitiveValue) || strings.Contains(output.String(), "bearer-token") {
		t.Fatalf("sensitive group exposed child data: %s", output.String())
	}
	record := decodeLogRecord(t, output.Bytes())
	for _, key := range []string{"payload", "authorization", "cause"} {
		if record[key] != redactedValue {
			t.Fatalf("%s = %#v, want redacted", key, record[key])
		}
	}
}

func TestErrorDefinitionRejectsUnsafeMetadata(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name    string
		code    string
		message string
	}{
		{name: "dynamic code", code: "worker.failure.private@example.com", message: "Worker failed"},
		{name: "uppercase code", code: "Worker.Runtime.Failed", message: "Worker failed"},
		{name: "multiline message", code: "worker.runtime.failed", message: "Worker failed\nsecret detail"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if recover() == nil {
					t.Fatal("MustDefineError did not reject unsafe metadata")
				}
			}()
			MustDefineError(testCase.code, testCase.message)
		})
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

type logValuerError struct {
	message string
}

func (err logValuerError) Error() string {
	return err.message
}

func (err logValuerError) LogValue() slog.Value {
	return slog.StringValue(err.message)
}

type cyclicError struct{}

func (*cyclicError) Error() string {
	return "sensitive cycle detail"
}

func (err *cyclicError) Unwrap() error {
	return err
}

type typedNilError struct{}

func (*typedNilError) Error() string {
	return "sensitive typed nil detail"
}

type panickingUnwrapError struct{}

func (panickingUnwrapError) Error() string {
	return "sensitive panic detail"
}

func (panickingUnwrapError) Unwrap() error {
	panic("sensitive panic detail")
}

func decodeLogRecord(t *testing.T, output []byte) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(output, &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	return record
}

func requireErrorDetails(t *testing.T, record map[string]any, key string) map[string]any {
	t.Helper()
	details, ok := record[key].(map[string]any)
	if !ok {
		t.Fatalf("%s = %#v, want structured safe diagnostic", key, record[key])
	}
	return details
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
