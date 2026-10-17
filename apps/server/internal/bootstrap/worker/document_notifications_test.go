package workerbootstrap

import (
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDocumentNotificationsHaveDurableFrequentRecovery(t *testing.T) {
	scheduler := &scheduleCapture{}
	mux := asynq.NewServeMux()
	require.NoError(t, registerDocumentNotificationDelivery(mux, scheduler, nil))
	require.Contains(t, scheduler.entries, scheduledTask{spec: "@every 5s", taskType: documentNotificationDeliveryTask})
	_, pattern := mux.Handler(asynq.NewTask(documentNotificationDeliveryTask, nil))
	require.Equal(t, documentNotificationDeliveryTask, pattern)
}
