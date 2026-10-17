package notifications

import (
	"context"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/stretchr/testify/require"
)

type documentDispatchStub struct {
	repositoryStub
	notification domain.Notification
	completed    bool
}

func (s *documentDispatchStub) DispatchDocumentNotifications(ctx context.Context, _ int, deliver func(context.Context, domain.Notification) error) (int, error) {
	if err := deliver(ctx, s.notification); err != nil {
		return 0, err
	}
	s.completed = true
	return 1, nil
}
func TestDocumentDeliveryUsesNormalChannelsAndRetainsFailedWakeups(t *testing.T) {
	_, notification := validNotification()
	notification.EntityType = domain.EntityTypeDocument
	notification.Type = domain.NotificationTypeMention
	repo := &documentDispatchStub{notification: notification}
	queue := &taskStub{pushErr: errors.New("queue unavailable")}
	service := newTestService(repo, queue)
	published := 0
	service.publishRealtime = func(context.Context, CoreNotification) error { published++; return nil }
	_, err := service.DispatchDocumentNotifications(t.Context(), 25)
	require.Error(t, err)
	require.False(t, repo.completed)
	require.Empty(t, queue.payloads)
	queue.pushErr = nil
	count, err := service.DispatchDocumentNotifications(t.Context(), 25)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.True(t, repo.completed)
	require.Equal(t, 2, published)
	require.Len(t, queue.payloads, 1)
	require.Equal(t, notification.RecipientID, queue.payloads[0].RecipientID)
	require.Equal(t, notification.ID, queue.pushPayloads[1].NotificationID)
}
func TestDocumentEmailOnlyDeliveryDoesNotPublishOrPush(t *testing.T) {
	_, notification := validNotification()
	notification.InAppEnabled = false
	repo := &documentDispatchStub{notification: notification}
	queue := &taskStub{}
	service := newTestService(repo, queue)
	service.publishRealtime = func(context.Context, CoreNotification) error {
		t.Fatal("email-only notification published")
		return nil
	}
	_, err := service.DispatchDocumentNotifications(t.Context(), 25)
	require.NoError(t, err)
	require.Empty(t, queue.pushPayloads)
	require.Len(t, queue.payloads, 1)
}
