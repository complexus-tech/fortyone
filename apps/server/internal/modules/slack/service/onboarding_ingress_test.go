package slack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type onboardingIngressRuntimeStub struct {
	*eventInboxCapture
	*eventStoreStub
}

type blockingOnboardingReceiptRepo struct {
	*onboardingReceiptRepoStub
	started chan struct{}
	release chan struct{}
}

func (r *blockingOnboardingReceiptRepo) HasSlackUserOnboardingReceipt(
	ctx context.Context,
	workspaceID uuid.UUID,
	slackTeamID, slackUserID string,
) (bool, error) {
	select {
	case r.started <- struct{}{}:
	default:
	}
	select {
	case <-r.release:
		return r.onboardingReceiptRepoStub.HasSlackUserOnboardingReceipt(ctx, workspaceID, slackTeamID, slackUserID)
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

// FindConversation resolves the method promoted by both embedded test stores.
// Event ingress must use the inbox's conversation state, while onboarding uses
// the same runtime value through OutboundStore.
func (s *onboardingIngressRuntimeStub) FindConversation(
	ctx context.Context,
	input conversationInput,
) (conversationRecord, error) {
	return s.eventInboxCapture.FindConversation(ctx, input)
}

func TestHandleCommandClaimsFirstUseGuideBeforeAcknowledgement(t *testing.T) {
	workspaceID := uuid.New()
	installation := firstUseIngressInstallation(workspaceID)
	linkedUserID := uuid.New()
	repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{
		workspace:      slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: installation,
		slackUserLinks: map[string]uuid.UUID{"T123:U123": linkedUserID},
	}}
	outbox := newEventStoreStub()
	outbox.processOutbound = false
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = outbox

	form := url.Values{}
	form.Set("team_id", "T123")
	form.Set("user_id", "U123")
	form.Set("channel_id", "C123")
	form.Set("trigger_id", "trigger")
	form.Set("text", "create task Ship onboarding")

	response, err := service.HandleCommand(context.Background(), []byte(form.Encode()))

	require.NoError(t, err)
	require.Empty(t, response.Text)
	assertSingleFirstUseIngressClaim(t, outbox, installation, "U123")
}

func TestHandleMessageActionClaimsFirstUseGuideBeforeAcknowledgement(t *testing.T) {
	workspaceID := uuid.New()
	installation := firstUseIngressInstallation(workspaceID)
	repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{slackWorkspace: installation}}
	outbox := newEventStoreStub()
	outbox.processOutbound = false
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = outbox

	payload, err := json.Marshal(map[string]any{
		"type":         "message_action",
		"trigger_id":   "trigger",
		"response_url": "https://hooks.slack.com/actions/T123/message",
		"team":         map[string]string{"id": "T123", "domain": "acme"},
		"channel":      map[string]string{"id": "C123", "name": "general"},
		"user":         map[string]string{"id": "U123", "username": "joseph"},
		"message":      map[string]string{"user": "U456", "text": "Ship onboarding", "ts": "171234.000100"},
	})
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payload))

	response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	assertSingleFirstUseIngressClaim(t, outbox, installation, "U123")
}

