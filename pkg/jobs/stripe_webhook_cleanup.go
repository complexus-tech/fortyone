package jobs

import (
	"context"
)

// purgeOldStripeWebhookEvents permanently deletes stripe webhook events older than 30 days
func (s *Scheduler) purgeOldStripeWebhookEvents(ctx context.Context) error {
	return PurgeOldStripeWebhookEvents(ctx, s.db, s.log)
}
