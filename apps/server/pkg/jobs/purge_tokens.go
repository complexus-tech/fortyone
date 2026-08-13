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

// PurgeExpiredTokens permanently deletes verification tokens older than 7 days
func PurgeExpiredTokens(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeExpiredTokens")
	defer span.End()

	log.Info(ctx, "Purging verification tokens older than 7 days")

	queries := []string{
		`DELETE FROM verification_tokens WHERE created_at < NOW() - INTERVAL '7 days'`,
		`DELETE FROM feedback_contributor_verifications WHERE expires_at < NOW() - INTERVAL '7 days'`,
		`DELETE FROM feedback_contributor_sessions
		 WHERE expires_at < NOW() - INTERVAL '7 days'
		    OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '7 days')`,
		`DELETE FROM feedback_contributor_unsubscribe_tokens
		 WHERE expires_at < NOW() - INTERVAL '7 days'
		    OR (consumed_at IS NOT NULL AND consumed_at < NOW() - INTERVAL '7 days')`,
		`DELETE FROM feedback_widget_assertion_nonces WHERE expires_at < NOW()`,
		`DELETE FROM feedback_widget_signing_secret_rotations
		 WHERE grace_expires_at < NOW() - INTERVAL '7 days'`,
	}

	var rowsAffected int64
	for _, query := range queries {
		result, err := db.ExecContext(ctx, query)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to delete expired tokens: %w", err)
		}
		deleted, err := result.RowsAffected()
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to get rows affected: %w", err)
		}
		rowsAffected += deleted
	}

	span.AddEvent("tokens_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", rowsAffected),
	))
	log.Info(ctx, fmt.Sprintf("Permanently deleted %d expired verification tokens", rowsAffected))
	return nil
}
