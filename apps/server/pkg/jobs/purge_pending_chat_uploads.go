package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

const pendingChatUploadRetention = 24 * time.Hour
const pendingChatUploadBatchSize = 100

type PendingChatUploadStore interface {
	RetirePendingChatUploads(context.Context, attachmentdomain.PendingChatUploadRetention) (int, error)
}

// PurgePendingChatUploads bounds each minute's maintenance work. Larger backlogs
// drain over subsequent runs without monopolizing object-deletion delivery.
func PurgePendingChatUploads(ctx context.Context, store PendingChatUploadStore, objects RetainedAttachmentObjectStore, log *logger.Logger) error {
	return purgePendingChatUploadsAt(ctx, store, objects, log, time.Now().UTC())
}

func purgePendingChatUploadsAt(ctx context.Context, store PendingChatUploadStore, objects RetainedAttachmentObjectStore, log *logger.Logger, now time.Time) error {
	if store == nil || objects == nil || log == nil {
		return errors.New("pending chat upload retention dependencies are required")
	}
	provider, container, err := objects.RetainedObjectStorage()
	if err != nil {
		return fmt.Errorf("resolve pending upload storage route: %w", err)
	}
	count, err := store.RetirePendingChatUploads(ctx, attachmentdomain.PendingChatUploadRetention{CreatedBefore: now.UTC().Add(-pendingChatUploadRetention), RetiredAt: now.UTC(), Limit: pendingChatUploadBatchSize, StorageProvider: provider, ContainerName: container})
	if err != nil {
		return fmt.Errorf("retire expired pending chat uploads: %w", err)
	}
	if count > 0 {
		log.Info(ctx, "retired expired pending chat associations", "count", count)
	}
	return nil
}
