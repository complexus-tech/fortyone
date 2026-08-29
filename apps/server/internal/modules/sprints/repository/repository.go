package sprintsrepository

import (
	sprintssql "github.com/complexus-tech/projects-api/internal/modules/sprints/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists sprints through the generated SQLC query set.
type Repository struct {
	queries    *sprintssql.Queries
	transactor platformdatabase.Transactor
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		queries:    sprintssql.New(pool),
		transactor: platformdatabase.NewTransactor(pool),
	}
}
