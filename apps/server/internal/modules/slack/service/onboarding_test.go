package slack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type onboardingReceiptRepoStub struct {
	*mockRepo
	delivered    bool
	err          error
	errOnce      error
	workspaceIDs []uuid.UUID
	teamIDs      []string
	userIDs      []string
}

type onboardingSignalStore struct {
	*eventStoreStub
	started chan struct{}
}

func (s *onboardingSignalStore) StartOutboundDelivery(
	ctx context.Context,
	input outboundDeliveryInput,
) (outboundDeliveryRecord, bool, error) {
	record, claimed, err := s.eventStoreStub.StartOutboundDelivery(ctx, input)
	select {
	case s.started <- struct{}{}:
	default:
	}
	return record, claimed, err
}

func (r *onboardingReceiptRepoStub) HasSlackUserOnboardingReceipt(
	_ context.Context,
	workspaceID uuid.UUID,
	slackTeamID, slackUserID string,
) (bool, error) {
	r.workspaceIDs = append(r.workspaceIDs, workspaceID)
	r.teamIDs = append(r.teamIDs, slackTeamID)
	r.userIDs = append(r.userIDs, slackUserID)
	if r.errOnce != nil {
		err := r.errOnce
		r.errOnce = nil
		return false, err
	}
	return r.delivered, r.err
}

func TestSendFirstInteractionGuidePostsOneDurablePrivateMessage(t *testing.T) {
	workspaceID := uuid.New()
	installGeneration := uuid.New()
	repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			BotUserID:         stringPointer("B123"),
			BotAccessToken:    "xoxb-token",
			InstallGeneration: installGeneration,
			IsActive:          true,
		},
	}}
	store := newEventStoreStub()
	service := newTestService(repo, nil, nil, Config{})
	service.outbound = store

	var posted map[string]any
	service.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/chat.postMessage", request.URL.String())
		require.Equal(t, "Bearer xoxb-token", request.Header.Get("Authorization"))
		require.NoError(t, json.NewDecoder(request.Body).Decode(&posted))
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.100"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	prepared, shouldSend, err := service.prepareFirstInteractionGuide(context.Background(), repo.slackWorkspace, " U123 ")
	require.NoError(t, err)
	require.True(t, shouldSend)
	err = service.sendPreparedFirstInteractionGuide(context.Background(), prepared)

	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{workspaceID}, repo.workspaceIDs)
	require.Equal(t, []string{"T123"}, repo.teamIDs)
	require.Equal(t, []string{"U123"}, repo.userIDs)
	require.Len(t, store.outboundInputs, 1)
	input := store.outboundInputs[0]
	require.Equal(t, slackOnboardingPurpose, input.Purpose)
	require.Equal(t, workspaceID, input.WorkspaceID)
	require.Nil(t, input.UserID)
	require.Equal(t, installGeneration, *input.InstallGeneration)
	require.Equal(t, "T123", input.ExternalWorkspaceID)
	require.Equal(t, "U123", input.ExternalRecipientUserID)
	require.Equal(t, "U123", input.ExternalChannelID)
	require.Empty(t, input.ExternalThreadID)
	require.Nil(t, input.ExpiresAt)
	require.Contains(t, input.IdempotencyKey, installGeneration.String())
	require.Contains(t, input.Content, "*Welcome to FortyOne in Slack*")
	require.Contains(t, input.Content, "mention <@B123>")
	require.Equal(t, "U123", posted["channel"])
	require.Equal(t, input.Content, posted["text"])
	require.NotEmpty(t, posted["client_msg_id"])
	require.Len(t, store.completedDeliveries, 1)
	require.Equal(t, "171.100", store.completedDeliveries[0].externalMessageID)
}

func TestSendFirstInteractionGuideStopsAtDurableReceipt(t *testing.T) {
	workspaceID := uuid.New()
	repo := &onboardingReceiptRepoStub{
		mockRepo:  &mockRepo{},
		delivered: true,
	}
	installation := slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		BotAccessToken:    "xoxb-token",
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
	store := newEventStoreStub()
	service := newTestService(repo, nil, nil, Config{})
	service.outbound = store

	_, shouldSend, err := service.prepareFirstInteractionGuide(context.Background(), installation, "U123")
	require.NoError(t, err)
	require.False(t, shouldSend)
	require.Len(t, repo.workspaceIDs, 1)
	require.Empty(t, store.outboundInputs)
}

