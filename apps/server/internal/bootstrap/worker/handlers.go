package workerbootstrap

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	chatsessionsrepository "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository"
	feedbackrepository "github.com/complexus-tech/projects-api/internal/modules/feedback/repository"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	mayarepository "github.com/complexus-tech/projects-api/internal/modules/maya/repository"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	subscriptionsrepository "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	teamsettingsrepository "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/taskhandlers"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type taskMuxDependencies struct {
	APIPublicURL           string
	Log                    *logger.Logger
	DatabasePool           *pgxpool.Pool
	Brevo                  *brevo.Service
	Mailer                 mailer.Service
	GitHub                 *github.Service
	Figma                  taskhandlers.FigmaWebhookProcessor
	Maya                   *maya.Service
	MayaRepository         *mayarepository.Repo
	Attachments            *attachments.Service
	EmailCopy              emailcopy.Generator
	EmailThreads           emailthread.GuidancePreparer
	Notifications          *notifications.Service
	WeeklyDigest           jobs.WeeklyDigestStore
	SlackEvents            taskhandlers.SlackEventProcessor
	EmailReplies           taskhandlers.EmailReplyProcessor
	EmailRecovery          taskhandlers.EmailReplyRecoverer
	Calendar               taskhandlers.CalendarSyncProcessor
	SystemUserID           uuid.UUID
	FeedbackTasks          *tasks.Service
	FeedbackOutbox         taskhandlers.FeedbackOutboxProcessor
	StoryScheduleOutbox    taskhandlers.StoryScheduleTransitionOutboxProcessor
	GoogleDriveRevocations taskhandlers.GoogleDriveRevocationProcessor
	InvitationOutbox       taskhandlers.InvitationOutboxProcessor
	FeedbackSecurityKey    string
	IdempotencyReceipts    taskhandlers.IdempotencyReceiptPurger
}

