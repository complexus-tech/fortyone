package attachments

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
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
	GetAttachmentByBlobName(ctx context.Context, blobName string) (CoreAttachment, error)
	GetAttachmentsByStoryID(ctx context.Context, storyID uuid.UUID) ([]CoreAttachment, error)
	StoryExistsInWorkspace(ctx context.Context, storyID, workspaceID uuid.UUID) (bool, error)
	LinkStoryMedia(ctx context.Context, storyID, attachmentID, createdBy, workspaceID uuid.UUID) error
	AuthorizeStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) (CoreAttachment, error)
	UnlinkStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) (bool, error)
	UpdateAttachmentStorageMetadata(ctx context.Context, blobName string, size int64, mimeType string) error
	DeleteAttachment(ctx context.Context, id uuid.UUID) error
	LinkAttachmentToStory(ctx context.Context, storyID, attachmentID uuid.UUID) error
}

type ImageOptimizationEnqueuer interface {
	EnqueueAttachmentImageOptimization(payload tasks.AttachmentImageOptimizationPayload) error
}

// Service manages attachment operations
type Service struct {
	log       *logger.Logger
	repo      Repository
	storage   storage.StorageService
	config    storage.Config
	optimizer ImageOptimizationEnqueuer
}

// New creates a new attachment service
func New(log *logger.Logger, repo Repository, storageService storage.StorageService, config storage.Config, optimizer ImageOptimizationEnqueuer) *Service {
	return &Service{
		log:       log,
		repo:      repo,
		storage:   storageService,
		config:    config,
		optimizer: optimizer,
	}
}

// UploadAttachment uploads a file and creates an attachment record
func (s *Service) UploadAttachment(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID, workspaceID uuid.UUID) (FileInfo, error) {
	return s.uploadAttachment(ctx, file, fileHeader, userID, workspaceID, nil)
}

// UploadDocumentMedia uploads an image or MP4 video for use inside a document.
// The content type is derived from the file bytes, not the multipart header.
func (s *Service) UploadDocumentMedia(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID, workspaceID uuid.UUID) (FileInfo, error) {
	return s.uploadAttachment(ctx, file, fileHeader, userID, workspaceID, validateInlineMediaContentType)
}

// UploadStoryMedia uploads an image or MP4 video and links it through the
// dedicated inline-media relation. The story is checked before storage is
// mutated so a cross-workspace or missing story cannot create an orphan.
func (s *Service) UploadStoryMedia(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	userID, storyID, workspaceID uuid.UUID,
) (FileInfo, error) {
	if userID == uuid.Nil || storyID == uuid.Nil || workspaceID == uuid.Nil {
		return FileInfo{}, ErrInvalidFile
	}

	exists, err := s.repo.StoryExistsInWorkspace(ctx, storyID, workspaceID)
	if err != nil {
		return FileInfo{}, fmt.Errorf("check story media workspace: %w", err)
	}
	if !exists {
		return FileInfo{}, ErrNotFound
	}

	fileInfo, err := s.uploadAttachment(
		ctx,
		file,
		fileHeader,
		userID,
		workspaceID,
		validateInlineMediaContentType,
	)
	if err != nil {
		return FileInfo{}, err
	}

	if err := s.repo.LinkStoryMedia(ctx, storyID, fileInfo.ID, userID, workspaceID); err != nil {
		if cleanupErr := s.DeleteAttachment(ctx, fileInfo.ID, userID); cleanupErr != nil {
			s.log.Error(ctx, "failed to clean up unlinked story media", "error", cleanupErr, "attachment_id", fileInfo.ID)
		}
		return FileInfo{}, fmt.Errorf("link story media: %w", err)
	}

	return fileInfo, nil
}

func validateInlineMediaContentType(contentType string) error {
	if isInlineMediaContentType(contentType) {
		return nil
	}
	return fmt.Errorf("%w: upload an image or MP4 video", ErrInvalidFileType)
}

func isInlineMediaContentType(contentType string) bool {
	return strings.HasPrefix(contentType, "image/") || contentType == "video/mp4"
}

