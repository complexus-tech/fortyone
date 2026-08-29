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

func TestUploadStoryMediaLinksDedicatedInlineAttachment(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	userID := uuid.New()
	imageBytes := append(
		[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
		make([]byte, 64)...,
	)
	repo := &attachmentRepositoryStub{
		storyExists:           true,
		storyMediaStoryID:     storyID,
		storyMediaWorkspaceID: workspaceID,
	}
	storageStub := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storageStub)

	uploaded, err := service.UploadStoryMedia(
		context.Background(),
		testMultipartFile{Reader: bytes.NewReader(imageBytes)},
		&multipart.FileHeader{
			Filename: "inline.png",
			Size:     int64(len(imageBytes)),
			Header:   map[string][]string{"Content-Type": {"image/png"}},
		},
		userID,
		storyID,
		workspaceID,
	)
	if err != nil {
		t.Fatalf("upload story media: %v", err)
	}
	if uploaded.ID == uuid.Nil || uploaded.MimeType != "image/png" {
		t.Fatalf("unexpected story media: %#v", uploaded)
	}
	if repo.linkedStoryMediaID != uploaded.ID || repo.linkedStoryMediaUser != userID {
		t.Fatalf("story media was not linked with its creator: attachment=%s creator=%s", repo.linkedStoryMediaID, repo.linkedStoryMediaUser)
	}
	if storageStub.uploadCount != 1 {
		t.Fatalf("upload count = %d, want 1", storageStub.uploadCount)
	}
}

func TestUploadStoryMediaRejectsInvalidMediaBeforeStorageAndLink(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	repo := &attachmentRepositoryStub{
		storyExists:           true,
		storyMediaStoryID:     storyID,
		storyMediaWorkspaceID: workspaceID,
	}
	storageStub := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storageStub)
	pdfBytes := []byte("%PDF-1.7\nnot inline media")

	_, err := service.UploadStoryMedia(
		context.Background(),
		testMultipartFile{Reader: bytes.NewReader(pdfBytes)},
		&multipart.FileHeader{
			Filename: "brief.pdf",
			Size:     int64(len(pdfBytes)),
			Header:   map[string][]string{"Content-Type": {"image/png"}},
		},
		uuid.New(),
		storyID,
		workspaceID,
	)
	if !errors.Is(err, ErrInvalidFileType) {
		t.Fatalf("expected invalid media type, got %v", err)
	}
	if storageStub.uploadCount != 0 || repo.linkedStoryMediaID != uuid.Nil {
		t.Fatalf("invalid media mutated storage or links: uploads=%d link=%s", storageStub.uploadCount, repo.linkedStoryMediaID)
	}
}

func TestUploadStoryMediaRejectsCrossWorkspaceStoryBeforeStorage(t *testing.T) {
	storyID := uuid.New()
	repo := &attachmentRepositoryStub{
		storyExists:           true,
		storyMediaStoryID:     storyID,
		storyMediaWorkspaceID: uuid.New(),
	}
	storageStub := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storageStub)
	imageBytes := append(
		[]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'},
		make([]byte, 64)...,
	)

	_, err := service.UploadStoryMedia(
		context.Background(),
		testMultipartFile{Reader: bytes.NewReader(imageBytes)},
		&multipart.FileHeader{Filename: "inline.png", Size: int64(len(imageBytes))},
		uuid.New(),
		storyID,
		uuid.New(),
	)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-workspace story to be hidden, got %v", err)
	}
	if storageStub.uploadCount != 0 || repo.linkedStoryMediaID != uuid.Nil {
		t.Fatal("cross-workspace story media mutated storage or links")
	}
}

func TestResolveStoryMediaRequiresExactStoryWorkspaceAndLink(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	attachmentID := uuid.New()
	repo := &attachmentRepositoryStub{
		storyMediaStoryID:     storyID,
		storyMediaWorkspaceID: workspaceID,
		storyMediaAttachment: CoreAttachment{
			ID:          attachmentID,
			WorkspaceID: workspaceID,
			Filename:    "walkthrough.mp4",
			BlobName:    "stored-walkthrough.mp4",
			MimeType:    "video/mp4",
		},
	}
	storageStub := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storageStub)

	resolved, err := service.ResolveStoryMediaAccessURL(
		context.Background(),
		storyID,
		attachmentID,
		workspaceID,
		2*time.Minute,
	)
	if err != nil {
		t.Fatalf("resolve story media: %v", err)
	}
	if resolved.URL != "https://storage.test/attachment" || storageStub.generatedExpiry != 2*time.Minute {
		t.Fatalf("unexpected resolved media: %#v expiry=%s", resolved, storageStub.generatedExpiry)
	}

	for name, deniedIdentity := range map[string][2]uuid.UUID{
		"different story":     {uuid.New(), workspaceID},
		"different workspace": {storyID, uuid.New()},
	} {
		_, err := service.ResolveStoryMediaAccessURL(
			context.Background(),
			deniedIdentity[0],
			attachmentID,
			deniedIdentity[1],
			2*time.Minute,
		)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("%s: expected media to be hidden, got %v", name, err)
		}
	}
	if storageStub.generateCount != 1 {
		t.Fatalf("unauthorized requests generated storage access; count=%d", storageStub.generateCount)
	}
}

