package storieshttp

import (
	"context"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type orphanedStoryMediaDeleter interface {
	DeleteOrphanedMedia(context.Context, uuid.UUID, uuid.UUID) error
}

// cleanUpLegacyHardDeleteMedia preserves the compatibility path for adapters
// that still return media for post-transaction deletion. The SQLC adapter sets
// the deferred flag because its transaction already retired metadata and
// durably enqueued object deletion.
func cleanUpLegacyHardDeleteMedia(
	ctx context.Context,
	deleter orphanedStoryMediaDeleter,
	log *logger.Logger,
	workspaceID uuid.UUID,
	result stories.HardBulkDeleteResult,
) {
	if result.AttachmentObjectDeletionDeferred || deleter == nil {
		return
	}
	for _, attachmentID := range result.OrphanedAttachmentIDs {
		if err := deleter.DeleteOrphanedMedia(ctx, attachmentID, workspaceID); err != nil && log != nil {
			log.Error(ctx, "failed to clean up permanently deleted story media", "error", err, "attachment_id", attachmentID)
		}
	}
}
