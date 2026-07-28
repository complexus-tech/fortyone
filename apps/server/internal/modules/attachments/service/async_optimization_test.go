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

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

type testMultipartFile struct {
	*bytes.Reader
}

func (testMultipartFile) Close() error {
	return nil
}

type attachmentRepositoryStub struct {
	attachment CoreAttachment
}

func (r *attachmentRepositoryStub) CreateAttachment(_ context.Context, attachment CoreAttachment) (CoreAttachment, error) {
	attachment.ID = uuid.New()
	attachment.CreatedAt = time.Now()
	r.attachment = attachment
	return attachment, nil
}

func (r *attachmentRepositoryStub) GetAttachmentByID(_ context.Context, _ uuid.UUID) (CoreAttachment, error) {
	return r.attachment, nil
}

func (r *attachmentRepositoryStub) GetAttachmentByBlobName(_ context.Context, blobName string) (CoreAttachment, error) {
	if r.attachment.BlobName != blobName {
		return CoreAttachment{}, ErrNotFound
	}
	return r.attachment, nil
}

func (r *attachmentRepositoryStub) GetAttachmentsByStoryID(_ context.Context, _ uuid.UUID) ([]CoreAttachment, error) {
	return nil, nil
}

func (r *attachmentRepositoryStub) UpdateAttachmentStorageMetadata(_ context.Context, blobName string, size int64, mimeType string) error {
	if r.attachment.BlobName != blobName {
		return ErrNotFound
	}
	r.attachment.Size = size
	r.attachment.MimeType = mimeType
	return nil
}

func (r *attachmentRepositoryStub) DeleteAttachment(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (r *attachmentRepositoryStub) LinkAttachmentToStory(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

type attachmentStorageStub struct {
	data        []byte
	contentType string
}

func (s *attachmentStorageStub) UploadFile(_ context.Context, _, _ string, data io.Reader, contentType string) (string, error) {
	uploaded, err := io.ReadAll(data)
	if err != nil {
		return "", err
	}
	s.data = uploaded
	s.contentType = contentType
	return "https://storage.test/attachment", nil
}

func (s *attachmentStorageStub) DownloadFile(_ context.Context, _, _ string) ([]byte, string, error) {
	return append([]byte(nil), s.data...), s.contentType, nil
}

func (s *attachmentStorageStub) GenerateAccessURL(_ context.Context, _, _ string, _ time.Duration) (string, error) {
	return "https://storage.test/attachment", nil
}

func (s *attachmentStorageStub) DeleteFile(_ context.Context, _, _ string) error {
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
	if len(optimizer.payloads) != 1 || optimizer.payloads[0].BlobName != uploaded.BlobName {
		t.Fatalf("expected one optimization task for %q, got %#v", uploaded.BlobName, optimizer.payloads)
	}

	if err := service.OptimizeStoredAttachment(context.Background(), uploaded.BlobName); err != nil {
		t.Fatalf("optimize stored attachment: %v", err)
	}
	if len(storageStub.data) >= len(original) {
		t.Fatalf("expected worker output to be smaller: optimized=%d original=%d", len(storageStub.data), len(original))
	}
	if repo.attachment.Size != int64(len(storageStub.data)) {
		t.Fatalf("metadata size = %d, want %d", repo.attachment.Size, len(storageStub.data))
	}
}
