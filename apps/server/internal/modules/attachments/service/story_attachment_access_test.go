package attachments

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestResolveStoryAttachmentScopesDocumentsAndImagesBeforeSigning(t *testing.T) {
	storyID, workspaceID, attachmentID := uuid.New(), uuid.New(), uuid.New()
	repo := &attachmentRepositoryStub{
		storyAttachmentStoryID: storyID, storyAttachmentWorkspaceID: workspaceID,
		storyAttachment: CoreAttachment{ID: attachmentID, WorkspaceID: workspaceID, Filename: "brief.pdf", BlobName: "brief", MimeType: "application/pdf"},
	}
	storage := &attachmentStorageStub{}
	service := newStoryMediaTestService(repo, storage)
	file, err := service.ResolveStoryAttachmentAccessURL(t.Context(), storyID, attachmentID, workspaceID, time.Minute)
	if err != nil || file.URL == "" || storage.generatedExpiry != time.Minute {
		t.Fatalf("resolve file=%#v error=%v", file, err)
	}
	for _, scope := range [][2]uuid.UUID{{uuid.New(), workspaceID}, {storyID, uuid.New()}} {
		if _, err := service.ResolveStoryAttachmentAccessURL(t.Context(), scope[0], attachmentID, scope[1], time.Minute); !errors.Is(err, ErrNotFound) {
			t.Fatalf("invalid scope resolved: %v", err)
		}
	}
	if storage.generateCount != 1 {
		t.Fatalf("unauthorized URL generated: %d", storage.generateCount)
	}
}
