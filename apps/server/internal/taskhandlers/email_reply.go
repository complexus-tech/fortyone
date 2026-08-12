package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleBrevoEmailReply(ctx context.Context, task *asynq.Task) error {
	var payload tasks.BrevoEmailReplyPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("decode Brevo email reply task: %w: %w", err, asynq.SkipRetry)
	}
	if h.emailReplies == nil {
		return errors.New("Brevo email reply processor is not configured")
	}
	if err := h.emailReplies.ProcessEvent(ctx, payload.ExternalWorkspaceID, payload.EventID); err != nil {
		return fmt.Errorf("process Brevo email reply: %w", err)
	}
	return nil
}

func (h *handlers) HandleBrevoEmailReplyRecovery(ctx context.Context, _ *asynq.Task) error {
	if h.emailRecovery == nil {
		return errors.New("Brevo email reply recovery is not configured")
	}
	recovered, err := h.emailRecovery.RecoverPendingEvents(ctx)
	if err != nil {
		return fmt.Errorf("recover Brevo email replies: %w", err)
	}
	if recovered > 0 && h.log != nil {
		h.log.Info(ctx, "Recovered Brevo email replies", "count", recovered)
	}
	return nil
}
