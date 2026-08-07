package attachments

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/google/uuid"
)

func TestUploadDocumentMediaUsesByteDetectedImageType(t *testing.T) {
	imageBytes := append(
		[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
		make([]byte, 64)...,
	)
	repo := &attachmentRepositoryStub{}
	storageStub := &attachmentStorageStub{}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "document-media-test"),
		repo,
		storageStub,
		storage.Config{AttachmentsBucket: "attachments"},
		nil,
	)

	file := testMultipartFile{Reader: bytes.NewReader(imageBytes)}
	header := &multipart.FileHeader{
		Filename: "diagram.png",
		Size:     int64(len(imageBytes)),
		Header:   map[string][]string{"Content-Type": {"application/pdf"}},
	}
	uploaded, err := service.UploadDocumentMedia(
		context.Background(),
		file,
		header,
		uuid.New(),
		uuid.New(),
	)

	if err != nil {
		t.Fatalf("upload document image: %v", err)
	}
	if uploaded.MimeType != "image/png" || storageStub.contentType != "image/png" {
		t.Fatalf("expected byte-detected image/png, response=%q storage=%q", uploaded.MimeType, storageStub.contentType)
	}
	if storageStub.uploadCount != 1 {
		t.Fatalf("upload count = %d, want 1", storageStub.uploadCount)
	}
}

func TestUploadDocumentMediaAcceptsByteDetectedMP4(t *testing.T) {
	videoBytes := []byte{
		0x00, 0x00, 0x00, 0x14,
		'f', 't', 'y', 'p',
		'm', 'p', '4', '2',
		0x00, 0x00, 0x00, 0x00,
		'i', 's', 'o', 'm',
	}
	repo := &attachmentRepositoryStub{}
	storageStub := &attachmentStorageStub{}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "document-media-test"),
		repo,
		storageStub,
		storage.Config{AttachmentsBucket: "attachments"},
		nil,
	)

	file := testMultipartFile{Reader: bytes.NewReader(videoBytes)}
	header := &multipart.FileHeader{
		Filename: "walkthrough.mp4",
		Size:     int64(len(videoBytes)),
		Header:   map[string][]string{"Content-Type": {"application/octet-stream"}},
	}
	uploaded, err := service.UploadDocumentMedia(
		context.Background(),
		file,
		header,
		uuid.New(),
		uuid.New(),
	)

	if err != nil {
		t.Fatalf("upload document video: %v", err)
	}
	if uploaded.MimeType != "video/mp4" || storageStub.contentType != "video/mp4" {
		t.Fatalf("expected byte-detected video/mp4, response=%q storage=%q", uploaded.MimeType, storageStub.contentType)
	}
}

func TestUploadDocumentMediaRejectsNonMediaBeforeStorage(t *testing.T) {
	pdfBytes := []byte("%PDF-1.7\nnot document media")
	repo := &attachmentRepositoryStub{}
	storageStub := &attachmentStorageStub{}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "document-media-test"),
		repo,
		storageStub,
		storage.Config{AttachmentsBucket: "attachments"},
		nil,
	)

	file := testMultipartFile{Reader: bytes.NewReader(pdfBytes)}
	header := &multipart.FileHeader{
		Filename: "brief.pdf",
		Size:     int64(len(pdfBytes)),
		Header:   map[string][]string{"Content-Type": {"image/png"}},
	}
	_, err := service.UploadDocumentMedia(
		context.Background(),
		file,
		header,
		uuid.New(),
		uuid.New(),
	)

	if !errors.Is(err, ErrInvalidFileType) {
		t.Fatalf("expected invalid file type, got %v", err)
	}
	if storageStub.uploadCount != 0 {
		t.Fatalf("rejected media reached storage %d times", storageStub.uploadCount)
	}
	if repo.attachment.ID != uuid.Nil {
		t.Fatalf("rejected media created attachment %#v", repo.attachment)
	}
}

func TestResolveAttachmentAccessURLScopesAttachmentToWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	attachmentID := uuid.New()
	repo := &attachmentRepositoryStub{attachment: CoreAttachment{
		ID:          attachmentID,
		WorkspaceID: workspaceID,
		Filename:    "clip.mp4",
		BlobName:    "stored-clip.mp4",
		MimeType:    "video/mp4",
	}}
	storageStub := &attachmentStorageStub{}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "document-media-test"),
		repo,
		storageStub,
		storage.Config{AttachmentsBucket: "attachments"},
		nil,
	)

	resolved, err := service.ResolveAttachmentAccessURL(context.Background(), attachmentID, workspaceID, 3*time.Minute)
	if err != nil {
		t.Fatalf("resolve attachment access URL: %v", err)
	}
	if resolved.URL != "https://storage.test/attachment" || storageStub.generatedExpiry != 3*time.Minute {
		t.Fatalf("unexpected access resolution: file=%#v expiry=%s", resolved, storageStub.generatedExpiry)
	}
	if storageStub.generatedFilename != "stored-clip.mp4" {
		t.Fatalf("generated access for %q", storageStub.generatedFilename)
	}

	_, err = service.ResolveAttachmentAccessURL(context.Background(), attachmentID, uuid.New(), 3*time.Minute)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-workspace media to be hidden, got %v", err)
	}
	if storageStub.generateCount != 1 {
		t.Fatalf("cross-workspace resolution generated object access; count=%d", storageStub.generateCount)
	}
}

func TestDeleteDocumentMediaUsesWorkspaceAuthorization(t *testing.T) {
	workspaceID := uuid.New()
	attachmentID := uuid.New()
	repo := &attachmentRepositoryStub{attachment: CoreAttachment{
		ID:          attachmentID,
		WorkspaceID: workspaceID,
		BlobName:    "document-image.png",
	}}
	storageStub := &attachmentStorageStub{}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "document-media-test"),
		repo,
		storageStub,
		storage.Config{AttachmentsBucket: "attachments"},
		nil,
	)

	if err := service.DeleteDocumentMedia(context.Background(), attachmentID, workspaceID); err != nil {
		t.Fatalf("delete document media: %v", err)
	}
	if repo.deletedAttachmentID != attachmentID || storageStub.deletedFilename != "document-image.png" {
		t.Fatalf("document media was not fully deleted: repo=%s storage=%q", repo.deletedAttachmentID, storageStub.deletedFilename)
	}

	repo.deletedAttachmentID = uuid.Nil
	storageStub.deletedFilename = ""
	if err := service.DeleteDocumentMedia(context.Background(), attachmentID, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-workspace cleanup to be hidden, got %v", err)
	}
	if repo.deletedAttachmentID != uuid.Nil || storageStub.deletedFilename != "" {
		t.Fatal("cross-workspace cleanup deleted document media")
	}
}
