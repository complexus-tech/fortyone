package agentadapter

import (
	"encoding/json"
	"testing"

	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/agentmentions"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatMayaMentionRequiresCanonicalUnquotedNode(t *testing.T) {
	t.Parallel()
	mayaID := uuid.New()
	mention := chatContentNode{Type: "mention", Attrs: map[string]any{"id": mayaID.String(), "label": "Maya", "mentionKind": "agent", "agentKey": "maya"}}
	for _, test := range []struct {
		name  string
		node  chatContentNode
		valid bool
	}{
		{"canonical", chatContentNode{Type: "paragraph", Content: []chatContentNode{mention}}, true},
		{"literal text", chatContentNode{Type: "paragraph", Content: []chatContentNode{{Type: "text", Text: "@Maya create a task"}}}, false},
		{"quoted", chatContentNode{Type: "blockquote", Content: []chatContentNode{mention}}, false},
		{"code", chatContentNode{Type: "codeBlock", Content: []chatContentNode{mention}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			content, err := json.Marshal(chatContentNode{Type: "doc", Content: []chatContentNode{test.node}})
			require.NoError(t, err)
			err = validateChatMayaMention(content, []string{"maya"}, mayaID)
			if test.valid {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, agentmentions.ErrInvalidMention)
			}
			require.ErrorIs(t, validateChatMayaMention(content, []string{"maya"}, uuid.New()), agentmentions.ErrInvalidMention)
		})
	}
}

func TestChatReplyUsesSafeRichDocumentAndNoModelMentions(t *testing.T) {
	t.Parallel()
	plain, rich, err := chatReplyContent("**Decision:** ship the checklist.\n\n- [PROD-42](https://fortyone.app/acme/work/PROD-42)\n- Ask [@Someone](mention:" + uuid.NewString() + ")\n\n<script>alert(1)</script>")
	require.NoError(t, err)
	require.Contains(t, plain, "Decision:")
	require.NotContains(t, plain, "alert(1)")
	var root chatContentNode
	require.NoError(t, json.Unmarshal(rich, &root))
	require.Equal(t, "doc", root.Type)
	require.NotEmpty(t, root.Content)
	for _, node := range root.Content {
		require.NotEqual(t, "text", node.Type, "whitespace cannot create invalid top-level editor nodes")
	}
	require.Contains(t, string(rich), `"type":"bold"`)
	require.Contains(t, string(rich), `"type":"link"`)
	require.NotContains(t, string(rich), `"type":"mention"`)
	require.NotContains(t, string(rich), "mention:")
}

func TestChatSourceHashIgnoresJSONBFormatting(t *testing.T) {
	t.Parallel()
	first := chatdomain.Message{Text: "hello", RichContent: json.RawMessage(`{"type":"doc","content":[]}`)}
	second := chatdomain.Message{Text: "hello", RichContent: json.RawMessage(`{ "content": [], "type": "doc" }`)}
	require.Equal(t, chatSourceHash(first), chatSourceHash(second))
	second.Text = "changed"
	require.NotEqual(t, chatSourceHash(first), chatSourceHash(second))
}
