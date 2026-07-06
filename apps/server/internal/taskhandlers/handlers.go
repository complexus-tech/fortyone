package taskhandlers

import (
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type handlers struct {
	log           *logger.Logger
	db            *sqlx.DB
	brevoService  *brevo.Service
	mailerService mailer.Service
	githubService *github.Service
	mayaService   *maya.Service
	systemUserID  uuid.UUID
}

// NewWorkerHandlers initializes the central task Handlers service.
func NewWorkerHandlers(log *logger.Logger, db *sqlx.DB, brevoService *brevo.Service, mailerService mailer.Service, githubService *github.Service, mayaService *maya.Service, systemUserID uuid.UUID) *handlers {
	return &handlers{
		log:           log,
		db:            db,
		brevoService:  brevoService,
		mailerService: mailerService,
		githubService: githubService,
		mayaService:   mayaService,
		systemUserID:  systemUserID,
	}
}
