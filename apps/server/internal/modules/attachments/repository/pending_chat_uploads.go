package attachmentsrepository

import (
	"context"
	"errors"
	"fmt"
	"slices"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	attachmentssql "github.com/complexus-tech/projects-api/internal/modules/attachments/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) RetirePendingChatUploads(ctx context.Context, batch attachmentdomain.PendingChatUploadRetention) (int, error) {
	if err := r.configured(); err != nil {
		return 0, err
	}
	if batch.Limit < 1 || batch.Limit > 500 || batch.CreatedBefore.IsZero() || !batch.CreatedBefore.Before(batch.RetiredAt) {
		return 0, errors.New("pending chat upload retention requires a valid cutoff and bounded batch")
	}
	count := 0
	err := r.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		rows, err := attachmentssql.New(tx).ExpirePendingChatUploads(ctx, attachmentssql.ExpirePendingChatUploadsParams{CreatedBefore: batch.CreatedBefore.UTC(), BatchLimit: int32(batch.Limit)})
		if err != nil {
			return fmt.Errorf("expire pending chat associations: %w", err)
		}
		byWorkspace := make(map[uuid.UUID][]uuid.UUID)
		workspaces := make([]uuid.UUID, 0)
		for _, row := range rows {
			if _, exists := byWorkspace[row.WorkspaceID]; !exists {
				workspaces = append(workspaces, row.WorkspaceID)
			}
			byWorkspace[row.WorkspaceID] = append(byWorkspace[row.WorkspaceID], row.AttachmentID)
		}
		slices.SortFunc(workspaces, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
		for _, workspaceID := range workspaces {
			if err := r.RetireUnreferencedAttachments(ctx, tx, byWorkspace[workspaceID], workspaceID, batch.StorageProvider, batch.ContainerName, batch.RetiredAt); err != nil {
				return err
			}
		}
		count = len(rows)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
