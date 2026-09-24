package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fileImportQueueStub struct {
	intents []SlackFileImportIntent
}

func (s *fileImportQueueStub) QueueSlackFileImport(_ context.Context, intent SlackFileImportIntent) error {
	s.intents = append(s.intents, intent)
	return nil
}

func TestConfirmedMayaFileAttachmentQueuesScopedImport(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	actorID := uuid.New()
	teamID := uuid.New()
	storyID := uuid.New()
	installationID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID: installationID, WorkspaceID: workspaceID, SlackTeamID: "T1",
			InstallGeneration: installGeneration, BotAccessToken: "xoxb-token", IsActive: true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
	}
	confirmer := &mutationConfirmerStub{result: StoryMutationResult{
		Status: "applied", Operation: StoryMutationAttachFile,
		StoryID: storyID, TeamID: teamID, Reference: "WEB-7", Title: "Design",
		Attachment: &StoryAttachmentSource{
			Provider: ProviderSlack, ExternalWorkspaceID: "T1", ChannelID: "C1",
			ThreadTS: "171.100", MessageTS: "171.101", FileID: "F1", Name: "design.png",
		},
	}}
	var providerRequest map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/chat.update", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&providerRequest))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	WithMutationConfirmer(confirmer)(service)
	queue := &fileImportQueueStub{}
	WithSlackFileImportQueue(queue)(service)
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions","team":{"id":"T1"},"user":{"id":"U1"},
		"channel":{"id":"C1"},"message":{"ts":"171.102","thread_ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"placeholder"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "signed-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	response, err := service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Len(t, queue.intents, 1)
	intent := queue.intents[0]
	require.Equal(t, workspaceID, intent.WorkspaceID)
	require.Equal(t, actorID, intent.ActorID)
	require.Equal(t, installationID, intent.InstallationID)
	require.Equal(t, installGeneration, intent.InstallGeneration)
	require.Equal(t, storyID, intent.StoryID)
	require.Equal(t, "U1", intent.SlackUserID)
	require.Equal(t, "F1", intent.FileID)
	require.Equal(t, "171.101", intent.MessageTS)
	require.NotEmpty(t, intent.IdempotencyKey)
	require.NotContains(t, intent.IdempotencyKey, "signed-token")
	require.Contains(t, providerRequest["text"], "being attached")
}
