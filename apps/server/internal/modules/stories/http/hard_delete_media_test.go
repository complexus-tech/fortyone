package storieshttp

import (
	"context"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCleanUpLegacyHardDeleteMediaSkipsDurablyDeferredObjects(t *testing.T) {
	t.Parallel()

	workspaceID, attachmentID := uuid.New(), uuid.New()
	deleter := &orphanedStoryMediaDeleterStub{}
	cleanUpLegacyHardDeleteMedia(
		context.Background(),
		deleter,
		nil,
		workspaceID,
		stories.HardBulkDeleteResult{
			OrphanedAttachmentIDs:            []uuid.UUID{attachmentID},
			AttachmentObjectDeletionDeferred: true,
		},
	)
	require.Empty(t, deleter.calls)
}

func TestCleanUpLegacyHardDeleteMediaPreservesCompatibilityPath(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	attachmentIDs := []uuid.UUID{uuid.New(), uuid.New()}
	deleter := &orphanedStoryMediaDeleterStub{}
	cleanUpLegacyHardDeleteMedia(
		context.Background(),
		deleter,
		nil,
		workspaceID,
		stories.HardBulkDeleteResult{OrphanedAttachmentIDs: attachmentIDs},
	)
	require.Equal(t, []orphanedStoryMediaDeleteCall{
		{attachmentID: attachmentIDs[0], workspaceID: workspaceID},
		{attachmentID: attachmentIDs[1], workspaceID: workspaceID},
	}, deleter.calls)
}

type orphanedStoryMediaDeleteCall struct {
	attachmentID uuid.UUID
	workspaceID  uuid.UUID
}

type orphanedStoryMediaDeleterStub struct {
	calls []orphanedStoryMediaDeleteCall
}

func (deleter *orphanedStoryMediaDeleterStub) DeleteOrphanedMedia(
	_ context.Context,
	attachmentID uuid.UUID,
	workspaceID uuid.UUID,
) error {
	deleter.calls = append(deleter.calls, orphanedStoryMediaDeleteCall{
		attachmentID: attachmentID,
		workspaceID:  workspaceID,
	})
	return nil
}
