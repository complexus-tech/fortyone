package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

// HandleUserOnboardingStart processes the user onboarding start task.
func (h *handlers) HandleUserOnboardingStart(ctx context.Context, t *asynq.Task) error {
	var p tasks.UserOnboardingStartPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal UserOnboardingStartPayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing UserOnboardingStart task",
		"user_id", p.UserID,
		"email", p.Email,
		"full_name", p.FullName,
		"task_id", t.ResultWriter().TaskID(),
	)

	// MailerLite integration
	if err := h.mailerLiteService.AddToOnboardingGroup(ctx, p.Email, p.FullName); err != nil {
		h.log.Error(ctx, "MailerLite integration failed",
			"error", err,
			"email", p.Email,
			"user_id", p.UserID,
			"task_id", t.ResultWriter().TaskID(),
		)
	} else {
		h.log.Info(ctx, "Successfully added user to MailerLite onboarding group",
			"email", p.Email,
			"user_id", p.UserID,
			"task_id", t.ResultWriter().TaskID(),
		)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed UserOnboardingStart task", "user_id", p.UserID)
	return nil
}
