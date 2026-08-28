package taskhandlers

import (
	"context"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

type SlackEventProcessor interface {
	ProcessWebhook(ctx context.Context, provider integrations.ProviderKey, inboxID uuid.UUID) error
}

type FigmaWebhookProcessor interface {
	ProcessWebhook(ctx context.Context, provider integrations.ProviderKey, inboxID uuid.UUID) error
}

type FigmaWebhookRecoverer interface {
	RecoverPendingWebhooks(ctx context.Context) (int, error)
}

type legacySlackEventProcessor interface {
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

// MayaTaskProcessor is the worker-owned subset of Maya's application service.
// Keeping this boundary narrow lets worker behavior be tested without a live
// planner, while the API and worker still share the same service instance.
type MayaTaskProcessor interface {
	ReconcileSchedule(context.Context, maya.ReconcileScheduleInput) error
	ProcessAssignmentBatch(context.Context, maya.ProcessAssignmentBatchInput) (maya.ProcessAssignmentBatchResult, error)
	RecoverScheduleOwnerships(context.Context, int) (int, error)
}

// MayaAssignmentStore owns the bounded read models used by assignment and
// workspace-schedule workers. Implementations must return stories ordered by
// ID and strictly after afterStoryID so the handlers can advance safely.
type MayaAssignmentStore interface {
	ListMayaAssignmentCandidates(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		int,
	) ([]mayadomain.AssignmentCandidateStory, error)
	ListWorkspaceScheduleCandidates(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		int,
	) ([]mayadomain.AssignmentCandidateStory, error)
}

// StorySyncReader is owned by the task use case. Bootstrap supplies the pgx
// story adapter, while the handler supplies an explicit system-actor scope for
// every read instead of reaching into a concrete repository or raw database.
type StorySyncReader interface {
	GetStoryForMutation(context.Context, storydomain.MutationScope, uuid.UUID) (storydomain.Story, error)
}

// NotificationDeliveryStore is owned by the email task use case. Bootstrap
// supplies the notifications service; worker code never owns notification SQL.
type NotificationDeliveryStore interface {
	GetEmailDelivery(context.Context, notificationsdomain.EmailNotificationQuery) (*notificationsdomain.EmailNotification, error)
	ListEmailDigest(context.Context, notificationsdomain.DeliveryScope) (*notificationsdomain.EmailDigest, error)
	ListDeliveryTeamIDs(context.Context, notificationsdomain.DeliveryScope) ([]uuid.UUID, error)
	MarkEmailSent(context.Context, notificationsdomain.DeliveryScope, []uuid.UUID) error
}

type handlers struct {
	log                    *logger.Logger
	brevoService           *brevo.Service
	mailerService          mailer.Service
	githubService          *github.Service
	storySyncReader        StorySyncReader
	mayaService            MayaTaskProcessor
	mayaAssignments        MayaAssignmentStore
	attachments            *attachments.Service
	emailCopy              emailcopy.Generator
	emailThreads           emailthread.GuidancePreparer
	notificationDeliveries NotificationDeliveryStore
	slackEvents            SlackEventProcessor
	slackCredentials       SlackCredentialBackfiller
	slackRecovery          SlackInboxRecoverer
	figmaWebhooks          FigmaWebhookProcessor
	figmaRecovery          FigmaWebhookRecoverer
	emailReplies           EmailReplyProcessor
	emailRecovery          EmailReplyRecoverer
	calendar               CalendarSyncProcessor
	systemUserID           uuid.UUID
	feedbackTasks          FeedbackDeliveryEnqueuer
	feedbackOutbox         FeedbackOutboxProcessor
	storyScheduleOutbox    StoryScheduleTransitionOutboxProcessor
	invitationOutbox       InvitationOutboxProcessor
	feedbackDeliveries     feedbackContributorDeliveryStore
	feedbackSecurityKey    string
}

// WorkerHandlerDependencies makes task composition explicit and prevents the
// central worker constructor from becoming an unsafe positional-argument list.
type WorkerHandlerDependencies struct {
	Log                    *logger.Logger
	Brevo                  *brevo.Service
	Mailer                 mailer.Service
	GitHub                 *github.Service
	StorySyncReader        StorySyncReader
	Maya                   MayaTaskProcessor
	MayaAssignments        MayaAssignmentStore
	Attachments            *attachments.Service
	EmailCopy              emailcopy.Generator
	EmailThreads           emailthread.GuidancePreparer
	NotificationDeliveries NotificationDeliveryStore
	SlackEvents            SlackEventProcessor
	FigmaWebhooks          FigmaWebhookProcessor
	EmailReplies           EmailReplyProcessor
	EmailRecovery          EmailReplyRecoverer
	Calendar               CalendarSyncProcessor
	SystemUserID           uuid.UUID
	FeedbackTasks          FeedbackDeliveryEnqueuer
	FeedbackOutbox         FeedbackOutboxProcessor
	FeedbackDeliveries     feedback.ContributorDeliveryStore
	StoryScheduleOutbox    StoryScheduleTransitionOutboxProcessor
	InvitationOutbox       InvitationOutboxProcessor
	FeedbackSecurityKey    string
}

// NewWorkerHandlers initializes the central task handlers service.
func NewWorkerHandlers(dependencies WorkerHandlerDependencies) *handlers {
	slackCredentials, _ := dependencies.SlackEvents.(SlackCredentialBackfiller)
	slackRecovery, _ := dependencies.SlackEvents.(SlackInboxRecoverer)
	figmaRecovery, _ := dependencies.FigmaWebhooks.(FigmaWebhookRecoverer)
	return &handlers{
		log:                    dependencies.Log,
		brevoService:           dependencies.Brevo,
		mailerService:          dependencies.Mailer,
		githubService:          dependencies.GitHub,
		storySyncReader:        dependencies.StorySyncReader,
		mayaService:            dependencies.Maya,
		mayaAssignments:        dependencies.MayaAssignments,
		attachments:            dependencies.Attachments,
		emailCopy:              dependencies.EmailCopy,
		emailThreads:           dependencies.EmailThreads,
		notificationDeliveries: dependencies.NotificationDeliveries,
		slackEvents:            dependencies.SlackEvents,
		slackCredentials:       slackCredentials,
		slackRecovery:          slackRecovery,
		figmaWebhooks:          dependencies.FigmaWebhooks,
		figmaRecovery:          figmaRecovery,
		emailReplies:           dependencies.EmailReplies,
		emailRecovery:          dependencies.EmailRecovery,
		calendar:               dependencies.Calendar,
		systemUserID:           dependencies.SystemUserID,
		feedbackTasks:          dependencies.FeedbackTasks,
		feedbackOutbox:         dependencies.FeedbackOutbox,
		storyScheduleOutbox:    dependencies.StoryScheduleOutbox,
		invitationOutbox:       dependencies.InvitationOutbox,
		feedbackDeliveries:     dependencies.FeedbackDeliveries,
		feedbackSecurityKey:    dependencies.FeedbackSecurityKey,
	}
}
