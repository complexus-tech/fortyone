package jobs

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const purgeDeletedChatSessionsQuery = `
	DELETE FROM chat_sessions AS session
	WHERE session.deleted_at IS NOT NULL
		AND session.deleted_at < NOW() - INTERVAL '30 days'
		AND NOT EXISTS (
			SELECT 1
			FROM chat_mutation_approval_executions AS execution
			WHERE execution.session_id = session.id
				AND execution.status IN ('ready', 'retry_ready', 'executing', 'failed_uncertain')
		)
`

// PurgeDeletedChatSessions permanently deletes chat sessions marked as deleted
// for 30+ days while retaining any session that still anchors an unresolved
// mutation approval quarantine.
func PurgeDeletedChatSessions(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedChatSessions")
	defer span.End()
	log.Info(ctx, "Purging chat sessions deleted for more than 30 days")

	result, err := db.ExecContext(ctx, purgeDeletedChatSessionsQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete chat sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("chat_sessions_deleted", trace.WithAttributes(
		attribute.Int("rows_affected", int(rowsAffected)),
	))
	log.Info(ctx, fmt.Sprintf("Permanently deleted %d chat sessions", rowsAffected))
	return nil
}
