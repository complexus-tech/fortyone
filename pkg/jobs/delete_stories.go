package jobs

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/web"
)

// purgeDeletedStories permanently deletes stories that have been marked as deleted for 30+ days
func (s *Scheduler) purgeDeletedStories(ctx context.Context) error {
	ctx, span := web.AddSpan(ctx, "jobs.purgeDeletedStories")
	defer span.End()

	s.log.Info(ctx, "Purging stories deleted for more than 30 days")
	countQuery := `
		SELECT COUNT(id)
		FROM stories
		WHERE deleted_at IS NOT NULL
		AND deleted_at < NOW() - INTERVAL '30 days'
	`

	var count int
	if err := s.db.GetContext(ctx, &count, countQuery); err != nil {
		return fmt.Errorf("failed to count stories to delete: %w", err)
	}

	s.log.Info(ctx, fmt.Sprintf("Found %d stories to be permanently deleted", count))

	if count == 0 {
		return nil
	}

	// Delete stories
	deleteQuery := `
		DELETE FROM stories
		WHERE deleted_at IS NOT NULL
		AND deleted_at < NOW() - INTERVAL '30 days'
	`

	result, err := s.db.ExecContext(ctx, deleteQuery)
	if err != nil {
		return fmt.Errorf("failed to delete stories: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	s.log.Info(ctx, fmt.Sprintf("Permanently deleted %d stories", rowsAffected))
	return nil
}