func (s *Service) uploadAttachment(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	validateContentType func(string) error,
) (FileInfo, error) {
	s.log.Info(ctx, "core.attachments.upload")
	ctx, span := web.AddSpan(ctx, "core.attachments.UploadAttachment")
	defer span.End()

	// Validate file
	if err := validate.Attachment(file, fileHeader); err != nil {
		span.RecordError(err)
		return FileInfo{}, attachmentValidationError(err)
	}

	// Generate a unique filename
	blobName := validate.GenerateFileName(fileHeader.Filename)
	contentType, err := detectFileContentType(file, fileHeader)
	if err != nil {
		span.RecordError(err)
		return FileInfo{}, fmt.Errorf("%w: failed to detect content type", ErrInvalidFile)
	}
	if validateContentType != nil {
		if err := validateContentType(contentType); err != nil {
			span.RecordError(err)
			return FileInfo{}, err
		}
	}

	// Upload to storage
	_, err = s.storage.UploadFile(
		ctx,
		s.config.AttachmentsBucket,
		blobName,
		file,
		contentType,
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
		MimeType:    contentType,
		UploadedBy:  userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		span.RecordError(err)
		// Try to clean up the blob since DB insert failed
		_ = s.storage.DeleteFile(ctx, s.config.AttachmentsBucket, blobName)
		return FileInfo{}, fmt.Errorf("failed to create attachment record: %w", err)
	}

	// Generate a presigned URL for the uploaded file (30 minutes)
	accessURL, err := s.storage.GenerateAccessURL(
		ctx,
		s.config.AttachmentsBucket,
		blobName,
		30*time.Minute,
	)
	if err != nil {
		span.RecordError(err)
		if cleanupErr := s.DeleteAttachment(ctx, attachment.ID, userID); cleanupErr != nil {
			s.log.Error(ctx, "failed to clean up attachment after access URL failure", "error", cleanupErr, "attachment_id", attachment.ID)
		}
		return FileInfo{}, fmt.Errorf("failed to generate access URL: %w", err)
	}

	s.maybeEnqueueImageOptimization(ctx, blobName, contentType)

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

func attachmentValidationError(err error) error {
	switch {
	case errors.Is(err, validate.ErrFileTooLarge):
		return fmt.Errorf("%w: maximum attachment size is 25 MB", ErrFileTooLarge)
	case errors.Is(err, validate.ErrInvalidFileType):
		return fmt.Errorf("%w: upload an image, MP4 video, PDF, Word, Excel, PowerPoint, text, or CSV file", ErrInvalidFileType)
	case errors.Is(err, validate.ErrEmptyFile):
		return fmt.Errorf("%w: the file is empty", ErrInvalidFile)
	case errors.Is(err, validate.ErrFileNameTooLong):
		return fmt.Errorf("%w: filename must be 255 characters or fewer", ErrInvalidFile)
	default:
		return fmt.Errorf("%w: %v", ErrInvalidFile, err)
	}
}

func detectFileContentType(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer[:n])
	if contentType == "application/octet-stream" {
		if headerContentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type")); headerContentType != "" {
			contentType = headerContentType
		}
	}

	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])), nil
}

func (s *Service) maybeEnqueueImageOptimization(ctx context.Context, blobName, contentType string) {
	if s.optimizer == nil || !isOptimizableImageType(contentType) {
		return
	}

	if err := s.optimizer.EnqueueAttachmentImageOptimization(tasks.AttachmentImageOptimizationPayload{
		BlobName: blobName,
	}); err != nil {
		s.log.Error(ctx, "failed to enqueue attachment image optimization", "error", err, "blob_name", blobName)
	}
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

	fileInfos := make([]FileInfo, 0, len(attachments))
	for _, attachment := range attachments {
		// Use the stored blob name instead of generating a new one
		blobName := attachment.BlobName

		// Generate a presigned URL for each file (15 minutes)
		accessURL, err := s.storage.GenerateAccessURL(
			ctx,
			s.config.AttachmentsBucket,
			blobName,
			15*time.Minute,
		)
		if err != nil {
			span.RecordError(err)
			s.log.Error(ctx, "failed to generate attachment access URL", "error", err, "attachment_id", attachment.ID)
			continue
		}

		fileInfos = append(fileInfos, FileInfo{
			ID:         attachment.ID,
			Filename:   attachment.Filename,
			BlobName:   blobName,
			Size:       attachment.Size,
			MimeType:   attachment.MimeType,
			URL:        accessURL,
			CreatedAt:  attachment.CreatedAt,
			UploadedBy: attachment.UploadedBy,
		})

	}

	return fileInfos, nil
}

