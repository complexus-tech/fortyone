package jobs

import (
	"context"
	"fmt"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PurgeDeletedFeedback permanently deletes feedback after its 30-day recovery window.
func PurgeDeletedFeedback(ctx context.Context, store feedback.MaintenanceStore, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedFeedback")
	defer span.End()
	log.Info(ctx, "Purging feedback deleted for more than 30 days")

	if store == nil {
		err := fmt.Errorf("feedback maintenance store is unavailable")
		span.RecordError(err)
		return err
	}
	result, err := store.PurgeDeletedFeedback(ctx, time.Now().UTC().Add(-30*24*time.Hour))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("purge deleted feedback: %w", err)
	}

	span.AddEvent("feedback_deleted", trace.WithAttributes(
		attribute.Int64("items_deleted", result.ItemsDeleted),
		attribute.Int64("contributors_deleted", result.ContributorsDeleted),
	))
	log.Info(ctx, "Permanently deleted feedback", "items", result.ItemsDeleted, "anonymous_contributors", result.ContributorsDeleted)
	return nil
}
