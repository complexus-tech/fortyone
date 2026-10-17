package tasks

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeAttachmentThumbnail = "attachments:thumbnail:generate"

type AttachmentThumbnailPayload struct {
	AttachmentID  uuid.UUID `json:"attachmentId"`
	WorkspaceID   uuid.UUID `json:"workspaceId"`
	SourceVersion string    `json:"sourceVersion"`
}

func (s *Service) EnqueueAttachmentThumbnail(payload AttachmentThumbnailPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode attachment thumbnail job: %w", err)
	}
	_, err = s.asynqClient.Enqueue(asynq.NewTask(TypeAttachmentThumbnail, data),
		asynq.Queue("automation"), asynq.MaxRetry(3), asynq.Timeout(time.Minute), asynq.Unique(time.Minute))
	if errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	return err
}
