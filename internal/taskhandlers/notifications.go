package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

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
	// 4. Send templated email using Brevo service

	// Example of how to use Brevo templated email (commented out for now):
	/*
		if h.brevoService != nil {
			// Example using the helper function for overdue story notification
			storyParams := brevo.StoryNotificationParams{
				NotificationEmailParams: brevo.NotificationEmailParams{
					UserName:      "John Doe",
					UserEmail:     "user@example.com",
					WorkspaceName: "My Workspace",
					WorkspaceURL:  "https://app.complexus.com/workspace/123",
					AppURL:        "https://app.complexus.com",
				},
				StoryTitle:    "Fix login bug",
				StoryURL:      "https://app.complexus.com/stories/123",
				DueDate:       "2024-01-15",
				DaysOverdue:   3,
				ProjectName:   "Authentication Project",
			}

			_, err := h.brevoService.SendOverdueStoryNotification(ctx, storyParams)
			if err != nil {
				h.log.Error(ctx, "Failed to send overdue story notification via Brevo",
					"error", err,
					"notification_id", p.NotificationID)
				// Don't return error - we don't want to retry email sending
			} else {
				h.log.Info(ctx, "Successfully sent overdue story notification via Brevo",
					"notification_id", p.NotificationID)
			}

			// Alternative: Using the generic SendNotificationEmail function
			// templateParams := map[string]interface{}{
			//     "USER_NAME":    "John Doe",
			//     "STORY_TITLE":  "Fix login bug",
			//     "DUE_DATE":     "2024-01-15",
			//     "DAYS_OVERDUE": 3,
			//     "STORY_URL":    "https://app.complexus.com/stories/123",
			// }
			//
			// _, err := h.brevoService.SendNotificationEmail(ctx,
			//     brevo.TemplateOverdueStory,
			//     "user@example.com",
			//     "John Doe",
			//     templateParams,
			//     []string{"notification", "overdue-story"})
		}
	*/

	// For now, just log that we would send an email
	h.log.Info(ctx, "Would send notification email",
		"notification_id", p.NotificationID,
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"task_id", t.ResultWriter().TaskID(),
		"action", "email_would_be_sent",
	)

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmail task",
		"notification_id", p.NotificationID,
		"task_id", t.ResultWriter().TaskID())
	return nil
}
