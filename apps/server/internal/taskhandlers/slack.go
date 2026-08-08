package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleSlackEvent(ctx context.Context, task *asynq.Task) error {
	if h.slackEvents == nil {
		return fmt.Errorf("Slack event processor is not configured: %w", asynq.SkipRetry)
	}

	var payload tasks.SlackEventPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode Slack event task: %w: %w", err, asynq.SkipRetry)
	}
	if strings.TrimSpace(payload.ExternalWorkspaceID) == "" || strings.TrimSpace(payload.EventID) == "" {
		return fmt.Errorf("Slack event task has an invalid payload: %w", asynq.SkipRetry)
	}

	if err := h.slackEvents.ProcessEvent(ctx, payload.ExternalWorkspaceID, payload.EventID); err != nil {
		h.log.Error(ctx, "Slack event processing failed", "event_id", payload.EventID, "error", err)
		return fmt.Errorf("process Slack event %s: %w", payload.EventID, err)
	}
	h.log.Info(ctx, "Slack event processed", "event_id", payload.EventID)
	return nil
}

func (h *handlers) HandleSlackCredentialBackfill(ctx context.Context, task *asynq.Task) error {
	if h.slackCredentials == nil {
		return fmt.Errorf("Slack credential backfiller is not configured: %w", asynq.SkipRetry)
	}
	upgraded, err := h.slackCredentials.BackfillLegacyCredentials(ctx)
	if err != nil {
		return fmt.Errorf("backfill legacy Slack credentials: %w", err)
	}
	h.log.Info(ctx, "Slack credential backfill completed", "upgraded", upgraded)
	return nil
}

func (h *handlers) HandleSlackInboxRecovery(ctx context.Context, task *asynq.Task) error {
	if h.slackRecovery == nil {
		return fmt.Errorf("Slack inbox recoverer is not configured: %w", asynq.SkipRetry)
	}
	recovered, err := h.slackRecovery.RecoverPendingEvents(ctx)
	if err != nil {
		return fmt.Errorf("recover pending Slack events: %w", err)
	}
	if recovered > 0 {
		h.log.Info(ctx, "Recovered pending Slack events", "recovered", recovered)
	}
	return nil
}
