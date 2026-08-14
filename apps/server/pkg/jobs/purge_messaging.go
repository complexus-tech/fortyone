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

// PurgeMessagingData removes expired nonces, redacts expired mutation
// proposals, and deletes provider message data after its operational retention
// window. Durable Maya email threads and messages are preserved so later
// replies retain the complete conversation; only their expired or revoked
// reply-address aliases are removed here.
func PurgeMessagingData(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeMessagingData")
	defer span.End()

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("begin messaging data cleanup: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	queries := []struct {
		name  string
		query string
	}{
		{
			name: "expired_nonces",
			query: `
				DELETE FROM messaging_nonces
				WHERE expires_at < NOW() - INTERVAL '1 day'
			`,
		},
		{
			name: "expired_story_mutation_confirmations",
			query: `
				UPDATE messaging_story_mutation_confirmations
				SET status = 'expired',
				    proposal = NULL,
				    applied_at = NULL,
				    expired_at = NOW(),
				    updated_at = NOW()
				WHERE expires_at <= NOW()
				  AND (
				      status = 'pending'
				      OR (
				          status = 'applied'
				          AND operation = 'create_stories'
				          AND proposal IS NOT NULL
				      )
				  )
			`,
		},
		{
			name: "old_deliveries",
			query: `
				DELETE FROM messaging_outbound_deliveries
				WHERE created_at < NOW() - INTERVAL '30 days'
			`,
		},
		{
			name: "old_inbound_events",
			query: `
				DELETE FROM messaging_inbound_events
				WHERE received_at < NOW() - INTERVAL '30 days'
			`,
		},
		{
			name: "completed_slack_uninstalls",
			query: `
				DELETE FROM slack_uninstall_outbox
				WHERE status = 'completed'
				  AND completed_at < NOW() - INTERVAL '30 days'
			`,
		},
		{
			name: "old_messages",
			query: `
				DELETE FROM messaging_messages
				WHERE created_at < NOW() - INTERVAL '30 days'
			`,
		},
		{
			name: "expired_email_reply_tokens",
			query: `
				DELETE FROM messaging_email_reply_tokens
				WHERE expires_at < NOW() - INTERVAL '1 day'
				   OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '1 day')
			`,
		},
		{
			name: "empty_conversations",
			query: `
				DELETE FROM messaging_conversations mc
				WHERE mc.updated_at < NOW() - INTERVAL '30 days'
				  AND NOT EXISTS (
					SELECT 1
					FROM messaging_messages mm
					WHERE mm.conversation_id = mc.id
				  )
			`,
		},
	}

	var total int64
	for _, operation := range queries {
		result, execErr := tx.ExecContext(ctx, operation.query)
		if execErr != nil {
			span.RecordError(execErr)
			return fmt.Errorf("delete %s: %w", operation.name, execErr)
		}
		rows, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			span.RecordError(rowsErr)
			return fmt.Errorf("count deleted %s: %w", operation.name, rowsErr)
		}
		total += rows
		span.AddEvent("messaging_data_deleted", trace.WithAttributes(
			attribute.String("kind", operation.name),
			attribute.Int64("rows_affected", rows),
		))
	}

	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("commit messaging data cleanup: %w", err)
	}
	log.Info(ctx, "Purged expired messaging data", "rows_affected", total)
	return nil
}
