package taskhandlers

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleAttachmentThumbnail(ctx context.Context, task *asynq.Task) error {
	var payload tasks.AttachmentThumbnailPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("invalid thumbnail job: %w", asynq.SkipRetry)
	}
	version, err := hex.DecodeString(payload.SourceVersion)
	if err != nil || len(version) != 32 || payload.AttachmentID == uuid.Nil || payload.WorkspaceID == uuid.Nil || h.attachments == nil {
		return fmt.Errorf("invalid thumbnail job identity: %w", asynq.SkipRetry)
	}
	return h.attachments.GenerateThumbnail(ctx, payload)
}
