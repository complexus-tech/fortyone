package attachments

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
	"os/exec"
	"testing"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type thumbnailRepositoryStub struct {
	attachmentRepositoryStub
	item        attachmentdomain.Thumbnail
	claimed     int
	retired     int
	completeErr error
}

func (r *thumbnailRepositoryStub) EnsureThumbnail(_ context.Context, item attachmentdomain.Thumbnail) (attachmentdomain.Thumbnail, error) {
	if r.item.ID != item.ID {
		r.item = item
		r.item.Status = "pending"
	}
	return r.item, nil
}
func (r *thumbnailRepositoryStub) ClaimThumbnail(_ context.Context, id, workspaceID uuid.UUID) (bool, error) {
	if id != r.item.ID || workspaceID != r.item.WorkspaceID || r.item.Status != "pending" {
		return false, nil
	}
	r.item.Status = "processing"
	r.claimed++
	return true, nil
}
func (r *thumbnailRepositoryStub) CompleteThumbnail(_ context.Context, _ attachmentdomain.Thumbnail, _ time.Time, size int32) error {
	if r.completeErr != nil {
		return r.completeErr
	}
	r.item.Status = "ready"
	r.item.ByteSize = &size
	return nil
}
func (r *thumbnailRepositoryStub) FailThumbnail(_ context.Context, _, _ uuid.UUID) error {
	r.item.Status = "failed"
	next := time.Now().Add(5 * time.Minute)
	r.item.RetryAfter = &next
	return nil
}
func (r *thumbnailRepositoryStub) ResetThumbnail(_ context.Context, _, _ uuid.UUID) error {
	r.item.Status = "pending"
	return nil
}
func (r *thumbnailRepositoryStub) RetireThumbnail(_ context.Context, _ attachmentdomain.Thumbnail) error {
	r.retired++
	return nil
}

type thumbnailQueueStub struct {
	imageOptimizerStub
	payloads []tasks.AttachmentThumbnailPayload
}

func (q *thumbnailQueueStub) EnqueueAttachmentThumbnail(payload tasks.AttachmentThumbnailPayload) error {
	q.payloads = append(q.payloads, payload)
	return nil
}

func thumbnailFixture(t *testing.T) (*Service, *thumbnailRepositoryStub, *attachmentStorageStub, *thumbnailQueueStub) {
	t.Helper()
	var imageData bytes.Buffer
	require.NoError(t, jpeg.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 1600, 900)), nil))
	file := CoreAttachment{ID: uuid.New(), WorkspaceID: uuid.New(), BlobName: "original.jpg", Size: int64(imageData.Len()), MimeType: "image/jpeg", UpdatedAt: time.Now(), ScanStatus: attachmentdomain.ScanStatusClean}
	repo := &thumbnailRepositoryStub{attachmentRepositoryStub: attachmentRepositoryStub{attachment: file}}
	objects := &attachmentStorageStub{data: imageData.Bytes(), contentType: "image/jpeg"}
	queue := &thumbnailQueueStub{}
	service := New(nil, repo, objects, storage.Config{Provider: "aws", AttachmentsBucket: "private"}, queue)
	return service, repo, objects, queue
}

func TestThumbnailGeneratedOncePerVersionAndReusedWithoutSourceDownload(t *testing.T) {
	s, repo, objects, queue := thumbnailFixture(t)
	file := repo.attachment
	_, err := s.ResolveThumbnail(t.Context(), file.ID, file.WorkspaceID, false)
	require.ErrorIs(t, err, ErrThumbnailPending)
	require.Len(t, queue.payloads, 1)
	require.NoError(t, s.GenerateThumbnail(t.Context(), queue.payloads[0]))
	require.Equal(t, 1, objects.uploadCount)
	config, _, err := image.DecodeConfig(bytes.NewReader(objects.data))
	require.NoError(t, err)
	require.Equal(t, 840, config.Width)
	require.LessOrEqual(t, config.Height, 840)
	require.NoError(t, s.GenerateThumbnail(t.Context(), queue.payloads[0]))
	access, err := s.ResolveThumbnail(t.Context(), file.ID, file.WorkspaceID, true)
	require.NoError(t, err)
	require.Equal(t, queue.payloads[0].SourceVersion, access.Version)
	require.Equal(t, 1, objects.uploadCount)
	require.Equal(t, 1, repo.claimed)
	require.Equal(t, time.Minute, objects.generatedExpiry)
	require.Contains(t, objects.generatedFilename, file.WorkspaceID.String())
}

