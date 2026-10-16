package attachments

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ResolveStoryAttachmentAccessURL authorizes the exact story relation before
// issuing fresh, short-lived object access for an embedded comment attachment.
func (s *Service) ResolveStoryAttachmentAccessURL(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID, expiry time.Duration) (FileInfo, error) {
	if storyID == uuid.Nil || attachmentID == uuid.Nil || workspaceID == uuid.Nil || expiry <= 0 {
		return FileInfo{}, ErrInvalidFile
	}
	attachment, err := s.repo.AuthorizeStoryAttachment(ctx, storyID, attachmentID, workspaceID)
	if err != nil {
		return FileInfo{}, err
	}
	if !attachment.AvailableForDownload() {
		return FileInfo{}, ErrNotFound
	}
	accessURL, err := s.storage.GenerateAccessURL(ctx, s.config.AttachmentsBucket, attachment.BlobName, expiry)
	if err != nil {
		return FileInfo{}, fmt.Errorf("generate story attachment access URL: %w", err)
	}
	return FileInfo{
		ID: attachment.ID, Filename: attachment.Filename, BlobName: attachment.BlobName,
		Size: attachment.Size, MimeType: attachment.MimeType, URL: accessURL,
		CreatedAt: attachment.CreatedAt, UploadedBy: attachment.UploadedBy,
	}, nil
}
