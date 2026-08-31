package workerbootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBuildSlackWebhookRuntimeClassifiesMissingSigningSecret(t *testing.T) {
	t.Parallel()

	pool := new(pgxpool.Pool)
	_, _, err := buildSlackWebhookRuntime(
		pool,
		slackrepository.New(pool),
		new(tasks.Service),
		slack.Config{WebhookPayloadSecret: "reviewed-test-payload-secret"},
	)
	if !errors.Is(err, slack.ErrSlackSigningSecretNotConfigured) {
		t.Fatalf("build Slack webhook runtime error = %v", err)
	}

	var output bytes.Buffer
	log := logger.NewWithJSON(&output, slog.LevelDebug, "worker-test")
	log.Error(context.Background(), "Slack initialization failed", "error", err)

	var record map[string]any
	if decodeErr := json.Unmarshal(output.Bytes(), &record); decodeErr != nil {
		t.Fatalf("decode worker diagnostic: %v", decodeErr)
	}
	errorDetails, ok := record["error"].(map[string]any)
	if !ok {
		t.Fatalf("error = %#v, want structured diagnostic", record["error"])
	}
	if errorDetails["code"] != "worker.slack.signing_secret_missing" {
		t.Fatalf("error code = %v", errorDetails["code"])
	}
	if errorDetails["safe_message"] != "SLACK_SIGNING_SECRET is not configured" {
		t.Fatalf("safe message = %v", errorDetails["safe_message"])
	}
}
