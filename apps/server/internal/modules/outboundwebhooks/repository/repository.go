package outboundwebhooksrepository

import (
	"errors"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhookssql "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotConfigured = errors.New("outbound webhooks repository is not configured")

type Repository struct {
	pool       *pgxpool.Pool
	queries    *outboundwebhookssql.Queries
	transactor platformdatabase.Transactor
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{
		pool:       pool,
		queries:    outboundwebhookssql.New(pool),
		transactor: platformdatabase.NewTransactor(pool),
	}
}

func (repository *Repository) configured() error {
	if repository == nil || repository.pool == nil || repository.queries == nil {
		return ErrNotConfigured
	}
	return nil
}

func mapReadError(err error, notFound error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return notFound
	}
	return err
}

func mapWriteError(err error) error {
	if platformdatabase.Classify(err) == platformdatabase.ErrorClassUniqueViolation {
		return errors.Join(outboundwebhooksdomain.ErrEndpointConflict, err)
	}
	return err
}
