package objectivesrepo

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
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

func (r *repo) List(ctx context.Context) ([]objectives.CoreObjective, error) {

	p := []dbObjective{
		{ID: uuid.New(), Name: "Project 1", Description: "This is project 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Project 2", Description: "This is project 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Project 3", Description: "This is project 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreObjectives(p), nil
}
