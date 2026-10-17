package publisher

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestChatNotificationsUseOnlyRecipientUpdateChannel(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	p := New(client, nil)
	now := time.Now().UTC()
	alert := ChatNotification{NotificationID: uuid.New(), EventID: uuid.New(), WorkspaceID: uuid.New(), UserID: uuid.New(), ActorID: uuid.New(), ConversationID: uuid.New(), MessageID: uuid.New(), CreatedAt: now, OccurredAt: now.Add(-time.Second), Title: "New chat message", Body: "There is a new message in a team conversation."}
	stream := client.Subscribe(t.Context(), UserUpdatesChannel(alert.UserID), "user-notifications:"+alert.UserID.String(), UserUpdatesChannel(uuid.New()))
	t.Cleanup(func() { _ = stream.Close() })
	for range 3 {
		_, err := stream.Receive(t.Context())
		require.NoError(t, err)
	}
	require.NoError(t, p.PublishChatNotification(t.Context(), alert))
	message, err := stream.ReceiveTimeout(t.Context(), time.Second)
	require.NoError(t, err)
	published, ok := message.(*redis.Message)
	require.True(t, ok)
	require.Equal(t, UserUpdatesChannel(alert.UserID), published.Channel)
	var received ChatNotification
	require.NoError(t, json.Unmarshal([]byte(published.Payload), &received))
	alert.Type = ChatNotificationType
	require.Equal(t, alert, received)
	require.Contains(t, published.Payload, `"threadRootId":null`)
	require.Empty(t, server.Keys(), "live alerts must not create another persistent event stream")
	require.Error(t, p.PublishChatNotification(t.Context(), ChatNotification{}))
	alert.ActorID = alert.UserID
	require.Error(t, p.PublishChatNotification(t.Context(), alert), "senders must never receive their own interruptions")
}
