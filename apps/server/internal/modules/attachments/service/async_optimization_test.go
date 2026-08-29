package attachments

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log/slog"
	"mime/multipart"
	"testing"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
)

type testMultipartFile struct {
	*bytes.Reader
}

func (testMultipartFile) Close() error {
	return nil
}

type attachmentRepositoryStub struct {
	attachment                 CoreAttachment
	deletedAttachmentID        uuid.UUID
	storyExists                bool
	storyAttachment            CoreAttachment
	storyAttachmentStoryID     uuid.UUID
	storyAttachmentWorkspaceID uuid.UUID
	storyMediaAttachment       CoreAttachment
	storyMediaStoryID          uuid.UUID
	storyMediaWorkspaceID      uuid.UUID
	linkedStoryMediaID         uuid.UUID
	linkedStoryMediaUser       uuid.UUID
	storyMediaUnlinkCount      int
	storyMediaOrphaned         bool
}

func (r *attachmentRepositoryStub) CreateAttachment(_ context.Context, attachment CoreAttachment) (CoreAttachment, error) {
	attachment.ID = uuid.New()
	attachment.CreatedAt = time.Now()
	r.attachment = attachment
	return attachment, nil
}
func (r *attachmentRepositoryStub) GetAttachmentByID(_ context.Context, id, workspaceID uuid.UUID) (CoreAttachment, error) {
	if r.attachment.ID != id || r.attachment.WorkspaceID != workspaceID {
		return CoreAttachment{}, ErrNotFound
	}
	return r.attachment, nil
}

func (r *attachmentRepositoryStub) GetAttachmentsByStoryID(_ context.Context, storyID, workspaceID uuid.UUID) ([]CoreAttachment, error) {
	if storyID != r.storyAttachmentStoryID || workspaceID != r.storyAttachmentWorkspaceID {
		return nil, ErrNotFound
	}
	if r.storyAttachment.ID == uuid.Nil {
		return nil, nil
	}
	return []CoreAttachment{r.storyAttachment}, nil
}

func (r *attachmentRepositoryStub) StoryExistsInWorkspace(_ context.Context, storyID, workspaceID uuid.UUID) (bool, error) {
	mediaMatch := storyID == r.storyMediaStoryID && workspaceID == r.storyMediaWorkspaceID
	attachmentMatch := storyID == r.storyAttachmentStoryID && workspaceID == r.storyAttachmentWorkspaceID
	return r.storyExists && (mediaMatch || attachmentMatch), nil
}

func (r *attachmentRepositoryStub) LinkStoryMedia(_ context.Context, storyID, attachmentID, createdBy, workspaceID uuid.UUID) error {
	if !r.storyExists || storyID != r.storyMediaStoryID || workspaceID != r.storyMediaWorkspaceID {
		return ErrNotFound
	}
	r.linkedStoryMediaID = attachmentID
	r.linkedStoryMediaUser = createdBy
	return nil
}

func (r *attachmentRepositoryStub) AuthorizeStoryMedia(_ context.Context, storyID, attachmentID, workspaceID uuid.UUID) (CoreAttachment, error) {
	if storyID != r.storyMediaStoryID ||
		workspaceID != r.storyMediaWorkspaceID ||
		attachmentID != r.storyMediaAttachment.ID {
		return CoreAttachment{}, ErrNotFound
	}
	return r.storyMediaAttachment, nil
}

func (r *attachmentRepositoryStub) AuthorizeStoryAttachment(_ context.Context, storyID, attachmentID, workspaceID uuid.UUID) (CoreAttachment, error) {
	if storyID != r.storyAttachmentStoryID ||
		workspaceID != r.storyAttachmentWorkspaceID ||
		attachmentID != r.storyAttachment.ID {
		return CoreAttachment{}, ErrNotFound
	}
	return r.storyAttachment, nil
}

func (r *attachmentRepositoryStub) UnlinkStoryMedia(_ context.Context, storyID, attachmentID, workspaceID uuid.UUID) (bool, error) {
	if storyID != r.storyMediaStoryID ||
		workspaceID != r.storyMediaWorkspaceID ||
		attachmentID != r.storyMediaAttachment.ID {
		return false, ErrNotFound
	}
	r.storyMediaUnlinkCount++
	if r.storyMediaOrphaned {
		r.deletedAttachmentID = attachmentID
	}
	return r.storyMediaOrphaned, nil
}

func (r *attachmentRepositoryStub) StartAttachmentOptimization(_ context.Context, attachmentID, workspaceID uuid.UUID, _ time.Duration) (CoreAttachment, error) {
	if r.attachment.ID != attachmentID || r.attachment.WorkspaceID != workspaceID {
		return CoreAttachment{}, ErrNotFound
	}
	r.attachment.OptimizationStatus = attachmentdomain.OptimizationProcessing
	return r.attachment, nil
}

func (r *attachmentRepositoryStub) CompleteAttachmentOptimization(_ context.Context, attachmentID, workspaceID uuid.UUID, size int64, mimeType string, status attachmentdomain.OptimizationStatus) error {
	if r.attachment.ID != attachmentID || r.attachment.WorkspaceID != workspaceID {
		return ErrNotFound
	}
	r.attachment.Size = size
	r.attachment.MimeType = mimeType
	r.attachment.OptimizationStatus = status
	return nil
}

