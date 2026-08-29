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

	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (s *Service) UploadProfileImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID uuid.UUID) (string, error) {
	s.log.Info(ctx, "core.attachments.UploadProfileImage")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.UploadProfileImage")
	defer span.End()

	if err := validate.ProfileImage(file, fileHeader); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("invalid profile image: %w", err)
	}

	blobName := validate.GenerateFileName(fileHeader.Filename)
	upload, err := prepareUploadFile(file, fileHeader, avatarImagePolicy)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to prepare profile image upload: %w", err)
	}

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
	if userID == uuid.Nil {
		return "", fmt.Errorf("user id is required")
	}
	if s.remoteImages == nil {
		return "", fmt.Errorf("remote image downloader is not configured")
	}
	download, err := s.remoteImages.Download(ctx, imageURL)
	if err != nil {
		if errors.Is(err, safehttp.ErrResponseTooLarge) {
			return "", validate.ErrFileTooLarge
		}
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	data := download.Body
	mimeType := http.DetectContentType(data)
	if !isAllowedImageType(mimeType) {
		return "", validate.ErrInvalidFileType
	}
	extension := imageExtensionForContentType(mimeType)
	if extension == "" {
		return "", validate.ErrInvalidFileType
	}

	blobName := validate.GenerateFileName("image" + extension)
	if optimized, ok := optimizeImageBytes(data, mimeType, avatarImagePolicy); ok {
		data = optimized.Data
		mimeType = optimized.ContentType
	}
	if _, err := s.storage.UploadFile(ctx, s.config.ProfilesBucket, blobName, bytes.NewReader(data), mimeType); err != nil {
		return "", fmt.Errorf("failed to upload profile image to storage: %w", err)
	}
	return blobName, nil
}

func (s *Service) DeleteProfileImage(ctx context.Context, avatarURL string) error {
	s.log.Info(ctx, "core.attachments.DeleteProfileImage")
	ctx, span := serviceTracer.Start(ctx, "core.attachments.DeleteProfileImage")
	defer span.End()

	if avatarURL == "" {
		return nil
	}
	blobName, err := s.getObjectNameFromURL(avatarURL, s.config.ProfilesBucket)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.storage.DeleteFile(ctx, s.config.ProfilesBucket, blobName); err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to delete profile image from storage", "error", err)
	}
	span.AddEvent("profile image deleted", trace.WithAttributes(
		attribute.String("avatar_url", avatarURL),
		attribute.String("blob_name", blobName),
	))
	return nil
}

func (s *Service) UploadWorkspaceLogo(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, workspaceID uuid.UUID) (string, error) {
	s.log.Info(ctx, "business.core.attachments.UploadWorkspaceLogo")
	ctx, span := serviceTracer.Start(ctx, "business.core.attachments.UploadWorkspaceLogo")
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
	span.AddEvent("workspace logo uploaded", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("blob_name", blobName),
	))
	return blobName, nil
}

func (s *Service) DeleteWorkspaceLogo(ctx context.Context, logoURL string) error {
	s.log.Info(ctx, "business.core.attachments.DeleteWorkspaceLogo")
	ctx, span := serviceTracer.Start(ctx, "business.core.attachments.DeleteWorkspaceLogo")
	defer span.End()

	if logoURL == "" {
		return nil
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
	span.AddEvent("workspace logo deleted", trace.WithAttributes(attribute.String("blob_name", blobName)))
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
	return preparedUpload{Data: data, ContentType: contentType}, nil
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
