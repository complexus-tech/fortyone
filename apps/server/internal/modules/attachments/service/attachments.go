package attachments

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var serviceTracer = otel.Tracer("github.com/complexus-tech/projects-api/internal/modules/attachments/service")

// Repository defines the storage interface for attachments
type Repository interface {
	CreateAttachment(ctx context.Context, attachment CoreAttachment) (CoreAttachment, error)
	GetAttachmentByID(ctx context.Context, id, workspaceID uuid.UUID) (CoreAttachment, error)
	GetAttachmentsByStoryID(ctx context.Context, storyID, workspaceID uuid.UUID) ([]CoreAttachment, error)
	StoryExistsInWorkspace(ctx context.Context, storyID, workspaceID uuid.UUID) (bool, error)
	AuthorizeStoryAttachment(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) (CoreAttachment, error)
	LinkStoryMedia(ctx context.Context, storyID, attachmentID, createdBy, workspaceID uuid.UUID) error
	AuthorizeStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) (CoreAttachment, error)
	UnlinkStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) (bool, error)
	DeleteAttachment(ctx context.Context, id, workspaceID uuid.UUID) error
	DeleteAttachmentIfUnreferenced(ctx context.Context, id, workspaceID uuid.UUID) (bool, error)
	LinkAttachmentToStory(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) error
	StartAttachmentOptimization(ctx context.Context, attachmentID, workspaceID uuid.UUID, lease time.Duration) (CoreAttachment, error)
	CompleteAttachmentOptimization(ctx context.Context, attachmentID, workspaceID uuid.UUID, size int64, mimeType string, status attachmentdomain.OptimizationStatus) error
	FailAttachmentOptimization(ctx context.Context, attachmentID, workspaceID uuid.UUID, reason string, queued bool) error
}

type ImageOptimizationEnqueuer interface {
	EnqueueAttachmentImageOptimization(payload tasks.AttachmentImageOptimizationPayload) error
}

type RemoteImageDownloader interface {
	Download(ctx context.Context, rawURL string) (safehttp.Download, error)
}

type Option func(*Service)

func WithRemoteImageDownloader(downloader RemoteImageDownloader) Option {
	return func(service *Service) {
		service.remoteImages = downloader
	}
}

// Service manages attachment operations
type Service struct {
	log          *logger.Logger
	repo         Repository
	storage      storage.StorageService
	config       storage.Config
	optimizer    ImageOptimizationEnqueuer
	remoteImages RemoteImageDownloader
}

