package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type emailReplyProcessorStub struct {
	externalWorkspaceID string
	eventID             string
	err                 error
}

func (stub *emailReplyProcessorStub) ProcessEvent(_ context.Context, externalWorkspaceID, eventID string) error {
	stub.externalWorkspaceID = externalWorkspaceID
	stub.eventID = eventID
	return stub.err
}

type emailReplyRecoveryStub struct {
	count int
	err   error
}

func (stub *emailReplyRecoveryStub) RecoverPendingEvents(context.Context) (int, error) {
	return stub.count, stub.err
}

func TestHandleBrevoEmailReplyDelegatesIdentityOnlyPayload(t *testing.T) {
	processor := &emailReplyProcessorStub{}
	handler := &handlers{emailReplies: processor}
	payload, err := json.Marshal(tasks.BrevoEmailReplyPayload{
		ExternalWorkspaceID: "workspace:thread",
		EventID:             "message-1",
	})
	require.NoError(t, err)
	task := asynq.NewTask(tasks.TypeBrevoEmailReply, payload)

	require.NoError(t, handler.HandleBrevoEmailReply(context.Background(), task))
	require.Equal(t, "workspace:thread", processor.externalWorkspaceID)
	require.Equal(t, "message-1", processor.eventID)
}

func TestHandleBrevoEmailReplyRecoverySurfacesFailure(t *testing.T) {
	handler := &handlers{emailRecovery: &emailReplyRecoveryStub{err: errors.New("queue unavailable")}}
	err := handler.HandleBrevoEmailReplyRecovery(context.Background(), asynq.NewTask(tasks.TypeBrevoEmailReplyRecovery, nil))
	require.ErrorContains(t, err, "queue unavailable")
}
