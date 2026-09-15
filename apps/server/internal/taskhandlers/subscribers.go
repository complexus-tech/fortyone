package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

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
