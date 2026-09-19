package notifications

import (
	"context"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type pushRepositoryStub struct {
	repositoryStub
	tokens []string
}

func (*pushRepositoryStub) RegisterPushDevice(context.Context, notificationsdomain.RegisterPushDevice) (notificationsdomain.PushDevice, error) {
	return notificationsdomain.PushDevice{}, nil
}

func (*pushRepositoryStub) UnregisterPushDevice(context.Context, uuid.UUID, string) error {
	return nil
}

func (stub *pushRepositoryStub) ListPushTokens(context.Context, uuid.UUID) ([]string, error) {
	return stub.tokens, nil
}

func (*pushRepositoryStub) GetPushDelivery(context.Context, uuid.UUID) (*notificationsdomain.PushDelivery, error) {
	return nil, nil
}

func (*pushRepositoryStub) MarkPushSent(context.Context, uuid.UUID, time.Time) error {
	return nil
}

func (*pushRepositoryStub) DisablePushDevices(context.Context, []string, time.Time) error {
	return nil
}

func TestSendTestPushQueuesAuthenticatedRecipient(t *testing.T) {
	t.Parallel()
	recipientID := uuid.New()
	repository := &pushRepositoryStub{tokens: []string{"ExponentPushToken[device]"}}
	queue := &taskStub{}
	service := newTestService(repository, queue)

	err := service.SendTestPush(context.Background(), recipientID)

	require.NoError(t, err)
	require.Len(t, queue.testPushes, 1)
	require.Equal(t, recipientID, queue.testPushes[0].RecipientID)
}

func TestSendTestPushRequiresRegisteredDevice(t *testing.T) {
	t.Parallel()
	service := newTestService(&pushRepositoryStub{}, &taskStub{})

	err := service.SendTestPush(context.Background(), uuid.New())

	require.ErrorIs(t, err, ErrConflict)
}
