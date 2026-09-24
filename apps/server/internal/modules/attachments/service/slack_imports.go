package attachments

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
)

// UploadSlackAttachment stores a Slack-hosted file after the importer has
// verified its declared size and streamed its exact bytes to a local file.
// Slack imports retain content validation but bypass direct-upload size caps.
func (s *Service) UploadSlackAttachment(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID, workspaceID uuid.UUID) (FileInfo, error) {
	return s.uploadAttachment(ctx, file, fileHeader, userID, workspaceID, uuid.Nil, validate.AttachmentWithoutSizeLimit, nil)
}

// UploadSlackAttachmentAndLinkToStory imports one file into one story.
func (s *Service) UploadSlackAttachmentAndLinkToStory(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID, storyID, workspaceID, importID uuid.UUID) (FileInfo, error) {
	if userID == uuid.Nil || storyID == uuid.Nil || workspaceID == uuid.Nil || importID == uuid.Nil {
		return FileInfo{}, ErrInvalidFile
	}

	exists, err := s.repo.StoryExistsInWorkspace(ctx, storyID, workspaceID)
	if err != nil {
		return FileInfo{}, fmt.Errorf("check story attachment workspace: %w", err)
	}
	if !exists {
		return FileInfo{}, ErrNotFound
	}

	attachment, err := s.repo.GetAttachmentBySlackImportID(ctx, importID, workspaceID)
	if err == nil {
		return s.linkSlackImportToStory(ctx, attachment, storyID, workspaceID)
	}
	if !errors.Is(err, ErrNotFound) {
		return FileInfo{}, fmt.Errorf("look up Slack import attachment: %w", err)
	}

	fileInfo, err := s.uploadAttachment(ctx, file, fileHeader, userID, workspaceID, importID, validate.AttachmentWithoutSizeLimit, nil)
	if err != nil {
		// A concurrent attempt may have inserted the same import while this
		// attempt uploaded its own blob. The upload path removes its losing
		// blob; use the attachment that won the uniqueness race.
		winner, lookupErr := s.repo.GetAttachmentBySlackImportID(ctx, importID, workspaceID)
		if lookupErr == nil {
			return s.linkSlackImportToStory(ctx, winner, storyID, workspaceID)
		}
		return FileInfo{}, err
	}
	if err := s.LinkAttachmentToStory(ctx, storyID, fileInfo.ID, workspaceID); err != nil {
		// Preserve the import-scoped record for a retry. A competing attempt
		// may have linked the same attachment concurrently.
		return FileInfo{}, fmt.Errorf("link Slack attachment to story: %w", err)
	}
	return fileInfo, nil
}

func (s *Service) linkSlackImportToStory(ctx context.Context, attachment CoreAttachment, storyID, workspaceID uuid.UUID) (FileInfo, error) {
	if err := s.LinkAttachmentToStory(ctx, storyID, attachment.ID, workspaceID); err != nil {
		return FileInfo{}, fmt.Errorf("link existing Slack attachment to story: %w", err)
	}
	accessURL, err := s.storage.GenerateAccessURL(ctx, s.config.AttachmentsBucket, attachment.BlobName, 30*time.Minute)
	if err != nil {
		return FileInfo{}, fmt.Errorf("generate existing Slack attachment access URL: %w", err)
	}
	s.maybeEnqueueImageOptimization(ctx, attachment)
	return FileInfo{
		ID:         attachment.ID,
		Filename:   attachment.Filename,
		BlobName:   attachment.BlobName,
		Size:       attachment.Size,
		MimeType:   attachment.MimeType,
		URL:        accessURL,
		CreatedAt:  attachment.CreatedAt,
		UploadedBy: attachment.UploadedBy,
	}, nil
}
