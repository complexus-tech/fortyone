package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

// HandleNotificationEmail processes the notification email task.
func (h *handlers) HandleNotificationEmail(ctx context.Context, t *asynq.Task) error {
	var p tasks.NotificationEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal NotificationEmailPayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing NotificationEmail task",
		"notification_id", p.NotificationID,
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"task_id", t.ResultWriter().TaskID(),
	)

	// TODO: In a real implementation, we would:
	// 1. Get notification from database to check if it's still unread
	// 2. Get user's email preferences for this notification type
	// 3. Get user's email address and name
	if err := h.brevoService.SendEmailNotification(ctx, 3, brevo.EmailNotificationParams{
		UserName:      "John Doe",
		UserEmail:     "user@example.com",
		WorkspaceName: "My Workspace",
		WorkspaceURL:  "https://app.complexus.com/workspace/123",
	}); err != nil {
		h.log.Error(ctx, "Failed to send notification email", "error", err, "task_id", t.ResultWriter().TaskID())
		return nil
	}

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmail task",
		"notification_id", p.NotificationID,
		"task_id", t.ResultWriter().TaskID())
	return nil
}
