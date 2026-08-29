package chatsessionsrepository

import (
	"context"
	"errors"

	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("chat session repository transactions are unavailable")

type repo struct {
	queries        chatsessionssql.Querier
	runTransaction func(context.Context, func(chatsessionssql.Querier) error) error
	log            *logger.Logger
}

// New constructs the native-pgx chat-session persistence adapter. Every SQL
// statement is owned by the module's generated SQLC query set.
func New(log *logger.Logger, pool *pgxpool.Pool) *repo {
	if pool == nil {
		return &repo{log: log}
	}

	queries := chatsessionssql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	return &repo{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(chatsessionssql.Querier) error) error {
			return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
		log: log,
	}
}

func (r *repo) withinTransaction(
	ctx context.Context,
	operation func(chatsessionssql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if r == nil || r.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return r.runTransaction(ctx, operation)
}
