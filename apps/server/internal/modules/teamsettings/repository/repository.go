package teamsettingsrepository

import (
	"context"
	"errors"

	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("team settings repository transactions are unavailable")

type repo struct {
	queries        teamsettingssql.Querier
	runTransaction func(context.Context, func(teamsettingssql.Querier) error) error
}

func New(pool *pgxpool.Pool) *repo {
	queries := teamsettingssql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(teamsettingssql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.Serializable,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries teamsettingssql.Querier) *repo {
	return &repo{queries: queries}
}

func (r *repo) withinTransaction(
	ctx context.Context,
	operation func(teamsettingssql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if r.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return mapDatabaseError(r.runTransaction(ctx, operation))
}
