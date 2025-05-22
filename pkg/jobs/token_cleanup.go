package jobs

import (
	"context"
)

// purgeExpiredTokens permanently deletes verification tokens older than 7 days
func (s *Scheduler) purgeExpiredTokens(ctx context.Context) error {
	return PurgeExpiredTokens(ctx, s.db, s.log)
}