func buildTaskMux(dependencies taskMuxDependencies) *asynq.ServeMux {
	chatSessions := chatsessionsrepository.New(dependencies.Log, dependencies.DatabasePool)
	storyStore := storiesrepository.New(dependencies.Log, dependencies.DatabasePool)
	feedbackStore := feedbackrepository.New(dependencies.Log, dependencies.DatabasePool)
	subscriptions := subscriptionsrepository.New(dependencies.DatabasePool)
	users := usersrepository.New(dependencies.DatabasePool)
	messaging := messagingrepository.New(dependencies.DatabasePool)
	objectiveGuidance := objectivesrepository.New(dependencies.DatabasePool)
	workspaceLifecycle := workspacesrepository.New(dependencies.DatabasePool)
	sprintAutomation := teamsettingsrepository.New(dependencies.DatabasePool)
	workerTaskService := taskhandlers.NewWorkerHandlers(taskhandlers.WorkerHandlerDependencies{
		Log:   dependencies.Log,
		Brevo: dependencies.Brevo, Mailer: dependencies.Mailer,
		GitHub: dependencies.GitHub, FigmaWebhooks: dependencies.Figma,
		StorySyncReader: storyStore, Maya: dependencies.Maya,
		MayaAssignments: dependencies.MayaRepository,
		Attachments:     dependencies.Attachments, EmailCopy: dependencies.EmailCopy,
		EmailThreads: dependencies.EmailThreads, NotificationDeliveries: dependencies.Notifications,
		RoutineDeliveries: notificationsrepository.New(dependencies.DatabasePool),
		EmailAvatars:      users, APIPublicURL: dependencies.APIPublicURL,
		BriefingSources: jobs.BriefingSources{Stories: storyStore, Objectives: objectiveGuidance, Weekly: dependencies.WeeklyDigest},
		SlackEvents:     dependencies.SlackEvents, EmailReplies: dependencies.EmailReplies,
		EmailRecovery: dependencies.EmailRecovery, Calendar: dependencies.Calendar,
		SystemUserID: dependencies.SystemUserID, FeedbackTasks: dependencies.FeedbackTasks,
		FeedbackOutbox:         dependencies.FeedbackOutbox,
		FeedbackDeliveries:     feedbackStore,
		StoryScheduleOutbox:    dependencies.StoryScheduleOutbox,
		GoogleDriveRevocations: dependencies.GoogleDriveRevocations,
		InvitationOutbox:       dependencies.InvitationOutbox,
		FeedbackSecurityKey:    dependencies.FeedbackSecurityKey,
	})
	storyAutomationHandlers := taskhandlers.NewStoryAutomationHandlers(
		taskhandlers.StoryAutomationHandlerDependencies{
			Log:          dependencies.Log,
			Store:        storyStore,
			SystemUserID: dependencies.SystemUserID,
		},
	)
	sprintAutomationHandlers := taskhandlers.NewSprintAutomationHandlers(
		taskhandlers.SprintAutomationHandlerDependencies{
			Log:   dependencies.Log,
			Store: sprintAutomation,
		},
	)
	storyRetentionHandlers := taskhandlers.NewStoryRetentionHandlers(
		taskhandlers.StoryRetentionHandlerDependencies{
			Log:     dependencies.Log,
			Store:   storyStore,
			Objects: dependencies.Attachments,
		},
	)
	mayaMaintenanceHandlers := taskhandlers.NewMayaMaintenanceHandlers(
		taskhandlers.MayaMaintenanceHandlerDependencies{
			Log:   dependencies.Log,
			Store: dependencies.MayaRepository,
		},
	)
	retentionHandlers := taskhandlers.NewRetentionHandlers(taskhandlers.RetentionHandlerDependencies{
		Log:                 dependencies.Log,
		ChatSessions:        chatSessions,
		VerificationTokens:  users,
		StripeWebhookEvents: subscriptions,
		MessagingData:       messaging,
		Feedback:            feedbackStore,
	})
	guidanceHandlers := taskhandlers.NewGuidanceHandlers(taskhandlers.GuidanceHandlerDependencies{
		Log:               dependencies.Log,
		Objectives:        objectiveGuidance,
		Stories:           storyStore,
		WeeklyDigest:      dependencies.WeeklyDigest,
		FeedbackDigest:    feedbackStore,
		Mailer:            dependencies.Mailer,
		CopyGenerator:     dependencies.EmailCopy,
		ThreadPreparation: dependencies.EmailThreads,
	})
	workspaceLifecycleHandlers := taskhandlers.NewWorkspaceLifecycleHandlers(
		taskhandlers.WorkspaceLifecycleHandlerDependencies{
			Log:    dependencies.Log,
			Store:  workspaceLifecycle,
			Mailer: dependencies.Mailer,
		},
	)
	userLifecycleHandlers := taskhandlers.NewUserLifecycleHandlers(
		taskhandlers.UserLifecycleHandlerDependencies{
			Log:    dependencies.Log,
			Store:  users,
			Mailer: dependencies.Mailer,
		},
	)
	strategyHandlers := taskhandlers.NewStrategyHandlers(taskhandlers.StrategyHandlerDependencies{
		Log:          dependencies.Log,
		Store:        objectiveGuidance,
		Notifier:     dependencies.Notifications,
		SystemUserID: dependencies.SystemUserID,
	})
	idempotencyCleanup := taskhandlers.NewIdempotencyCleanupHandler(
		dependencies.Log,
		dependencies.IdempotencyReceipts,
	)

	mux := asynq.NewServeMux()

	// Existing handlers
	mux.HandleFunc(tasks.TypeUserOnboardingStart, workerTaskService.HandleUserOnboardingStart)
	mux.HandleFunc(tasks.TypeWorkspaceTrialStart, workerTaskService.HandleWorkspaceTrialStart)
	mux.HandleFunc(tasks.TypeWorkspaceTrialEnd, workerTaskService.HandleWorkspaceTrialEnd)
	mux.HandleFunc(tasks.TypeSubscriberUpdate, workerTaskService.HandleSubscriberUpdate)
	mux.HandleFunc(tasks.TypeNotificationEmail, workerTaskService.HandleNotificationEmail)
	mux.HandleFunc(tasks.TypeNotificationEmailDigest, workerTaskService.HandleNotificationEmailDigest)
	mux.HandleFunc(tasks.TypeFeedbackContributorDelivery, workerTaskService.HandleFeedbackContributorDelivery)
	mux.HandleFunc(tasks.TypeFeedbackContributorDeliveryRecovery, workerTaskService.HandleFeedbackContributorDeliveryRecovery)
	mux.HandleFunc(tasks.TypeFeedbackOutboxDispatch, workerTaskService.HandleFeedbackOutboxDispatch)
	mux.HandleFunc(tasks.TypeGitHubStorySync, workerTaskService.HandleGitHubStorySync)
	mux.HandleFunc(tasks.TypeGitHubWebhook, workerTaskService.HandleGitHubWebhook)
	mux.HandleFunc(tasks.TypeGitHubWebhookRecovery, workerTaskService.HandleGitHubWebhookRecovery)
	mux.HandleFunc(tasks.TypeFigmaWebhook, workerTaskService.HandleFigmaWebhook)
	mux.HandleFunc(tasks.TypeFigmaWebhookRecovery, workerTaskService.HandleFigmaWebhookRecovery)
	mux.HandleFunc(tasks.TypeMayaBatchAssignment, workerTaskService.HandleMayaBatchAssignment)
	mux.HandleFunc(tasks.TypeMayaScheduleRecovery, workerTaskService.HandleMayaScheduleRecovery)
	mux.HandleFunc(tasks.TypeAttachmentImageOptimization, workerTaskService.HandleAttachmentImageOptimization)
	mux.HandleFunc(tasks.TypeSlackEvent, workerTaskService.HandleSlackEvent)
	mux.HandleFunc(tasks.TypeSlackCredentialBackfill, workerTaskService.HandleSlackCredentialBackfill)
	mux.HandleFunc(tasks.TypeSlackInboxRecovery, workerTaskService.HandleSlackInboxRecovery)
	mux.HandleFunc(tasks.TypeBrevoEmailReply, workerTaskService.HandleBrevoEmailReply)
	mux.HandleFunc(tasks.TypeBrevoEmailReplyRecovery, workerTaskService.HandleBrevoEmailReplyRecovery)
	mux.HandleFunc(tasks.TypeCalendarSync, workerTaskService.HandleCalendarSync)
	mux.HandleFunc(tasks.TypeCalendarWatchRenewal, workerTaskService.HandleCalendarWatchRenewal)
	mux.HandleFunc(tasks.TypeCalendarScheduleReconcile, workerTaskService.HandleCalendarScheduleReconcile)
	mux.HandleFunc(tasks.TypeCalendarWorkspaceScheduleBatch, workerTaskService.HandleCalendarWorkspaceScheduleBatch)
	mux.HandleFunc(tasks.TypeCalendarScheduleOutbox, workerTaskService.HandleCalendarScheduleOutboxDispatch)
	mux.HandleFunc(tasks.TypeGoogleDriveRevocationDispatch, workerTaskService.HandleGoogleDriveRevocationDispatch)
	mux.HandleFunc(tasks.TypeStoryScheduleTransitionOutbox, workerTaskService.HandleStoryScheduleTransitionOutboxDispatch)
	mux.HandleFunc(tasks.TypeInvitationOutboxDispatch, workerTaskService.HandleInvitationOutboxDispatch)

	// Cleanup handlers
	mux.HandleFunc(tasks.TypeTokenCleanup, retentionHandlers.HandleTokenCleanup)
	mux.HandleFunc(tasks.TypeDeleteStories, storyRetentionHandlers.HandleDeletedStories)
	mux.HandleFunc(tasks.TypeAttachmentObjectDeletions, storyRetentionHandlers.HandleAttachmentObjectDeletions)
	mux.HandleFunc(tasks.TypeDeleteFeedback, retentionHandlers.HandleDeleteFeedback)
	mux.HandleFunc(tasks.TypeWebhookCleanup, retentionHandlers.HandleWebhookCleanup)
	mux.HandleFunc(tasks.TypeChatSessionsCleanup, retentionHandlers.HandleChatSessionsCleanup)
	mux.HandleFunc(tasks.TypeMessagingCleanup, retentionHandlers.HandleMessagingCleanup)
	mux.HandleFunc(tasks.TypeAPIIdempotencyCleanup, idempotencyCleanup.Handle)
	mux.HandleFunc(tasks.TypeWorkspaceCleanup, workspaceLifecycleHandlers.HandleWorkspaceCleanup)

	// Automation handlers
	mux.HandleFunc(tasks.TypeSprintAutoCreation, sprintAutomationHandlers.HandleSprintAutoCreation)
	mux.HandleFunc(tasks.TypeStoryAutoArchive, storyAutomationHandlers.HandleStoryAutoArchive)
	mux.HandleFunc(tasks.TypeStoryAutoClose, storyAutomationHandlers.HandleStoryAutoClose)
	mux.HandleFunc(tasks.TypeSprintStoryMigration, storyAutomationHandlers.HandleSprintStoryMigration)
	mux.HandleFunc(tasks.TypeMayaWorkFocusInference, mayaMaintenanceHandlers.HandleWorkFocusInference)
	// Route legacy queued guidance jobs through the same daily claim during rollout.
	mux.HandleFunc(tasks.TypeMorningBriefing, workerTaskService.HandleMorningBriefing)
	mux.HandleFunc(tasks.TypeOverdueStoriesEmail, workerTaskService.HandleMorningBriefing)
	mux.HandleFunc(tasks.TypeObjectiveOverdueEmail, workerTaskService.HandleMorningBriefing)
	mux.HandleFunc(tasks.TypeWeeklyDigestEmail, workerTaskService.HandleMorningBriefing)
	mux.HandleFunc(tasks.TypeFeedbackDigestEmail, guidanceHandlers.HandleFeedbackDigestEmail)
	mux.HandleFunc(tasks.TypeStrategyCommunications, strategyHandlers.HandleStrategyCommunications)
	mux.HandleFunc(tasks.TypeDisableInactiveAutomation, sprintAutomationHandlers.HandleDisableInactiveAutomation)

	// Lifecycle management handlers
	mux.HandleFunc(tasks.TypeWorkspaceInactivityWarning, workspaceLifecycleHandlers.HandleWorkspaceInactivityWarning)
	mux.HandleFunc(tasks.TypeUserInactivityWarning, userLifecycleHandlers.HandleUserInactivityWarning)
	mux.HandleFunc(tasks.TypeWorkspaceDeletion, workspaceLifecycleHandlers.HandleWorkspaceDeletion)
	mux.HandleFunc(tasks.TypeUserDeactivation, userLifecycleHandlers.HandleUserDeactivation)

	return mux
}
