package searchrepository

import (
	"errors"

	searchsql "github.com/complexus-tech/projects-api/internal/modules/search/repository/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryNotConfigured = errors.New("search repository is not configured")

type repo struct {
	queries searchsql.Querier
}

// New creates a search repository backed by the shared native pgx pool.
func New(pool *pgxpool.Pool) *repo {
	if pool == nil {
		return &repo{}
	}

	return &repo{queries: searchsql.New(pool)}
}

func (r *repo) ready() error {
	if r == nil || r.queries == nil {
		return errRepositoryNotConfigured
	}
	return nil
}
