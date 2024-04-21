package teamsrepo

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

func (r *repo) List(ctx context.Context) ([]teams.CoreTeam, error) {

	p := []dbTeam{
		{ID: uuid.New(), Name: "Team 1", Description: "This is team 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Team 2", Description: "This is team 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Team 3", Description: "This is team 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreTeams(p), nil
}
