package objectivesrepository

import (
	"errors"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

// Repository errors
var (
	ErrNotFound = errors.New("objective not found")
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
