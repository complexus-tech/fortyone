package developercredentialsrepository

import (
	"context"
	"errors"

	developercredentialssql "github.com/complexus-tech/projects-api/internal/modules/developercredentials/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("developer credential repository transactions are unavailable")

type Store struct {
	queries        developercredentialssql.Querier
	runTransaction func(context.Context, func(developercredentialssql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Store {
	queries := developercredentialssql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	store := newWithQueries(queries)
	store.runTransaction = func(ctx context.Context, operation func(developercredentialssql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.Serializable,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return store
}

func newWithQueries(queries developercredentialssql.Querier) *Store {
	return &Store{queries: queries}
}

func (store *Store) withinTransaction(
	ctx context.Context,
	operation func(developercredentialssql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if store.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return mapDatabaseError(store.runTransaction(ctx, operation))
}
