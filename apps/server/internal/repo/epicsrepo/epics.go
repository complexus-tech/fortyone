package epicsrepo

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/epics"
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

func (r *repo) List(ctx context.Context) ([]epics.CoreEpic, error) {

	p := []dbEpic{
		{ID: uuid.New(), Name: "Epic 1", Description: "This is epic 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Epic 2", Description: "This is epic 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Epic 3", Description: "This is epic 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreEpics(p), nil
}
