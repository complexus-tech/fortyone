package attachments

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

const ThumbnailMaxBytes = 2 << 20
const thumbnailMaxSourceBytes = 25 << 20
const thumbnailRendererVersion = "first-page-jpeg-v1"

var (
	ErrThumbnailPending     = errors.New("attachment thumbnail is being prepared")
	ErrThumbnailUnavailable = errors.New("attachment thumbnail is unavailable")
	// The worker's global concurrency also handles unrelated jobs. Keep expensive
	// image/PDF work to two at a time independently of that setting.
	thumbnailRenderSlots = make(chan struct{}, 2)
)

type thumbnailRepository interface {
	EnsureThumbnail(context.Context, attachmentdomain.Thumbnail) (attachmentdomain.Thumbnail, error)
	ClaimThumbnail(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	CompleteThumbnail(context.Context, attachmentdomain.Thumbnail, time.Time, int32) error
	FailThumbnail(context.Context, uuid.UUID, uuid.UUID) error
	ResetThumbnail(context.Context, uuid.UUID, uuid.UUID) error
	RetireThumbnail(context.Context, attachmentdomain.Thumbnail) error
}

type thumbnailEnqueuer interface {
	EnqueueAttachmentThumbnail(tasks.AttachmentThumbnailPayload) error
}

type ThumbnailAccess struct {
	URL     string
	Version string
	Size    int32
}

func thumbnailVersion(attachment CoreAttachment) string {
	value := fmt.Sprintf("%s:%s:%d:%s:%s", thumbnailRendererVersion, attachment.ID,
		attachment.Size, attachment.MimeType, attachment.UpdatedAt.UTC().Format(time.RFC3339Nano))
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func (s *Service) thumbnailIdentity(id, workspaceID uuid.UUID, version string) attachmentdomain.Thumbnail {
	return attachmentdomain.Thumbnail{
		ID: uuid.NewSHA1(id, []byte(version)), AttachmentID: id, WorkspaceID: workspaceID,
		SourceVersion: version, Provider: s.config.Provider, Container: s.config.AttachmentsBucket,
		BlobName: fmt.Sprintf("thumbnails/%s/%s/%s.jpg", workspaceID, id, version),
	}
}

func thumbnailApplicable(file CoreAttachment) bool {
	if !file.AvailableForDownload() || file.Size < 1 || file.Size > thumbnailMaxSourceBytes {
		return false
	}
	switch normalizeContentType(file.MimeType) {
	case "application/pdf", "image/jpeg", "image/png", "image/webp", "image/gif":
		return file.OptimizationStatus != attachmentdomain.OptimizationQueued && file.OptimizationStatus != attachmentdomain.OptimizationProcessing
	default:
		return false
	}
}

// ResolveThumbnail is called only after resource authorization. A signed URL
// lasts one minute, and HTTP adapters must not cache the authorization result.
func (s *Service) ResolveThumbnail(ctx context.Context, id, workspaceID uuid.UUID, refresh bool) (ThumbnailAccess, error) {
	repo, ok := s.repo.(thumbnailRepository)
	queue, queueOK := s.optimizer.(thumbnailEnqueuer)
	if !ok || !queueOK {
		return ThumbnailAccess{}, ErrThumbnailUnavailable
	}
	file, err := s.repo.GetAttachmentByID(ctx, id, workspaceID)
	if err != nil {
		return ThumbnailAccess{}, err
	}
	if !thumbnailApplicable(file) {
		return ThumbnailAccess{}, ErrThumbnailUnavailable
	}
	item, err := repo.EnsureThumbnail(ctx, s.thumbnailIdentity(id, workspaceID, thumbnailVersion(file)))
	if err != nil {
		return ThumbnailAccess{}, fmt.Errorf("resolve attachment thumbnail: %w", err)
	}
	if item.Status == "ready" && item.ByteSize != nil {
		url, err := s.storage.GenerateAccessURL(ctx, item.Container, item.BlobName, time.Minute)
		if err != nil {
			return ThumbnailAccess{}, ErrThumbnailUnavailable
		}
		return ThumbnailAccess{URL: url, Version: item.SourceVersion, Size: *item.ByteSize}, nil
	}
	if item.Status == "failed" && item.RetryAfter != nil && item.RetryAfter.After(time.Now()) {
		if !refresh {
			return ThumbnailAccess{}, ErrThumbnailUnavailable
		}
		if err := repo.ResetThumbnail(ctx, item.ID, workspaceID); err != nil {
			return ThumbnailAccess{}, err
		}
	}
	if err := queue.EnqueueAttachmentThumbnail(tasks.AttachmentThumbnailPayload{
		AttachmentID: id, WorkspaceID: workspaceID, SourceVersion: item.SourceVersion,
	}); err != nil {
		return ThumbnailAccess{}, fmt.Errorf("enqueue attachment thumbnail: %w", err)
	}
	return ThumbnailAccess{}, ErrThumbnailPending
}

// GenerateThumbnail is idempotent across workers. The database lease owns one
// source version; the final write rejects changed/deleted/infected originals.
func (s *Service) GenerateThumbnail(ctx context.Context, payload tasks.AttachmentThumbnailPayload) error {
	repo, ok := s.repo.(thumbnailRepository)
	if !ok {
		return ErrThumbnailUnavailable
	}
	item := s.thumbnailIdentity(payload.AttachmentID, payload.WorkspaceID, payload.SourceVersion)
	file, err := s.repo.GetAttachmentByID(ctx, payload.AttachmentID, payload.WorkspaceID)
	if errors.Is(err, ErrNotFound) || (err == nil && thumbnailVersion(file) != payload.SourceVersion) {
		return repo.RetireThumbnail(ctx, item)
	}
	if err != nil {
		return err
	}
	if !thumbnailApplicable(file) {
		return repo.FailThumbnail(ctx, item.ID, item.WorkspaceID)
	}
	claimed, err := repo.ClaimThumbnail(ctx, item.ID, item.WorkspaceID)
	if err != nil || !claimed {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	select {
	case thumbnailRenderSlots <- struct{}{}:
		defer func() { <-thumbnailRenderSlots }()
	case <-ctx.Done():
		return s.failThumbnail(item, ctx.Err())
	}
	data, _, err := s.storage.DownloadFile(ctx, item.Container, file.BlobName, thumbnailMaxSourceBytes)
	if err != nil {
		return s.failThumbnail(item, errors.New("download thumbnail source"))
	}
	output, err := renderAttachmentThumbnail(ctx, data, normalizeContentType(file.MimeType))
	if err != nil {
		return s.failThumbnail(item, err)
	}
	if _, err := s.storage.UploadFile(ctx, item.Container, item.BlobName, bytes.NewReader(output), "image/jpeg"); err != nil {
		return s.failThumbnail(item, errors.New("store attachment thumbnail"))
	}
	if err := repo.CompleteThumbnail(ctx, item, file.UpdatedAt, int32(len(output))); err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if retireErr := repo.RetireThumbnail(cleanup, item); retireErr != nil {
			return errors.Join(err, retireErr)
		}
		if errors.Is(err, attachmentdomain.ErrStateConflict) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Service) failThumbnail(item attachmentdomain.Thumbnail, cause error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	repo := s.repo.(thumbnailRepository)
	if err := repo.FailThumbnail(ctx, item.ID, item.WorkspaceID); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}
