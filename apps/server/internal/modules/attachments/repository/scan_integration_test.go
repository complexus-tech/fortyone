//go:build integration

package attachmentsrepository

import (
	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateAttachmentPersistsCleanScanEvidence(t *testing.T) {
	database := testkit.NewPostgres(t)
	fixture := newAttachmentFixture(t, t.Context(), database.Pool)
	repo := New(database.Pool)
	scannedAt := time.Now().UTC().Truncate(time.Microsecond)
	file, err := repo.CreateAttachment(t.Context(), attachmentdomain.Attachment{
		Filename: "recording.webm", BlobName: uuid.NewString() + ".webm", Size: 128, MimeType: "audio/webm", UploadedBy: fixture.userA, WorkspaceID: fixture.workspaceA,
		ScanStatus: attachmentdomain.ScanStatusClean, ScanCompletedAt: &scannedAt,
	})
	require.NoError(t, err)
	require.Equal(t, attachmentdomain.ScanStatusClean, file.ScanStatus)
	require.NotNil(t, file.ScanCompletedAt)
	require.True(t, scannedAt.Equal(*file.ScanCompletedAt))
	stored, err := repo.GetAttachmentByID(t.Context(), file.ID, fixture.workspaceA)
	require.NoError(t, err)
	require.Equal(t, attachmentdomain.ScanStatusClean, stored.ScanStatus)
	require.NotNil(t, stored.ScanCompletedAt)
	require.True(t, scannedAt.Equal(*stored.ScanCompletedAt))
	unscanned := createFixtureAttachment(t, t.Context(), repo, fixture.workspaceA, fixture.userA, attachmentdomain.OptimizationNotRequested)
	require.Equal(t, attachmentdomain.ScanStatusUnscanned, unscanned.ScanStatus)
	require.Nil(t, unscanned.ScanCompletedAt, "ordinary uploads must not manufacture a scan timestamp")
}
