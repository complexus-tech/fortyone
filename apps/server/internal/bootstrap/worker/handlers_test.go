package workerbootstrap

import (
	"testing"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestBuildTaskMuxRegistersAttachmentImageOptimization(t *testing.T) {
	mux := buildTaskMux(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, uuid.Nil, nil, nil, nil, "")

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeAttachmentImageOptimization, nil))

	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeAttachmentImageOptimization, pattern)
}

func TestBuildTaskMuxRegistersBrevoEmailReplyTasks(t *testing.T) {
	mux := buildTaskMux(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, uuid.Nil, nil, nil, nil, "")

	for _, taskType := range []string{tasks.TypeBrevoEmailReply, tasks.TypeBrevoEmailReplyRecovery} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersFeedbackDeliveryAndRecovery(t *testing.T) {
	mux := buildTaskMux(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, uuid.Nil, nil, nil, nil, "")
	for _, taskType := range []string{tasks.TypeFeedbackContributorDelivery, tasks.TypeFeedbackContributorDeliveryRecovery, tasks.TypeFeedbackOutboxDispatch} {
		handler, pattern := mux.Handler(asynq.NewTask(taskType, nil))
		require.NotNil(t, handler)
		require.Equal(t, taskType, pattern)
	}
}

func TestBuildTaskMuxRegistersMayaScheduleRecovery(t *testing.T) {
	mux := buildTaskMux(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, uuid.Nil, nil, nil, nil, "")

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeMayaScheduleRecovery, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeMayaScheduleRecovery, pattern)
}

func TestBuildTaskMuxRegistersStoryScheduleTransitionOutbox(t *testing.T) {
	mux := buildTaskMux(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, uuid.Nil, nil, nil, nil, "")

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeStoryScheduleTransitionOutbox, nil))
	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeStoryScheduleTransitionOutbox, pattern)
}
