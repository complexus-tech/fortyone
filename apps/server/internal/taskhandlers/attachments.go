package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

func (h *handlers) HandleAttachmentImageOptimization(ctx context.Context, task *asynq.Task) error {
	var payload tasks.AttachmentImageOptimizationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal attachment image optimization payload: %w: %w", err, asynq.SkipRetry)
	}
	if strings.TrimSpace(payload.BlobName) == "" {
		return fmt.Errorf("attachment image optimization blob name is required: %w", asynq.SkipRetry)
	}
	if h.attachments == nil {
		return fmt.Errorf("attachments service is required: %w", asynq.SkipRetry)
	}

	err := h.attachments.OptimizeStoredAttachment(ctx, payload.BlobName)
	if err == nil {
		h.log.Info(ctx, "attachment image optimized", "blob_name", payload.BlobName)
		return nil
	}

	if errors.Is(err, attachments.ErrImageOptimizationNotApplicable) ||
		errors.Is(err, attachments.ErrImageOptimizationSkipped) {
		h.log.Info(ctx, "attachment image optimization skipped", "blob_name", payload.BlobName, "reason", err)
		return nil
	}

	h.log.Error(ctx, "attachment image optimization failed", "blob_name", payload.BlobName, "error", err)
	return fmt.Errorf("optimize attachment image: %w", err)
}
