package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	figmaprovider "github.com/complexus-tech/projects-api/internal/modules/figma"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleFigmaWebhook(ctx context.Context, task *asynq.Task) error {
	if h.figmaWebhooks == nil {
		return fmt.Errorf("figma webhook processor is not configured: %w", asynq.SkipRetry)
	}
	var payload tasks.FigmaWebhookPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode Figma webhook task: %w: %w", err, asynq.SkipRetry)
	}
	if payload.InboxID == uuid.Nil {
		return fmt.Errorf("figma webhook task has an invalid inbox payload: %w", asynq.SkipRetry)
	}
	if err := h.figmaWebhooks.ProcessWebhook(
		ctx,
		figmaprovider.ProviderKey,
		payload.InboxID,
	); err != nil {
		return fmt.Errorf("process Figma webhook inbox %s: %w", payload.InboxID, err)
	}
	return nil
}

func (h *handlers) HandleFigmaWebhookRecovery(ctx context.Context, _ *asynq.Task) error {
	if h.figmaRecovery == nil {
		return fmt.Errorf("figma inbox recoverer is not configured: %w", asynq.SkipRetry)
	}
	recovered, err := h.figmaRecovery.RecoverPendingWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("recover pending Figma webhooks: %w", err)
	}
	if recovered > 0 && h.log != nil {
		h.log.Info(ctx, "recovered pending Figma webhooks", "recovered", recovered)
	}
	return nil
}
