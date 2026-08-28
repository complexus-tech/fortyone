package usersrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
)

var errUserMaintenanceUnavailable = errors.New("user maintenance repository is not configured")

// PurgeExpiredVerificationTokens permanently removes one bounded batch of
// tokens whose expiry is older than the application-provided retention cutoff.
func (r *repo) PurgeExpiredVerificationTokens(
	ctx context.Context,
	retainedBefore time.Time,
	batchSize int,
) (int64, error) {
	if r == nil || r.queries == nil {
		return 0, errUserMaintenanceUnavailable
	}
	if retainedBefore.IsZero() {
		return 0, errors.New("verification token retention cutoff is required")
	}
	if batchSize <= 0 {
		return 0, errors.New("verification token purge batch size must be positive")
	}
	databaseBatchSize, err := safecast.Int32(batchSize)
	if err != nil {
		return 0, fmt.Errorf("validate verification token purge batch size: %w", err)
	}

	deleted, err := r.queries.PurgeExpiredVerificationTokens(ctx, usersql.PurgeExpiredVerificationTokensParams{
		RetainedBefore: retainedBefore.UTC(),
		BatchSize:      databaseBatchSize,
	})
	if err != nil {
		return 0, fmt.Errorf("purge expired verification tokens: %w", err)
	}
	return deleted, nil
}

// DeactivateInactiveUsers deactivates one bounded batch only when both the
// inactivity threshold and the independently measured warning grace period
// have elapsed.
func (r *repo) DeactivateInactiveUsers(
	ctx context.Context,
	inactiveBefore time.Time,
	warningSentBefore time.Time,
	deactivatedAt time.Time,
	batchSize int,
) (int64, error) {
	if r == nil || r.queries == nil {
		return 0, errUserMaintenanceUnavailable
	}
	if inactiveBefore.IsZero() || warningSentBefore.IsZero() || deactivatedAt.IsZero() {
		return 0, errors.New("user deactivation cutoffs and mutation time are required")
	}
	if !inactiveBefore.Before(deactivatedAt) || !warningSentBefore.Before(deactivatedAt) {
		return 0, errors.New("user deactivation cutoffs must precede the mutation time")
	}
	if batchSize <= 0 {
		return 0, errors.New("user deactivation batch size must be positive")
	}
	databaseBatchSize, err := safecast.Int32(batchSize)
	if err != nil {
		return 0, fmt.Errorf("validate user deactivation batch size: %w", err)
	}

	inactiveBefore = inactiveBefore.UTC()
	deactivated, err := r.queries.DeactivateInactiveUsers(ctx, usersql.DeactivateInactiveUsersParams{
		DeactivatedAt:     deactivatedAt.UTC(),
		InactiveBefore:    &inactiveBefore,
		WarningSentBefore: warningSentBefore.UTC(),
		BatchSize:         databaseBatchSize,
	})
	if err != nil {
		return 0, fmt.Errorf("deactivate inactive users: %w", err)
	}
	return deactivated, nil
}
