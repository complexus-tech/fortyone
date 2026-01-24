package taskhandlers

import (
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/jmoiron/sqlx"
)

type handlers struct {
	log           *logger.Logger
	db            *sqlx.DB
	brevoService  *brevo.Service
	mailerService mailer.Service
}

// NewWorkerHandlers initializes the central task Handlers service.
func NewWorkerHandlers(log *logger.Logger, db *sqlx.DB, brevoService *brevo.Service, mailerService mailer.Service) *handlers {
	return &handlers{
		log:           log,
		db:            db,
		brevoService:  brevoService,
		mailerService: mailerService,
	}
}
