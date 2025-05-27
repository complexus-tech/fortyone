package taskhandlers

import (
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type handlers struct {
	log          *logger.Logger
	db           *sqlx.DB
	brevoService *brevo.Service
}

// NewWorkerHandlers initializes the central task Handlers service.
func NewWorkerHandlers(log *logger.Logger, db *sqlx.DB, brevoService *brevo.Service) *handlers {
	return &handlers{
		log:          log,
		db:           db,
		brevoService: brevoService,
	}
}
