package taskhandlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type storyScheduleTransitionOutboxProcessorStub struct {
	called bool
	err    error
}

func (p *storyScheduleTransitionOutboxProcessorStub) DispatchReadyScheduleTransitionOutbox(context.Context) (int, error) {
	p.called = true
	return 1, p.err
}

func TestHandleStoryScheduleTransitionOutboxDispatchDelegatesToProcessor(t *testing.T) {
	t.Parallel()

	processor := &storyScheduleTransitionOutboxProcessorStub{}
	handler := &handlers{storyScheduleOutbox: processor}

	require.NoError(t, handler.HandleStoryScheduleTransitionOutboxDispatch(context.Background(), asynq.NewTask("test", nil)))
	require.True(t, processor.called)
}

func TestHandleStoryScheduleTransitionOutboxDispatchReturnsProcessorError(t *testing.T) {
	t.Parallel()

	processor := &storyScheduleTransitionOutboxProcessorStub{err: errors.New("database unavailable")}
	handler := &handlers{storyScheduleOutbox: processor}

	err := handler.HandleStoryScheduleTransitionOutboxDispatch(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "database unavailable")
}

func TestHandleStoryScheduleTransitionOutboxDispatchRequiresProcessor(t *testing.T) {
	t.Parallel()

	err := (&handlers{}).HandleStoryScheduleTransitionOutboxDispatch(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "is not configured")
}