func TestThumbnailRejectsOtherWorkspaceAndUnsafeAttachmentStates(t *testing.T) {
	s, repo, objects, queue := thumbnailFixture(t)
	_, err := s.ResolveThumbnail(t.Context(), repo.attachment.ID, uuid.New(), false)
	require.ErrorIs(t, err, ErrNotFound)
	for _, status := range []attachmentdomain.ScanStatus{attachmentdomain.ScanStatusPending, attachmentdomain.ScanStatusInfected, attachmentdomain.ScanStatusFailed} {
		repo.attachment.ScanStatus = status
		_, err := s.ResolveThumbnail(t.Context(), repo.attachment.ID, repo.attachment.WorkspaceID, false)
		require.ErrorIs(t, err, ErrThumbnailUnavailable)
	}
	require.Empty(t, queue.payloads)
	require.Zero(t, objects.generateCount)
}

func TestThumbnailVersionChangeAndDeletedSourceRetireStaleJob(t *testing.T) {
	s, repo, objects, queue := thumbnailFixture(t)
	_, _ = s.ResolveThumbnail(t.Context(), repo.attachment.ID, repo.attachment.WorkspaceID, false)
	version := thumbnailVersion(repo.attachment)
	repo.attachment.UpdatedAt = repo.attachment.UpdatedAt.Add(time.Second)
	require.NotEqual(t, version, thumbnailVersion(repo.attachment))
	require.NoError(t, s.GenerateThumbnail(t.Context(), queue.payloads[0]))
	require.Equal(t, 1, repo.retired)
	require.Zero(t, objects.uploadCount)
	repo.attachment.ID = uuid.Nil
	require.NoError(t, s.GenerateThumbnail(t.Context(), queue.payloads[0]))
	require.Equal(t, 2, repo.retired)
}

func TestThumbnailSourceMutationDuringUploadRetiresDerivedObject(t *testing.T) {
	s, repo, objects, queue := thumbnailFixture(t)
	_, _ = s.ResolveThumbnail(t.Context(), repo.attachment.ID, repo.attachment.WorkspaceID, false)
	repo.completeErr = attachmentdomain.ErrStateConflict
	require.NoError(t, s.GenerateThumbnail(t.Context(), queue.payloads[0]))
	require.Equal(t, 1, objects.uploadCount)
	require.Equal(t, 1, repo.retired)
}

func TestThumbnailFailedPreviewCanBeExplicitlyRetried(t *testing.T) {
	s, repo, _, queue := thumbnailFixture(t)
	_, _ = s.ResolveThumbnail(t.Context(), repo.attachment.ID, repo.attachment.WorkspaceID, false)
	require.NoError(t, repo.FailThumbnail(t.Context(), repo.item.ID, repo.item.WorkspaceID))
	_, err := s.ResolveThumbnail(t.Context(), repo.attachment.ID, repo.attachment.WorkspaceID, false)
	require.ErrorIs(t, err, ErrThumbnailUnavailable)
	require.Len(t, queue.payloads, 1)
	_, err = s.ResolveThumbnail(t.Context(), repo.attachment.ID, repo.attachment.WorkspaceID, true)
	require.ErrorIs(t, err, ErrThumbnailPending)
	require.Len(t, queue.payloads, 2)
}

func TestThumbnailRenderBoundsAndCanceledContext(t *testing.T) {
	_, err := renderAttachmentThumbnail(t.Context(), []byte("not an image"), "image/png")
	require.ErrorIs(t, err, ErrThumbnailUnavailable)
	_, err = renderAttachmentThumbnail(t.Context(), make([]byte, thumbnailMaxSourceBytes+1), "image/png")
	require.ErrorIs(t, err, ErrThumbnailUnavailable)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, _, objects, _ := thumbnailFixture(t)
	_, err = renderAttachmentThumbnail(ctx, objects.data, "image/jpeg")
	require.True(t, errors.Is(err, context.Canceled))
}

func TestPDFThumbnailUsesFirstPageAndBoundedRaster(t *testing.T) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("Poppler installed in production worker; local pdftoppm is not available")
	}
	// Poppler accepts this minimal document and reconstructs the omitted xref.
	pdf := []byte("%PDF-1.4\n1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << >> >> endobj\ntrailer << /Root 1 0 R >>\n%%EOF")
	output, err := renderAttachmentThumbnail(t.Context(), pdf, "application/pdf")
	require.NoError(t, err)
	config, format, err := image.DecodeConfig(bytes.NewReader(output))
	require.NoError(t, err)
	require.Equal(t, "jpeg", format)
	require.LessOrEqual(t, config.Width, 840)
	require.Equal(t, 840, config.Height)
}