// New creates a new attachment service
func New(log *logger.Logger, repo Repository, storageService storage.StorageService, config storage.Config, optimizer ImageOptimizationEnqueuer, options ...Option) *Service {
	service := &Service{
		log:          log,
		repo:         repo,
		storage:      storageService,
		config:       config,
		optimizer:    optimizer,
		remoteImages: newDefaultRemoteImageDownloader(),
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func newDefaultRemoteImageDownloader() RemoteImageDownloader {
	downloader, err := safehttp.NewDownloader(safehttp.Config{
		MaxResponseBytes: validate.MaxProfileImageSize,
		Timeout:          10 * time.Second,
	})
	if err != nil {
		panic("invalid built-in remote image downloader configuration: " + err.Error())
	}
	return downloader
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
		if cleanupErr := s.DeleteAttachment(ctx, fileInfo.ID, workspaceID, userID); cleanupErr != nil {
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
	ctx, span := serviceTracer.Start(ctx, "core.attachments.UploadAttachment")
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
		Filename:           fileHeader.Filename,
		BlobName:           blobName,
		Size:               fileHeader.Size,
		MimeType:           contentType,
		UploadedBy:         userID,
		WorkspaceID:        workspaceID,
		ScanStatus:         attachmentdomain.ScanStatusUnscanned,
		OptimizationStatus: optimizationStatus(contentType, s.optimizer != nil),
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
		if cleanupErr := s.DeleteAttachment(ctx, attachment.ID, workspaceID, userID); cleanupErr != nil {
			s.log.Error(ctx, "failed to clean up attachment after access URL failure", "error", cleanupErr, "attachment_id", attachment.ID)
		}
		return FileInfo{}, fmt.Errorf("failed to generate access URL: %w", err)
	}

	s.maybeEnqueueImageOptimization(ctx, attachment)

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

func optimizationStatus(contentType string, optimizerConfigured bool) attachmentdomain.OptimizationStatus {
	if !optimizerConfigured || !isOptimizableImageType(contentType) {
		return attachmentdomain.OptimizationNotRequested
	}
	return attachmentdomain.OptimizationQueued
}

func (s *Service) maybeEnqueueImageOptimization(ctx context.Context, attachment CoreAttachment) {
	if s.optimizer == nil || attachment.OptimizationStatus != attachmentdomain.OptimizationQueued {
		return
	}

	if err := s.optimizer.EnqueueAttachmentImageOptimization(tasks.AttachmentImageOptimizationPayload{
		AttachmentID: attachment.ID,
		WorkspaceID:  attachment.WorkspaceID,
	}); err != nil {
		s.log.Error(ctx, "failed to enqueue attachment image optimization", "error", err, "attachment_id", attachment.ID)
		if stateErr := s.repo.FailAttachmentOptimization(
			ctx,
			attachment.ID,
			attachment.WorkspaceID,
			"failed to enqueue image optimization",
			true,
		); stateErr != nil && !errors.Is(stateErr, attachmentdomain.ErrStateConflict) {
			s.log.Error(ctx, "failed to record attachment optimization enqueue failure", "error", stateErr, "attachment_id", attachment.ID)
		}
	}
}

// GetAttachmentsForStory gets all attachments for a story
func (s *Service) GetAttachmentsForStory(ctx context.Context, storyID, workspaceID uuid.UUID) ([]FileInfo, error) {
	s.log.Info(ctx, "core.attachments.getForStory")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.GetAttachmentsForStory")
	defer span.End()

	if storyID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, ErrInvalidFile
	}
	exists, err := s.repo.StoryExistsInWorkspace(ctx, storyID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("check attachment story workspace: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	// Get attachments from database
	attachments, err := s.repo.GetAttachmentsByStoryID(ctx, storyID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	fileInfos := make([]FileInfo, 0, len(attachments))
	for _, attachment := range attachments {
		if !attachment.AvailableForDownload() {
			continue
		}
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

	attachment, err := s.repo.GetAttachmentByID(ctx, id, workspaceID)
	if err != nil {
		return FileInfo{}, err
	}
	if !attachment.AvailableForDownload() {
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
	if !attachment.AvailableForDownload() || !isInlineMediaContentType(attachment.MimeType) {
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
func (s *Service) DeleteAttachment(ctx context.Context, id, workspaceID, userID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.delete")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.DeleteAttachment")
	defer span.End()

	// Get attachment to check ownership and get the filename
	if id == uuid.Nil || workspaceID == uuid.Nil || userID == uuid.Nil {
		return ErrInvalidFile
	}
	attachment, err := s.repo.GetAttachmentByID(ctx, id, workspaceID)
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

// DeleteStoryAttachment removes an attachment only after authorizing the exact
// story, attachment, and workspace relation. Workspace admins may remove any
// attachment in their workspace; other users may remove only their uploads.
func (s *Service) DeleteStoryAttachment(
	ctx context.Context,
	storyID, attachmentID, workspaceID, userID uuid.UUID,
	isAdmin bool,
) error {
	s.log.Info(ctx, "core.attachments.deleteStoryAttachment")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.DeleteStoryAttachment")
	defer span.End()

	if storyID == uuid.Nil || attachmentID == uuid.Nil || workspaceID == uuid.Nil || userID == uuid.Nil {
		return ErrInvalidFile
	}

	attachment, err := s.repo.AuthorizeStoryAttachment(ctx, storyID, attachmentID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if attachment.UploadedBy != userID && !isAdmin {
		span.RecordError(ErrUnauthorized)
		return ErrUnauthorized
	}

	return s.deleteStoredAttachment(ctx, span, attachment)
}

// DeleteDocumentMedia deletes an attachment after a document handler has
// authorized the editor and confirms the attachment belongs to the workspace.
func (s *Service) DeleteDocumentMedia(ctx context.Context, id, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.deleteDocumentMedia")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.DeleteDocumentMedia")
	defer span.End()
	return s.deleteOrphanedMedia(ctx, span, id, workspaceID)
}

// DeleteOrphanedMedia removes media whose owning feature has already deleted
// its final database relation. The caller must determine orphan status inside
// the same transaction that removes that relation.
func (s *Service) DeleteOrphanedMedia(ctx context.Context, id, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.deleteOrphanedMedia")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.DeleteOrphanedMedia")
	defer span.End()
	return s.deleteOrphanedMedia(ctx, span, id, workspaceID)
}

func (s *Service) deleteOrphanedMedia(ctx context.Context, span trace.Span, id, workspaceID uuid.UUID) error {
	attachment, err := s.repo.GetAttachmentByID(ctx, id, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	deleted, err := s.repo.DeleteAttachmentIfUnreferenced(ctx, attachment.ID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if !deleted {
		return nil
	}
	return s.deleteStoredObject(ctx, span, attachment)
}

// DeleteStoryMedia unlinks only the exact story-media relation. Dedicated
// storage is removed when no document, generic story attachment, or other
// inline story still references the attachment.
func (s *Service) DeleteStoryMedia(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.deleteStoryMedia")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.DeleteStoryMedia")
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

	return s.deleteStoredObject(ctx, span, attachment)
}

func (s *Service) deleteStoredAttachment(ctx context.Context, span trace.Span, attachment CoreAttachment) error {
	// Use the stored blob name if available, otherwise generate it
	blobName := attachment.BlobName
	// Delete from database first
	err := s.repo.DeleteAttachment(ctx, attachment.ID, attachment.WorkspaceID)
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

func (s *Service) deleteStoredObject(ctx context.Context, span trace.Span, attachment CoreAttachment) error {
	if err := s.storage.DeleteFile(ctx, s.config.AttachmentsBucket, attachment.BlobName); err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to delete orphaned blob from storage", "error", err, "attachment_id", attachment.ID)
	}
	span.AddEvent("attachment deleted", trace.WithAttributes(
		attribute.String("attachment_id", attachment.ID.String()),
	))
	return nil
}

// LinkAttachmentToStory links an attachment to a story
func (s *Service) LinkAttachmentToStory(ctx context.Context, storyID, attachmentID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "core.attachments.linkToStory")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.LinkAttachmentToStory")
	defer span.End()

	// Link attachment to story
	err := s.repo.LinkAttachmentToStory(ctx, storyID, attachmentID, workspaceID)
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
	ctx, span := serviceTracer.Start(ctx, "core.attachments.UploadAndLinkToStory")
	defer span.End()

	// First upload the attachment
	fileInfo, err := s.UploadAttachment(ctx, file, fileHeader, userID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return FileInfo{}, err
	}

	// Then link it to the story
	err = s.LinkAttachmentToStory(ctx, storyID, fileInfo.ID, workspaceID)
	if err != nil {
		span.RecordError(err)
		// Try to clean up the attachment since linking failed
		_ = s.DeleteAttachment(ctx, fileInfo.ID, workspaceID, userID)
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
