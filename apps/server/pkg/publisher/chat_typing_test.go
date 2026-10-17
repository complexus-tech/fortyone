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

func TestChatTypingPublishesOnlyToRecipientLiveChannel(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	p := New(client, nil)
	activity := ChatTyping{WorkspaceID: uuid.New(), UserID: uuid.New(), ConversationID: uuid.New(), ActorID: uuid.New(), FullName: "Joseph", IsTyping: true, ExpiresAt: time.Now().UTC().Add(6 * time.Second)}
	stream := client.Subscribe(t.Context(), UserUpdatesChannel(activity.UserID))
	t.Cleanup(func() { _ = stream.Close() })
	_, err := stream.Receive(t.Context())
	require.NoError(t, err)
	require.NoError(t, p.PublishChatTyping(t.Context(), activity))
	message, err := stream.ReceiveTimeout(t.Context(), time.Second)
	require.NoError(t, err)
	published, ok := message.(*redis.Message)
	require.True(t, ok)
	var received ChatTyping
	require.NoError(t, json.Unmarshal([]byte(published.Payload), &received))
	activity.Type = ChatTypingType
	require.Equal(t, activity, received)
	require.Contains(t, published.Payload, `"threadRootId":null`)
	require.Empty(t, server.Keys(), "typing creates no stream or history records")
	require.Error(t, p.PublishChatTyping(t.Context(), ChatTyping{}))
}
