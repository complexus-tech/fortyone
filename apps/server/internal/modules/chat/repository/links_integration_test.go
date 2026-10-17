//go:build integration

package chatrepository

import (
	"encoding/json"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatLinkResourcesTrackVisibleRichContentAcrossEdits(t *testing.T) {
	f := newFixture(t)
	message, err := f.service.CreateMessage(f.ctx, domain.CreateMessageCommand{
		Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(),
		Text:        "https://fortyone.app/code https://fortyone.app/label maya@example.com",
		RichContent: json.RawMessage(`{"type":"doc","content":[{"type":"codeBlock","content":[{"type":"text","text":"https://fortyone.app/code"}]},{"type":"paragraph","content":[{"type":"text","text":"https://fortyone.app/label","marks":[{"type":"link","attrs":{"href":"https://fortyone.app/destination"}}]},{"type":"text","text":" maya@example.com"}]}]}`),
	})
	require.NoError(t, err)
	query := domain.SearchQuery{Scope: f.scope, ConversationID: f.room, Kind: "links"}
	resources, err := f.service.Resources(f.ctx, query)
	require.NoError(t, err)
	urls := make([]string, 0, len(resources.Items))
	for _, item := range resources.Items {
		require.NotNil(t, item.URL)
		urls = append(urls, *item.URL)
	}
	require.ElementsMatch(t, []string{"https://fortyone.app/destination", "mailto:maya@example.com"}, urls)
	_, err = f.service.UpdateMessage(f.ctx, domain.UpdateMessageCommand{Scope: f.scope, MessageID: message.ID, Version: message.Version, Text: "www.fortyone.app/updated"})
	require.NoError(t, err)
	resources, err = f.service.Resources(f.ctx, query)
	require.NoError(t, err)
	require.Len(t, resources.Items, 1)
	require.Equal(t, "https://www.fortyone.app/updated", *resources.Items[0].URL)
}
