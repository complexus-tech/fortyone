// Package googledriverepository implements Google Drive persistence with
// module-local sqlc queries and native pgx transactions.
package googledriverepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/semaphore"
)

var errTransactionsUnavailable = errors.New("Google Drive repository transactions are unavailable")

const maxConcurrentProviderLifecycleSessions int32 = 4

type Repository struct {
	queries         googledrivesql.Querier
	pool            *pgxpool.Pool
	runTransaction  func(context.Context, func(googledrivesql.Querier) error) error
	runProviderLock func(context.Context, string, func(context.Context) error) error
	now             func() time.Time
}

type providerLifecycleSessionContextKey struct{}

// providerLifecycleSession owns one reserved PostgreSQL session for every
// provider advisory lock in a nested lifecycle operation. Reusing the outer
// session is important: reserving one pool connection per nested lock can
// exhaust the pool before the operation can open its short database
// transaction.
type providerLifecycleSession struct {
	pool       *pgxpool.Pool
	connection *pgxpool.Conn
	queries    *googledrivesql.Queries
	tainted    bool
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{now: time.Now}
	}
	queries := googledrivesql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	return &Repository{
		queries: queries,
		pool:    pool,
		now:     time.Now,
		runTransaction: func(ctx context.Context, operation func(googledrivesql.Querier) error) error {
			return transactor.WithinTransaction(ctx, pgx.TxOptions{
				IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
			}, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
		runProviderLock: providerLifecycleLockRunner(pool),
	}
}

func (repository *Repository) withinTransaction(ctx context.Context, operation func(googledrivesql.Querier) error) error {
	if repository == nil || repository.queries == nil || operation == nil {
		return errTransactionsUnavailable
	}
	if session := repository.providerLifecycleSession(ctx); session != nil {
		transactor := platformdatabase.NewTransactor(session.connection)
		return mapDatabaseError(transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(session.queries.WithTx(tx))
		}))
	}
	if repository.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return mapDatabaseError(repository.runTransaction(ctx, operation))
}

func (repository *Repository) queriesForContext(ctx context.Context) googledrivesql.Querier {
	if session := repository.providerLifecycleSession(ctx); session != nil {
		return session.queries
	}
	return repository.queries
}

func (repository *Repository) providerLifecycleSession(ctx context.Context) *providerLifecycleSession {
	if repository == nil || repository.pool == nil {
		return nil
	}
	return providerLifecycleSessionFromContext(ctx, repository.pool)
}

func (repository *Repository) currentTime() time.Time {
	if repository == nil || repository.now == nil {
		return time.Now().UTC()
	}
	return repository.now().UTC()
}

func providerLifecycleLockRunner(pool *pgxpool.Pool) func(context.Context, string, func(context.Context) error) (resultErr error) {
	// Provider lifecycle operations hold one PostgreSQL session across network
	// I/O so their session-scoped advisory locks remain valid. Bound those
	// sessions independently of pgxpool and leave connection headroom for
	// ordinary API and worker traffic. A one-connection pool cannot preserve
	// headroom, but still admits one operation so revocation cleanup can run.
	admission := semaphore.NewWeighted(int64(providerLifecycleAdmissionLimit(pool.Config().MaxConns)))

	return func(ctx context.Context, lockKey string, operation func(context.Context) error) (resultErr error) {
		if nested := providerLifecycleSessionFromContext(ctx, pool); nested != nil {
			return nested.withLock(ctx, lockKey, operation)
		}

		if err := admission.Acquire(ctx, 1); err != nil {
			return fmt.Errorf("wait for Google Drive provider lifecycle capacity: %w", err)
		}
		defer admission.Release(1)

		connection, err := pool.Acquire(ctx)
		if err != nil {
			return fmt.Errorf("reserve Google Drive provider lifecycle lock connection: %w", err)
		}
		session := &providerLifecycleSession{
			pool: pool, connection: connection, queries: googledrivesql.New(connection),
		}
		locked := false

		defer func() {
			if locked && !session.tainted {
				resultErr = errors.Join(resultErr, session.release(ctx, lockKey))
			}
			if !session.tainted {
				connection.Release()
				return
			}

			// Never return a session with a possibly-held advisory lock to the
			// pool. Closing a hijacked connection makes PostgreSQL release it.
			closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			rawConnection := connection.Hijack()
			resultErr = errors.Join(resultErr, rawConnection.Close(closeCtx))
		}()

		if err := session.acquire(ctx, lockKey); err != nil {
			return err
		}
		locked = true
		return operation(context.WithValue(ctx, providerLifecycleSessionContextKey{}, session))
	}
}