func (r *attachmentRepositoryStub) FailAttachmentOptimization(_ context.Context, attachmentID, workspaceID uuid.UUID, _ string, _ bool) error {
	if r.attachment.ID != attachmentID || r.attachment.WorkspaceID != workspaceID {
		return ErrNotFound
	}
	r.attachment.OptimizationStatus = attachmentdomain.OptimizationFailed
	return nil
}

func (r *attachmentRepositoryStub) DeleteAttachment(_ context.Context, id, workspaceID uuid.UUID) error {
	if r.attachment.ID != uuid.Nil && (r.attachment.ID != id || r.attachment.WorkspaceID != workspaceID) {
		return ErrNotFound
	}
	r.deletedAttachmentID = id
	return nil
}

func (r *attachmentRepositoryStub) DeleteAttachmentIfUnreferenced(_ context.Context, id, workspaceID uuid.UUID) (bool, error) {
	if r.attachment.ID != id || r.attachment.WorkspaceID != workspaceID {
		return false, ErrNotFound
	}
	r.deletedAttachmentID = id
	return true, nil
}

func (r *attachmentRepositoryStub) LinkAttachmentToStory(_ context.Context, _, _, _ uuid.UUID) error {
	return nil
}

type attachmentStorageStub struct {
	data              []byte
	contentType       string
	uploadCount       int
	generateCount     int
	generatedExpiry   time.Duration
	generatedFilename string
	deletedFilename   string
	downloadLimit     int64
}

func (s *attachmentStorageStub) UploadFile(_ context.Context, _, _ string, data io.Reader, contentType string) (string, error) {
	uploaded, err := io.ReadAll(data)
	if err != nil {
		return "", err
	}
	s.data = uploaded
	s.contentType = contentType
	s.uploadCount++
	return "https://storage.test/attachment", nil
}

func (s *attachmentStorageStub) DownloadFile(_ context.Context, _, _ string, maxBytes int64) ([]byte, string, error) {
	s.downloadLimit = maxBytes
	return append([]byte(nil), s.data...), s.contentType, nil
}

func (s *attachmentStorageStub) GenerateAccessURL(_ context.Context, _, filename string, expiry time.Duration) (string, error) {
	s.generateCount++
	s.generatedFilename = filename
	s.generatedExpiry = expiry
	return "https://storage.test/attachment", nil
}

func (s *attachmentStorageStub) DeleteFile(_ context.Context, _, filename string) error {
	s.deletedFilename = filename
	return nil
}

func (s *attachmentStorageStub) GetPublicURL(_ context.Context, _, _ string) (string, error) {
	return "https://storage.test/attachment", nil
}

type imageOptimizerStub struct {
	payloads []tasks.AttachmentImageOptimizationPayload
}

func (s *imageOptimizerStub) EnqueueAttachmentImageOptimization(payload tasks.AttachmentImageOptimizationPayload) error {
	s.payloads = append(s.payloads, payload)
	return nil
}

func TestAttachmentUploadEnqueuesOptimizationAndWorkerCompressesStoredImage(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2400, 1400))
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			source.Set(x, y, color.RGBA{
				R: uint8(x % 255),
				G: uint8(y % 255),
				B: uint8((x + y) % 255),
				A: 255,
			})
		}
	}

	var input bytes.Buffer
	if err := jpeg.Encode(&input, source, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode source jpeg: %v", err)
	}
	original := append([]byte(nil), input.Bytes()...)

	repo := &attachmentRepositoryStub{}
	storageStub := &attachmentStorageStub{}
	optimizer := &imageOptimizerStub{}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "attachments-test"),
		repo,
		storageStub,
		storage.Config{AttachmentsBucket: "attachments"},
		optimizer,
	)

	file := testMultipartFile{Reader: bytes.NewReader(original)}
	header := &multipart.FileHeader{
		Filename: "large-photo.jpg",
		Size:     int64(len(original)),
		Header:   map[string][]string{"Content-Type": {"image/jpeg"}},
	}
	uploaded, err := service.UploadAttachment(
		context.Background(),
		file,
		header,
		uuid.New(),
		uuid.New(),
	)
	if err != nil {
		t.Fatalf("upload attachment: %v", err)
	}

	if len(storageStub.data) != len(original) {
		t.Fatalf("upload should store original bytes before the worker runs: stored=%d original=%d", len(storageStub.data), len(original))
	}
	if len(optimizer.payloads) != 1 ||
		optimizer.payloads[0].AttachmentID != uploaded.ID ||
		optimizer.payloads[0].WorkspaceID != repo.attachment.WorkspaceID {
		t.Fatalf("expected one tenant-scoped optimization task for %s, got %#v", uploaded.ID, optimizer.payloads)
	}

	if err := service.OptimizeStoredAttachment(context.Background(), uploaded.ID, repo.attachment.WorkspaceID); err != nil {
		t.Fatalf("optimize stored attachment: %v", err)
	}
	if storageStub.downloadLimit != validate.MaxAttachmentSize {
		t.Fatalf("download limit = %d, want %d", storageStub.downloadLimit, validate.MaxAttachmentSize)
	}
	if len(storageStub.data) >= len(original) {
		t.Fatalf("expected worker output to be smaller: optimized=%d original=%d", len(storageStub.data), len(original))
	}
	if repo.attachment.Size != int64(len(storageStub.data)) {
		t.Fatalf("metadata size = %d, want %d", repo.attachment.Size, len(storageStub.data))
	}
}