// ResolveAttachmentAccessURL returns fresh, short-lived object access for an
// attachment after confirming that it belongs to the requested workspace.
func (s *Service) ResolveAttachmentAccessURL(ctx context.Context, id, workspaceID uuid.UUID, expiry time.Duration) (FileInfo, error) {
	if id == uuid.Nil || workspaceID == uuid.Nil || expiry <= 0 {
		return FileInfo{}, ErrInvalidFile
	}

	attachment, err := s.repo.GetAttachmentByID(ctx, id)
	if err != nil {
		return FileInfo{}, err
	}
	if attachment.WorkspaceID != workspaceID {
		return FileInfo{}, ErrNotFound
	}

	accessURL, err := s.storage.GenerateAccessURL(
		ctx,
		s.config.AttachmentsBucket,
		attachment.BlobName,
		expiry,
	)
	if err != nil {
		return FileInfo{}, fmt.Errorf("failed to generate attachment access URL: %w", err)
	}

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

// ResolveStoryMediaAccessURL returns fresh, short-lived object access only for
// an image or MP4 linked to the exact story and workspace.
func (s *Service) ResolveStoryMediaAccessURL(
	ctx context.Context,
	storyID, attachmentID, workspaceID uuid.UUID,
	expiry time.Duration,
) (FileInfo, error) {
	if storyID == uuid.Nil || attachmentID == uuid.Nil || workspaceID == uuid.Nil || expiry <= 0 {
		return FileInfo{}, ErrInvalidFile
	}

	attachment, err := s.repo.AuthorizeStoryMedia(ctx, storyID, attachmentID, workspaceID)
	if err != nil {
		return FileInfo{}, err
	}
	if !isInlineMediaContentType(attachment.MimeType) {
		return FileInfo{}, ErrNotFound
	}

	accessURL, err := s.storage.GenerateAccessURL(
		ctx,
		s.config.AttachmentsBucket,
		attachment.BlobName,
		expiry,
	)
	if err != nil {
		return FileInfo{}, fmt.Errorf("generate story media access URL: %w", err)
	}

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

	return s.deleteStoredAttachment(ctx, span, attachment)
}

// DeleteDocumentMedia deletes an attachment after a document handler has
// authorized the editor and confirms the attachment belongs to the workspace.
func (s *Service) DeleteDocumentMedia(ctx context.Context, id, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.deleteDocumentMedia")
	ctx, span := web.AddSpan(ctx, "core.attachments.DeleteDocumentMedia")
	defer span.End()
	return s.deleteOrphanedMedia(ctx, span, id, workspaceID)
}

// DeleteOrphanedMedia removes media whose owning feature has already deleted
// its final database relation. The caller must determine orphan status inside
// the same transaction that removes that relation.
func (s *Service) DeleteOrphanedMedia(ctx context.Context, id, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.deleteOrphanedMedia")
	ctx, span := web.AddSpan(ctx, "core.attachments.DeleteOrphanedMedia")
	defer span.End()
	return s.deleteOrphanedMedia(ctx, span, id, workspaceID)
}

func (s *Service) deleteOrphanedMedia(ctx context.Context, span trace.Span, id, workspaceID uuid.UUID) error {
	attachment, err := s.repo.GetAttachmentByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if attachment.WorkspaceID != workspaceID {
		span.RecordError(ErrNotFound)
		return ErrNotFound
	}

	return s.deleteStoredAttachment(ctx, span, attachment)
}

// DeleteStoryMedia unlinks only the exact story-media relation. Dedicated
// storage is removed when no document, generic story attachment, or other
// inline story still references the attachment.
func (s *Service) DeleteStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.deleteStoryMedia")
	ctx, span := web.AddSpan(ctx, "core.attachments.DeleteStoryMedia")
	defer span.End()

	if storyID == uuid.Nil || attachmentID == uuid.Nil || workspaceID == uuid.Nil {
		return ErrInvalidFile
	}

	attachment, err := s.repo.AuthorizeStoryMedia(ctx, storyID, attachmentID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if !isInlineMediaContentType(attachment.MimeType) {
		span.RecordError(ErrNotFound)
		return ErrNotFound
	}

	isOrphaned, err := s.repo.UnlinkStoryMedia(ctx, storyID, attachmentID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if !isOrphaned {
		return nil
	}

	return s.deleteStoredAttachment(ctx, span, attachment)
}

func (s *Service) deleteStoredAttachment(ctx context.Context, span trace.Span, attachment CoreAttachment) error {
	// Use the stored blob name if available, otherwise generate it
	blobName := attachment.BlobName
	// Delete from database first
	err := s.repo.DeleteAttachment(ctx, attachment.ID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete from storage
	err = s.storage.DeleteFile(ctx, s.config.AttachmentsBucket, blobName)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to delete blob from storage", "error", err)
		// We don't return this error since the DB record is already deleted
	}

	span.AddEvent("attachment deleted", trace.WithAttributes(
		attribute.String("attachment_id", attachment.ID.String()),
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

// UploadProfileImage uploads a profile image and returns the blob name.
func (s *Service) UploadProfileImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID) (string, error) {
	s.log.Info(ctx, "core.attachments.UploadProfileImage")
	ctx, span := web.AddSpan(ctx, "core.attachments.UploadProfileImage")
	defer span.End()

	// Validate file using your existing validator
	if err := validate.ProfileImage(file, fileHeader); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("invalid profile image: %w", err)
	}

	// Generate a unique filename for profile images
	blobName := validate.GenerateFileName(fileHeader.Filename)
	upload, err := prepareUploadFile(file, fileHeader, avatarImagePolicy)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to prepare profile image upload: %w", err)
	}

	// Upload to storage profile images container
	_, err = s.storage.UploadFile(
		ctx,
		s.config.ProfilesBucket,
		blobName,
		bytes.NewReader(upload.Data),
		upload.ContentType,
	)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to upload profile image to storage: %w", err)
	}

	span.AddEvent("profile image uploaded", trace.WithAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("filename", fileHeader.Filename),
		attribute.String("blob_name", blobName),
	))

	return blobName, nil
}

