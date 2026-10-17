package attachmentsrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	attachmentssql "github.com/complexus-tech/projects-api/internal/modules/attachments/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) LockTeamDeletionAttachments(ctx context.Context, tx pgx.Tx, teamID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	if tx == nil || teamID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, errors.New("team attachment deletion transaction and scope are required")
	}
	ids, err := attachmentssql.New(tx).LockTeamDeletionAttachments(ctx, attachmentssql.LockTeamDeletionAttachmentsParams{
		TeamID: teamID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("lock team attachment candidates: %w", err)
	}
	return ids, nil
}

func (repository *Repository) RetireTeamDeletionAttachments(
	ctx context.Context, tx pgx.Tx, ids []uuid.UUID, workspaceID uuid.UUID,
	provider, container string, deletedAt time.Time,
) error {
	return repository.RetireUnreferencedAttachments(ctx, tx, ids, workspaceID, provider, container, deletedAt)
}

// RetireUnreferencedAttachments preserves uploads with any surviving consumer.
// Metadata retirement and retryable storage deletion are committed together.
func (repository *Repository) RetireUnreferencedAttachments(
	ctx context.Context, tx pgx.Tx, ids []uuid.UUID, workspaceID uuid.UUID,
	provider, container string, deletedAt time.Time,
) error {
	if tx == nil || workspaceID == uuid.Nil || strings.TrimSpace(provider) == "" ||
		strings.TrimSpace(container) == "" || deletedAt.IsZero() {
		return errors.New("attachment retirement storage route and transaction are required")
	}
	queries := attachmentssql.New(tx)
	const batchSize = 500
	for start := 0; start < len(ids); start += batchSize {
		end := min(start+batchSize, len(ids))
		_, err := queries.RetireUnreferencedAttachments(ctx, attachmentssql.RetireUnreferencedAttachmentsParams{
			AttachmentIds: ids[start:end], WorkspaceID: workspaceID,
			StorageProvider: provider, ContainerName: container, DeletedAt: deletedAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("retire attachments and enqueue object deletion: %w", err)
		}
	}
	return nil
}