func TestHandleMessageActionStartsModalWhileFirstUseClaimIsPending(t *testing.T) {
	workspaceID := uuid.New()
	linkedUserID := uuid.New()
	installation := firstUseIngressInstallation(workspaceID)
	installation.BotAccessToken = "xoxb-token"
	repo := &blockingOnboardingReceiptRepo{
		onboardingReceiptRepoStub: &onboardingReceiptRepoStub{mockRepo: &mockRepo{
			slackWorkspace: installation,
			slackUserLinks: map[string]uuid.UUID{"T123:U123": linkedUserID},
		}},
		started: make(chan struct{}, 1),
		release: make(chan struct{}),
	}
	outbox := newEventStoreStub()
	outbox.processOutbound = false
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	service.outbound = outbox
	modalOpened := make(chan struct{}, 1)
	modalUpdated := make(chan struct{}, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.String() {
		case "https://slack.com/api/views.open":
			modalOpened <- struct{}{}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true,"view":{"id":"V123"}}`)),
				Header:     make(http.Header),
			}, nil
		case "https://slack.com/api/views.update":
			modalUpdated <- struct{}{}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		default:
			return nil, errors.New("unexpected Slack endpoint")
		}
	})}

	payload, err := json.Marshal(map[string]any{
		"type":         "message_action",
		"trigger_id":   "trigger",
		"response_url": "https://hooks.slack.com/actions/T123/message",
		"team":         map[string]string{"id": "T123", "domain": "acme"},
		"channel":      map[string]string{"id": "C123", "name": "general"},
		"user":         map[string]string{"id": "U123", "username": "joseph"},
		"message":      map[string]string{"user": "U456", "text": "Ship onboarding", "ts": "171234.000100"},
	})
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payload))

	result := make(chan error, 1)
	go func() {
		_, handleErr := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
		result <- handleErr
	}()

	select {
	case <-repo.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for the first-use claim")
	}
	select {
	case <-modalOpened:
	case <-time.After(time.Second):
		close(repo.release)
		t.Fatal("modal work did not start while the first-use claim was pending")
	}
	close(repo.release)
	require.NoError(t, <-result)
	select {
	case <-modalUpdated:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for modal hydration to finish")
	}
	assertSingleFirstUseIngressClaim(t, outbox, installation, "U123")
}

func TestHandleEventsClaimsFirstUseGuideForDirectMessage(t *testing.T) {
	workspaceID := uuid.New()
	installation := firstUseIngressInstallation(workspaceID)
	repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{slackWorkspace: installation}}
	queue := &eventQueueStub{}
	runtime := &onboardingIngressRuntimeStub{
		eventInboxCapture: &eventInboxCapture{},
		eventStoreStub:    newEventStoreStub(),
	}
	runtime.processOutbound = false
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	WithEventRuntime(queue, runtime)(service)

	response, err := service.HandleEvents(
		context.Background(),
		[]byte(directMessageEvent("Ev-first-use", "What am I working on?")),
	)

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	assertSingleFirstUseIngressClaim(t, runtime.eventStoreStub, installation, "U1")
	require.Len(t, queue.payloads, 1)
	require.Equal(t, "slack", queue.payloads[0].Provider)
	require.NotEqual(t, uuid.Nil, queue.payloads[0].InboxID)
}

func TestHandleEventsDoesNotClaimFirstUseGuideForIgnoredMessages(t *testing.T) {
	tests := []struct {
		name      string
		eventBody string
	}{
		{
			name:      "unsupported channel root",
			eventBody: `{"type":"event_callback","team_id":"T123","event_id":"Ev-root","event":{"type":"message","channel_type":"channel","user":"U123","channel":"C123","ts":"10.1","text":"unrelated root"}}`,
		},
		{
			name:      "unsubscribed channel thread",
			eventBody: `{"type":"event_callback","team_id":"T123","event_id":"Ev-thread","event":{"type":"message","channel_type":"channel","user":"U123","channel":"C123","ts":"10.2","thread_ts":"10.1","text":"unrelated reply"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			installation := firstUseIngressInstallation(uuid.New())
			repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{slackWorkspace: installation}}
			queue := &eventQueueStub{}
			runtime := &onboardingIngressRuntimeStub{
				eventInboxCapture: &eventInboxCapture{},
				eventStoreStub:    newEventStoreStub(),
			}
			runtime.processOutbound = false
			service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
			WithEventRuntime(queue, runtime)(service)

			response, err := service.HandleEvents(context.Background(), []byte(test.eventBody))

			require.NoError(t, err)
			require.Empty(t, response.Challenge)
			require.Empty(t, runtime.outboundInputs)
			require.Empty(t, queue.payloads)
		})
	}
}

func TestLinkSlackAccountClaimsFirstUseGuideAsFallback(t *testing.T) {
	workspaceID := uuid.New()
	userID := uuid.New()
	installation := firstUseIngressInstallation(workspaceID)
	repo := &onboardingReceiptRepoStub{mockRepo: &mockRepo{
		workspace:      slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme"},
		slackWorkspace: installation,
	}}
	outbox := newEventStoreStub()
	outbox.processOutbound = false
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{
		WebsiteURL: "https://fortyone.app",
	})
	service.outbound = outbox

	link, err := service.buildSlackUserLinkURL(context.Background(), workspaceID, "T123", "U999")
	require.NoError(t, err)
	parsedLink, err := url.Parse(link)
	require.NoError(t, err)

	_, err = service.LinkSlackAccount(
		context.Background(),
		workspaceID,
		userID,
		parsedLink.Query().Get("slack_link_token"),
	)

	require.NoError(t, err)
	require.Equal(t, userID, repo.slackUserLinks["T123:U999"])
	assertSingleFirstUseIngressClaim(t, outbox, installation, "U999")
}

func firstUseIngressInstallation(workspaceID uuid.UUID) slackrepository.SlackWorkspaceRecord {
	return slackrepository.SlackWorkspaceRecord{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T123",
		InstallGeneration: uuid.New(),
		IsActive:          true,
	}
}

func assertSingleFirstUseIngressClaim(
	t *testing.T,
	outbox *eventStoreStub,
	installation slackrepository.SlackWorkspaceRecord,
	slackUserID string,
) {
	t.Helper()
	require.Len(t, outbox.outboundInputs, 1)
	input := outbox.outboundInputs[0]
	require.Equal(t, slackOnboardingPurpose, input.Purpose)
	require.Equal(t, installation.WorkspaceID, input.WorkspaceID)
	require.Equal(t, installation.InstallGeneration, *input.InstallGeneration)
	require.Equal(t, installation.SlackTeamID, input.ExternalWorkspaceID)
	require.Equal(t, slackUserID, input.ExternalRecipientUserID)
	require.Equal(t, slackUserID, input.ExternalChannelID)
	require.Contains(t, input.IdempotencyKey, installation.InstallGeneration.String())
}
