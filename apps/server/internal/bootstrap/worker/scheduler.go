package workerbootstrap

import (
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
)

type scheduleRegistrar interface {
	Register(spec string, task *asynq.Task, opts ...asynq.Option) (entryID string, err error)
}

func registerSchedules(scheduler scheduleRegistrar) error {
	_, err := scheduler.Register(
		"*/1 * * * *",
		asynq.NewTask(tasks.TypeFeedbackOutboxDispatch, nil),
		asynq.Queue("notifications"),
	)
	if err != nil {
		return fmt.Errorf("failed to register feedback outbox dispatch task: %w", err)
	}

	_, err = scheduler.Register(
		"*/1 * * * *",
		asynq.NewTask(tasks.TypeFeedbackContributorDeliveryRecovery, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register feedback delivery recovery task: %w", err)
	}

	_, err = scheduler.Register(
		"17 * * * *", // Hourly; only renews missing channels or channels expiring within 24 hours.
		asynq.NewTask(tasks.TypeCalendarWatchRenewal, nil),
		asynq.Queue("integrations"),
	)
	if err != nil {
		return fmt.Errorf("failed to register calendar watch renewal task: %w", err)
	}

	_, err = scheduler.Register(
		"*/1 * * * *",
		asynq.NewTask(tasks.TypeCalendarScheduleOutbox, nil),
		asynq.Queue("integrations"),
	)
	if err != nil {
		return fmt.Errorf("failed to register calendar schedule outbox task: %w", err)
	}

	_, err = scheduler.Register(
		"*/1 * * * *",
		asynq.NewTask(tasks.TypeStoryScheduleTransitionOutbox, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register story schedule transition outbox task: %w", err)
	}

	_, err = scheduler.Register(
		"*/1 * * * *",
		asynq.NewTask(tasks.TypeMayaScheduleRecovery, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register Maya schedule recovery task: %w", err)
	}

	_, err = scheduler.Register(
		"@daily",
		asynq.NewTask(tasks.TypeDeleteStories, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register delete stories task: %w", err)
	}

	_, err = scheduler.Register(
		"30 0 * * *",
		asynq.NewTask(tasks.TypeDeleteFeedback, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register delete feedback task: %w", err)
	}

	_, err = scheduler.Register(
		"@weekly", // Sunday 00:00 AM
		asynq.NewTask(tasks.TypeTokenCleanup, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register token cleanup task: %w", err)
	}

	_, err = scheduler.Register(
		"0 3 * * 3", // Wednesday 3:00 AM
		asynq.NewTask(tasks.TypeWebhookCleanup, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register webhook cleanup task: %w", err)
	}

	_, err = scheduler.Register(
		"0 */4 * * *", // Every 4 hours
		asynq.NewTask(tasks.TypeWorkspaceCleanup, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register workspace cleanup task: %w", err)
	}

	_, err = scheduler.Register(
		"10 1 * * *", // Every day at 1:10 AM (avoids Sunday collision)
		asynq.NewTask(tasks.TypeSprintAutoCreation, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register sprint auto-creation task: %w", err)
	}

	_, err = scheduler.Register(
		"0 4 * * 5", // Friday 4:00 AM
		asynq.NewTask(tasks.TypeStoryAutoArchive, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register story auto-archive task: %w", err)
	}

	_, err = scheduler.Register(
		"0 5 * * 6", // Saturday 5:00 AM
		asynq.NewTask(tasks.TypeStoryAutoClose, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register story auto-close task: %w", err)
	}

	_, err = scheduler.Register(
		"0 1 * * *", // Daily at 1:00 AM
		asynq.NewTask(tasks.TypeSprintStoryMigration, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register sprint story migration task: %w", err)
	}

	_, err = scheduler.Register(
		"30 1 * * *", // Daily at 1:30 AM, after sprint migration
		asynq.NewTask(tasks.TypeMayaWorkFocusInference, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register Maya work-focus inference task: %w", err)
	}

	_, err = scheduler.Register(
		"0 * * * *", // Hourly
		asynq.NewTask(tasks.TypeMayaBatchAssignment, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register Maya batch assignment task: %w", err)
	}

	_, err = scheduler.Register(
		"0 9 * * *", // Daily at 9:00 AM
		asynq.NewTask("overdue:stories:email", nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register overdue stories email task: %w", err)
	}

	_, err = scheduler.Register(
		"0 10 * * 1", // Monday at 10:00 AM
		asynq.NewTask("overdue:objectives:email", nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register objective overdue email task: %w", err)
	}

	_, err = scheduler.Register(
		"0 8 * * 1", // Monday at 8:00 AM
		asynq.NewTask(tasks.TypeWeeklyDigestEmail, nil),
		asynq.Queue("notifications"),
	)
	if err != nil {
		return fmt.Errorf("failed to register weekly digest email task: %w", err)
	}

	_, err = scheduler.Register(
		"0 * * * *", // Hourly; each recipient is evaluated in their local timezone
		asynq.NewTask(tasks.TypeFeedbackDigestEmail, nil),
		asynq.Queue("notifications"),
	)
	if err != nil {
		return fmt.Errorf("failed to register feedback digest email task: %w", err)
	}

	_, err = scheduler.Register(
		"0 * * * *", // Hourly; strategy communications are evaluated in each recipient's local timezone
		asynq.NewTask(tasks.TypeStrategyCommunications, nil),
		asynq.Queue("notifications"),
	)
	if err != nil {
		return fmt.Errorf("failed to register strategy communications task: %w", err)
	}

	_, err = scheduler.Register(
		"0 1 * * 0", // Sunday 01:00 AM
		asynq.NewTask(tasks.TypeWorkspaceInactivityWarning, nil),
		asynq.Queue("notifications"),
	)
	if err != nil {
		return fmt.Errorf("failed to register workspace inactivity warning task: %w", err)
	}

	_, err = scheduler.Register(
		"0 2 * * 0", // Sunday 02:00 AM
		asynq.NewTask(tasks.TypeUserInactivityWarning, nil),
		asynq.Queue("notifications"),
	)
	if err != nil {
		return fmt.Errorf("failed to register user inactivity warning task: %w", err)
	}

	_, err = scheduler.Register(
		"0 3 * * 0", // Sunday 03:00 AM
		asynq.NewTask(tasks.TypeWorkspaceDeletion, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register workspace deletion task: %w", err)
	}

	_, err = scheduler.Register(
		"0 4 * * 0", // Sunday 04:00 AM
		asynq.NewTask(tasks.TypeUserDeactivation, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register user deactivation task: %w", err)
	}

	_, err = scheduler.Register(
		"0 6 1 * *", // 1st of month at 6:00 AM
		asynq.NewTask(tasks.TypeDisableInactiveAutomation, nil),
		asynq.Queue("automation"),
	)
	if err != nil {
		return fmt.Errorf("failed to register disable inactive automation task: %w", err)
	}

	_, err = scheduler.Register(
		"0 2 * * 2", // Tuesday 2:00 AM (quiet day)
		asynq.NewTask(tasks.TypeChatSessionsCleanup, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register chat sessions cleanup task: %w", err)
	}

	_, err = scheduler.Register(
		"30 2 * * *", // Daily at 2:30 AM
		asynq.NewTask(tasks.TypeMessagingCleanup, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register messaging cleanup task: %w", err)
	}

	_, err = scheduler.Register(
		"15 2 * * *", // Daily at 2:15 AM
		asynq.NewTask(tasks.TypeSlackCredentialBackfill, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register Slack credential backfill task: %w", err)
	}

	_, err = scheduler.Register(
		"*/1 * * * *", // Every minute
		asynq.NewTask(tasks.TypeSlackInboxRecovery, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register Slack inbox recovery task: %w", err)
	}

	_, err = scheduler.Register(
		"*/1 * * * *", // Every minute
		asynq.NewTask(tasks.TypeBrevoEmailReplyRecovery, nil),
		asynq.Queue("cleanup"),
	)
	if err != nil {
		return fmt.Errorf("failed to register Brevo email reply inbox recovery task: %w", err)
	}

	return nil
}
