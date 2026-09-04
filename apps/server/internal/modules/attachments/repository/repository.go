package attachmentsrepository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	attachmentssql "github.com/complexus-tech/projects-api/internal/modules/attachments/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool       *pgxpool.Pool
	queries    *attachmentssql.Queries
	transactor platformdatabase.Transactor
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{
		pool:       pool,
		queries:    attachmentssql.New(pool),
		transactor: platformdatabase.NewTransactor(pool),
	}
}

func (repository *Repository) CreateAttachment(
	ctx context.Context,
	attachment attachmentdomain.Attachment,
) (attachmentdomain.Attachment, error) {
	if err := repository.configured(); err != nil {
		return attachmentdomain.Attachment{}, err
	}
	var uploaderID *uuid.UUID
	if attachment.UploadedBy != uuid.Nil {
		uploaderID = &attachment.UploadedBy
	}
	row, err := repository.queries.CreateAttachment(ctx, attachmentssql.CreateAttachmentParams{
		Filename:           attachment.Filename,
		BlobName:           attachment.BlobName,
		Size:               attachment.Size,
		MimeType:           attachment.MimeType,
		UploadedBy:         uploaderID,
		WorkspaceID:        attachment.WorkspaceID,
		ScanStatus:         string(defaultScanStatus(attachment.ScanStatus)),
		OptimizationStatus: string(defaultOptimizationStatus(attachment.OptimizationStatus)),
	})
	if err != nil {
		return attachmentdomain.Attachment{}, fmt.Errorf("create attachment: %w", err)
	}
	return toDomain(row), nil
}

func (repository *Repository) GetAttachmentByID(
	ctx context.Context,
	attachmentID, workspaceID uuid.UUID,
) (attachmentdomain.Attachment, error) {
	if err := repository.configured(); err != nil {
		return attachmentdomain.Attachment{}, err
	}
	row, err := repository.queries.GetWorkspaceAttachment(ctx, attachmentssql.GetWorkspaceAttachmentParams{
		AttachmentID: attachmentID,
		WorkspaceID:  workspaceID,
	})
	if err != nil {
		return attachmentdomain.Attachment{}, mapNotFound("get workspace attachment", err)
	}
	return toDomain(row), nil
}

