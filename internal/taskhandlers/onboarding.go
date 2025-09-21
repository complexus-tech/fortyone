package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

var (
	BrevoOnboardingList = int64(6)
	BrevoTrialList      = int64(12)
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

	// Brevo integration
	_, err := h.brevoService.CreateOrUpdateContact(ctx, brevo.CreateOrUpdateContactRequest{
		Email: p.Email,
		Attributes: map[string]any{
			"NAME": p.FullName,
		},
		ListIDs: []int64{BrevoOnboardingList},
	})
	if err != nil {
		return fmt.Errorf("failed to create/update contact: %w", err)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed UserOnboardingStart task", "user_id", p.UserID)
	return nil
}
