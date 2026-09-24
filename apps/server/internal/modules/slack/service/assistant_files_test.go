package slack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/stretchr/testify/require"
)

func TestAssistantAvailableFilesUsesCurrentMessageAndThreadWithoutExternalFiles(t *testing.T) {
	t.Parallel()
	event := normalizedSlackEvent{
		TeamID: "T1", ChannelID: "C1", ThreadTS: "171.100", MessageTS: "171.103",
		Files: []slackMessageFile{{ID: "F3", Name: "current.png"}},
	}
	files := assistantAvailableFiles(event, []slackMessageFile{
		{ID: "F1", Name: "earlier.pdf", MessageTS: "171.101"},
		{ID: "F2", Name: "external", IsExternal: true, MessageTS: "171.102"},
		{ID: "F3", Name: "duplicate.png", MessageTS: "171.103"},
	})
	require.Len(t, files, 2)
	require.Equal(t, "F3", files[0].FileID)
	require.Equal(t, "171.103", files[0].MessageTS)
	require.Equal(t, "F1", files[1].FileID)
	require.Equal(t, "171.101", files[1].MessageTS)
	for _, file := range files {
		require.Equal(t, "T1", file.ExternalWorkspaceID)
		require.Equal(t, "C1", file.ChannelID)
		require.Equal(t, "171.100", file.ThreadTS)
	}
	turn, err := assistantFileReferenceTurn(files)
	require.NoError(t, err)
	require.Contains(t, turn.Text, `"file_id":"F1"`)
	require.NotContains(t, turn.Text, "https://")
	require.NotContains(t, turn.Text, "external")
}

func TestSlackFileLanguageRequestsThreadHydration(t *testing.T) {
	t.Parallel()
	for _, prompt := range []string{"Attach the design file", "Upload that document to WEB-7", "Use the photo above"} {
		require.True(t, slackPromptRequestsThreadContext(prompt), prompt)
	}
	require.False(t, slackPromptRequestsThreadContext(strings.Repeat("status ", 2)))
}

func TestSlackThreadReferenceIncludesFileOnlyMessagesEvenWhenTextWasPersisted(t *testing.T) {
	t.Parallel()
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/conversations.replies", request.URL.Path)
		_, _ = w.Write([]byte(`{"ok":true,"messages":[
			{"ts":"171.100","user":"U1","text":"Start"},
			{"ts":"171.101","thread_ts":"171.100","user":"U2","text":"","subtype":"file_share","files":[{"id":"F1","name":"draft.pdf","mode":"hosted"}]}
		]}`))
	}))
	defer provider.Close()
	client := newSlackWebClient(provider.Client())
	client.baseURL = provider.URL
	reference, err := (&EventProcessor{webClient: client}).loadSlackThreadReference(
		context.Background(), "xoxb-token",
		slackrepository.SlackWorkspaceRecord{SlackTeamDomain: "acme"},
		normalizedSlackEvent{ChannelID: "C1", ThreadTS: "171.100", MessageTS: "171.102"},
		map[string]struct{}{"171.101": {}},
	)
	require.NoError(t, err)
	require.Len(t, reference.Files, 1)
	require.Equal(t, "F1", reference.Files[0].ID)
	require.Equal(t, "171.101", reference.Files[0].MessageTS)
	require.NotContains(t, reference.Turn.Text, `"timestamp":"171.101"`)
}
