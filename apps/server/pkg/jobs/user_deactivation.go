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

// ProcessUserDeactivation deactivates users that received inactivity warnings 30+ days ago and are still inactive
func ProcessUserDeactivation(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.ProcessUserDeactivation")
	defer span.End()

	log.Info(ctx, "Deactivating users inactive for 8+ months with 30-day grace period")

	deactivateQuery := `
		UPDATE users 
		SET is_active = false
		WHERE last_login_at < NOW() - INTERVAL '8 months 30 days'
		AND is_active = true
		AND u.is_system = false
		AND inactivity_warning_sent_at IS NOT NULL
	`
	result, err := db.ExecContext(ctx, deactivateQuery)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to deactivate users: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	span.AddEvent("users_deactivated", trace.WithAttributes(
		attribute.Int("rows_affected", int(rowsAffected)),
	))
	log.Info(ctx, fmt.Sprintf("Deactivated %d users", rowsAffected))
	return nil
}
