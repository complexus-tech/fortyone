package workerbootstrap

import (
	"testing"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestBuildTaskMuxRegistersAttachmentImageOptimization(t *testing.T) {
	mux := buildTaskMux(nil, nil, nil, nil, nil, nil, nil, nil, nil, uuid.Nil)

	handler, pattern := mux.Handler(asynq.NewTask(tasks.TypeAttachmentImageOptimization, nil))

	require.NotNil(t, handler)
	require.Equal(t, tasks.TypeAttachmentImageOptimization, pattern)
}
