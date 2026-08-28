// Package figmarepository implements Figma persistence with module-local sqlc
// queries and native pgx transactions.
package figmarepository

import (
	"context"
	"errors"
	"time"

	figmasql "github.com/complexus-tech/projects-api/internal/modules/figma/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errTransactionsUnavailable = errors.New("figma repository transactions are unavailable")

type Repository struct {
	queries        figmasql.Querier
	runTransaction func(context.Context, func(figmasql.Querier) error) error
	now            func() time.Time
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{now: time.Now}
	}
	queries := figmasql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(
		ctx context.Context,
		operation func(figmasql.Querier) error,
	) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries figmasql.Querier) *Repository {
	repository := &Repository{queries: queries, now: time.Now}
	if queries != nil {
		repository.runTransaction = func(
			ctx context.Context,
			operation func(figmasql.Querier) error,
		) error {
			return operation(queries)
		}
	}
	return repository
}

func (repository *Repository) withinTransaction(
	ctx context.Context,
	operation func(figmasql.Querier) error,
) error {
	if repository == nil || repository.queries == nil || repository.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return mapDatabaseError(repository.runTransaction(ctx, operation))
}

func (repository *Repository) currentTime() time.Time {
	if repository == nil || repository.now == nil {
		return time.Now().UTC()
	}
	return repository.now().UTC()
}