func TestSendFirstInteractionGuideTreatsConcurrentClaimAsInFlight(t *testing.T) {
	repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{}}
	installation := slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T123",
		BotAccessToken:    "xoxb-token",
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
	store := newEventStoreStub()
	store.outboundErr = errors.Join(ErrOutboundDeliveryBusy, &messagingrepository.LeaseBusyError{})
	service := newTestService(repo, nil, nil, Config{})
	service.outbound = store

	_, shouldSend, err := service.prepareFirstInteractionGuide(context.Background(), installation, "U123")
	require.NoError(t, err)
	require.False(t, shouldSend)
	require.Len(t, store.outboundInputs, 1)
	require.Empty(t, store.completedDeliveries)
	require.Empty(t, store.failedDeliveries)
}

func TestFirstInteractionGuideProviderIdentitySurvivesReinstall(t *testing.T) {
	workspaceID := uuid.New()
	repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{}}
	service := newTestService(repo, nil, nil, Config{})
	service.outbound = newEventStoreStub()
	firstInstallation := slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
	secondInstallation := firstInstallation
	secondInstallation.InstallGeneration = uuid.New()

	first, claimed, err := service.prepareFirstInteractionGuide(context.Background(), firstInstallation, "U123")
	require.NoError(t, err)
	require.True(t, claimed)
	second, claimed, err := service.prepareFirstInteractionGuide(context.Background(), secondInstallation, "U123")
	require.NoError(t, err)
	require.True(t, claimed)

	require.NotEqual(t, first.idempotencyKey, second.idempotencyKey)
	require.Equal(t, first.providerKey, second.providerKey)
	require.Equal(t, deterministicSlackMessageID(first.providerKey), deterministicSlackMessageID(second.providerKey))
	require.NotContains(t, first.idempotencyKey, "T123")
	require.NotContains(t, first.idempotencyKey, "U123")
}

func TestDispatchFirstInteractionGuideRetriesAfterPreparationDeadline(t *testing.T) {
	installation := slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T123",
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
	repo := &onboardingReceiptRepoStub{
		mockRepo: &mockRepo{},
		errOnce:  context.DeadlineExceeded,
	}
	store := &onboardingSignalStore{
		eventStoreStub: newEventStoreStub(),
		started:        make(chan struct{}, 1),
	}
	store.processOutbound = false
	service := newTestService(repo, nil, nil, Config{})
	service.outbound = store

	service.dispatchFirstInteractionGuide(context.Background(), installation, "U123")

	select {
	case <-store.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the detached preparation retry")
	}
	require.Len(t, repo.workspaceIDs, 2)
	require.Equal(t, slackOnboardingPurpose, store.outboundInputs[0].Purpose)
}

func TestBuildSlackFirstInteractionGuideDescribesSupportedMayaWorkflows(t *testing.T) {
	guide := buildSlackFirstInteractionGuide("B123")

	for _, expected := range []string{
		"*Welcome to FortyOne in Slack*",
		"mention <@B123>",
		"`/fortyone [title]`",
		"create or update a story",
		"wait for you to confirm",
		"permission-aware preview",
		"sync back to the Request",
		"work you're allowed to access",
	} {
		require.Contains(t, guide, expected)
	}
	require.NotContains(t, guide, "when Maya is enabled")
	require.NotContains(t, guide, "when workflow actions are enabled")
	require.NotContains(t, guide, "depend on your plan and workspace settings")
	require.NotContains(t, guide, "calendar")
	require.NotContains(t, guide, "GitHub")
}

func TestBuildSlackFirstInteractionGuideFallsBackWithoutBotUserID(t *testing.T) {
	guide := buildSlackFirstInteractionGuide("")

	require.Contains(t, guide, "mention the FortyOne app")
	require.Contains(t, guide, "`/fortyone [title]`")
	require.Contains(t, guide, "*Make changes with Maya*")
}

func TestSlackEventCountsAsFirstInteraction(t *testing.T) {
	for _, kind := range []slackEventKind{
		slackEventKindMention,
		slackEventKindDirect,
		slackEventKindChannelThread,
		slackEventKindLinkShared,
		slackEventKindEntityDetails,
	} {
		require.True(t, slackEventCountsAsFirstInteraction(kind), kind)
	}
	for _, kind := range []slackEventKind{
		slackEventKindUninstalled,
		slackEventKindRevoked,
	} {
		require.False(t, slackEventCountsAsFirstInteraction(kind), kind)
	}
}
