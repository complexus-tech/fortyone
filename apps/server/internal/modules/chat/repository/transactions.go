package chatrepository

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

const (
	chatTransactionAttempts = 8
	chatRetryInitialDelay   = 10 * time.Millisecond
	chatRetryMaximumDelay   = 160 * time.Millisecond
)

func (r *Repository) WithinUnit(ctx context.Context, operation func(context.Context) error) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if _, ok := platformdatabase.UnitTransaction(ctx); ok {
		return operation(ctx)
	}
	if r.pool == nil {
		return errors.New("chat transaction pool is unavailable")
	}
	return withinRetryableChatUnit(ctx, r.pool, operation)
}

// A serialization failure aborts the whole transaction, including its receipt
// and outbox. Start a fresh Serializable unit for every retry, retaining the
// caller's creation ID. Do not replay an ambiguous COMMIT/transport failure.
func withinRetryableChatUnit(ctx context.Context, beginner platformdatabase.Beginner, operation func(context.Context) error) error {
	var err error
	for attempt := range chatTransactionAttempts {
		if err := ctx.Err(); err != nil {
			return err
		}
		err = platformdatabase.WithinUnitOfWork(ctx, beginner, func(unit context.Context, _ pgx.Tx) error {
			return operation(unit)
		})
		if !platformdatabase.IsRetryableTransactionError(err) || attempt == chatTransactionAttempts-1 {
			return err
		}
		if err := waitForChatRetry(ctx, chatRetryDelay(attempt)); err != nil {
			return err
		}
	}
	return err
}

func chatRetryDelay(attempt int) time.Duration {
	ceiling := min(chatRetryInitialDelay<<min(attempt, 4), chatRetryMaximumDelay)
	// Equal jitter avoids immediate synchronized retries while keeping all
	// seven backoffs together at or below 630ms; request deadlines still take priority.
	return ceiling/2 + time.Duration(rand.Int64N(int64(ceiling/2)+1))
}

func waitForChatRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