func (repository *Repository) GetAttachmentsByStoryID(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
) ([]attachmentdomain.Attachment, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListStoryAttachments(ctx, attachmentssql.ListStoryAttachmentsParams{
		StoryID: storyID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list story attachments: %w", err)
	}
	attachments := make([]attachmentdomain.Attachment, 0, len(rows))
	for _, row := range rows {
		attachments = append(attachments, toDomain(row))
	}
	return attachments, nil
}

func (repository *Repository) StoryExistsInWorkspace(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
) (bool, error) {
	if err := repository.configured(); err != nil {
		return false, err
	}
	exists, err := repository.queries.StoryExistsInWorkspace(ctx, attachmentssql.StoryExistsInWorkspaceParams{
		StoryID: storyID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("check story workspace: %w", err)
	}
	return exists, nil
}

func (repository *Repository) AuthorizeStoryAttachment(
	ctx context.Context,
	storyID, attachmentID, workspaceID uuid.UUID,
) (attachmentdomain.Attachment, error) {
	if err := repository.configured(); err != nil {
		return attachmentdomain.Attachment{}, err
	}
	row, err := repository.queries.AuthorizeWorkspaceStoryAttachment(ctx, attachmentssql.AuthorizeWorkspaceStoryAttachmentParams{
		StoryID: storyID, AttachmentID: attachmentID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return attachmentdomain.Attachment{}, mapNotFound("authorize story attachment", err)
	}
	return toDomain(row), nil
}

func (repository *Repository) LinkAttachmentToStory(
	ctx context.Context,
	storyID, attachmentID, workspaceID uuid.UUID,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	_, err := repository.queries.LinkWorkspaceStoryAttachment(ctx, attachmentssql.LinkWorkspaceStoryAttachmentParams{
		StoryID: storyID, AttachmentID: attachmentID, WorkspaceID: workspaceID,
	})
	return mapNotFound("link story attachment", err)
}

func (repository *Repository) LinkStoryMedia(
	ctx context.Context,
	storyID, attachmentID, createdBy, workspaceID uuid.UUID,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	_, err := repository.queries.LinkWorkspaceStoryMedia(ctx, attachmentssql.LinkWorkspaceStoryMediaParams{
		StoryID: storyID, AttachmentID: attachmentID, CreatedBy: createdBy, WorkspaceID: workspaceID,
	})
	return mapNotFound("link story media", err)
}

func (repository *Repository) AuthorizeStoryMedia(
	ctx context.Context,
	storyID, attachmentID, workspaceID uuid.UUID,
) (attachmentdomain.Attachment, error) {
	if err := repository.configured(); err != nil {
		return attachmentdomain.Attachment{}, err
	}
	row, err := repository.queries.AuthorizeWorkspaceStoryMedia(ctx, attachmentssql.AuthorizeWorkspaceStoryMediaParams{
		StoryID: storyID, AttachmentID: attachmentID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return attachmentdomain.Attachment{}, mapNotFound("authorize story media", err)
	}
	return toDomain(row), nil
}

func (repository *Repository) UnlinkStoryMedia(
	ctx context.Context,
	storyID, attachmentID, workspaceID uuid.UUID,
) (bool, error) {
	if err := repository.configured(); err != nil {
		return false, err
	}
	isOrphaned := false
	err := repository.transactor.WithinTransaction(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		queries := attachmentssql.New(tx)
		_, err := queries.UnlinkWorkspaceStoryMedia(ctx, attachmentssql.UnlinkWorkspaceStoryMediaParams{
			StoryID: storyID, AttachmentID: attachmentID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return mapNotFound("unlink story media", err)
		}
		_, err = queries.DeleteUnreferencedWorkspaceAttachment(ctx, attachmentssql.DeleteUnreferencedWorkspaceAttachmentParams{
			AttachmentID: attachmentID, WorkspaceID: workspaceID,
		})
		switch {
		case err == nil:
			isOrphaned = true
			return nil
		case errors.Is(err, pgx.ErrNoRows):
			return nil
		default:
			return fmt.Errorf("delete unreferenced story media: %w", err)
		}
	})
	if err != nil {
		return false, err
	}
	return isOrphaned, nil
}

func (repository *Repository) DeleteAttachment(
	ctx context.Context,
	attachmentID, workspaceID uuid.UUID,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	count, err := repository.queries.DeleteWorkspaceAttachment(ctx, attachmentssql.DeleteWorkspaceAttachmentParams{
		AttachmentID: attachmentID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("delete workspace attachment: %w", err)
	}
	if count == 0 {
		return attachmentdomain.ErrNotFound
	}
	return nil
}

func (repository *Repository) DeleteAttachmentIfUnreferenced(
	ctx context.Context,
	attachmentID, workspaceID uuid.UUID,
) (bool, error) {
	if err := repository.configured(); err != nil {
		return false, err
	}
	_, err := repository.queries.DeleteUnreferencedWorkspaceAttachment(ctx, attachmentssql.DeleteUnreferencedWorkspaceAttachmentParams{
		AttachmentID: attachmentID, WorkspaceID: workspaceID,
	})
	if err == nil {
		return true, nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		if _, lookupErr := repository.GetAttachmentByID(ctx, attachmentID, workspaceID); lookupErr != nil {
			return false, lookupErr
		}
		return false, nil
	}
	return false, fmt.Errorf("delete unreferenced attachment: %w", err)
}

func (repository *Repository) StartAttachmentOptimization(
	ctx context.Context,
	attachmentID, workspaceID uuid.UUID,
	lease time.Duration,
) (attachmentdomain.Attachment, error) {
	if err := repository.configured(); err != nil {
		return attachmentdomain.Attachment{}, err
	}
	leaseSeconds := int64(lease / time.Second)
	if leaseSeconds < 1 || leaseSeconds > math.MaxInt32 {
		return attachmentdomain.Attachment{}, fmt.Errorf("invalid attachment optimization lease")
	}
	row, err := repository.queries.StartWorkspaceAttachmentOptimization(ctx, attachmentssql.StartWorkspaceAttachmentOptimizationParams{
		AttachmentID: attachmentID, WorkspaceID: workspaceID, LeaseSeconds: int32(leaseSeconds),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return attachmentdomain.Attachment{}, attachmentdomain.ErrStateConflict
		}
		return attachmentdomain.Attachment{}, fmt.Errorf("start attachment optimization: %w", err)
	}
	return toDomain(row), nil
}

func (repository *Repository) CompleteAttachmentOptimization(
	ctx context.Context,
	attachmentID, workspaceID uuid.UUID,
	size int64,
	mimeType string,
	status attachmentdomain.OptimizationStatus,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	if status != attachmentdomain.OptimizationSucceeded && status != attachmentdomain.OptimizationSkipped {
		return fmt.Errorf("invalid terminal attachment optimization status %q", status)
	}
	count, err := repository.queries.CompleteWorkspaceAttachmentOptimization(ctx, attachmentssql.CompleteWorkspaceAttachmentOptimizationParams{
		AttachmentID: attachmentID, WorkspaceID: workspaceID, Size: size, MimeType: mimeType,
		OptimizationStatus: string(status),
	})
	if err != nil {
		return fmt.Errorf("complete attachment optimization: %w", err)
	}
	if count == 0 {
		return attachmentdomain.ErrStateConflict
	}
	return nil
}

func (repository *Repository) FailAttachmentOptimization(
	ctx context.Context,
	attachmentID, workspaceID uuid.UUID,
	reason string,
	queued bool,
) error {
	if err := repository.configured(); err != nil {
		return err
	}
	reason = boundedFailure(reason)
	params := attachmentssql.FailWorkspaceAttachmentOptimizationParams{
		AttachmentID: attachmentID, WorkspaceID: workspaceID, ErrorMessage: &reason,
	}
	var (
		count int64
		err   error
	)
	if queued {
		count, err = repository.queries.FailQueuedWorkspaceAttachmentOptimization(ctx, attachmentssql.FailQueuedWorkspaceAttachmentOptimizationParams(params))
	} else {
		count, err = repository.queries.FailWorkspaceAttachmentOptimization(ctx, params)
	}
	if err != nil {
		return fmt.Errorf("fail attachment optimization: %w", err)
	}
	if count == 0 {
		return attachmentdomain.ErrStateConflict
	}
	return nil
}

func (repository *Repository) configured() error {
	if repository == nil || repository.pool == nil || repository.queries == nil {
		return attachmentdomain.ErrNotConfigured
	}
	return nil
}

func mapNotFound(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return attachmentdomain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return nil
}

func defaultScanStatus(status attachmentdomain.ScanStatus) attachmentdomain.ScanStatus {
	if status == "" {
		return attachmentdomain.ScanStatusUnscanned
	}
	return status
}

func defaultOptimizationStatus(status attachmentdomain.OptimizationStatus) attachmentdomain.OptimizationStatus {
	if status == "" {
		return attachmentdomain.OptimizationNotRequested
	}
	return status
}

func boundedFailure(reason string) string {
	const limit = 512
	if len(reason) <= limit {
		return reason
	}
	return reason[:limit]
}
