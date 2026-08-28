package subscriptionsrepository

import (
	"errors"

	subscriptionssql "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryNotConfigured = errors.New("subscriptions repository is not configured")

type Repository struct {
	queries subscriptionssql.Querier
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{queries: subscriptionssql.New(pool)}
}

func (repository *Repository) configured() error {
	if repository == nil || repository.queries == nil {
		return errRepositoryNotConfigured
	}
	return nil
}
