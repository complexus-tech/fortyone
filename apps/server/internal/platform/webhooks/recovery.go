package webhooks

import (
	"context"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/google/uuid"
)

type RecoveryPolicy struct {
	ClaimLimit        int32
	MaxAttempts       int32
	PendingAge        time.Duration
	FailedAge         time.Duration
	ProcessingLease   time.Duration
	RecoveryBaseDelay time.Duration
	RecoveryMaxShift  int32
}

func DefaultRecoveryPolicy() RecoveryPolicy {
	return RecoveryPolicy{
		ClaimLimit:        100,
		MaxAttempts:       20,
		PendingAge:        30 * time.Second,
		FailedAge:         5 * time.Minute,
		ProcessingLease:   10 * time.Minute,
		RecoveryBaseDelay: 30 * time.Second,
		RecoveryMaxShift:  8,
	}
}

func (policy RecoveryPolicy) Validate() error {
	if policy.ClaimLimit < 1 || policy.ClaimLimit > 500 || policy.MaxAttempts < 1 || policy.MaxAttempts > 100 {
		return ErrInvalidRequest
	}
	if !wholeSecondsWithin(policy.PendingAge, 24*time.Hour) ||
		!wholeSecondsWithin(policy.FailedAge, 30*24*time.Hour) ||
		!wholeSecondsWithin(policy.ProcessingLease, 24*time.Hour) ||
		!wholeSecondsWithin(policy.RecoveryBaseDelay, 24*time.Hour) {
		return ErrInvalidRequest
	}
	if policy.RecoveryMaxShift < 0 || policy.RecoveryMaxShift > 20 {
		return ErrInvalidRequest
	}
	return nil
}

func wholeSecondsWithin(value, maximum time.Duration) bool {
	return value >= time.Second && value <= maximum && value%time.Second == 0
}

type RecoveryReport struct {
	Claimed    int
	Dispatched int
	Released   int
}

func (gateway *Gateway) Recover(
	ctx context.Context,
	provider integrations.ProviderKey,
	policy RecoveryPolicy,
) (RecoveryReport, error) {
	if gateway == nil || gateway.inbox == nil {
		return RecoveryReport{}, ErrNotConfigured
	}
	if err := policy.Validate(); err != nil {
		return RecoveryReport{}, err
	}
	runtime, err := gateway.runtimes.require(provider)
	if err != nil {
		return RecoveryReport{}, err
	}
	now := gateway.config.Now().UTC()
	records, err := gateway.inbox.ClaimRecoverable(ctx, provider, policy, now)
	if err != nil {
		return RecoveryReport{}, fmt.Errorf("claim recoverable webhook deliveries: %w", err)
	}
	report := RecoveryReport{Claimed: len(records)}
	var dispatchFailed bool
	for _, record := range records {
		if err := runtime.Dispatcher.Enqueue(ctx, Task{InboxID: record.ID, Provider: provider}); err == nil {
			if markErr := gateway.inbox.MarkQueued(ctx, record.ID, now); markErr != nil {
				return report, fmt.Errorf("record recovered webhook queue handoff: %w", markErr)
			}
			report.Dispatched++
			continue
		}
		dispatchFailed = true
		if releaseErr := gateway.inbox.ReleaseRecovery(ctx, record.ID, record.RecoveryGeneration, now); releaseErr != nil {
			return report, fmt.Errorf("release webhook recovery claim: %w", releaseErr)
		}
		report.Released++
	}
	if dispatchFailed {
		return report, ErrDispatchUnavailable
	}
	return report, nil
}

func (gateway *Gateway) ExpirePayloads(ctx context.Context, limit int32) ([]uuid.UUID, error) {
	if gateway == nil || gateway.inbox == nil {
		return nil, ErrNotConfigured
	}
	if limit < 1 || limit > 1000 {
		return nil, ErrInvalidRequest
	}
	return gateway.inbox.ExpirePayloads(ctx, gateway.config.Now().UTC(), limit)
}