func TestGetAttachmentsForStoryRejectsCrossWorkspaceBeforeSigning(t *testing.T) {
	storyID := uuid.New()
	repo := &attachmentRepositoryStub{
		storyExists:                true,
		storyAttachmentStoryID:     storyID,
		storyAttachmentWorkspaceID: uuid.New(),
	}
	storageStub := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storageStub)

	_, err := service.GetAttachmentsForStory(context.Background(), storyID, uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-workspace story to be hidden, got %v", err)
	}
	if storageStub.generateCount != 0 {
		t.Fatalf("generated %d signed URLs for an unauthorized story", storageStub.generateCount)
	}
}

func TestDeleteStoryAttachmentEnforcesExactRelationAndOwnership(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	uploaderID := uuid.New()
	attachmentID := uuid.New()
	repo := &attachmentRepositoryStub{
		storyAttachmentStoryID:     storyID,
		storyAttachmentWorkspaceID: workspaceID,
		storyAttachment: CoreAttachment{
			ID:          attachmentID,
			UploadedBy:  uploaderID,
			WorkspaceID: workspaceID,
			BlobName:    "story-attachment.pdf",
		},
	}
	storageStub := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storageStub)

	if err := service.DeleteStoryAttachment(context.Background(), storyID, attachmentID, workspaceID, uuid.New(), false); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected non-owner deletion to be forbidden, got %v", err)
	}
	if repo.deletedAttachmentID != uuid.Nil {
		t.Fatal("unauthorized deletion reached persistence")
	}

	if err := service.DeleteStoryAttachment(context.Background(), storyID, attachmentID, workspaceID, uuid.New(), true); err != nil {
		t.Fatalf("workspace admin delete: %v", err)
	}
	if repo.deletedAttachmentID != attachmentID {
		t.Fatalf("deleted attachment = %s, want %s", repo.deletedAttachmentID, attachmentID)
	}

	repo.deletedAttachmentID = uuid.Nil
	if err := service.DeleteStoryAttachment(context.Background(), storyID, attachmentID, uuid.New(), uploaderID, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-workspace attachment to be hidden, got %v", err)
	}
	if repo.deletedAttachmentID != uuid.Nil {
		t.Fatal("cross-workspace deletion reached persistence")
	}
}

func TestDeleteStoryMediaRequiresExactLinkBeforeCleanup(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	attachmentID := uuid.New()
	repo := &attachmentRepositoryStub{
		storyMediaStoryID:     storyID,
		storyMediaWorkspaceID: workspaceID,
		storyMediaOrphaned:    true,
		storyMediaAttachment: CoreAttachment{
			ID:          attachmentID,
			WorkspaceID: workspaceID,
			BlobName:    "inline-image.png",
			MimeType:    "image/png",
		},
	}
	storageStub := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storageStub)

	if err := service.DeleteStoryMedia(context.Background(), uuid.New(), attachmentID, workspaceID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected cross-story cleanup to be hidden, got %v", err)
	}
	if repo.storyMediaUnlinkCount != 0 || storageStub.deletedFilename != "" {
		t.Fatal("cross-story cleanup unlinked or deleted media")
	}

	if err := service.DeleteStoryMedia(context.Background(), storyID, attachmentID, workspaceID); err != nil {
		t.Fatalf("delete story media: %v", err)
	}
	if repo.storyMediaUnlinkCount != 1 || repo.deletedAttachmentID != attachmentID {
		t.Fatalf("story media cleanup did not remove the relation and record: unlinks=%d deleted=%s", repo.storyMediaUnlinkCount, repo.deletedAttachmentID)
	}
	if storageStub.deletedFilename != "inline-image.png" {
		t.Fatalf("deleted storage object = %q", storageStub.deletedFilename)
	}
}

func newStoryMediaTestService(repo Repository, storageService storage.StorageService) *Service {
	return New(
		logger.NewWithText(io.Discard, slog.LevelError, "story-media-test"),
		repo,
		storageService,
		storage.Config{AttachmentsBucket: "attachments"},
		nil,
	)
}
