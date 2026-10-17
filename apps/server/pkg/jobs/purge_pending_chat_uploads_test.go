package jobs

import (
	"context"
	"testing"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/stretchr/testify/require"
)

func TestPendingChatUploadCleanupUses24HourCutoffAndOneBoundedBatch(t *testing.T) {
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.FixedZone("CAT", 7200))
	store := &pendingChatUploadStub{}
	objects := &retainedAttachmentObjectStoreStub{provider: "aws", container: "attachments"}
	require.NoError(t, purgePendingChatUploadsAt(t.Context(), store, objects, newTestJobLogger(), now))
	require.Equal(t, 1, store.calls)
	require.Equal(t, now.UTC().Add(-24*time.Hour), store.batch.CreatedBefore)
	require.Equal(t, now.UTC(), store.batch.RetiredAt)
	require.Equal(t, 100, store.batch.Limit)
	require.Equal(t, "aws", store.batch.StorageProvider)
	require.Equal(t, "attachments", store.batch.ContainerName)
	require.Empty(t, objects.deletedBlobNames)
}

type pendingChatUploadStub struct {
	calls int
	batch attachmentdomain.PendingChatUploadRetention
}

func (s *pendingChatUploadStub) RetirePendingChatUploads(_ context.Context, batch attachmentdomain.PendingChatUploadRetention) (int, error) {
	s.calls++
	s.batch = batch
	return batch.Limit, nil
}
