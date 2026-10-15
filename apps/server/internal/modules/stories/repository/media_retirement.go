package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

type mediaRetirementQueries interface {
	DeleteUnreferencedStoryRetentionAttachments(context.Context, storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams) ([]storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow, error)
	InsertAttachmentObjectDeletionOutbox(context.Context, storyreadsql.InsertAttachmentObjectDeletionOutboxParams) (int64, error)
}

// The caller owns the mutation transaction. Rechecking every reference before
// retiring metadata prevents a shared file from being deleted; the existing
// fenced worker handles object deletion after commit, including process crashes.
func retireReconciledStoryMedia(ctx context.Context, queries mediaRetirementQueries, route attachmentObjectStorageRoute, workspaceID uuid.UUID, candidates []uuid.UUID, now time.Time) error {
	route, err := validateInteractiveHardDeleteStorageRoute(route)
	if err != nil {
		return err
	}
	expected := make(map[uuid.UUID]struct{}, len(candidates))
	for _, id := range candidates {
		expected[id] = struct{}{}
	}
	retired, err := queries.DeleteUnreferencedStoryRetentionAttachments(ctx, storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsParams{AttachmentIds: candidates})
	if err != nil {
		return fmt.Errorf("retire reconciled story media: %w", err)
	}
	seen := make(map[uuid.UUID]struct{}, len(retired))
	for _, attachment := range retired {
		_, candidate := expected[attachment.AttachmentID]
		_, duplicate := seen[attachment.AttachmentID]
		if attachment.AttachmentID == uuid.Nil || attachment.WorkspaceID != workspaceID || !candidate || duplicate || strings.TrimSpace(attachment.BlobName) == "" || len(attachment.BlobName) > maximumBlobNameLength {
			return errors.New("reconciled story media returned invalid storage metadata")
		}
		seen[attachment.AttachmentID] = struct{}{}
		inserted, err := queries.InsertAttachmentObjectDeletionOutbox(ctx, storyreadsql.InsertAttachmentObjectDeletionOutboxParams{
			AttachmentID: attachment.AttachmentID, WorkspaceID: workspaceID, StorageProvider: route.provider, ContainerName: route.container, BlobName: attachment.BlobName, EnqueuedAt: now.UTC(),
		})
		if err != nil {
			return fmt.Errorf("enqueue reconciled story media deletion: %w", err)
		}
		if inserted != 1 {
			return fmt.Errorf("enqueue reconciled story media deletion: inserted %d rows, want 1", inserted)
		}
	}
	return nil
}
