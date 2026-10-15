package storiesrepository

import (
	"errors"
	"testing"
	"time"

	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReconciledMediaDeletionRetiresOnlyUnreferencedFilesAndDurablyQueuesObjects(t *testing.T) {
	workspace, shared, orphan := uuid.New(), uuid.New(), uuid.New()
	queries := &interactiveHardDeleteQueryStub{retiredRows: []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{{AttachmentID: orphan, WorkspaceID: workspace, BlobName: "opaque-object"}}, insertRows: 1}
	now := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	require.NoError(t, retireReconciledStoryMedia(t.Context(), queries, attachmentObjectStorageRoute{provider: "aws", container: "attachments"}, workspace, []uuid.UUID{shared, orphan}, now))
	require.Equal(t, []uuid.UUID{shared, orphan}, queries.deleteAttachmentParams.AttachmentIds)
	require.Equal(t, []string{"delete_attachments", "insert_outbox"}, queries.calls)
	require.Len(t, queries.insertParams, 1)
	require.Equal(t, orphan, queries.insertParams[0].AttachmentID)
	require.Equal(t, workspace, queries.insertParams[0].WorkspaceID)
	require.Equal(t, "opaque-object", queries.insertParams[0].BlobName)
	require.Equal(t, now, queries.insertParams[0].EnqueuedAt)
}

func TestReconciledMediaOutboxFailureAbortsTheOwningMutation(t *testing.T) {
	workspace, attachment := uuid.New(), uuid.New()
	unavailable := errors.New("outbox unavailable")
	for _, test := range []struct {
		name      string
		row       storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow
		insertErr error
	}{
		{"enqueue failure", storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{AttachmentID: attachment, WorkspaceID: workspace, BlobName: "opaque"}, unavailable},
		{"foreign workspace", storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{AttachmentID: attachment, WorkspaceID: uuid.New(), BlobName: "opaque"}, nil},
		{"unexpected attachment", storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{AttachmentID: uuid.New(), WorkspaceID: workspace, BlobName: "opaque"}, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			queries := &interactiveHardDeleteQueryStub{retiredRows: []storyreadsql.DeleteUnreferencedStoryRetentionAttachmentsRow{test.row}, insertRows: 1, insertErr: test.insertErr}
			err := retireReconciledStoryMedia(t.Context(), queries, attachmentObjectStorageRoute{provider: "aws", container: "attachments"}, workspace, []uuid.UUID{attachment}, time.Now())
			require.Error(t, err)
			if test.insertErr != nil {
				require.ErrorIs(t, err, unavailable)
			} else {
				require.Empty(t, queries.insertParams)
			}
		})
	}
}
