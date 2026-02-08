package workerbootstrap

import (
	"github.com/complexus-tech/projects-api/internal/taskhandlers"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

func buildTaskMux(log *logger.Logger, db *sqlx.DB, brevoService *brevo.Service, mailerService mailer.Service, systemUserID uuid.UUID) *asynq.ServeMux {
	workerTaskService := taskhandlers.NewWorkerHandlers(log, db, brevoService, mailerService)
	cleanupHandlers := taskhandlers.NewCleanupHandlers(log, db, mailerService, systemUserID)

	mux := asynq.NewServeMux()

	// Existing handlers
	mux.HandleFunc(tasks.TypeUserOnboardingStart, workerTaskService.HandleUserOnboardingStart)
	mux.HandleFunc(tasks.TypeWorkspaceTrialStart, workerTaskService.HandleWorkspaceTrialStart)
	mux.HandleFunc(tasks.TypeWorkspaceTrialEnd, workerTaskService.HandleWorkspaceTrialEnd)
	mux.HandleFunc(tasks.TypeSubscriberUpdate, workerTaskService.HandleSubscriberUpdate)
	mux.HandleFunc(tasks.TypeNotificationEmail, workerTaskService.HandleNotificationEmail)

	// Cleanup handlers
	mux.HandleFunc(tasks.TypeTokenCleanup, cleanupHandlers.HandleTokenCleanup)
	mux.HandleFunc(tasks.TypeDeleteStories, cleanupHandlers.HandleDeleteStories)
	mux.HandleFunc(tasks.TypeWebhookCleanup, cleanupHandlers.HandleWebhookCleanup)
	mux.HandleFunc(tasks.TypeChatSessionsCleanup, cleanupHandlers.HandleChatSessionsCleanup)
	mux.HandleFunc(tasks.TypeWorkspaceCleanup, cleanupHandlers.HandleWorkspaceCleanup)

	// Automation handlers
	mux.HandleFunc(tasks.TypeSprintAutoCreation, cleanupHandlers.HandleSprintAutoCreation)
	mux.HandleFunc(tasks.TypeStoryAutoArchive, cleanupHandlers.HandleStoryAutoArchive)
	mux.HandleFunc(tasks.TypeStoryAutoClose, cleanupHandlers.HandleStoryAutoClose)
	mux.HandleFunc(tasks.TypeSprintStoryMigration, cleanupHandlers.HandleSprintStoryMigration)
	mux.HandleFunc("overdue:stories:email", cleanupHandlers.HandleOverdueStoriesEmail)
	mux.HandleFunc("overdue:objectives:email", cleanupHandlers.HandleObjectiveOverdueEmail)
	mux.HandleFunc(tasks.TypeDisableInactiveAutomation, cleanupHandlers.HandleDisableInactiveAutomation)

	// Lifecycle management handlers
	mux.HandleFunc(tasks.TypeWorkspaceInactivityWarning, cleanupHandlers.HandleWorkspaceInactivityWarning)
	mux.HandleFunc(tasks.TypeUserInactivityWarning, cleanupHandlers.HandleUserInactivityWarning)
	mux.HandleFunc(tasks.TypeWorkspaceDeletion, cleanupHandlers.HandleWorkspaceDeletion)
	mux.HandleFunc(tasks.TypeUserDeactivation, cleanupHandlers.HandleUserDeactivation)

	return mux
}
