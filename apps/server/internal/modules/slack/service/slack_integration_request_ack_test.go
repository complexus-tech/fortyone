package slack

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAcceptIntegrationRequestPostsLinkerAndCanonicalStoryCode(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{
		teamMembers: []slackrepository.TeamMemberRecord{{
			UserID:   actorID,
			Username: "joseph",
			FullName: "Joseph Mukorivo",
			Email:    "joseph@example.com",
		}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T1",
			BotAccessToken:    "xoxb-token",
			InstallGeneration: installGeneration,
			IsActive:          true,
		},
	}
	store := newEventStoreStub()
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	err := service.AcceptIntegrationRequest(context.Background(), integrationRequest{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		TeamID:      teamID,
		Provider:    providerSlack,
		Metadata: map[string]any{
			"slack_channel_id": "C1",
			"slack_message_ts": "171.100",
			"slack_team_id":    "T1",
			"workspace_slug":   "acme",
			"team_code":        "web",
		},
	}, singleStory{
		ID:         uuid.New(),
		SequenceID: 123,
		Title:      "Fix login bug",
		Team:       teamID,
		Reporter:   &actorID,
		CreatedAt:  time.Unix(1_700_000_000, 0),
		UpdatedAt:  time.Unix(1_700_000_000, 0),
	})

	require.NoError(t, err)
	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "Joseph Mukorivo linked a request to <https://acme.fortyone.app/work/WEB-123|WEB-123>", store.outboundInputs[0].Content)
	require.Equal(t, "C1", store.outboundInputs[0].ExternalChannelID)
	require.Equal(t, "171.100", store.outboundInputs[0].ExternalThreadID)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Len(t, providerPayload.Metadata.Entities, 1)
	entity := providerPayload.Metadata.Entities[0]
	require.Equal(t, "https://acme.fortyone.app/work/WEB-123", entity.URL)
	require.NotContains(t, entity.EntityPayload.Fields, "description")
	require.NotContains(t, entity.EntityPayload.Fields, "created_by")
	require.NotContains(t, entity.EntityPayload.Fields, "date_created")
	require.NotContains(t, entity.EntityPayload.Fields, "date_updated")
	require.False(t, *providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestPostSlackTaskAckFallbackKeepsLifecycleCopyAndSuppressesClassicUnfurls(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	reporterID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		IsActive:          true,
	}}
	store := newEventStoreStub()
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	service.postSlackTaskAck(
		context.Background(),
		workspaceID,
		installGeneration,
		"slack:request:fallback",
		requestSourceContext{SlackTeamID: "T1", SlackChannelID: "C1", SlackMessageTS: "171.100"},
		"xoxb-token",
		"acme",
		"WEB",
		"Joseph Mukorivo",
		"",
		slackStoryReceiptActionLinkedRequest,
		singleStory{ID: uuid.New(), SequenceID: 123, Team: teamID, Reporter: &reporterID},
	)

	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "Joseph Mukorivo linked a request to <https://acme.app.example.com/work/WEB-123|WEB-123>", store.outboundInputs[0].Content)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Nil(t, providerPayload.Metadata)
	require.NotNil(t, providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlLinks)
	require.NotNil(t, providerPayload.UnfurlMedia)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestPostSlackCreationAckCancelsWhenInstallationGenerationChanges(t *testing.T) {
	store := newEventStoreStub()
	workspaceID := uuid.New()
	originalGeneration := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("stale creation acknowledgement reached Slack")
		return nil, errors.New("unexpected provider call")
	})}

	service.postSlackCreationAck(
		context.Background(),
		workspaceID,
		originalGeneration,
		"slack:view:V1:confirmation",
		requestSourceContext{SlackTeamID: "T1", SlackChannelID: "C1", SlackMessageTS: "171.100"},
		"xoxb-old-token",
		"Task created in FortyOne.",
	)

	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, originalGeneration, *store.outboundInputs[0].InstallGeneration)
	require.Len(t, store.cancelledDeliveries, 1)
	require.Empty(t, store.completedDeliveries)
}