func providerLifecycleAdmissionLimit(maxConnections int32) int32 {
	if maxConnections <= 1 {
		return 1
	}
	return min(maxConcurrentProviderLifecycleSessions, maxConnections-1)
}

func providerLifecycleSessionFromContext(
	ctx context.Context,
	pool *pgxpool.Pool,
) *providerLifecycleSession {
	session, _ := ctx.Value(providerLifecycleSessionContextKey{}).(*providerLifecycleSession)
	if session == nil || session.pool != pool {
		return nil
	}
	return session
}

func (session *providerLifecycleSession) withLock(
	ctx context.Context,
	lockKey string,
	operation func(context.Context) error,
) (resultErr error) {
	if session.tainted {
		return errors.New("Google Drive provider lifecycle lock session is unavailable")
	}
	if err := session.acquire(ctx, lockKey); err != nil {
		return err
	}
	defer func() {
		if !session.tainted {
			resultErr = errors.Join(resultErr, session.release(ctx, lockKey))
		}
	}()
	return operation(ctx)
}

func (session *providerLifecycleSession) acquire(ctx context.Context, lockKey string) error {
	if err := session.queries.AcquireGoogleDriveProviderLifecycleLock(
		ctx,
		googledrivesql.AcquireGoogleDriveProviderLifecycleLockParams{LockKey: lockKey},
	); err != nil {
		// A transport failure can leave the server-side session in an unknown
		// state. Tainting guarantees it will be closed rather than pooled.
		session.tainted = true
		return fmt.Errorf("acquire Google Drive provider lifecycle lock: %w", err)
	}
	return nil
}

func (session *providerLifecycleSession) release(ctx context.Context, lockKey string) error {
	unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	released, err := session.queries.ReleaseGoogleDriveProviderLifecycleLock(
		unlockCtx,
		googledrivesql.ReleaseGoogleDriveProviderLifecycleLockParams{LockKey: lockKey},
	)
	if err == nil && released {
		return nil
	}
	session.tainted = true
	if err == nil {
		err = errors.New("provider lifecycle lock was not held by its reserved connection")
	}
	return fmt.Errorf("release Google Drive provider lifecycle lock: %w", err)
}

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	for _, domainError := range []error{
		domain.ErrNotFound, domain.ErrConflict, domain.ErrAccountOwned, domain.ErrForbidden, domain.ErrInvalidInput,
	} {
		if errors.Is(err, domainError) {
			return err
		}
	}
	switch platformdatabase.Classify(err) {
	case platformdatabase.ErrorClassUniqueViolation,
		platformdatabase.ErrorClassSerializationFailure,
		platformdatabase.ErrorClassDeadlock:
		return fmt.Errorf("%w: %v", domain.ErrConflict, err)
	case platformdatabase.ErrorClassForeignKeyViolation,
		platformdatabase.ErrorClassNotNullViolation,
		platformdatabase.ErrorClassCheckViolation:
		return fmt.Errorf("%w: %v", domain.ErrForbidden, err)
	default:
		return err
	}
}

func requireAffected(rows int64, err error, zeroError error) error {
	if err != nil {
		return mapDatabaseError(err)
	}
	if rows != 1 {
		return zeroError
	}
	return nil
}
