package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeAttachmentImageOptimization = "attachments:image:optimize"

type AttachmentImageOptimizationPayload struct {
	AttachmentID uuid.UUID `json:"attachmentId"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`
}

func (s *Service) EnqueueAttachmentImageOptimization(payload AttachmentImageOptimizationPayload) error {
	ctx := context.Background()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("tasks: marshal %s payload: %w", TypeAttachmentImageOptimization, err)
	}

	task := asynq.NewTask(
		TypeAttachmentImageOptimization,
		payloadBytes,
		asynq.Queue("automation"),
		asynq.MaxRetry(5),
	)

	info, err := s.asynqClient.Enqueue(task)
	if err != nil {
		s.log.Error(ctx, "failed to enqueue attachment image optimization", "error", err, "attachment_id", payload.AttachmentID)
		return fmt.Errorf("tasks: enqueue %s task: %w", TypeAttachmentImageOptimization, err)
	}

	s.log.Info(ctx, "attachment image optimization enqueued", "task_id", info.ID, "attachment_id", payload.AttachmentID)
	return nil
}
