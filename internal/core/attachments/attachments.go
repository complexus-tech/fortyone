package attachments

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Repository defines the storage interface for attachments
type Repository interface {
	CreateAttachment(ctx context.Context, attachment CoreAttachment) (CoreAttachment, error)
	GetAttachmentByID(ctx context.Context, id uuid.UUID) (CoreAttachment, error)
	GetAttachmentsByStoryID(ctx context.Context, storyID uuid.UUID) ([]CoreAttachment, error)
	DeleteAttachment(ctx context.Context, id uuid.UUID) error
	LinkAttachmentToStory(ctx context.Context, storyID, attachmentID uuid.UUID) error
}

// Service manages attachment operations
type Service struct {
	log     *logger.Logger
	repo    Repository
	storage storage.StorageService
	config  storage.Config
}

// New creates a new attachment service
func New(log *logger.Logger, repo Repository, storageService storage.StorageService, config storage.Config) *Service {
	return &Service{
		log:     log,
		repo:    repo,
		storage: storageService,
		config:  config,
	}
}

// UploadAttachment uploads a file and creates an attachment record
func (s *Service) UploadAttachment(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID, workspaceID uuid.UUID) (FileInfo, error) {
	s.log.Info(ctx, "core.attachments.upload")
	ctx, span := web.AddSpan(ctx, "core.attachments.UploadAttachment")
	defer span.End()

	// Validate file
	if err := validate.Attachment(file, fileHeader); err != nil {
		span.RecordError(err)
		return FileInfo{}, fmt.Errorf("invalid file: %w", err)
	}

	// Generate a unique filename
	blobName := validate.GenerateFileName(fileHeader.Filename)

	containerName, err := s.getContainerName("attachments")
	if err != nil {
		span.RecordError(err)
		return FileInfo{}, err
	}

	// Upload to storage
	_, err = s.storage.UploadFile(
		ctx,
		containerName,
		blobName,
		file,
		fileHeader.Header.Get("Content-Type"),
	)
	if err != nil {
		span.RecordError(err)
		return FileInfo{}, fmt.Errorf("failed to upload to storage: %w", err)
	}

	// Create attachment record in database
	attachment, err := s.repo.CreateAttachment(ctx, CoreAttachment{
		Filename:    fileHeader.Filename,
		BlobName:    blobName,
		Size:        fileHeader.Size,
		MimeType:    fileHeader.Header.Get("Content-Type"),
		UploadedBy:  userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		span.RecordError(err)
		// Try to clean up the blob since DB insert failed
		_ = s.storage.DeleteFile(ctx, containerName, blobName)
		return FileInfo{}, fmt.Errorf("failed to create attachment record: %w", err)
	}

	// Generate a presigned URL for the uploaded file (30 minutes)
	accessURL, err := s.storage.GenerateAccessURL(
		ctx,
		containerName,
		blobName,
		30*time.Minute,
	)
	if err != nil {
		span.RecordError(err)
		return FileInfo{}, fmt.Errorf("failed to generate access URL: %w", err)
	}

	span.AddEvent("attachment created", trace.WithAttributes(
		attribute.String("attachment_id", attachment.ID.String()),
		attribute.String("filename", fileHeader.Filename),
	))

	return FileInfo{
		ID:         attachment.ID,
		Filename:   attachment.Filename,
		BlobName:   blobName,
		Size:       attachment.Size,
		MimeType:   attachment.MimeType,
		URL:        accessURL,
		CreatedAt:  attachment.CreatedAt,
		UploadedBy: attachment.UploadedBy,
	}, nil
}

// GetAttachmentsForStory gets all attachments for a story
func (s *Service) GetAttachmentsForStory(ctx context.Context, storyID uuid.UUID) ([]FileInfo, error) {
	s.log.Info(ctx, "core.attachments.getForStory")
	ctx, span := web.AddSpan(ctx, "core.attachments.GetAttachmentsForStory")
	defer span.End()

	// Get attachments from database
	attachments, err := s.repo.GetAttachmentsByStoryID(ctx, storyID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	containerName, err := s.getContainerName("attachments")
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	fileInfos := make([]FileInfo, len(attachments))
	for i, attachment := range attachments {
		// Use the stored blob name instead of generating a new one
		blobName := attachment.BlobName

		// Generate a presigned URL for each file (15 minutes)
		accessURL, err := s.storage.GenerateAccessURL(
			ctx,
			containerName,
			blobName,
			15*time.Minute,
		)
		if err != nil {
			span.RecordError(err)
			continue
		}

		fileInfos[i] = FileInfo{
			ID:         attachment.ID,
			Filename:   attachment.Filename,
			BlobName:   blobName,
			Size:       attachment.Size,
			MimeType:   attachment.MimeType,
			URL:        accessURL,
			CreatedAt:  attachment.CreatedAt,
			UploadedBy: attachment.UploadedBy,
		}

	}

	return fileInfos, nil
}

// DeleteAttachment deletes an attachment
func (s *Service) DeleteAttachment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.delete")
	ctx, span := web.AddSpan(ctx, "core.attachments.DeleteAttachment")
	defer span.End()

	// Get attachment to check ownership and get the filename
	attachment, err := s.repo.GetAttachmentByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Verify ownership - allow the uploader or workspace admin to delete
	if attachment.UploadedBy != userID {
		// TODO: check if user is workspace admin
		span.RecordError(ErrUnauthorized)
		return ErrUnauthorized
	}

	// Use the stored blob name if available, otherwise generate it
	blobName := attachment.BlobName
	// Delete from database first
	err = s.repo.DeleteAttachment(ctx, id)
	if err != nil {
		span.RecordError(err)
		return err
	}

	containerName, err := s.getContainerName("attachments")
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete from storage
	err = s.storage.DeleteFile(ctx, containerName, blobName)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to delete blob from storage", "error", err)
		// We don't return this error since the DB record is already deleted
	}

	span.AddEvent("attachment deleted", trace.WithAttributes(
		attribute.String("attachment_id", id.String()),
	))

	return nil
}

