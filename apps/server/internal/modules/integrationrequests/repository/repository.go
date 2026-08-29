package integrationrequestsrepository

import (
	"context"
	"errors"

	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryNotConfigured = errors.New("integration request repository is not configured")

type Repo struct {
	queries           integrationrequestssql.Querier
	withinTransaction func(context.Context, func(integrationrequestssql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repo {
	if pool == nil {
		return &Repo{}
	}
	queries := integrationrequestssql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.withinTransaction = func(ctx context.Context, operation func(integrationrequestssql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries integrationrequestssql.Querier) *Repo {
	repository := &Repo{queries: queries}
	if queries != nil {
		repository.withinTransaction = func(_ context.Context, operation func(integrationrequestssql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (r *Repo) configured() error {
	if r == nil || r.queries == nil || r.withinTransaction == nil {
		return errRepositoryNotConfigured
	}
	return nil
}
