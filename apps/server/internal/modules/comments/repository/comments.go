package commentsrepository

import (
	"context"
	"errors"

	commentsql "github.com/complexus-tech/projects-api/internal/modules/comments/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("comments repository transactions are unavailable")

type Repository struct {
	log            *logger.Logger
	queries        commentsql.Querier
	runTransaction func(context.Context, func(commentsql.Querier) error) error
}

func New(log *logger.Logger, db *pgxpool.Pool) *Repository {
	if db == nil {
		return &Repository{log: log}
	}
	queries := commentsql.New(db)
	transactor := platformdatabase.NewTransactor(db)
	repository := newWithQueries(log, queries)
	repository.runTransaction = func(ctx context.Context, operation func(commentsql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(log *logger.Logger, queries commentsql.Querier) *Repository {
	repository := &Repository{log: log, queries: queries}
	if queries != nil {
		repository.runTransaction = func(ctx context.Context, operation func(commentsql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (r *Repository) withinTransaction(
	ctx context.Context,
	operation func(commentsql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if r == nil || r.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return r.runTransaction(ctx, operation)
}
