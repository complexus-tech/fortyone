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
	request := integrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: providerSlack,
		Metadata: map[string]any{"slack_user_id": "U-requester"},
	}
	thread := ProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: providerSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "D-requester", ExternalThreadID: "1710000000.001",
	}
	input := CreateIntegrationRequestCommentInput{AuthorID: actorID}

	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, input)

	require.NoError(t, err)
	require.Equal(t, "U-requester", prepared.ExternalRecipientUserID)
	payload, err := DecodeSlackProviderPayload(prepared.ProviderPayload)
	require.NoError(t, err)
	require.NotNil(t, payload.Authorization)
	require.Equal(t, []uuid.UUID{teamID}, payload.Authorization.AllowedTeamIDs)
	require.Equal(t, &actorID, payload.Authorization.ActorUserID)
	require.Equal(t, slackDeliveryAuthorizationScopeActorMembership, payload.Authorization.Scope)
}

func TestPrepareIntegrationRequestCommentCarriesLinkedSlackAuthor(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	generation := uuid.New()
	repo := &mockRepo{slackUserLinks: map[string]uuid.UUID{"T1:UAUTHOR": actorID}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	request := integrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: providerSlack,
	}
	thread := ProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: providerSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "C1", ExternalThreadID: "1710000000.001",
	}

	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, CreateIntegrationRequestCommentInput{AuthorID: actorID})

	require.NoError(t, err)
	payload, err := DecodeSlackProviderPayload(prepared.ProviderPayload)
	require.NoError(t, err)
	require.Equal(t, "UAUTHOR", payload.AuthorSlackUserID)
}

func TestFormatSlackIntegrationRequestCommentUsesLinkedIdentityOrSafeAuthorFallback(t *testing.T) {
	tests := []struct {
		name          string
		authorName    string
		authorSlackID string
		body          string
		want          string
	}{
		{
			name:          "linked Slack identity",
			authorName:    "Joseph Mukorivo",
			authorSlackID: "U123ABC",
			body:          "Please take a look",
			want:          "<@U123ABC> via FortyOne: Please take a look",
		},
		{
			name:       "unlinked FortyOne user",
			authorName: "A <team> member",
			body:       "Please take a look",
			want:       "A &lt;team&gt; member via FortyOne: Please take a look",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, formatSlackIntegrationRequestComment(test.authorName, test.authorSlackID, test.body))
		})
	}
}

func TestDeliverIntegrationRequestCommentSendsToBoundPublicThreadWithoutChannelAudienceMapping(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	generation := uuid.New()
	request := integrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: providerSlack,
	}
	thread := ProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: providerSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "C1", ExternalThreadID: "1710000000.001",
	}
	idempotencyKey := "integration-request-comment:" + uuid.NewString()
	comment := IntegrationRequestComment{
		ID: uuid.New(), WorkspaceID: workspaceID, ThreadID: thread.ID,
		Direction:    integrationrequests.CommentDirectionOutbound,
		AuthorUserID: &actorID, OutboundIdempotencyKey: &idempotencyKey,
		AuthorName: "Joseph Mukorivo", Body: "Reply from FortyOne",
	}
	repo := &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID: uuid.New(), WorkspaceID: workspaceID, SlackTeamID: "T1",
			BotAccessToken: "xoxb-token", InstallGeneration: generation, IsActive: true,
		},
		team:              slackrepository.TeamRecord{ID: teamID},
		teamMembers:       []slackrepository.TeamMemberRecord{{UserID: actorID}},
		slackUserLinks:    map[string]uuid.UUID{"T1:UAUTHOR": actorID},
		authorizedTeamIDs: []uuid.UUID{},
	}
	providerCalls := 0
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		providerCalls++
		require.Equal(t, "/chat.postMessage", request.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&body))
		require.Equal(t, "C1", body["channel"])
		require.Equal(t, "1710000000.001", body["thread_ts"])
		require.Equal(t, "<@UAUTHOR> via FortyOne: "+comment.Body, body["text"])
		_, _ = w.Write([]byte(`{"ok":true,"ts":"1710000000.002"}`))
	}))
	defer provider.Close()
	store := newEventStoreStub()
	store.deliveryContent = stringPointer(comment.Body)
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	service.client = provider.Client()
	service.webClient = newSlackWebClient(service.client)
	service.webClient.baseURL = provider.URL
	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, CreateIntegrationRequestCommentInput{AuthorID: actorID})
	require.NoError(t, err)

	err = service.DeliverIntegrationRequestComment(context.Background(), request, thread, comment, prepared)

	require.NoError(t, err)
	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "<@UAUTHOR> via FortyOne: "+comment.Body, *store.deliveryContent)
	require.Equal(t, 1, providerCalls)
	require.Empty(t, store.cancelledDeliveries)
	require.Len(t, store.completedDeliveries, 1)
	payload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{teamID}, payload.Authorization.AllowedTeamIDs)
	require.Equal(t, &actorID, payload.Authorization.ActorUserID)
	require.Equal(t, slackDeliveryAuthorizationScopeActorMembership, payload.Authorization.Scope)
}

