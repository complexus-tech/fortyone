package attachmentsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) EnsureThumbnail(ctx context.Context, item attachmentdomain.Thumbnail) (attachmentdomain.Thumbnail, error) {
	if err := r.configured(); err != nil {
		return item, err
	}
	err := r.pool.QueryRow(ctx, `SELECT thumbnail_id, status, retry_after, byte_size
		FROM public.attachment_thumbnails WHERE attachment_id = $1 AND workspace_id = $2 AND source_version = $3`,
		item.AttachmentID, item.WorkspaceID, item.SourceVersion).Scan(&item.ID, &item.Status, &item.RetryAfter, &item.ByteSize)
	if err == nil || !errors.Is(err, pgx.ErrNoRows) {
		return item, err
	}
	// Tenant-scoped INSERT SELECT prevents a worker or caller from pairing a
	// valid attachment with a different workspace.
	err = r.pool.QueryRow(ctx, `INSERT INTO public.attachment_thumbnails
		(thumbnail_id, attachment_id, workspace_id, source_version, storage_provider, container_name, blob_name)
		SELECT $1, attachment_id, workspace_id, $4, $5, $6, $7 FROM public.attachments
		WHERE attachment_id = $2 AND workspace_id = $3
		ON CONFLICT (attachment_id, source_version) DO UPDATE SET source_version = EXCLUDED.source_version
		RETURNING thumbnail_id, status, retry_after, byte_size`, item.ID, item.AttachmentID, item.WorkspaceID,
		item.SourceVersion, item.Provider, item.Container, item.BlobName).Scan(&item.ID, &item.Status, &item.RetryAfter, &item.ByteSize)
	if errors.Is(err, pgx.ErrNoRows) {
		return item, attachmentdomain.ErrNotFound
	}
	return item, err
}

func (r *Repository) ClaimThumbnail(ctx context.Context, id, workspaceID uuid.UUID) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	result, err := r.pool.Exec(ctx, `UPDATE public.attachment_thumbnails SET status = 'processing',
		lease_expires_at = CURRENT_TIMESTAMP + INTERVAL '1 minute', retry_after = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE thumbnail_id = $1 AND workspace_id = $2 AND
		(status = 'pending' OR (status = 'failed' AND retry_after <= CURRENT_TIMESTAMP)
		OR (status = 'processing' AND lease_expires_at <= CURRENT_TIMESTAMP))`, id, workspaceID)
	return err == nil && result.RowsAffected() == 1, err
}

func (r *Repository) CompleteThumbnail(ctx context.Context, item attachmentdomain.Thumbnail, sourceUpdatedAt time.Time, bytes int32) error {
	if err := r.configured(); err != nil {
		return err
	}
	result, err := r.pool.Exec(ctx, `UPDATE public.attachment_thumbnails AS thumbnail SET status = 'ready',
		byte_size = $4, lease_expires_at = NULL, retry_after = NULL, updated_at = CURRENT_TIMESTAMP
		FROM public.attachments AS attachment
		WHERE thumbnail.thumbnail_id = $1 AND thumbnail.workspace_id = $2 AND thumbnail.status = 'processing'
		AND attachment.attachment_id = thumbnail.attachment_id AND attachment.workspace_id = thumbnail.workspace_id
		AND attachment.updated_at = $3 AND attachment.scan_status NOT IN ('pending', 'infected', 'failed')`,
		item.ID, item.WorkspaceID, sourceUpdatedAt, bytes)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return attachmentdomain.ErrStateConflict
	}
	// Superseded derivatives retire through the same durable object deletion
	// mechanism as attachment deletion.
	_, err = r.pool.Exec(ctx, `DELETE FROM public.attachment_thumbnails WHERE attachment_id = $1
		AND workspace_id = $2 AND thumbnail_id <> $3
		AND created_at < (SELECT created_at FROM public.attachment_thumbnails WHERE thumbnail_id = $3 AND workspace_id = $2)`, item.AttachmentID, item.WorkspaceID, item.ID)
	return err
}

func (r *Repository) FailThumbnail(ctx context.Context, id, workspaceID uuid.UUID) error {
	if err := r.configured(); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE public.attachment_thumbnails SET status = 'failed',
		lease_expires_at = NULL, retry_after = CURRENT_TIMESTAMP + INTERVAL '5 minutes', updated_at = CURRENT_TIMESTAMP
		WHERE thumbnail_id = $1 AND workspace_id = $2`, id, workspaceID)
	return err
}

func (r *Repository) ResetThumbnail(ctx context.Context, id, workspaceID uuid.UUID) error {
	if err := r.configured(); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE public.attachment_thumbnails SET status = 'pending', retry_after = NULL,
		updated_at = CURRENT_TIMESTAMP WHERE thumbnail_id = $1 AND workspace_id = $2 AND status = 'failed'`, id, workspaceID)
	return err
}

// RetireThumbnail also reactivates a completed deletion. This closes the race
// where the original is deleted while a worker is uploading its derivative.
func (r *Repository) RetireThumbnail(ctx context.Context, item attachmentdomain.Thumbnail) error {
	if err := r.configured(); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO public.attachment_object_deletion_outbox
		(attachment_id, workspace_id, storage_provider, container_name, blob_name)
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT (attachment_id) DO UPDATE SET
		status = 'pending', attempt_count = 0, next_attempt_at = CURRENT_TIMESTAMP,
		processing_started_at = NULL, claim_token = NULL, completed_at = NULL, last_error = NULL, updated_at = CURRENT_TIMESTAMP`,
		item.ID, item.WorkspaceID, item.Provider, item.Container, item.BlobName)
	if err != nil {
		return fmt.Errorf("retire thumbnail object: %w", err)
	}
	_, err = r.pool.Exec(ctx, `DELETE FROM public.attachment_thumbnails WHERE thumbnail_id = $1 AND workspace_id = $2`, item.ID, item.WorkspaceID)
	return err
}
