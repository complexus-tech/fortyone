package chatadapter

import (
	"errors"
	"testing"
	"time"

	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestReactionDeliveryUsesExistingReceiptAndChatAlert(t *testing.T) {
	author, reactor, thread := uuid.New(), uuid.New(), uuid.New()
	event := chat.Event{
		ID: uuid.New(), WorkspaceID: uuid.New(), ConversationID: uuid.New(),
		MessageID: uuid.New(), ActorID: reactor, ThreadRootID: &thread,
		Type: "reaction.updated", CreatedAt: time.Now().UTC(),
	}
	store := &eventStoreStub{
		targets: []chat.NotificationRecipient{
			{UserID: author, Reason: "reaction"},
			{UserID: reactor, Reason: "reaction"},
			{UserID: author, Reason: "reaction"},
		},
		content: chat.NotificationContent{
			ConversationKind: "team", ConversationName: "Product", AuthorName: "Joseph",
			ActorName: "Sarah", ReactionEmoji: "🙌🏾", Text: "Ready to ship",
		},
	}
	delivery := &deliveryStub{alertErr: errors.New("redis unavailable")}
	dispatcher := Dispatcher{Store: store, Updates: delivery, Notifications: delivery}
	require.Error(t, dispatcher.deliver(t.Context(), event))
	require.Len(t, delivery.notices, 1, "only the author gets a receipt")
	notice := delivery.notices[0]
	require.NoError(t, notice.Validate())
	require.Equal(t, author, notice.RecipientID)
	require.Equal(t, reactor, notice.ActorID)
	require.True(t, notice.Message.EmailSuppressed)
	require.NotNil(t, notice.Message.ChatReaction)
	require.Equal(t, event.ID, notice.Message.ChatReaction.EventID)
	require.Equal(t, "🙌🏾", notice.Message.ChatReaction.Emoji)
	require.Equal(t, "Sarah reacted 🙌🏾 to your message: Ready to ship", notice.Message.Variables["message"].Value)

	delivery.alertErr = nil
	require.NoError(t, dispatcher.deliver(t.Context(), event))
	require.Len(t, delivery.receipts, 1, "an SSE retry must reuse the persisted receipt")
	require.Len(t, delivery.alerts, 1)
	alert := delivery.alerts[0]
	require.Equal(t, author, alert.UserID)
	require.Equal(t, event.ID, alert.EventID)
	require.Equal(t, event.MessageID, alert.MessageID)
	require.Equal(t, event.ThreadRootID, alert.ThreadRootID)
	require.Equal(t, notice.Message.Variables["message"].Value, alert.Body)
}

func TestLegacyAndRemovalReactionEventsCannotCreateAlerts(t *testing.T) {
	store := &eventStoreStub{
		targets: []chat.NotificationRecipient{{UserID: uuid.New(), Reason: "reaction"}},
		content: chat.NotificationContent{ActorName: "Sarah"},
	}
	delivery := &deliveryStub{}
	err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).deliver(t.Context(), chat.Event{Type: "reaction.updated"})
	require.NoError(t, err)
	require.Empty(t, delivery.notices)
	require.Empty(t, delivery.alerts)
}