// LinkAttachmentToStory links an attachment to a story
func (s *Service) LinkAttachmentToStory(ctx context.Context, storyID, attachmentID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.linkToStory")
	ctx, span := web.AddSpan(ctx, "core.attachments.LinkAttachmentToStory")
	defer span.End()

	// Link attachment to story
	err := s.repo.LinkAttachmentToStory(ctx, storyID, attachmentID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	span.AddEvent("attachment linked to story", trace.WithAttributes(
		attribute.String("attachment_id", attachmentID.String()),
		attribute.String("story_id", storyID.String()),
	))

	return nil
}

// UploadAndLinkToStory uploads a file and links it to a story in a single operation
func (s *Service) UploadAndLinkToStory(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID, storyID uuid.UUID, workspaceID uuid.UUID) (FileInfo, error) {
	s.log.Info(ctx, "core.attachments.uploadAndLinkToStory")
	ctx, span := web.AddSpan(ctx, "core.attachments.UploadAndLinkToStory")
	defer span.End()

	// First upload the attachment
	fileInfo, err := s.UploadAttachment(ctx, file, fileHeader, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return FileInfo{}, err
	}

	// Then link it to the story
	err = s.LinkAttachmentToStory(ctx, storyID, fileInfo.ID)
	if err != nil {
		span.RecordError(err)
		// Try to clean up the attachment since linking failed
		_ = s.DeleteAttachment(ctx, fileInfo.ID, userID)
		return FileInfo{}, fmt.Errorf("failed to link attachment to story: %w", err)
	}

	span.AddEvent("attachment uploaded and linked to story", trace.WithAttributes(
		attribute.String("attachment_id", fileInfo.ID.String()),
		attribute.String("story_id", storyID.String()),
		attribute.String("filename", fileHeader.Filename),
	))

	return fileInfo, nil
}

// UploadProfileImage uploads a profile image and returns the image URL
func (s *Service) UploadProfileImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID) (string, error) {
	s.log.Info(ctx, "core.attachments.uploadProfileImage")
	ctx, span := web.AddSpan(ctx, "core.attachments.UploadProfileImage")
	defer span.End()

	// Validate file using your existing validator
	if err := validate.ProfileImage(file, fileHeader); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("invalid profile image: %w", err)
	}

	// Generate a unique filename for profile images
	blobName := fmt.Sprintf("profile_%s_%s", userID.String(), validate.GenerateFileName(fileHeader.Filename))

	containerName, err := s.getContainerName("profile-images")
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	// Upload to storage profile images container
	_, err = s.storage.UploadFile(
		ctx,
		containerName,
		blobName,
		file,
		fileHeader.Header.Get("Content-Type"),
	)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to upload profile image to storage: %w", err)
	}

	// Generate permanent URL (container must be set to public access)
	imageURL, err := s.storage.GetPublicURL(ctx, containerName, blobName)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to generate profile image URL: %w", err)
	}

	span.AddEvent("profile image uploaded", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("filename", fileHeader.Filename),
		attribute.String("image_url", imageURL),
	))

	return imageURL, nil
}

