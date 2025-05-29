package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeNotificationEmail = "notification:email:send"

type NotificationEmailPayload struct {
	NotificationID uuid.UUID `json:"notificationId"`
	RecipientID    uuid.UUID `json:"recipientId"`
	WorkspaceID    uuid.UUID `json:"workspaceId"`
}

// EnqueueNotificationEmail enqueues a task to send an email notification after 1 hour delay.
func (s *Service) EnqueueNotificationEmail(payload NotificationEmailPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue NotificationEmail task",
		"notification_id", payload.NotificationID,
		"recipient_id", payload.RecipientID,
		"workspace_id", payload.WorkspaceID)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal NotificationEmailPayload", "error", err,
			"notification_id", payload.NotificationID)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeNotificationEmail, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("notifications"),
		asynq.MaxRetry(2),
		// asynq.ProcessIn(time.Hour), // 1-hour delay
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeNotificationEmail, payloadBytes, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue NotificationEmail task", "error", err,
			"notification_id", payload.NotificationID)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeNotificationEmail, err)
	}

	s.log.Info(ctx, "Successfully enqueued NotificationEmail task",
		"task_id", info.ID,
		"queue", info.Queue,
		"notification_id", payload.NotificationID,
		"process_in", "1 hour")
	return info, nil
}
