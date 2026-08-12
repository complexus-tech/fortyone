package jobs

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// PurgeDeletedFeedback permanently deletes feedback after its 30-day recovery window.
func PurgeDeletedFeedback(ctx context.Context, db *sqlx.DB, log *logger.Logger) error {
	ctx, span := web.AddSpan(ctx, "jobs.PurgeDeletedFeedback")
	defer span.End()
	log.Info(ctx, "Purging feedback deleted for more than 30 days")

	var result struct {
		ItemsDeleted        int `db:"items_deleted"`
		ContributorsDeleted int `db:"contributors_deleted"`
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("start feedback purge: %w", err)
	}
	defer tx.Rollback()

	var contributorIDs []uuid.UUID
	if err := tx.SelectContext(ctx, &contributorIDs, feedbackContributorsToPurgeQuery()); err != nil {
		span.RecordError(err)
		return fmt.Errorf("select feedback contributors to purge: %w", err)
	}
	contributorIDs = uniqueUUIDs(contributorIDs)

	err = tx.GetContext(ctx, &result, `
		WITH deleted_items AS (
			DELETE FROM feedback_items
			WHERE deleted_at IS NOT NULL
				AND deleted_at < NOW() - INTERVAL '30 days'
			RETURNING contributor_id
		)
		SELECT
			COUNT(*) AS items_deleted,
			0 AS contributors_deleted
		FROM deleted_items
	`)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete feedback: %w", err)
	}
	if len(contributorIDs) > 0 {
		if err := tx.GetContext(ctx, &result.ContributorsDeleted, `
			WITH deleted_contributors AS (
				DELETE FROM feedback_contributors contributor
				WHERE contributor.id = ANY(CAST($1 AS uuid[]))
					AND contributor.kind = 'anonymous'
					AND NOT EXISTS (
						SELECT 1
						FROM feedback_items retained
						WHERE retained.contributor_id = contributor.id
					)
				RETURNING contributor.id
			)
			SELECT COUNT(*) FROM deleted_contributors
		`, pq.Array(contributorIDs)); err != nil {
			span.RecordError(err)
			return fmt.Errorf("delete orphaned anonymous contributors: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("commit feedback purge: %w", err)
	}

	span.AddEvent("feedback_deleted", trace.WithAttributes(
		attribute.Int("items_deleted", result.ItemsDeleted),
		attribute.Int("contributors_deleted", result.ContributorsDeleted),
	))
	log.Info(ctx, "Permanently deleted feedback", "items", result.ItemsDeleted, "anonymous_contributors", result.ContributorsDeleted)
	return nil
}

func feedbackContributorsToPurgeQuery() string {
	return `
		SELECT contributor_id
		FROM feedback_items
		WHERE deleted_at IS NOT NULL
			AND deleted_at < NOW() - INTERVAL '30 days'
		FOR UPDATE
	`
}

func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	unique := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}
