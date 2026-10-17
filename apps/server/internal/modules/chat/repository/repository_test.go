package chatrepository

import (
	"encoding/json"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreationHashCanonicalizesJSONAndIDSetOrder(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	command := domain.CreateMessageCommand{Text: "Original", RichContent: json.RawMessage(`{"type":"doc","attrs":{"b":2,"a":1}}`), MentionedUserIDs: []uuid.UUID{a, b}}
	first, err := creationHash(command)
	require.NoError(t, err)
	command.RichContent = json.RawMessage(`{ "attrs": {"a":1,"b":2}, "type":"doc" }`)
	command.MentionedUserIDs = []uuid.UUID{b, a}
	second, err := creationHash(command)
	require.NoError(t, err)
	require.Equal(t, first, second)
	command.Text = "Different"
	third, err := creationHash(command)
	require.NoError(t, err)
	require.NotEqual(t, first, third)
}

func TestMessageCursorCannotCrossConversationOrThread(t *testing.T) {
	conversation, root := uuid.New(), uuid.New()
	encoded := encodeCursor(cursor{ConversationID: conversation, ThreadRootID: &root, Sequence: 42})
	sequence, err := sequenceCursor(encoded, conversation, &root)
	require.NoError(t, err)
	require.EqualValues(t, 42, sequence)
	_, err = sequenceCursor(encoded, uuid.New(), &root)
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = sequenceCursor(encoded, conversation, nil)
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = sequenceCursor("not-a-cursor", conversation, &root)
	require.ErrorIs(t, err, domain.ErrInvalid)
}

func TestLinkResourcesNeverFetchOrAcceptActiveSchemes(t *testing.T) {
	links := extractLinks("Ignored rich-text mirror", json.RawMessage(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"See https://fortyone.app and https://fortyone.app."},{"type":"text","text":"Unsafe link","marks":[{"type":"link","attrs":{"href":"javascript:alert(1)"}}]},{"type":"text","text":"Documentation","marks":[{"type":"link","attrs":{"href":"https://fortyone.app/docs"}}]}]}]}`))
	require.Equal(t, []string{"https://fortyone.app/", "https://fortyone.app/docs"}, links)
}
