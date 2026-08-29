package webhooksrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	webhooksql "github.com/complexus-tech/projects-api/internal/platform/webhooks/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) ClaimRecoverable(
	ctx context.Context,
	provider integrations.ProviderKey,
	policy webhooks.RecoveryPolicy,
	now time.Time,
) ([]webhooks.Record, error) {
	if repository == nil || repository.queries == nil {
		return nil, webhooks.ErrNotConfigured
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ClaimRecoverableDeliveries(ctx, webhooksql.ClaimRecoverableDeliveriesParams{
		Now:                 now.UTC(),
		Provider:            string(provider),
		MaxAttempts:         policy.MaxAttempts,
		PendingAgeSeconds:   int64(policy.PendingAge / time.Second),
		FailedAgeSeconds:    int64(policy.FailedAge / time.Second),
		LeaseSeconds:        int64(policy.ProcessingLease / time.Second),
		RecoveryBaseSeconds: int64(policy.RecoveryBaseDelay / time.Second),
		RecoveryMaxShift:    policy.RecoveryMaxShift,
		ClaimLimit:          policy.ClaimLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("claim recoverable webhook deliveries: %w", err)
	}
	records := make([]webhooks.Record, 0, len(rows))
	for _, row := range rows {
		record, err := toRecord(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (repository *Repository) ReleaseRecovery(
	ctx context.Context,
	id uuid.UUID,
	generation int32,
	releasedAt time.Time,
) error {
	if repository == nil || repository.queries == nil {
		return webhooks.ErrNotConfigured
	}
	rows, err := repository.queries.ReleaseDeliveryRecovery(ctx, webhooksql.ReleaseDeliveryRecoveryParams{
		ReleasedAt:         releasedAt.UTC(),
		ID:                 id,
		RecoveryGeneration: generation,
	})
	return requireAffected("release webhook delivery recovery", rows, err)
}

func (repository *Repository) Start(
	ctx context.Context,
	id uuid.UUID,
	now time.Time,
	lease time.Duration,
) (webhooks.Record, bool, error) {
	if repository == nil || repository.queries == nil {
		return webhooks.Record{}, false, webhooks.ErrNotConfigured
	}
	if lease < time.Second || lease > 24*time.Hour || lease%time.Second != 0 {
		return webhooks.Record{}, false, webhooks.ErrInvalidRequest
	}
	row, err := repository.queries.TryStartDelivery(ctx, webhooksql.TryStartDeliveryParams{
		Now:          now.UTC(),
		ID:           id,
		LeaseSeconds: int64(lease / time.Second),
	})
	if err == nil {
		record, mapErr := toRecord(row)
		return record, true, mapErr
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return webhooks.Record{}, false, fmt.Errorf("start webhook delivery: %w", err)
	}
	record, err := repository.GetByID(ctx, id)
	if err != nil {
		return webhooks.Record{}, false, err
	}
	switch {
	case record.Status.Terminal():
		return record, false, nil
	case record.Status == webhooks.StatusProcessing:
		return record, false, webhooks.ErrLeaseBusy
	default:
		return record, false, webhooks.ErrInvalidState
	}
}

func (repository *Repository) Complete(
	ctx context.Context,
	id uuid.UUID,
	status webhooks.Status,
	outcomeCode string,
	completedAt time.Time,
) error {
	if repository == nil || repository.queries == nil {
		return webhooks.ErrNotConfigured
	}
	if err := webhooks.ValidateCompletion(status, outcomeCode); err != nil {
		return err
	}
	rows, err := repository.queries.CompleteDelivery(ctx, webhooksql.CompleteDeliveryParams{
		Status:      string(status),
		SafeMessage: outcomeCode,
		CompletedAt: completedAt.UTC(),
		ID:          id,
	})
	return requireAffected("complete webhook delivery", rows, err)
}

func (repository *Repository) ExpirePayloads(
	ctx context.Context,
	now time.Time,
	limit int32,
) ([]uuid.UUID, error) {
	if repository == nil || repository.queries == nil {
		return nil, webhooks.ErrNotConfigured
	}
	if limit < 1 || limit > 1000 {
		return nil, webhooks.ErrInvalidRequest
	}
	ids, err := repository.queries.ExpireDeliveryPayloads(ctx, webhooksql.ExpireDeliveryPayloadsParams{
		Now:         now.UTC(),
		ExpiryLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("expire webhook delivery payloads: %w", err)
	}
	return ids, nil
}
