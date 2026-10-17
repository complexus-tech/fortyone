package attachments

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) uploadAttachment(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	userID uuid.UUID,
	workspaceID uuid.UUID,
	importID uuid.UUID,
	validateFile func(multipart.File, *multipart.FileHeader) error,
	validateContentType func(string) error,
) (FileInfo, error) {
	return s.uploadAttachmentWithDetector(ctx, file, fileHeader, userID, workspaceID, importID, validateFile, validateContentType, detectFileContentType)
}

func (s *Service) uploadAttachmentWithDetector(
	ctx context.Context,
	file multipart.File,
	fileHeader *multipart.FileHeader,
	userID, workspaceID, importID uuid.UUID,
	validateFile func(multipart.File, *multipart.FileHeader) error,
	validateContentType func(string) error,
	detectContentType func(multipart.File, *multipart.FileHeader) (string, error),
) (FileInfo, error) {
	s.log.Info(ctx, "core.attachments.upload")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.UploadAttachment")
	defer span.End()

	// Validate file
	if err := validateFile(file, fileHeader); err != nil {
		span.RecordError(err)
		return FileInfo{}, attachmentValidationError(err)
	}

	// Generate a unique filename
	blobName := validate.GenerateFileName(fileHeader.Filename)
	contentType, err := detectContentType(file, fileHeader)
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

	// Storage receives the complete validated object. MIME checks are not a
	// malware scan, so new uploads retain the truthful unscanned status.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return FileInfo{}, ErrInvalidFile
	}
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

	platformdatabase.AfterRollback(ctx, func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if err := s.storage.DeleteFile(cleanupCtx, s.config.AttachmentsBucket, blobName); err != nil {
			s.log.Error(cleanupCtx, "failed to clean up rolled back attachment object", "error", err)
		}
	})

	// Create attachment record in database
	attachmentInput := CoreAttachment{
		Filename:          fileHeader.Filename,
		BlobName:          blobName,
		Size:              fileHeader.Size,
		MimeType:          contentType,
		UploadedBy:        userID,
		WorkspaceID:       workspaceID,
		SlackFileImportID: importID,
		ScanStatus:        attachmentdomain.ScanStatusUnscanned,
		// The current optimizer reads the complete object into memory and has a
		// 25 MB download ceiling. Larger Slack imports remain available as
		// original files instead of enqueueing a job that cannot succeed.
		OptimizationStatus: optimizationStatus(contentType, s.optimizer != nil && fileHeader.Size <= validate.MaxAttachmentSize),
	}
	createAttachment := s.repo.CreateAttachment
	if importID != uuid.Nil {
		createAttachment = s.repo.CreateSlackImportAttachment
	}
	attachment, err := createAttachment(ctx, attachmentInput)
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
		if importID == uuid.Nil {
			if cleanupErr := s.DeleteAttachment(ctx, attachment.ID, workspaceID, userID); cleanupErr != nil {
				s.log.Error(ctx, "failed to clean up attachment after access URL failure", "error", cleanupErr, "attachment_id", attachment.ID)
			}
		}
		return FileInfo{}, fmt.Errorf("failed to generate access URL: %w", err)
	}

	platformdatabase.AfterCommit(ctx, func() { s.maybeEnqueueImageOptimization(ctx, attachment) })

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
	return validate.AttachmentContentType(file, fileHeader)
}
