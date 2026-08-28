package feedbackrepository

import (
	"context"
	"errors"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	feedbacksql "github.com/complexus-tech/projects-api/internal/modules/feedback/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("feedback repository transactions are unavailable")

type Repo struct {
	log            *logger.Logger
	queries        feedbacksql.Querier
	runTransaction func(context.Context, pgx.TxOptions, func(feedbacksql.Querier) error) error
}

func New(log *logger.Logger, pool *pgxpool.Pool) *Repo {
	if pool == nil {
		return &Repo{log: log}
	}
	queries := feedbacksql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(log, queries)
	repository.runTransaction = func(ctx context.Context, options pgx.TxOptions, operation func(feedbacksql.Querier) error) error {
		return transactor.WithinTransaction(ctx, options, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(log *logger.Logger, queries feedbacksql.Querier) *Repo {
	repository := &Repo{log: log, queries: queries}
	if queries != nil {
		repository.runTransaction = func(_ context.Context, _ pgx.TxOptions, operation func(feedbacksql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (r *Repo) withinTransaction(
	ctx context.Context,
	options pgx.TxOptions,
	operation func(feedbacksql.Querier) error,
) error {
	if operation == nil {
		return platformdatabase.ErrNilTransactionOperation
	}
	if r == nil || r.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return r.runTransaction(ctx, options, operation)
}

func normalizeError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return feedback.ErrNotFound
	}
	return err
}

func requireRowsAffected(count int64) error {
	if count == 0 {
		return feedback.ErrNotFound
	}
	return nil
}
