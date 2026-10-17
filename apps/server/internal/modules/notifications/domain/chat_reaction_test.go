package notifications

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatReactionSourceIsStoredButNotExposed(t *testing.T) {
	t.Parallel()
	message := NotificationMessage{
		Template:     "{actor} reacted {emoji} to your message",
		ChatReaction: &ChatReaction{EventID: uuid.New(), Emoji: "👍"},
	}
	stored, err := json.Marshal(message)
	require.NoError(t, err)
	var decoded NotificationMessage
	require.NoError(t, json.Unmarshal(stored, &decoded))
	require.Equal(t, message, decoded)
	public := decoded.Public()
	require.Nil(t, public.ChatReaction)
	require.NotNil(t, decoded.ChatReaction)
	encoded, err := json.Marshal(public)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "chatReaction")
	require.Equal(t, message.Template, public.Template)
}

func TestChatReactionRequiresChatSourceIdentity(t *testing.T) {
	t.Parallel()
	base := NewNotification{
		RecipientID: uuid.New(), WorkspaceID: uuid.New(), EntityID: uuid.New(), ActorID: uuid.New(),
		Type: NotificationTypeChatMessage, EntityType: EntityTypeChatMessage, Title: "New reaction",
		Message: NotificationMessage{ChatReaction: &ChatReaction{EventID: uuid.New(), Emoji: "👍"}},
	}
	require.NoError(t, base.Validate())
	for _, test := range []struct {
		name   string
		modify func(*NewNotification)
	}{
		{"missing event", func(input *NewNotification) { input.Message.ChatReaction.EventID = uuid.Nil }},
		{"empty emoji", func(input *NewNotification) { input.Message.ChatReaction.Emoji = "  " }},
		{"non-chat type", func(input *NewNotification) {
			input.Type, input.EntityType = NotificationTypeStoryUpdate, EntityTypeStory
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := base
			reaction := *base.Message.ChatReaction
			input.Message.ChatReaction = &reaction
			test.modify(&input)
			require.ErrorIs(t, input.Validate(), ErrInvalid)
		})
	}
}
