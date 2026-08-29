package objectivesrepository

import (
	"context"
	"errors"

	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("objectives repository transactions are unavailable")

type Repository struct {
	queries        objectivessql.Querier
	runTransaction func(context.Context, func(objectivessql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	queries := objectivessql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(objectivessql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries objectivessql.Querier) *Repository {
	repository := &Repository{queries: queries}
	if queries != nil {
		repository.runTransaction = func(ctx context.Context, operation func(objectivessql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (repository *Repository) withinTransaction(
	ctx context.Context,
	operation func(objectivessql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if repository == nil || repository.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return mapDatabaseError(repository.runTransaction(ctx, operation))
}
