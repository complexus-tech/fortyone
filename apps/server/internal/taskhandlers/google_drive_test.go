package taskhandlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type googleDriveRevocationProcessorStub struct {
	calls int
	err   error
}

func (processor *googleDriveRevocationProcessorStub) DispatchPendingRevocations(context.Context) (int, error) {
	processor.calls++
	return 1, processor.err
}

func TestHandleGoogleDriveRevocationDispatchDelegatesToProcessor(t *testing.T) {
	t.Parallel()

	processor := &googleDriveRevocationProcessorStub{}
	handler := NewWorkerHandlers(WorkerHandlerDependencies{GoogleDriveRevocations: processor})

	require.NoError(t, handler.HandleGoogleDriveRevocationDispatch(t.Context(), asynq.NewTask("test", nil)))
	require.Equal(t, 1, processor.calls)
}

func TestHandleGoogleDriveRevocationDispatchReturnsProcessorError(t *testing.T) {
	t.Parallel()

	processorErr := errors.New("database unavailable")
	handler := NewWorkerHandlers(WorkerHandlerDependencies{
		GoogleDriveRevocations: &googleDriveRevocationProcessorStub{err: processorErr},
	})

	err := handler.HandleGoogleDriveRevocationDispatch(t.Context(), asynq.NewTask("test", nil))

	require.ErrorIs(t, err, processorErr)
}

func TestHandleGoogleDriveRevocationDispatchRequiresProcessor(t *testing.T) {
	t.Parallel()

	err := (&handlers{}).HandleGoogleDriveRevocationDispatch(t.Context(), asynq.NewTask("test", nil))

	require.ErrorContains(t, err, "not configured")
}
