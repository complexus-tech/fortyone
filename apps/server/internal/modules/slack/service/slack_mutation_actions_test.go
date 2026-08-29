package slack

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHandleMutationActionRechecksScopeAndReplacesButtonsWithLinkedReceipt(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T1",
			BotAccessToken:    "xoxb-token",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
	}
	confirmer := &mutationConfirmerStub{result: StoryMutationResult{
		Status:    "applied",
		Operation: storyMutationCreate,
		StoryID:   uuid.New(),
		Reference: "web-123",
		TeamID:    teamID,
		Title:     "Fix login",
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
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions",
		"team":{"id":"T1"},
		"user":{"id":"U1","name":"joseph"},
		"channel":{"id":"C1"},
		"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"opaque-token"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	response, err := service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, []string{"opaque-token"}, confirmer.tokens)
	require.Len(t, confirmer.scopes, 1)
	require.Equal(t, workspaceID, confirmer.scopes[0].WorkspaceID)
	require.Equal(t, actorID, confirmer.scopes[0].UserID)
	require.Equal(t, []uuid.UUID{teamID}, confirmer.scopes[0].AllowedTeamIDs)
	require.True(t, confirmer.scopes[0].AllowMutations)
	require.Equal(t, "Joseph Mukorivo created <https://acme.fortyone.app/work/WEB-123|WEB-123>", providerRequest["text"])
	require.Empty(t, providerRequest["blocks"])
}

func TestHandleMutationActionRendersItemizedBatchReceipt(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T1",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
	}
	confirmer := &mutationConfirmerStub{result: StoryMutationResult{
		Status:    "applied",
		Operation: storyMutationCreateBatch,
		TeamID:    teamID,
		Items: []storyMutationItemResult{
			{Index: 0, Status: "applied", StoryID: uuid.New(), Reference: "WEB-123", TeamID: teamID, Title: "Add Microsoft auth", Priority: "High"},
			{Index: 1, Status: "applied", StoryID: uuid.New(), Reference: "WEB-124", TeamID: teamID, Title: "Add TikTok", Priority: "No Priority"},
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
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions","team":{"id":"T1"},"user":{"id":"U1","name":"joseph"},
		"channel":{"id":"C1"},"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"placeholder"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-batch-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	response, err := service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, "Joseph Mukorivo created 2 stories:\n• <https://acme.fortyone.app/work/WEB-123|WEB-123> — Add Microsoft auth\n• <https://acme.fortyone.app/work/WEB-124|WEB-124> — Add TikTok", providerRequest["text"])
	require.Empty(t, providerRequest["blocks"])
}

func TestHandleMutationActionKeepsRetryControlsAndShowsPartialBatchProgress(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T1",
			BotAccessToken: "xoxb-token",
			IsActive:       true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
	}
	confirmer := &mutationConfirmerStub{
		result: StoryMutationResult{
			Status:    "partial",
			Operation: storyMutationCreateBatch,
			TeamID:    teamID,
			Items: []storyMutationItemResult{{
				Index: 0, Status: "applied", StoryID: uuid.New(), Reference: "WEB-123",
				TeamID: teamID, Title: "Add Microsoft auth", Priority: "High",
			}},
		},
		err: errors.New("temporary story provider failure"),
	}
	var providerRequest map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/chat.update", request.URL.Path)
		require.NoError(t, json.NewDecoder(request.Body).Decode(&providerRequest))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	WithMutationConfirmer(confirmer)(service)
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions","team":{"id":"T1"},"user":{"id":"U1","name":"joseph"},
		"channel":{"id":"C1"},"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"placeholder"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-batch-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	response, err := service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(
		t,
		"Joseph Mukorivo created 1 of the proposed stories before FortyOne hit an error:\n• <https://acme.fortyone.app/work/WEB-123|WEB-123> — Add Microsoft auth\nSelect *Retry remaining* to try again. Already-created stories will not be duplicated.",
		providerRequest["text"],
	)
	blocks, ok := providerRequest["blocks"].([]any)
	require.True(t, ok)
	require.Len(t, blocks, 2, "partial progress must retain the retry and cancel controls")
	actions, ok := blocks[1].(map[string]any)
	require.True(t, ok)
	elements, ok := actions["elements"].([]any)
	require.True(t, ok)
	require.Len(t, elements, 1)
	confirm, ok := elements[0].(map[string]any)
	require.True(t, ok)
	buttonText, ok := confirm["text"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Retry remaining", buttonText["text"])
	encodedValue, ok := confirm["value"].(string)
	require.True(t, ok)
	decodedValue, err := decodeSlackMutationActionValue(encodedValue)
	require.NoError(t, err)
	require.Equal(t, "opaque-batch-token", decodedValue.Token)
}

func TestHandleMutationActionIgnoresLegacyDisabledSettings(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T1",
			BotAccessToken:    "xoxb-token",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
		slackUserLinks:    map[string]uuid.UUID{"T1:U1": actorID},
		authorizedTeamIDs: []uuid.UUID{teamID},
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			FullName: "Joseph Mukorivo",
		}},
		agentSettings: slackrepository.AgentSettingsRecord{
			Guidance: "configured",
		},
	}
	confirmer := &mutationConfirmerStub{result: StoryMutationResult{
		Status:    "applied",
		Operation: storyMutationCreate,
		StoryID:   uuid.New(),
		Reference: "WEB-123",
		TeamID:    teamID,
		Title:     "Fix login",
	}}
	var providerRequest map[string]any
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		require.NoError(t, json.NewDecoder(request.Body).Decode(&providerRequest))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer provider.Close()
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	WithMutationConfirmer(confirmer)(service)
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL

	var payload interactionPayload
	require.NoError(t, json.Unmarshal([]byte(`{
		"type":"block_actions","team":{"id":"T1"},"user":{"id":"U1"},
		"channel":{"id":"C1"},"message":{"ts":"171.100"},
		"actions":[{"action_id":"fortyone_confirm_story_mutation","value":"opaque-token"}]
	}`), &payload))
	actionValue, err := encodeSlackMutationActionValue("U1", "opaque-token")
	require.NoError(t, err)
	payload.Actions[0].Value = actionValue

	_, err = service.handleMutationAction(context.Background(), payload)
	require.NoError(t, err)
	require.Equal(t, []string{"opaque-token"}, confirmer.tokens)
	require.Contains(t, providerRequest["text"], "created")
	require.Empty(t, providerRequest["blocks"])
}