func (s *Service) UploadProfileImageFromURL(ctx context.Context, imageURL string, userID uuid.UUID) (string, error) {
	if strings.TrimSpace(imageURL) == "" {
		return "", fmt.Errorf("image url is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build image request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	if resp.ContentLength > 0 && resp.ContentLength > validate.MaxProfileImageSize {
		return "", validate.ErrFileTooLarge
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, validate.MaxProfileImageSize+1))
	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}
	if int64(len(data)) > validate.MaxProfileImageSize {
		return "", validate.ErrFileTooLarge
	}

	mimeType := http.DetectContentType(data)
	if !isAllowedImageType(mimeType) {
		return "", validate.ErrInvalidFileType
	}

	ext := imageExtensionForContentType(mimeType)
	if ext == "" {
		return "", validate.ErrInvalidFileType
	}

	filename := "image" + ext
	blobName := validate.GenerateFileName(filename)

	if optimized, ok := optimizeImageBytes(data, mimeType, avatarImagePolicy); ok {
		data = optimized.Data
		mimeType = optimized.ContentType
	}

	if _, err := s.storage.UploadFile(ctx, s.config.ProfilesBucket, blobName, bytes.NewReader(data), mimeType); err != nil {
		return "", fmt.Errorf("failed to upload profile image to storage: %w", err)
	}

	return blobName, nil
}

