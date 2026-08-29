package actorsrepository

import (
	"context"
	"errors"
	"fmt"

	actorssql "github.com/complexus-tech/projects-api/internal/platform/actors/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSystemActorNotFound = errors.New("active system actor not found")

type Repository struct {
	queries *actorssql.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: actorssql.New(pool)}
}

func (repository *Repository) FindActiveSystemActorByEmail(
	ctx context.Context,
	email string,
) (uuid.UUID, error) {
	actorID, err := repository.queries.FindActiveSystemActorByEmail(
		ctx,
		actorssql.FindActiveSystemActorByEmailParams{Email: email},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrSystemActorNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("find active system actor: %w", err)
	}
	return actorID, nil
}
