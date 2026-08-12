package taskhandlers

import (
	"context"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SlackEventProcessor interface {
	ProcessEvent(ctx context.Context, externalWorkspaceID, eventID string) error
}

type SlackCredentialBackfiller interface {
	BackfillLegacyCredentials(ctx context.Context) (int, error)
}

type SlackInboxRecoverer interface {
	RecoverPendingEvents(ctx context.Context) (int, error)
}

type handlers struct {
	log              *logger.Logger
	db               *sqlx.DB
	brevoService     *brevo.Service
	mailerService    mailer.Service
	githubService    *github.Service
	mayaService      *maya.Service
	attachments      *attachments.Service
	emailCopy        emailcopy.Generator
	slackEvents      SlackEventProcessor
	slackCredentials SlackCredentialBackfiller
	slackRecovery    SlackInboxRecoverer
	systemUserID     uuid.UUID
}

// NewWorkerHandlers initializes the central task Handlers service.
func NewWorkerHandlers(log *logger.Logger, db *sqlx.DB, brevoService *brevo.Service, mailerService mailer.Service, githubService *github.Service, mayaService *maya.Service, attachmentsService *attachments.Service, emailCopy emailcopy.Generator, slackEvents SlackEventProcessor, systemUserID uuid.UUID) *handlers {
	slackCredentials, _ := slackEvents.(SlackCredentialBackfiller)
	slackRecovery, _ := slackEvents.(SlackInboxRecoverer)
	return &handlers{
		log:              log,
		db:               db,
		brevoService:     brevoService,
		mailerService:    mailerService,
		githubService:    githubService,
		mayaService:      mayaService,
		attachments:      attachmentsService,
		emailCopy:        emailCopy,
		slackEvents:      slackEvents,
		slackCredentials: slackCredentials,
		slackRecovery:    slackRecovery,
		systemUserID:     systemUserID,
	}
}
