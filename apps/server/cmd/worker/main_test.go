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

func TestLogWorkerFailureWritesSafeDiagnostic(t *testing.T) {
	t.Parallel()

	const rawCause = "dial postgres://admin:sensitive-password@private.example/customer"
	var output bytes.Buffer
	log := logger.NewWithJSON(&output, slog.LevelDebug, service)

	logWorkerFailure(
		context.Background(),
		log,
		"bootstrap",
		workerBootstrapFailure.Wrap(errors.New(rawCause)),
	)

	if strings.Contains(output.String(), rawCause) || strings.Contains(output.String(), "sensitive-password") {
		t.Fatalf("worker log exposed raw failure: %s", output.String())
	}
	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode worker log: %v", err)
	}
	if record["phase"] != "bootstrap" || record["version"] != version {
		t.Fatalf("worker diagnostic context = %#v", record)
	}
	errorDetails, ok := record["error"].(map[string]any)
	if !ok {
		t.Fatalf("worker error = %#v, want structured diagnostic", record["error"])
	}
	if errorDetails["code"] != "worker.bootstrap.failed" || errorDetails["safe_message"] != "Worker failed during startup" {
		t.Fatalf("worker error diagnostic = %#v", errorDetails)
	}
}
