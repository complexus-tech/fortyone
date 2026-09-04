package workerbootstrap

import (
	"testing"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestBuildTaskMuxRegistersAttachmentImageOptimization(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeAttachmentImageOptimization, nil))

	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeAttachmentImageOptimization, pattern)
}

func TestBuildTaskMuxRegistersBrevoEmailReplyTasks(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	for _, taskType := range []string{tasks.TypeBrevoEmailReply, tasks.TypeBrevoEmailReplyRecovery} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersFigmaWebhookTasks(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	for _, taskType := range []string{tasks.TypeFigmaWebhook, tasks.TypeFigmaWebhookRecovery} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersFeedbackDeliveryAndRecovery(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})
	for _, taskType := range []string{tasks.TypeFeedbackContributorDelivery, tasks.TypeFeedbackContributorDeliveryRecovery, tasks.TypeFeedbackOutboxDispatch} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersMayaScheduleRecovery(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeMayaScheduleRecovery, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeMayaScheduleRecovery, pattern)
}

func TestBuildTaskMuxRegistersStoryScheduleTransitionOutbox(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeStoryScheduleTransitionOutbox, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeStoryScheduleTransitionOutbox, pattern)
}

func TestBuildTaskMuxRegistersGoogleDriveRevocationDispatch(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeGoogleDriveRevocationDispatch, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeGoogleDriveRevocationDispatch, pattern)
}

func TestBuildTaskMuxRegistersCalendarWorkspaceScheduleBatch(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeCalendarWorkspaceScheduleBatch, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeCalendarWorkspaceScheduleBatch, pattern)
}

func TestBuildTaskMuxRegistersInvitationOutbox(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeInvitationOutboxDispatch, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeInvitationOutboxDispatch, pattern)
}

func TestBuildTaskMuxRegistersSprintAutomation(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeSprintAutoCreation, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeSprintAutoCreation, pattern)
}

func TestBuildTaskMuxRegistersAPIIdempotencyReceiptCleanup(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeAPIIdempotencyCleanup, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeAPIIdempotencyCleanup, pattern)
}

func TestBuildTaskMuxRegistersRetentionHandlers(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	for _, taskType := range []string{
		tasks.TypeTokenCleanup,
		tasks.TypeChatSessionsCleanup,
		tasks.TypeWebhookCleanup,
		tasks.TypeMessagingCleanup,
		tasks.TypeDeleteFeedback,
	} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersGuidanceHandlers(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	for _, taskType := range []string{
		tasks.TypeOverdueStoriesEmail,
		tasks.TypeObjectiveOverdueEmail,
		tasks.TypeWeeklyDigestEmail,
		tasks.TypeFeedbackDigestEmail,
	} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersUserLifecycleHandlers(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	for _, taskType := range []string{
		tasks.TypeUserInactivityWarning,
		tasks.TypeUserDeactivation,
	} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersStrategyHandlers(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeStrategyCommunications, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeStrategyCommunications, pattern)
}

func TestBuildTaskMuxRegistersWorkspaceLifecycleHandlers(t *testing.T) {
	mux := buildTaskMux(taskMuxDependencies{})

	for _, taskType := range []string{
		tasks.TypeWorkspaceInactivityWarning,
		tasks.TypeWorkspaceDeletion,
		tasks.TypeWorkspaceCleanup,
	} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}
