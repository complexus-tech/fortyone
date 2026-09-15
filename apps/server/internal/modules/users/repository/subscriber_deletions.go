package usersrepository

import (
	"context"
	"errors"
	"strings"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/semaphore"
)

type SubscriberDeletionRepository struct {
	pool      *pgxpool.Pool
	queries   *usersql.Queries
	admission *semaphore.Weighted
}

type subscriberSessionKey struct{}
type subscriberSession struct {
	repository *SubscriberDeletionRepository
	queries    *usersql.Queries
}

func NewSubscriberDeletionRepository(pool *pgxpool.Pool) *SubscriberDeletionRepository {
	return &SubscriberDeletionRepository{pool: pool, queries: usersql.New(pool), admission: semaphore.NewWeighted(2)}
}

func (r *SubscriberDeletionRepository) forContext(ctx context.Context) *usersql.Queries {
	if session, ok := ctx.Value(subscriberSessionKey{}).(*subscriberSession); ok && session.repository == r {
		return session.queries
	}
	return r.queries
}

func (r *SubscriberDeletionRepository) Claim(ctx context.Context, now, expires time.Time) (usersdomain.SubscriberDeletion, bool, error) {
	token := uuid.New()
	row, err := r.forContext(ctx).ClaimAccountSubscriberDeletion(ctx, usersql.ClaimAccountSubscriberDeletionParams{ClaimedAt: now, LeaseExpiresAt: &expires, LeaseToken: &token})
	if errors.Is(err, pgx.ErrNoRows) {
		return usersdomain.SubscriberDeletion{}, false, nil
	}
	if err != nil {
		return usersdomain.SubscriberDeletion{}, false, err
	}
	return usersdomain.SubscriberDeletion{ID: row.ID, Email: row.Email, LeaseToken: token, AttemptCount: int(row.AttemptCount)}, true, nil
}

func (r *SubscriberDeletionRepository) Renew(ctx context.Context, id, token uuid.UUID, now, expires time.Time) (bool, error) {
	rows, err := r.forContext(ctx).RenewAccountSubscriberDeletion(ctx, usersql.RenewAccountSubscriberDeletionParams{ID: id, LeaseToken: &token, CheckedAt: &now, LeaseExpiresAt: &expires})
	return rows == 1, err
}

func (r *SubscriberDeletionRepository) Complete(ctx context.Context, id, token uuid.UUID) error {
	rows, err := r.forContext(ctx).CompleteAccountSubscriberDeletion(ctx, usersql.CompleteAccountSubscriberDeletionParams{ID: id, LeaseToken: &token})
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("subscriber cleanup claim is no longer current")
	}
	return nil
}

func (r *SubscriberDeletionRepository) Retry(ctx context.Context, id, token uuid.UUID, next, released time.Time) error {
	rows, err := r.forContext(ctx).RetryAccountSubscriberDeletion(ctx, usersql.RetryAccountSubscriberDeletionParams{ID: id, LeaseToken: &token, NextAttemptAt: next, ReleasedAt: released})
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("subscriber cleanup claim is no longer current")
	}
	return nil
}

func (r *SubscriberDeletionRepository) ActiveAccount(ctx context.Context, email string) (usersdomain.AccountSubscriber, bool, error) {
	row, err := r.forContext(ctx).GetActiveAccountSubscriber(ctx, usersql.GetActiveAccountSubscriberParams{Email: email})
	if errors.Is(err, pgx.ErrNoRows) {
		return usersdomain.AccountSubscriber{}, false, nil
	}
	if err != nil {
		return usersdomain.AccountSubscriber{}, false, err
	}
	return usersdomain.AccountSubscriber{UserID: row.UserID, Email: row.Email, FullName: row.FullName, CleanupPending: row.CleanupPending}, true, nil
}

// A reserved session (not a transaction) spans provider I/O. All reads reuse it,
// and uncertain lock acquisition/release closes it rather than pooling a lock.
func (r *SubscriberDeletionRepository) WithinSubscriberLifecycle(ctx context.Context, email string, operation func(context.Context) error) (resultErr error) {
	if strings.TrimSpace(email) == "" || operation == nil {
		return errors.New("subscriber lifecycle requires an address and operation")
	}
	if err := r.admission.Acquire(ctx, 1); err != nil {
		return err
	}
	defer r.admission.Release(1)
	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	queries := usersql.New(connection)
	locked := false
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if locked {
			released, err := queries.ReleaseAccountSubscriberLifecycle(cleanupCtx, usersql.ReleaseAccountSubscriberLifecycleParams{Email: email})
			if err == nil && released {
				connection.Release()
				return
			}
			resultErr = errors.Join(resultErr, errors.New("subscriber lifecycle lock release failed"))
		}
		resultErr = errors.Join(resultErr, connection.Hijack().Close(cleanupCtx))
	}()
	if err := queries.AcquireAccountSubscriberLifecycle(ctx, usersql.AcquireAccountSubscriberLifecycleParams{Email: email}); err != nil {
		return err
	}
	locked = true
	return operation(context.WithValue(ctx, subscriberSessionKey{}, &subscriberSession{repository: r, queries: queries}))
}
