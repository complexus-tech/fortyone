package labelsrepository

import (
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		log: log,
		db:  db,
	}
}
