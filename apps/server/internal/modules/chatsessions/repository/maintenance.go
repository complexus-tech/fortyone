package chatsessionsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

var errChatSessionMaintenanceUnavailable = errors.New("chat session maintenance repository is not configured")

// PurgeDeletedChatSessions permanently removes one bounded batch of expired
// soft-deleted sessions. The generated query retains sessions that still own
// unresolved mutation approvals and skips rows held by another worker.
func (r *repo) PurgeDeletedChatSessions(
	ctx context.Context,
	deletedBefore time.Time,
	batchSize int,
) (int64, error) {
	if r == nil || r.queries == nil {
		return 0, errChatSessionMaintenanceUnavailable
	}
	if deletedBefore.IsZero() {
		return 0, errors.New("chat session deletion cutoff is required")
	}
	if batchSize <= 0 {
		return 0, errors.New("chat session purge batch size must be positive")
	}
	databaseBatchSize, err := safecast.Int32(batchSize)
	if err != nil {
		return 0, fmt.Errorf("validate chat session purge batch size: %w", err)
	}

	deletedBefore = deletedBefore.UTC()
	deleted, err := r.queries.PurgeDeletedChatSessions(ctx, chatsessionssql.PurgeDeletedChatSessionsParams{
		DeletedBefore: &deletedBefore,
		BatchSize:     databaseBatchSize,
	})
	if err != nil {
		return 0, fmt.Errorf("purge deleted chat sessions: %w", err)
	}
	return deleted, nil
}
