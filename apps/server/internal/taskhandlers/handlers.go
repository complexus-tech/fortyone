package taskhandlers

import (
	"context"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
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

type EmailReplyProcessor interface {
	ProcessEvent(ctx context.Context, externalWorkspaceID, eventID string) error
}

type EmailReplyRecoverer interface {
	RecoverPendingEvents(ctx context.Context) (int, error)
}

type FeedbackDeliveryEnqueuer interface {
	EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload) error
}

type FeedbackOutboxProcessor interface {
	DispatchReadyOutboxEvents(context.Context) error
}

type handlers struct {
	log                *logger.Logger
	db                 *sqlx.DB
	brevoService       *brevo.Service
	mailerService      mailer.Service
	githubService      *github.Service
	mayaService        *maya.Service
	attachments        *attachments.Service
	emailCopy          emailcopy.Generator
	emailThreads       emailthread.GuidancePreparer
	slackEvents        SlackEventProcessor
	slackCredentials   SlackCredentialBackfiller
	slackRecovery      SlackInboxRecoverer
	emailReplies       EmailReplyProcessor
	emailRecovery      EmailReplyRecoverer
	calendar           CalendarSyncProcessor
	systemUserID       uuid.UUID
	feedbackTasks      FeedbackDeliveryEnqueuer
	feedbackOutbox     FeedbackOutboxProcessor
	feedbackDeliveries feedbackContributorDeliveryStore
	feedbackAuthSecret string
}

// NewWorkerHandlers initializes the central task Handlers service.
func NewWorkerHandlers(log *logger.Logger, db *sqlx.DB, brevoService *brevo.Service, mailerService mailer.Service, githubService *github.Service, mayaService *maya.Service, attachmentsService *attachments.Service, emailCopy emailcopy.Generator, emailThreads emailthread.GuidancePreparer, slackEvents SlackEventProcessor, emailReplies EmailReplyProcessor, emailRecovery EmailReplyRecoverer, calendar CalendarSyncProcessor, systemUserID uuid.UUID, feedbackTasks FeedbackDeliveryEnqueuer, feedbackOutbox FeedbackOutboxProcessor, feedbackAuthSecret string) *handlers {
	slackCredentials, _ := slackEvents.(SlackCredentialBackfiller)
	slackRecovery, _ := slackEvents.(SlackInboxRecoverer)
	return &handlers{
		log:                log,
		db:                 db,
		brevoService:       brevoService,
		mailerService:      mailerService,
		githubService:      githubService,
		mayaService:        mayaService,
		attachments:        attachmentsService,
		emailCopy:          emailCopy,
		emailThreads:       emailThreads,
		slackEvents:        slackEvents,
		slackCredentials:   slackCredentials,
		slackRecovery:      slackRecovery,
		emailReplies:       emailReplies,
		emailRecovery:      emailRecovery,
		calendar:           calendar,
		systemUserID:       systemUserID,
		feedbackTasks:      feedbackTasks,
		feedbackOutbox:     feedbackOutbox,
		feedbackDeliveries: &databaseFeedbackContributorDeliveryStore{db: db},
		feedbackAuthSecret: feedbackAuthSecret,
	}
}
