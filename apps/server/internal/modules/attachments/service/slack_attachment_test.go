package attachments

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"mime/multipart"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/google/uuid"
)

type slackAttachmentRepositoryStub struct {
	*attachmentRepositoryStub
	linkedStoryID      uuid.UUID
	linkedAttachmentID uuid.UUID
	linkedWorkspaceID  uuid.UUID
}

func (r *slackAttachmentRepositoryStub) LinkAttachmentToStory(_ context.Context, storyID, attachmentID, workspaceID uuid.UUID) error {
	r.linkedStoryID = storyID
	r.linkedAttachmentID = attachmentID
	r.linkedWorkspaceID = workspaceID
	return nil
}

func TestUploadSlackAttachmentAndLinkToStoryUsesAttachmentPipeline(t *testing.T) {
	var source bytes.Buffer
	if err := jpeg.Encode(&source, image.NewRGBA(image.Rect(0, 0, 2, 2)), nil); err != nil {
		t.Fatalf("encode JPEG: %v", err)
	}
	storyID, workspaceID, actorID := uuid.New(), uuid.New(), uuid.New()
	importID := uuid.New()
	repo := &slackAttachmentRepositoryStub{attachmentRepositoryStub: &attachmentRepositoryStub{
		storyExists:                true,
		storyAttachmentStoryID:     storyID,
		storyAttachmentWorkspaceID: workspaceID,
	}}
	storageStub := &attachmentStorageStub{}
	optimizer := &imageOptimizerStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "slack-attachment-test"), repo, storageStub, storage.Config{AttachmentsBucket: "attachments"}, optimizer)
	file := testMultipartFile{Reader: bytes.NewReader(source.Bytes())}
	header := &multipart.FileHeader{Filename: "photo.jpg", Size: int64(source.Len())}

	uploaded, err := service.UploadSlackAttachmentAndLinkToStory(context.Background(), file, header, actorID, storyID, workspaceID, importID)
	if err != nil {
		t.Fatalf("UploadSlackAttachmentAndLinkToStory() error = %v", err)
	}
	if uploaded.ID == uuid.Nil || repo.linkedAttachmentID != uploaded.ID || repo.linkedStoryID != storyID || repo.linkedWorkspaceID != workspaceID {
		t.Fatalf("Slack attachment was not linked to requested story: %+v", repo)
	}
	if repo.attachment.UploadedBy != actorID || repo.attachment.WorkspaceID != workspaceID || repo.attachment.Size != int64(source.Len()) {
		t.Fatalf("stored attachment metadata = %+v", repo.attachment)
	}
	if storageStub.uploadCount != 1 || storageStub.contentType != "image/jpeg" || len(optimizer.payloads) != 1 {
		t.Fatalf("storage uploads = %d, content type = %q, optimization tasks = %d", storageStub.uploadCount, storageStub.contentType, len(optimizer.payloads))
	}
	if repo.attachment.SlackFileImportID != importID {
		t.Fatalf("stored Slack import ID = %s, want %s", repo.attachment.SlackFileImportID, importID)
	}

	// A recovered intent may repeat the import after the attachment was already
	// stored. It should link the existing record without uploading again.
	again, err := service.UploadSlackAttachmentAndLinkToStory(context.Background(), file, header, actorID, storyID, workspaceID, importID)
	if err != nil || again.ID != uploaded.ID || storageStub.uploadCount != 1 || repo.createCount != 1 {
		t.Fatalf("idempotent retry: attachment = %s, error = %v, uploads = %d, creates = %d", again.ID, err, storageStub.uploadCount, repo.createCount)
	}
}

func TestUploadSlackAttachmentChecksStoryBeforeStorage(t *testing.T) {
	storageStub := &attachmentStorageStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "slack-attachment-test"), &attachmentRepositoryStub{}, storageStub, storage.Config{AttachmentsBucket: "attachments"}, nil)
	file := testMultipartFile{Reader: bytes.NewReader([]byte("notes"))}
	_, err := service.UploadSlackAttachmentAndLinkToStory(context.Background(), file, &multipart.FileHeader{Filename: "notes.txt", Size: 5}, uuid.New(), uuid.New(), uuid.New(), uuid.New())
	if err != ErrNotFound || storageStub.uploadCount != 0 {
		t.Fatalf("error = %v, storage uploads = %d", err, storageStub.uploadCount)
	}
}
