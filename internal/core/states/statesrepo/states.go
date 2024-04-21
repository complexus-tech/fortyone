package statesrepo

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/states"
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

func (r *repo) List(ctx context.Context) ([]states.CoreState, error) {

	p := []dbState{
		{ID: uuid.New(), Name: "State 1", Description: "This is state 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "State 2", Description: "This is state 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "State 3", Description: "This is state 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreStates(p), nil
}
