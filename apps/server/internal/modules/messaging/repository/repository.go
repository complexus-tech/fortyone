package messagingrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = pgx.ErrNoRows

var ErrLeaseBusy = errors.New("messaging lease is busy")

const (
	messagingLeaseDuration    = 2 * time.Minute
	messagingLeaseRetryMargin = 5 * time.Second
	inboundRecoveryBaseDelay  = 10 * time.Minute
	inboundRecoveryMaxShift   = 5
)

// LeaseBusyError indicates that another worker still owns an active processing
// lease. Callers must retry rather than treating the existing row as completed.
type LeaseBusyError struct {
	Resource   string
	RetryAfter time.Duration
}

func (e *LeaseBusyError) Error() string {
	return fmt.Sprintf("%s: %v; retry after %s", e.Resource, ErrLeaseBusy, e.RetryAfter)
}

func (e *LeaseBusyError) Unwrap() error {
	return ErrLeaseBusy
}

// LeaseRetryAfter extracts the safe retry delay from a busy messaging lease.
func LeaseRetryAfter(err error) (time.Duration, bool) {
	var busy *LeaseBusyError
	if !errors.As(err, &busy) {
		return 0, false
	}
	if busy.RetryAfter <= 0 {
		return messagingLeaseDuration + messagingLeaseRetryMargin, true
	}
	return busy.RetryAfter, true
}

func newLeaseBusyError(resource string) *LeaseBusyError {
	return &LeaseBusyError{
		Resource:   resource,
		RetryAfter: messagingLeaseDuration + messagingLeaseRetryMargin,
	}
}

// Repository is the module-owned persistence boundary. Generated SQL stays
// internal to this package; callers receive stable domain records instead of
// database row types.
type Repository struct {
	pool           *pgxpool.Pool
	queries        messagingsql.Querier
	runTransaction func(context.Context, func(messagingsql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	queries := messagingsql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.pool = pool
	repository.runTransaction = func(ctx context.Context, operation func(messagingsql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries messagingsql.Querier) *Repository {
	repository := &Repository{queries: queries}
	if queries != nil {
		repository.runTransaction = func(ctx context.Context, operation func(messagingsql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (repository *Repository) configured() bool {
	return repository != nil && repository.queries != nil
}

func (repository *Repository) withinTransaction(
	ctx context.Context,
	operation func(messagingsql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if repository == nil || repository.queries == nil || repository.runTransaction == nil {
		return errors.New("messaging repository transactions are unavailable")
	}
	return repository.runTransaction(ctx, operation)
}

func requireAffectedRows(affected int64, operation string) error {
	if affected == 0 {
		return fmt.Errorf("%s: %w", operation, ErrNotFound)
	}
	return nil
}
