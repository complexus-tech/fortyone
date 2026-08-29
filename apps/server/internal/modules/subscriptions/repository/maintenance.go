package subscriptionsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	subscriptionssql "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

// PurgeTerminalStripeWebhookEvents permanently removes one bounded batch of
// old processed or failed receipts. Processing leases are excluded by the
// generated query's lifecycle fence.
func (repository *Repository) PurgeTerminalStripeWebhookEvents(
	ctx context.Context,
	terminalBefore time.Time,
	batchSize int,
) (int64, error) {
	if err := repository.configured(); err != nil {
		return 0, err
	}
	if terminalBefore.IsZero() {
		return 0, errors.New("stripe webhook terminal retention cutoff is required")
	}
	if batchSize <= 0 {
		return 0, errors.New("stripe webhook purge batch size must be positive")
	}
	databaseBatchSize, err := safecast.Int32(batchSize)
	if err != nil {
		return 0, fmt.Errorf("validate Stripe webhook purge batch size: %w", err)
	}

	terminalBefore = terminalBefore.UTC()
	deleted, err := repository.queries.PurgeTerminalStripeWebhookEvents(
		ctx,
		subscriptionssql.PurgeTerminalStripeWebhookEventsParams{
			TerminalBefore: &terminalBefore,
			BatchSize:      databaseBatchSize,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("purge terminal Stripe webhook events: %w", err)
	}
	return deleted, nil
}
