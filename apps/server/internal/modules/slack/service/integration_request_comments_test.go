package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPrepareIntegrationRequestCommentFreezesSelectedTeamActorAndDMRecipient(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	generation := uuid.New()
	service := &Service{}
	request := integrationrequests.CoreIntegrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: integrationrequests.ProviderSlack,
		Metadata: map[string]any{"slack_user_id": "U-requester"},
	}
	thread := integrationrequests.CoreProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: integrationrequests.ProviderSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "D-requester", ExternalThreadID: "1710000000.001",
	}
	input := integrationrequests.CoreCreateCommentInput{AuthorID: actorID}

	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, input)

	require.NoError(t, err)
	require.Equal(t, "U-requester", prepared.ExternalRecipientUserID)
	payload, err := DecodeSlackProviderPayload(prepared.ProviderPayload)
	require.NoError(t, err)
	require.NotNil(t, payload.Authorization)
	require.Equal(t, []uuid.UUID{teamID}, payload.Authorization.AllowedTeamIDs)
	require.Equal(t, &actorID, payload.Authorization.ActorUserID)
}

func TestDeliverIntegrationRequestCommentRevalidatesPublicAudienceBeforeFirstSend(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	generation := uuid.New()
	request := integrationrequests.CoreIntegrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: integrationrequests.ProviderSlack,
	}
	thread := integrationrequests.CoreProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: integrationrequests.ProviderSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "C1", ExternalThreadID: "1710000000.001",
	}
	idempotencyKey := "integration-request-comment:" + uuid.NewString()
	comment := integrationrequests.CoreIntegrationRequestComment{
		ID: uuid.New(), WorkspaceID: workspaceID, ThreadID: thread.ID,
		Direction:    integrationrequests.CommentDirectionOutbound,
		AuthorUserID: &actorID, OutboundIdempotencyKey: &idempotencyKey,
		Body: "Reply from FortyOne",
	}
	repo := &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID: uuid.New(), WorkspaceID: workspaceID, SlackTeamID: "T1",
			BotAccessToken: "xoxb-token", InstallGeneration: generation, IsActive: true,
		},
		team:              slackrepository.TeamRecord{ID: teamID},
		teamMembers:       []slackrepository.TeamMemberRecord{{UserID: actorID}},
		authorizedTeamIDs: []uuid.UUID{},
	}
	store := newEventStoreStub()
	store.deliveryContent = stringPointer(comment.Body)
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, integrationrequests.CoreCreateCommentInput{AuthorID: actorID})
	require.NoError(t, err)

	err = service.DeliverIntegrationRequestComment(context.Background(), request, thread, comment, prepared)

	require.NoError(t, err)
	require.Len(t, store.outboundInputs, 1)
	require.Len(t, store.cancelledDeliveries, 1)
	require.Empty(t, store.completedDeliveries)
	payload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{teamID}, payload.Authorization.AllowedTeamIDs)
	require.Equal(t, &actorID, payload.Authorization.ActorUserID)
}

func TestDeliverIntegrationRequestCommentRevalidatesDMRecipientBeforeFirstSend(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	recipientID := uuid.New()
	generation := uuid.New()
	request := integrationrequests.CoreIntegrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: integrationrequests.ProviderSlack,
		Metadata: map[string]any{"slack_user_id": "U-requester"},
	}
	thread := integrationrequests.CoreProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: integrationrequests.ProviderSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "D-requester", ExternalThreadID: "1710000000.001",
	}
	idempotencyKey := "integration-request-comment:" + uuid.NewString()
	comment := integrationrequests.CoreIntegrationRequestComment{
		ID: uuid.New(), WorkspaceID: workspaceID, ThreadID: thread.ID,
		Direction:    integrationrequests.CommentDirectionOutbound,
		AuthorUserID: &actorID, OutboundIdempotencyKey: &idempotencyKey,
		Body: "Reply from FortyOne",
	}
	repo := &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID: uuid.New(), WorkspaceID: workspaceID, SlackTeamID: "T1",
			BotAccessToken: "xoxb-token", InstallGeneration: generation, IsActive: true,
		},
		team:           slackrepository.TeamRecord{ID: teamID},
		teamMembers:    []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackUserLinks: map[string]uuid.UUID{"T1:U-requester": recipientID},
	}
	store := newEventStoreStub()
	store.deliveryContent = stringPointer(comment.Body)
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, integrationrequests.CoreCreateCommentInput{AuthorID: actorID})
	require.NoError(t, err)

	err = service.DeliverIntegrationRequestComment(context.Background(), request, thread, comment, prepared)

	require.NoError(t, err)
	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "U-requester", store.outboundInputs[0].ExternalRecipientUserID)
	require.Len(t, store.cancelledDeliveries, 1)
	require.Empty(t, store.completedDeliveries)
}

func TestDeliverIntegrationRequestCommentSendsOnceFromDurableAuthorizedDelivery(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	generation := uuid.New()
	request := integrationrequests.CoreIntegrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: integrationrequests.ProviderSlack,
	}
	thread := integrationrequests.CoreProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: integrationrequests.ProviderSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "C-current", ExternalThreadID: "1710000000.009",
	}
	idempotencyKey := "integration-request-comment:" + uuid.NewString()
	comment := integrationrequests.CoreIntegrationRequestComment{
		ID: uuid.New(), WorkspaceID: workspaceID, ThreadID: thread.ID,
		Direction:    integrationrequests.CommentDirectionOutbound,
		AuthorUserID: &actorID, OutboundIdempotencyKey: &idempotencyKey,
		Body: "Reply from FortyOne",
	}
	repo := &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID: uuid.New(), WorkspaceID: workspaceID, SlackTeamID: "T1",
			BotAccessToken: "xoxb-token", InstallGeneration: generation, IsActive: true,
		},
		team:              slackrepository.TeamRecord{ID: teamID},
		teamMembers:       []slackrepository.TeamMemberRecord{{UserID: actorID}},
		authorizedTeamIDs: []uuid.UUID{teamID},
	}
	providerCalls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		providerCalls++
		require.Equal(t, "/chat.postMessage", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		require.Equal(t, "C1", body["channel"])
		require.Equal(t, "1710000000.001", body["thread_ts"])
		require.Equal(t, "Reply from the durable outbox", body["text"])
		_, _ = w.Write([]byte(`{"ok":true,"ts":"1710000000.002"}`))
	}))
	defer provider.Close()
	store := newEventStoreStub()
	store.deliveryChannelID = "C1"
	store.deliveryThreadID = stringPointer("1710000000.001")
	store.deliveryContent = stringPointer("Reply from the durable outbox")
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL
	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, integrationrequests.CoreCreateCommentInput{AuthorID: actorID})
	require.NoError(t, err)

	require.NoError(t, service.DeliverIntegrationRequestComment(context.Background(), request, thread, comment, prepared))
	require.Equal(t, 1, providerCalls)
	require.Len(t, store.completedDeliveries, 1)
	require.Empty(t, store.cancelledDeliveries)
}
