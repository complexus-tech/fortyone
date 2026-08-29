package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	slackprovider "github.com/complexus-tech/projects-api/internal/modules/slack"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleSlackEvent(ctx context.Context, task *asynq.Task) error {
	if h.slackEvents == nil {
		return fmt.Errorf("slack event processor is not configured: %w", asynq.SkipRetry)
	}

	var payload tasks.SlackEventPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode Slack event task: %w: %w", err, asynq.SkipRetry)
	}
	provider := integrations.ProviderKey(strings.TrimSpace(payload.Provider))
	if provider != "" || payload.InboxID != uuid.Nil {
		if provider != slackprovider.ProviderKey || payload.InboxID == uuid.Nil ||
			strings.TrimSpace(payload.ExternalWorkspaceID) != "" || strings.TrimSpace(payload.EventID) != "" || payload.RecoveryAttempt != 0 {
			return fmt.Errorf("slack webhook task has an invalid inbox payload: %w", asynq.SkipRetry)
		}
		if err := h.slackEvents.ProcessWebhook(ctx, provider, payload.InboxID); err != nil {
			h.log.Error(ctx, "Slack webhook processing failed", "provider", provider, "inbox_id", payload.InboxID, "error", err)
			return fmt.Errorf("process Slack webhook inbox %s: %w", payload.InboxID, err)
		}
		h.log.Info(ctx, "Slack webhook processed", "provider", provider, "inbox_id", payload.InboxID)
		return nil
	}

	externalWorkspaceID := strings.TrimSpace(payload.ExternalWorkspaceID)
	eventID := strings.TrimSpace(payload.EventID)
	legacyProcessor, compatible := h.slackEvents.(legacySlackEventProcessor)
	if !compatible || externalWorkspaceID == "" || eventID == "" || payload.RecoveryAttempt < 0 {
		return fmt.Errorf("slack event task has an invalid legacy payload: %w", asynq.SkipRetry)
	}
	if err := legacyProcessor.ProcessEvent(ctx, externalWorkspaceID, eventID); err != nil {
		h.log.Error(ctx, "Slack event processing failed", "event_id", payload.EventID, "error", err)
		return fmt.Errorf("process Slack event %s: %w", eventID, err)
	}
	h.log.Info(ctx, "Slack event processed", "event_id", eventID, "legacy_task", true)
	return nil
}

func (h *handlers) HandleSlackCredentialBackfill(ctx context.Context, task *asynq.Task) error {
	if h.slackCredentials == nil {
		return fmt.Errorf("slack credential backfiller is not configured: %w", asynq.SkipRetry)
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
		return fmt.Errorf("slack inbox recoverer is not configured: %w", asynq.SkipRetry)
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
