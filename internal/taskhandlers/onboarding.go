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

	// TODO: Implement MailerLite integration here.
	// When the mailerliteClient is added to the Handlers struct and initialized:
	// 1. Use h.mailerliteClient to interact with the MailerLite API.
	// 2. Add user (p.Email, p.FullName) to the appropriate MailerLite group/workflow.
	//    Example (conceptual, actual SDK calls will vary):
	//    subscriberData := mailerlite.Subscriber{ Email: p.Email, Fields: map[string]string{"name": p.FullName} }
	//    _, err := h.mailerliteClient.Subscribers.Create(ctx, subscriberData)
	//    if err != nil {
	//        h.log.Error(ctx, "MailerLite subscriber creation failed", "error", err, "email", p.Email)
	//        return err // Potentially retryable
	//    }

	h.log.Info(ctx, "HANDLER: Successfully processed UserOnboardingStart task (simulation)", "user_id", p.UserID)
	return nil
}
