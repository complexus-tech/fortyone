package attachmentsrepository

import (
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

// Repository provides access to the attachments store
type Repository struct {
	log *logger.Logger
	db  *sqlx.DB
}

// New creates a new attachments repository
func New(log *logger.Logger, db *sqlx.DB) *Repository {
	return &Repository{
		log: log,
		db:  db,
	}
}
