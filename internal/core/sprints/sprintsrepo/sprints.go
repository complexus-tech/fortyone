package sprintsrepo

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
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

func (r *repo) List(ctx context.Context) ([]sprints.CoreSprint, error) {

	p := []dbSprint{
		{ID: uuid.New(), Name: "Sprint 1", Description: "This is sprint 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Sprint 2", Description: "This is sprint 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Sprint 3", Description: "This is sprint 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreSprints(p), nil
}
