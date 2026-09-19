package taskhandlers

import (
	"context"
	"encoding/json"
	"testing"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/expopush"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type pushDeliveryStoreStub struct {
	delivery *notificationsdomain.PushDelivery
	disabled []string
	marked   uuid.UUID
}

func (store *pushDeliveryStoreStub) GetPushDelivery(context.Context, uuid.UUID) (*notificationsdomain.PushDelivery, error) {
	return store.delivery, nil
}

func (store *pushDeliveryStoreStub) ListPushTokens(context.Context, uuid.UUID) ([]string, error) {
	if store.delivery == nil {
		return nil, nil
	}
	return store.delivery.Tokens, nil
}

func (store *pushDeliveryStoreStub) DisablePushDevices(_ context.Context, tokens []string) error {
	store.disabled = append([]string(nil), tokens...)
	return nil
}

func (store *pushDeliveryStoreStub) MarkPushSent(_ context.Context, notificationID uuid.UUID) error {
	store.marked = notificationID
	return nil
}

type pushSenderStub struct {
	messages []expopush.Message
	result   expopush.Result
}

func (sender *pushSenderStub) Send(_ context.Context, messages []expopush.Message) (expopush.Result, error) {
	sender.messages = append(sender.messages, messages...)
	return sender.result, nil
}

func TestNotificationPushUsesPublicNavigationMetadataAndDisablesDeadTokens(t *testing.T) {
	t.Parallel()
	notificationID, recipientID, workspaceID, entityID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	invalidToken := "ExponentPushToken[invalid-device]"
	store := &pushDeliveryStoreStub{delivery: &notificationsdomain.PushDelivery{
		NotificationID: notificationID, RecipientID: recipientID, WorkspaceID: workspaceID,
		WorkspaceSlug: "fortyone", EntityType: notificationsdomain.EntityTypeStory, EntityID: entityID,
		Title: "Story updated", Message: json.RawMessage(`{"template":"{actor} updated it","variables":{"actor":{"value":"Maya","type":"actor"}}}`),
		Tokens: []string{"ExponentPushToken[active-device]", invalidToken},
	}}
	sender := &pushSenderStub{result: expopush.Result{InvalidTokens: []string{invalidToken}}}
	handler := &handlers{pushDeliveries: store, pushSender: sender}
	payload, err := json.Marshal(tasks.NotificationPushPayload{NotificationID: notificationID})
	require.NoError(t, err)

	err = handler.HandleNotificationPush(context.Background(), asynq.NewTask(tasks.TypeNotificationPush, payload))

	require.NoError(t, err)
	require.Equal(t, notificationID, store.marked)
	require.Equal(t, []string{invalidToken}, store.disabled)
	require.Len(t, sender.messages, 2)
	require.Equal(t, "Maya updated it", sender.messages[0].Body)
	require.Equal(t, recipientID.String(), sender.messages[0].Data["recipientId"])
	require.Equal(t, "fortyone", sender.messages[0].Data["workspaceSlug"])
	require.Equal(t, entityID.String(), sender.messages[0].Data["entityId"])
}

func TestNotificationPushTestUsesRegisteredDevices(t *testing.T) {
	t.Parallel()
	recipientID := uuid.New()
	store := &pushDeliveryStoreStub{delivery: &notificationsdomain.PushDelivery{
		Tokens: []string{"ExponentPushToken[test-device]"},
	}}
	sender := &pushSenderStub{}
	handler := &handlers{pushDeliveries: store, pushSender: sender}
	payload, err := json.Marshal(tasks.NotificationPushTestPayload{RecipientID: recipientID})
	require.NoError(t, err)

	err = handler.HandleNotificationPushTest(context.Background(), asynq.NewTask(tasks.TypeNotificationPushTest, payload))

	require.NoError(t, err)
	require.Len(t, sender.messages, 1)
	require.Equal(t, "FortyOne", sender.messages[0].Title)
	require.Equal(t, "Push notifications are working.", sender.messages[0].Body)
	require.Equal(t, "test", sender.messages[0].Data["kind"])
	require.Equal(t, recipientID.String(), sender.messages[0].Data["recipientId"])
}
