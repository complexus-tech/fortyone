package jobs

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// purgeExpiredTokens permanently deletes verification tokens older than 7 days
func (s *Scheduler) purgeExpiredTokens(ctx context.Context) error {
	ctx, span := web.AddSpan(ctx, "jobs.purgeExpiredTokens")
	defer span.End()

	s.log.Info(ctx, "Purging verification tokens older than 7 days")

	deleteQuery := `
		DELETE FROM verification_tokens
		WHERE created_at < NOW() - INTERVAL '7 days'
	`

	result, err := s.db.ExecContext(ctx, deleteQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("tokens_deleted", trace.WithAttributes(
		attribute.Int64("rows_affected", rowsAffected),
	))
	s.log.Info(ctx, fmt.Sprintf("Permanently deleted %d expired verification tokens", rowsAffected))
	return nil
}