// DeleteProfileImage deletes a profile image from storage
func (s *Service) DeleteProfileImage(ctx context.Context, avatarURL string) error {
	s.log.Info(ctx, "core.attachments.DeleteProfileImage")
	ctx, span := web.AddSpan(ctx, "core.attachments.DeleteProfileImage")
	defer span.End()

	if avatarURL == "" {
		return nil // Nothing to delete
	}

	blobName, err := s.getObjectNameFromURL(avatarURL, s.config.ProfilesBucket)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Delete from storage
	err = s.storage.DeleteFile(ctx, s.config.ProfilesBucket, blobName)
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

// UploadWorkspaceLogo uploads a workspace logo and returns the blob name.
func (s *Service) UploadWorkspaceLogo(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, workspaceID uuid.UUID) (string, error) {
	s.log.Info(ctx, "business.core.attachments.UploadWorkspaceLogo")
	ctx, span := web.AddSpan(ctx, "business.core.attachments.UploadWorkspaceLogo")
	defer span.End()

	if err := validate.WorkspaceLogo(file, fileHeader); err != nil {
		return "", fmt.Errorf("workspace logo validation failed: %w", err)
	}

	blobName := validate.GenerateFileName(fileHeader.Filename)
	upload, err := prepareUploadFile(file, fileHeader, avatarImagePolicy)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to prepare workspace logo upload: %w", err)
	}

	if _, err := s.storage.UploadFile(ctx, s.config.LogosBucket, blobName, bytes.NewReader(upload.Data), upload.ContentType); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to upload workspace logo: %w", err)
	}

	span.AddEvent("workspace logo uploaded.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("blob_name", blobName),
	))

	return blobName, nil
}

func (s *Service) DeleteWorkspaceLogo(ctx context.Context, logoURL string) error {
	s.log.Info(ctx, "business.core.attachments.DeleteWorkspaceLogo")
	ctx, span := web.AddSpan(ctx, "business.core.attachments.DeleteWorkspaceLogo")
	defer span.End()

	if logoURL == "" {
		return nil // Nothing to delete
	}

	blobName, err := s.getObjectNameFromURL(logoURL, s.config.LogosBucket)
	if err != nil {
		span.RecordError(err)
		return err
	}

	if err := s.storage.DeleteFile(ctx, s.config.LogosBucket, blobName); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete workspace logo: %w", err)
	}

	span.AddEvent("workspace logo deleted.", trace.WithAttributes(
		attribute.String("blob_name", blobName),
	))

	return nil
}

func (s *Service) ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error) {
	if strings.TrimSpace(avatar) == "" {
		return "", nil
	}
	if isHTTPURL(avatar) {
		return avatar, nil
	}
	return s.storage.GenerateAccessURL(ctx, s.config.ProfilesBucket, avatar, expiry)
}

func (s *Service) ResolveWorkspaceLogoURL(ctx context.Context, logo string, expiry time.Duration) (string, error) {
	if strings.TrimSpace(logo) == "" {
		return "", nil
	}
	if isHTTPURL(logo) {
		return logo, nil
	}
	return s.storage.GenerateAccessURL(ctx, s.config.LogosBucket, logo, expiry)
}

func (s *Service) getObjectNameFromURL(fileURL, container string) (string, error) {
	parsed, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid file URL: %w", err)
	}

	if parsed.Scheme == "" && parsed.Host == "" {
		path := strings.TrimPrefix(fileURL, "/")
		if path == "" {
			return "", fmt.Errorf("invalid file URL format")
		}
		return path, nil
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

func isAllowedImageType(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

type preparedUpload struct {
	Data        []byte
	ContentType string
}

func prepareUploadFile(file multipart.File, fileHeader *multipart.FileHeader, policy imageOptimizationPolicy) (preparedUpload, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return preparedUpload{}, err
	}

	contentType := fileHeader.Header.Get("Content-Type")
	detectedContentType := http.DetectContentType(data)
	if isAllowedImageType(detectedContentType) {
		contentType = detectedContentType
	}
	if contentType == "" {
		contentType = detectedContentType
	}

	if optimized, ok := optimizeImageBytes(data, contentType, policy); ok {
		data = optimized.Data
		contentType = optimized.ContentType
	}

	return preparedUpload{
		Data:        data,
		ContentType: contentType,
	}, nil
}

func imageExtensionForContentType(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func isHTTPURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}