func TestDeliverIntegrationRequestCommentCancelsWhenAuthorLosesRequestTeamAccess(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	generation := uuid.New()
	request := integrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: providerSlack,
	}
	thread := ProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: providerSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "C1", ExternalThreadID: "1710000000.001",
	}
	idempotencyKey := "integration-request-comment:" + uuid.NewString()
	comment := IntegrationRequestComment{
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
		teamMembers:       []slackrepository.TeamMemberRecord{},
		authorizedTeamIDs: []uuid.UUID{teamID},
	}
	store := newEventStoreStub()
	store.deliveryContent = stringPointer(comment.Body)
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, CreateIntegrationRequestCommentInput{AuthorID: actorID})
	require.NoError(t, err)

	err = service.DeliverIntegrationRequestComment(context.Background(), request, thread, comment, prepared)

	require.NoError(t, err)
	require.Len(t, store.outboundInputs, 1)
	require.Len(t, store.cancelledDeliveries, 1)
	require.Empty(t, store.completedDeliveries)
}

func TestDeliverIntegrationRequestCommentRevalidatesDMRecipientBeforeFirstSend(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	requestID := uuid.New()
	actorID := uuid.New()
	recipientID := uuid.New()
	generation := uuid.New()
	request := integrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: providerSlack,
		Metadata: map[string]any{"slack_user_id": "U-requester"},
	}
	thread := ProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: providerSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "D-requester", ExternalThreadID: "1710000000.001",
	}
	idempotencyKey := "integration-request-comment:" + uuid.NewString()
	comment := IntegrationRequestComment{
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
	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, CreateIntegrationRequestCommentInput{AuthorID: actorID})
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
	request := integrationRequest{
		ID: requestID, WorkspaceID: workspaceID, TeamID: teamID,
		Provider: providerSlack,
	}
	thread := ProviderThread{
		ID: uuid.New(), WorkspaceID: workspaceID, IntegrationRequestID: requestID,
		TeamID: teamID, Provider: providerSlack,
		ExternalWorkspaceID: "T1", InstallationGeneration: &generation,
		ExternalChannelID: "C-current", ExternalThreadID: "1710000000.009",
	}
	idempotencyKey := "integration-request-comment:" + uuid.NewString()
	comment := IntegrationRequestComment{
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
	prepared, err := service.PrepareIntegrationRequestComment(context.Background(), request, thread, CreateIntegrationRequestCommentInput{AuthorID: actorID})
	require.NoError(t, err)

	require.NoError(t, service.DeliverIntegrationRequestComment(context.Background(), request, thread, comment, prepared))
	require.Equal(t, 1, providerCalls)
	require.Len(t, store.completedDeliveries, 1)
	require.Empty(t, store.cancelledDeliveries)
}
