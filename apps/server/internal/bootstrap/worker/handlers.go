package workerbootstrap

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	"github.com/complexus-tech/projects-api/internal/taskhandlers"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

func buildTaskMux(log *logger.Logger, db *sqlx.DB, brevoService *brevo.Service, mailerService mailer.Service, githubService *github.Service, mayaService *maya.Service, attachmentsService *attachments.Service, emailCopy emailcopy.Generator, emailThreads emailthread.GuidancePreparer, notificationsService *notifications.Service, slackEvents taskhandlers.SlackEventProcessor, emailReplies taskhandlers.EmailReplyProcessor, emailRecovery taskhandlers.EmailReplyRecoverer, calendar taskhandlers.CalendarSyncProcessor, systemUserID uuid.UUID, feedbackTasks *tasks.Service, feedbackOutbox taskhandlers.FeedbackOutboxProcessor, feedbackAuthSecret string) *asynq.ServeMux {
	workerTaskService := taskhandlers.NewWorkerHandlers(log, db, brevoService, mailerService, githubService, mayaService, attachmentsService, emailCopy, emailThreads, slackEvents, emailReplies, emailRecovery, calendar, systemUserID, feedbackTasks, feedbackOutbox, feedbackAuthSecret)
	cleanupHandlers := taskhandlers.NewCleanupHandlers(log, db, mailerService, emailCopy, emailThreads, systemUserID, notificationsService)

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
	mux.HandleFunc(tasks.TypeCalendarScheduleOutbox, workerTaskService.HandleCalendarScheduleOutboxDispatch)

	// Cleanup handlers
	mux.HandleFunc(tasks.TypeTokenCleanup, cleanupHandlers.HandleTokenCleanup)
	mux.HandleFunc(tasks.TypeDeleteStories, cleanupHandlers.HandleDeleteStories)
	mux.HandleFunc(tasks.TypeDeleteFeedback, cleanupHandlers.HandleDeleteFeedback)
	mux.HandleFunc(tasks.TypeWebhookCleanup, cleanupHandlers.HandleWebhookCleanup)
	mux.HandleFunc(tasks.TypeChatSessionsCleanup, cleanupHandlers.HandleChatSessionsCleanup)
	mux.HandleFunc(tasks.TypeMessagingCleanup, cleanupHandlers.HandleMessagingCleanup)
	mux.HandleFunc(tasks.TypeWorkspaceCleanup, cleanupHandlers.HandleWorkspaceCleanup)

	// Automation handlers
	mux.HandleFunc(tasks.TypeSprintAutoCreation, cleanupHandlers.HandleSprintAutoCreation)
	mux.HandleFunc(tasks.TypeStoryAutoArchive, cleanupHandlers.HandleStoryAutoArchive)
	mux.HandleFunc(tasks.TypeStoryAutoClose, cleanupHandlers.HandleStoryAutoClose)
	mux.HandleFunc(tasks.TypeSprintStoryMigration, cleanupHandlers.HandleSprintStoryMigration)
	mux.HandleFunc(tasks.TypeMayaWorkFocusInference, cleanupHandlers.HandleMayaWorkFocusInference)
	mux.HandleFunc("overdue:stories:email", cleanupHandlers.HandleOverdueStoriesEmail)
	mux.HandleFunc("overdue:objectives:email", cleanupHandlers.HandleObjectiveOverdueEmail)
	mux.HandleFunc(tasks.TypeWeeklyDigestEmail, cleanupHandlers.HandleWeeklyDigestEmail)
	mux.HandleFunc(tasks.TypeFeedbackDigestEmail, cleanupHandlers.HandleFeedbackDigestEmail)
	mux.HandleFunc(tasks.TypeStrategyCommunications, cleanupHandlers.HandleStrategyCommunications)
	mux.HandleFunc(tasks.TypeDisableInactiveAutomation, cleanupHandlers.HandleDisableInactiveAutomation)

	// Lifecycle management handlers
	mux.HandleFunc(tasks.TypeWorkspaceInactivityWarning, cleanupHandlers.HandleWorkspaceInactivityWarning)
	mux.HandleFunc(tasks.TypeUserInactivityWarning, cleanupHandlers.HandleUserInactivityWarning)
	mux.HandleFunc(tasks.TypeWorkspaceDeletion, cleanupHandlers.HandleWorkspaceDeletion)
	mux.HandleFunc(tasks.TypeUserDeactivation, cleanupHandlers.HandleUserDeactivation)

	return mux
}
