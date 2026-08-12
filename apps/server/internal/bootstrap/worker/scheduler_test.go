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
