package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeNotificationEmail       = "notification:email:send"
	TypeNotificationEmailDigest = "notification:email:digest"
	TypeWeeklyDigestEmail       = "email:digest:weekly"
	TypeFeedbackDigestEmail     = "feedback:email:digest"

	notificationEmailDigestDelay = time.Hour
	// The uniqueness lock must outlive the scheduled delay. Asynq releases it
	// immediately after successful processing, so this does not extend the next digest.
	notificationEmailDigestUniqueTTL = 2 * time.Hour
)

type NotificationEmailPayload struct {
	NotificationID uuid.UUID `json:"notificationId"`
	RecipientID    uuid.UUID `json:"recipientId"`
	WorkspaceID    uuid.UUID `json:"workspaceId"`
}

type NotificationEmailDigestPayload struct {
	RecipientID uuid.UUID `json:"recipientId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
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
		asynq.ProcessIn(time.Hour), // 1-hour delay
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

// EnqueueNotificationEmailDigest enqueues one coalesced email task per recipient and workspace.
func (s *Service) EnqueueNotificationEmailDigest(payload NotificationEmailDigestPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	ctx := context.Background()
	s.log.Info(ctx, "Attempting to enqueue NotificationEmailDigest task",
		"recipient_id", payload.RecipientID,
		"workspace_id", payload.WorkspaceID)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.log.Error(ctx, "Failed to marshal NotificationEmailDigestPayload", "error", err,
			"recipient_id", payload.RecipientID,
			"workspace_id", payload.WorkspaceID)
		return nil, fmt.Errorf("tasks: failed to marshal %s payload: %w", TypeNotificationEmailDigest, err)
	}

	defaultOpts := []asynq.Option{
		asynq.Queue("notifications"),
		asynq.MaxRetry(2),
		asynq.ProcessIn(notificationEmailDigestDelay),
		asynq.Unique(notificationEmailDigestUniqueTTL),
	}

	finalOpts := append(defaultOpts, opts...)
	task := asynq.NewTask(TypeNotificationEmailDigest, payloadBytes, finalOpts...)

	info, err := s.asynqClient.Enqueue(task)
	if errors.Is(err, asynq.ErrDuplicateTask) {
		s.log.Info(ctx, "NotificationEmailDigest task already queued",
			"recipient_id", payload.RecipientID,
			"workspace_id", payload.WorkspaceID)
		return nil, nil
	}
	if err != nil {
		s.log.Error(ctx, "Failed to enqueue NotificationEmailDigest task", "error", err,
			"recipient_id", payload.RecipientID,
			"workspace_id", payload.WorkspaceID)
		return nil, fmt.Errorf("tasks: failed to enqueue %s task: %w", TypeNotificationEmailDigest, err)
	}

	s.log.Info(ctx, "Successfully enqueued NotificationEmailDigest task",
		"task_id", info.ID,
		"queue", info.Queue,
		"recipient_id", payload.RecipientID,
		"workspace_id", payload.WorkspaceID,
		"process_in", notificationEmailDigestDelay.String())
	return info, nil
}
