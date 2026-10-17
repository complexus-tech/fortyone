package chatadapter

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"

	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNotificationCopyIdentifiesSourceAndActualMessage(t *testing.T) {
	for _, test := range []struct {
		name    string
		content chat.NotificationContent
		title   string
		body    string
	}{
		{"direct", chat.NotificationContent{ConversationKind: "dm", AuthorName: "Sarah Jones", Text: "Can you review this?"}, "Sarah Jones", "Can you review this?"},
		{"team", chat.NotificationContent{ConversationKind: "team", ConversationName: "Product", AuthorName: "Sarah Jones", Text: "Ready for review"}, "Product", "Sarah Jones: Ready for review"},
		{"plain syntax", chat.NotificationContent{AuthorName: "Sarah", Text: "Use <input> and {message} literally"}, "Sarah", "Use <input> and {message} literally"},
		{"rich mention", chat.NotificationContent{AuthorName: "Sarah", Text: "stale plain mirror", RichContent: json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"id":"private-id","label":"Joseph"}},{"type":"text","text":", please review","marks":[{"type":"bold"}]}]},{"type":"paragraph","content":[{"type":"text","text":"the launch plan","marks":[{"type":"link","attrs":{"href":"https://example.test"}}]}]}]}`)}, "Sarah", "@Joseph, please review the launch plan"},
		{"voice", chat.NotificationContent{AuthorName: "Sarah", Attachments: []chat.Attachment{{Kind: "voice-note", FileName: "recording.webm"}}}, "Sarah", "Sent a voice message"},
		{"image", chat.NotificationContent{AuthorName: "Sarah", Attachments: []chat.Attachment{{MediaType: "image/png", FileName: "shot.png"}}}, "Sarah", "Sent an image"},
		{"file", chat.NotificationContent{AuthorName: "Sarah", Attachments: []chat.Attachment{{FileName: "Launch plan.pdf"}}}, "Sarah", "Shared Launch plan.pdf"},
		{"files", chat.NotificationContent{AuthorName: "Sarah", Attachments: []chat.Attachment{{}, {}}}, "Sarah", "Sent 2 attachments"},
		{"caption", chat.NotificationContent{AuthorName: "Sarah", Text: "Here is the design", Attachments: []chat.Attachment{{FileName: "design.png"}}}, "Sarah", "Here is the design"},
		{"direct reaction", chat.NotificationContent{ConversationKind: "dm", AuthorName: "Joseph", ActorName: "Sarah", ReactionEmoji: "👍", Text: "Ready to ship"}, "Sarah", "Sarah reacted 👍 to your message: Ready to ship"},
		{"team reaction", chat.NotificationContent{ConversationKind: "team", ConversationName: "Product", AuthorName: "Joseph", ActorName: "Sarah", ReactionEmoji: "🙌🏾", Text: "Ready to ship"}, "Product", "Sarah reacted 🙌🏾 to your message: Ready to ship"},
		{"attachment reaction", chat.NotificationContent{ConversationKind: "dm", ActorName: "Sarah", ReactionEmoji: "✅", Attachments: []chat.Attachment{{FileName: "design.png"}}}, "Sarah", "Sarah reacted ✅ to your message"},
		{"missing reactor", chat.NotificationContent{ReactionEmoji: "👀", Text: "Launch"}, "Someone", "Someone reacted 👀 to your message: Launch"},
	} {
		t.Run(test.name, func(t *testing.T) {
			title, body := notificationCopy(test.content)
			require.Equal(t, test.title, title)
			require.Equal(t, test.body, body)
		})
	}
}

func TestNotificationCopyBoundsUTF8WithoutLosingThePreview(t *testing.T) {
	title, body := notificationCopy(chat.NotificationContent{
		ConversationKind: "team", ConversationName: strings.Repeat("界", 100),
		AuthorName: "Sarah", Text: strings.Repeat("👋", 400),
	})
	require.LessOrEqual(t, len(title), 256)
	require.LessOrEqual(t, len(body), 1024)
	require.True(t, utf8.ValidString(title))
	require.True(t, utf8.ValidString(body))
	require.True(t, strings.HasSuffix(title, "…"))
	require.True(t, strings.HasSuffix(body, "…"))
	require.True(t, strings.HasPrefix(body, "Sarah: 👋"))
}

func TestNotificationCopySkipsUnavailableMessageContent(t *testing.T) {
	for _, lookupErr := range []error{chat.ErrNotFound, chat.ErrConflict} {
		store := &eventStoreStub{
			targets:    []chat.NotificationRecipient{{UserID: uuid.New(), Reason: "message"}},
			contentErr: lookupErr,
		}
		delivery := &deliveryStub{}
		err := (Dispatcher{Store: store, Updates: delivery, Notifications: delivery}).deliver(t.Context(), chat.Event{})
		if lookupErr == chat.ErrNotFound {
			require.NoError(t, err)
		} else {
			require.ErrorIs(t, err, lookupErr)
		}
		require.Empty(t, delivery.notices)
		require.Empty(t, delivery.alerts)
	}
}
