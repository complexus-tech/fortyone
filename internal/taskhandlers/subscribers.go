package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

// HandleSubscriberUpdate processes the subscriber update task.
func (h *handlers) HandleSubscriberUpdate(ctx context.Context, t *asynq.Task) error {
	var p tasks.SubscriberUpdatePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal SubscriberUpdatePayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing UserOnboardingStart task",
		"email", p.Email,
		"full_name", p.FullName,
		"task_id", t.ResultWriter().TaskID(),
	)

	// MailerLite integration
	subscriberID, err := h.mailerLiteService.CreateOrUpdateSubscriber(ctx, p.Email, p.FullName)
	if err != nil {
		return fmt.Errorf("failed to create/update subscriber: %w", err)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed SubscriberUpdate task", "email", p.Email, "subscriber_id", subscriberID)
	return nil
}
