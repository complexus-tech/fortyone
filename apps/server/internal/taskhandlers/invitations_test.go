package taskhandlers

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type invitationOutboxProcessorStub struct {
	called bool
	err    error
}

func (p *invitationOutboxProcessorStub) DispatchReady(context.Context) error {
	p.called = true
	return p.err
}

func TestHandleInvitationOutboxDispatch(t *testing.T) {
	t.Parallel()

	processor := &invitationOutboxProcessorStub{}
	handler := &handlers{invitationOutbox: processor}
	require.NoError(t, handler.HandleInvitationOutboxDispatch(context.Background(), asynq.NewTask("test", nil)))
	require.True(t, processor.called)

	processor.err = errors.New("database unavailable")
	err := handler.HandleInvitationOutboxDispatch(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "database unavailable")
}

func TestHandleInvitationOutboxDispatchRequiresProcessor(t *testing.T) {
	t.Parallel()

	err := (&handlers{}).HandleInvitationOutboxDispatch(context.Background(), asynq.NewTask("test", nil))
	require.ErrorContains(t, err, "unavailable")
}
