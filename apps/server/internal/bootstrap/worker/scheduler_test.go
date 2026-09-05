package workerbootstrap

import (
	"testing"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type scheduleCapture struct {
	entries []scheduledTask
}

type scheduledTask struct {
	spec     string
	taskType string
}

func TestGuidanceHasNoSeparateEmailSchedules(t *testing.T) {
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	for _, entry := range scheduler.entries {
		require.NotContains(t, []string{tasks.TypeMorningBriefing, tasks.TypeWeeklyDigestEmail, tasks.TypeOverdueStoriesEmail, tasks.TypeObjectiveOverdueEmail}, entry.taskType)
	}
}

func TestRegisterSchedulesDispatchesInvitationOutboxEveryMinute(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeInvitationOutboxDispatch,
	})
}

func TestRegisterSchedulesDispatchesOutboundWebhooksPromptly(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "@every 5s", taskType: tasks.TypeOutboundWebhookDispatch,
	})
}

func (s *scheduleCapture) Register(spec string, task *asynq.Task, _ ...asynq.Option) (string, error) {
	s.entries = append(s.entries, scheduledTask{spec: spec, taskType: task.Type()})
	return task.Type(), nil
}

func TestRegisterSchedulesRunsBrevoEmailReplyRecoveryEveryMinute(t *testing.T) {
	t.Parallel()

	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))

	require.Contains(t, scheduler.entries, scheduledTask{
		spec:     "*/1 * * * *",
		taskType: tasks.TypeBrevoEmailReplyRecovery,
	})
}

func TestRegisterSchedulesRecoversFigmaWebhooksEveryMinute(t *testing.T) {
	t.Parallel()

	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))

	require.Contains(t, scheduler.entries, scheduledTask{
		spec:     "*/1 * * * *",
		taskType: tasks.TypeFigmaWebhookRecovery,
	})
}

func TestRegisterSchedulesRecoversFeedbackDeliveryOutboxEveryMinute(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeFeedbackContributorDeliveryRecovery,
	})
}

func TestRegisterSchedulesDispatchesFeedbackOutboxEveryMinute(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeFeedbackOutboxDispatch,
	})
}

func TestRegisterSchedulesChecksCalendarWatchRenewalsHourly(t *testing.T) {
	t.Parallel()

	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec:     "17 * * * *",
		taskType: tasks.TypeCalendarWatchRenewal,
	})
}

func TestRegisterSchedulesDispatchesCalendarScheduleOutboxEveryMinute(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeCalendarScheduleOutbox,
	})
}

func TestRegisterSchedulesDispatchesGoogleDriveRevocationsEveryMinute(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeGoogleDriveRevocationDispatch,
	})
}

func TestRegisterSchedulesDispatchesStoryScheduleTransitionOutboxEveryMinute(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeStoryScheduleTransitionOutbox,
	})
}

func TestRegisterSchedulesRecoversMayaSchedulesEveryMinute(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeMayaScheduleRecovery,
	})
}

func TestRegisterSchedulesPurgesAPIIdempotencyReceiptsHourly(t *testing.T) {
	t.Parallel()
	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "7 * * * *", taskType: tasks.TypeAPIIdempotencyCleanup,
	})
}

func TestRegisterSchedulesDeletesRetainedAttachmentObjectsEveryMinute(t *testing.T) {
	t.Parallel()

	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))
	require.Contains(t, scheduler.entries, scheduledTask{
		spec: "*/1 * * * *", taskType: tasks.TypeAttachmentObjectDeletions,
	})
}

func TestRegisterSchedulesRunsRetentionPolicies(t *testing.T) {
	t.Parallel()

	scheduler := &scheduleCapture{}
	require.NoError(t, registerSchedules(scheduler))

	for _, expected := range []scheduledTask{
		{spec: "@weekly", taskType: tasks.TypeTokenCleanup},
		{spec: "0 2 * * 2", taskType: tasks.TypeChatSessionsCleanup},
		{spec: "0 3 * * 3", taskType: tasks.TypeWebhookCleanup},
		{spec: "0 4 * * 0", taskType: tasks.TypeUserDeactivation},
	} {
		require.Contains(t, scheduler.entries, expected)
	}
}
