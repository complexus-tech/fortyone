package jobs

import (
	"context"
)

// purgeDeletedStories permanently deletes stories that have been marked as deleted for 30+ days
func (s *Scheduler) purgeDeletedStories(ctx context.Context) error {
	return PurgeDeletedStories(ctx, s.db, s.log)
}
