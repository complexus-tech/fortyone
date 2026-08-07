package calendarrepository

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{log: log, db: db}
}

// lockCalendarUser serializes calendar mutations for one workspace member.
// This closes races between manual blocks, provider snapshots, and reconnects.
func lockCalendarUser(ctx context.Context, tx *sqlx.Tx, workspaceID, userID uuid.UUID) error {
	const query = `
		SELECT pg_advisory_xact_lock(
			hashtextextended(
				CONCAT(CAST($1 AS text), ':', CAST($2 AS text)),
				0
			)
		)
	`
	if _, err := tx.ExecContext(ctx, query, workspaceID, userID); err != nil {
		return fmt.Errorf("lock calendar user: %w", err)
	}
	return nil
}