// DeleteProfileImage deletes a profile image from storage
func (s *Service) DeleteProfileImage(ctx context.Context, avatarURL string) error {
	s.log.Info(ctx, "core.attachments.deleteProfileImage")
	ctx, span := web.AddSpan(ctx, "core.attachments.DeleteProfileImage")
	defer span.End()

	if avatarURL == "" {
		return nil // Nothing to delete
	}

	containerName, err := s.getContainerName("profile-images")
	if err != nil {
		span.RecordError(err)
		return err
	}

	blobName, err := s.getObjectNameFromURL(avatarURL, containerName)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete from storage
	err = s.storage.DeleteFile(ctx, containerName, blobName)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to delete profile image from storage", "error", err)
		// Don't return error as this is not critical
	}

	span.AddEvent("profile image deleted", trace.WithAttributes(
		attribute.String("avatar_url", avatarURL),
		attribute.String("blob_name", blobName),
	))

	return nil
}

func (s *Service) UploadWorkspaceLogo(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, workspaceID uuid.UUID) (string, error) {
	s.log.Info(ctx, "business.core.attachments.upload_workspace_logo")
	ctx, span := web.AddSpan(ctx, "business.core.attachments.UploadWorkspaceLogo")
	defer span.End()

	if err := validate.WorkspaceLogo(file, fileHeader); err != nil {
		return "", fmt.Errorf("workspace logo validation failed: %w", err)
	}

	blobName := fmt.Sprintf("%s/%s", workspaceID.String(), fileHeader.Filename)
	containerName, err := s.getContainerName("workspace-logos")
	if err != nil {
		span.RecordError(err)
		return "", err
	}

	_, err = s.storage.UploadFile(ctx, containerName, blobName, file, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to upload workspace logo: %w", err)
	}

	url, err := s.storage.GetPublicURL(ctx, containerName, blobName)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to generate workspace logo URL: %w", err)
	}

	span.AddEvent("workspace logo uploaded.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("url", url),
	))

	return url, nil
}

func (s *Service) DeleteWorkspaceLogo(ctx context.Context, logoURL string) error {
	s.log.Info(ctx, "business.core.attachments.delete_workspace_logo")
	ctx, span := web.AddSpan(ctx, "business.core.attachments.DeleteWorkspaceLogo")
	defer span.End()

	if logoURL == "" {
		return nil // Nothing to delete
	}

	containerName, err := s.getContainerName("workspace-logos")
	if err != nil {
		span.RecordError(err)
		return err
	}

	blobName, err := s.getObjectNameFromURL(logoURL, containerName)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.storage.DeleteFile(ctx, containerName, blobName); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete workspace logo: %w", err)
	}

	span.AddEvent("workspace logo deleted.", trace.WithAttributes(
		attribute.String("blob_name", blobName),
	))

	return nil
}

func (s *Service) getContainerName(fileType string) (string, error) {
	switch s.config.Provider {
	case "azure":
		switch fileType {
		case "attachments":
			return s.config.Azure.AttachmentsContainer, nil
		case "profile-images":
			return s.config.Azure.ProfileImagesContainer, nil
		case "workspace-logos":
			return s.config.Azure.WorkspaceLogosContainer, nil
		}
	case "aws":
		switch fileType {
		case "attachments":
			return s.config.AWS.AttachmentsBucket, nil
		case "profile-images":
			return s.config.AWS.ProfileImagesBucket, nil
		case "workspace-logos":
			return s.config.AWS.WorkspaceLogosBucket, nil
		}
	default:
		return "", fmt.Errorf("unsupported storage provider: %s", s.config.Provider)
	}

	return "", fmt.Errorf("unsupported file type: %s", fileType)
}

func (s *Service) getObjectNameFromURL(fileURL, container string) (string, error) {
	parsed, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid file URL: %w", err)
	}

	path := strings.TrimPrefix(parsed.Path, "/")
	if path == "" {
		return "", fmt.Errorf("invalid file URL format")
	}

	if s.config.Provider == "azure" {
		prefix := container + "/"
		if !strings.HasPrefix(path, prefix) {
			return "", fmt.Errorf("invalid file URL format")
		}
		return strings.TrimPrefix(path, prefix), nil
	}

	return path, nil
}
