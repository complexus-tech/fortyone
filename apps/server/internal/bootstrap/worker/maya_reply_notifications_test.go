package workerbootstrap

import (
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestMayaReplyNotificationsHaveDurableFrequentRecovery(t *testing.T) {
	scheduler := &scheduleCapture{}
	mux := asynq.NewServeMux()
	require.NoError(t, registerMayaReplyNotificationDelivery(mux, scheduler, nil))
	require.Contains(t, scheduler.entries, scheduledTask{spec: "@every 5s", taskType: mayaReplyNotificationTask})
	_, pattern := mux.Handler(asynq.NewTask(mayaReplyNotificationTask, nil))
	require.Equal(t, mayaReplyNotificationTask, pattern)
}
