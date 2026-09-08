package logger

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestLoggerIncludesReadableWrappedAndJoinedErrors(t *testing.T) {
	err := fmt.Errorf("create story update notifications: %w", errors.Join(
		fmt.Errorf("recipient: %w", errors.New("story read is not permitted: actor not found in context")),
		errors.New(`duplicate key value violates unique constraint "notifications_recipient_workspace_entity_unique" (SQLSTATE 23505)`),
	))
	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelDebug, "logger-test")
	log.Error(context.Background(), "failed to process message", "error", err)
	details := requireErrorDetails(t, decodeLogRecord(t, output.Bytes()), "error")
	if details["message"] != err.Error() {
		t.Fatalf("readable error = %v, want complete operation and causes", details["message"])
	}
}

func TestReadableErrorRedactsCredentialsAndPayloads(t *testing.T) {
	for _, sensitive := range []string{
		"https://example.com/callback?code=private-code",
		"postgres://user:private-password@example.com/db",
		"private@example.com",
		"Bearer private-token",
		"password=private-password",
		`{"client_secret":"private value with spaces"}`,
		"api_key='private value'",
		"access_token=private-token",
		"response body: private user content",
		"DETAIL: Key (email)=(private@example.com) already exists.",
		"-----BEGIN PRIVATE KEY-----\nprivate-key\n-----END PRIVATE KEY-----",
		"ghp_privateTokenValue",
	} {
		t.Run(sensitive, func(t *testing.T) {
			got := readableErrorMessage(fmt.Errorf("operation failed: %s", sensitive))
			if strings.Contains(got, "private") || !strings.Contains(got, redactedValue) || !strings.HasPrefix(got, "operation failed: ") {
				t.Fatalf("redacted message = %q", got)
			}
		})
	}
}

func TestReadableErrorBoundsOutputAndRecoversPanics(t *testing.T) {
	got := readableErrorMessage(errors.New(strings.Repeat("界", maxErrorMessageBytes)))
	if !utf8.ValidString(got) || !strings.HasSuffix(got, "… [truncated]") || len(got) > maxErrorMessageBytes+20 {
		t.Fatalf("error message was not bounded safely: bytes=%d", len(got))
	}
	if got := readableErrorMessage(panickingMessageError{}); got != "error message unavailable" {
		t.Fatalf("panicking error message = %q", got)
	}
}

func TestLoggerIncludesClassifiedErrorCause(t *testing.T) {
	definition := MustDefineError("api.configuration.invalid", "API configuration is invalid")
	err := definition.Wrap(errors.New("APP_DB_SSL_MODE must be verify-full in production"))
	var output bytes.Buffer
	log := NewWithJSON(&output, slog.LevelError, "logger-test")
	log.Error(context.Background(), "startup failed", "error", err)
	details := requireErrorDetails(t, decodeLogRecord(t, output.Bytes()), "error")
	causes, ok := details["causes"].([]any)
	if !ok || len(causes) != 1 || causes[0] != "APP_DB_SSL_MODE must be verify-full in production" {
		t.Fatalf("classified error causes = %#v", details["causes"])
	}
}

type panickingMessageError struct{}

func (panickingMessageError) Error() string { panic("private error contents") }
