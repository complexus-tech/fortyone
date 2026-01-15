package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/brevo"
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

	// Brevo integration
	_, err := h.brevoService.CreateOrUpdateContact(ctx, brevo.CreateOrUpdateContactRequest{
		Email: p.Email,
		Attributes: map[string]any{
			"NAME": p.FullName,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create/update contact: %w", err)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed ContactUpdate task", "email", p.Email)
	return nil
}

// HandleSubscriberDelete processes the subscriber delete task.
func (h *handlers) HandleSubscriberDelete(ctx context.Context, t *asynq.Task) error {
	var p tasks.SubscriberDeletePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal SubscriberDeletePayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing SubscriberDelete task", "email", p.Email, "task_id", t.ResultWriter().TaskID())

	// Brevo integration
	err := h.brevoService.DeleteContact(ctx, brevo.DeleteContactRequest{
		Email: p.Email,
	})
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	h.log.Info(ctx, "HANDLER: Successfully processed SubscriberDelete task", "email", p.Email)
	return nil
}
