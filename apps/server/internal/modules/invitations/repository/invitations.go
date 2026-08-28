package invitationsrepository

import (
	"context"
	"errors"

	invitationsql "github.com/complexus-tech/projects-api/internal/modules/invitations/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("invitation repository transactions are unavailable")

type repo struct {
	queries        invitationsql.Querier
	runTransaction func(context.Context, func(invitationsql.Querier) error) error
}

func New(pool *pgxpool.Pool) *repo {
	queries := invitationsql.New(pool)
	transactor := database.NewTransactor(pool)
	return &repo{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(invitationsql.Querier) error) error {
			return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
				return operation(queries.WithTx(tx))
			})
		},
	}
}

func (r *repo) withinTransaction(
	ctx context.Context,
	operation func(invitationsql.Querier) error,
) error {
	if operation == nil {
		return database.ErrNilTransactionOperation
	}
	if r.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return r.runTransaction(ctx, operation)
}
