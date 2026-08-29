package adminrepository

import (
	"context"
	"errors"
	"fmt"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("admin repository transactions are unavailable")

// Repository owns the admin authorization and transaction boundary. Every
// operation uses the same transaction for its live actor check and its reads,
// mutations, and immutable audit writes.
type Repository struct {
	queries        adminsql.Querier
	runTransaction func(context.Context, func(adminsql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}

	queries := adminsql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(adminsql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

// newWithQueries is the narrow unit-test seam. Production always composes New
// with a pgx pool so database transactions remain repository-owned.
func newWithQueries(queries adminsql.Querier) *Repository {
	repository := &Repository{queries: queries}
	if queries != nil {
		repository.runTransaction = func(ctx context.Context, operation func(adminsql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (repository *Repository) withinTransaction(
	ctx context.Context,
	operation func(adminsql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if repository == nil || repository.queries == nil || repository.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return mapDatabaseError(repository.runTransaction(ctx, operation))
}

func (repository *Repository) withActiveInternalAdmin(
	ctx context.Context,
	actorID uuid.UUID,
	operation func(adminsql.Querier) error,
) error {
	return repository.withinTransaction(ctx, func(queries adminsql.Querier) error {
		if _, err := queries.LockActiveInternalAdmin(ctx, adminsql.LockActiveInternalAdminParams{
			ActorID: actorID,
		}); errors.Is(err, pgx.ErrNoRows) {
			return admindomain.ErrForbidden
		} else if err != nil {
			return fmt.Errorf("authorize active internal admin: %w", err)
		}
		return operation(queries)
	})
}

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	for _, domainErr := range []error{
		admindomain.ErrForbidden,
		admindomain.ErrNotFound,
		admindomain.ErrConflict,
		admindomain.ErrInvalidAction,
		admindomain.ErrInvalidPagination,
		admindomain.ErrSelfMutation,
		admindomain.ErrInvalidTrialEndsOn,
	} {
		if errors.Is(err, domainErr) {
			return err
		}
	}

	switch platformdatabase.Classify(err) {
	case platformdatabase.ErrorClassSerializationFailure, platformdatabase.ErrorClassDeadlock,
		platformdatabase.ErrorClassUniqueViolation:
		return fmt.Errorf("%w: %v", admindomain.ErrConflict, err)
	case platformdatabase.ErrorClassForeignKeyViolation, platformdatabase.ErrorClassNotNullViolation,
		platformdatabase.ErrorClassCheckViolation:
		return fmt.Errorf("%w: %v", admindomain.ErrInvalidAction, err)
	default:
		return err
	}
}
