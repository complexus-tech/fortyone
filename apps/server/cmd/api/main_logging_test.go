package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
)

func TestLogAPIFailureWritesSafeDiagnostic(t *testing.T) {
	t.Parallel()

	const rawCause = "provider response contained bearer-token and private@example.com"
	var output bytes.Buffer
	log := logger.NewWithJSON(&output, slog.LevelDebug, service)

	logAPIFailure(
		context.Background(),
		log,
		"migration",
		apiMigrationFailure.Wrap(errors.New(rawCause)),
	)

	if strings.Contains(output.String(), rawCause) || strings.Contains(output.String(), "bearer-token") {
		t.Fatalf("API log exposed raw failure: %s", output.String())
	}
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode API log: %v", err)
	}
	if record["phase"] != "migration" || record["version"] != version {
		t.Fatalf("API diagnostic context = %#v", record)
	}
	errorDetails, ok := record["error"].(map[string]any)
	if !ok {
		t.Fatalf("API error = %#v, want structured diagnostic", record["error"])
	}
	if errorDetails["code"] != "api.migrations.failed" || errorDetails["safe_message"] != "Database migrations failed" {
		t.Fatalf("API error diagnostic = %#v", errorDetails)
	}
}

func TestAPIFallbackPreservesSubsystemDiagnostic(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	log := logger.NewWithJSON(&output, slog.LevelDebug, service)
	err := apiRedisUnavailableFailure.Wrap(errors.New("redis.internal:6379 authentication failed"))

	logAPIFailure(context.Background(), log, "runtime", apiRuntimeFailure.WrapIfUnclassified(err))

	var record map[string]any
	if decodeErr := json.Unmarshal(output.Bytes(), &record); decodeErr != nil {
		t.Fatalf("decode API log: %v", decodeErr)
	}
	errorDetails, ok := record["error"].(map[string]any)
	if !ok {
		t.Fatalf("API error = %#v, want structured diagnostic", record["error"])
	}
	if errorDetails["code"] != "api.redis.unavailable" {
		t.Fatalf("API subsystem error code = %v", errorDetails["code"])
	}
}
