package slack

import (
	"context"
	"net/http"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestVerifyRequest(t *testing.T) {
	secret := "secret"
	service := newTestService(nil, nil, nil, Config{SigningSecret: secret})
	body := []byte("payload=test")
	timestamp := "1700000000"

	headers := http.Header{}
	headers.Set("X-Slack-Request-Timestamp", timestamp)
	headers.Set("X-Slack-Signature", slackSignature(secret, timestamp, body))

	err := service.VerifyRequest(body, headers)
	require.NoError(t, err)
}

func TestHandleEventsDropsUnknownSlackInstallationBeforePersistence(t *testing.T) {
	repo := &mockRepo{getSlackTeamErr: slackdomain.ErrNotFound}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(directMessageEvent("Ev-unknown-team", "private message")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsDropsUnrelatedChannelMessagesBeforePersistence(t *testing.T) {
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       uuid.New(),
		SlackTeamID:       "T1",
		InstallGeneration: uuid.New(),
	}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(`{"type":"event_callback","team_id":"T1","event_id":"Ev-root","event":{"type":"message","channel_type":"channel","user":"U1","channel":"C1","ts":"10.1","text":"unrelated channel message"}}`))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsPersistsCandidateChannelThreadReply(t *testing.T) {
	workspaceID := uuid.New()
	installGeneration := uuid.New()
	linkedUserID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}, slackUserLinks: map[string]uuid.UUID{"T1:U1": linkedUserID}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{conversation: conversationRecord{
		ID:        uuid.New(),
		UpdatedAt: time.Unix(1_700_000_000, 0).UTC(),
	}}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-thread", "U1", "what about urgent work?")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Len(t, inbox.conversationLookups, 1)
	require.Equal(t, linkedUserID, inbox.conversationLookups[0].UserID)
	require.Equal(t, "10.1", inbox.conversationLookups[0].ExternalThreadID)
	require.Len(t, queue.payloads, 1)
	require.Equal(t, "slack", queue.payloads[0].Provider)
	require.NotEqual(t, uuid.Nil, queue.payloads[0].InboxID)
	require.Len(t, queue.requests, 1)
	require.Contains(t, string(queue.requests[0].Body), `"event_id":"Ev-thread"`)
}

func TestHandleEventsPersistsExactRequestThreadReplyFromUnlinkedSlackActor(t *testing.T) {
	workspaceID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: installGeneration,
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}}
	requests := &mockRequestStore{threadMatch: true}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-unlinked-request-thread", "U-external", "Customer confirmed")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Empty(t, inbox.conversationLookups)
	require.Len(t, queue.payloads, 1)
	require.Equal(t, "slack", queue.payloads[0].Provider)
	require.NotEqual(t, uuid.Nil, queue.payloads[0].InboxID)
	require.Equal(t, workspaceID, requests.threadLookup.WorkspaceID)
	require.Equal(t, installGeneration, requests.threadLookup.InstallationGeneration)
	require.Equal(t, "T1", requests.threadLookup.ExternalWorkspaceID)
	require.Equal(t, "C1", requests.threadLookup.ExternalChannelID)
	require.Equal(t, "10.1", requests.threadLookup.ExternalThreadID)
}

func TestHandleEventsDropsUnsubscribedChannelThreadBeforePersistence(t *testing.T) {
	workspaceID := uuid.New()
	linkedUserID := uuid.New()
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		InstallGeneration: uuid.New(),
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}, slackUserLinks: map[string]uuid.UUID{"T1:U1": linkedUserID}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-unsubscribed", "U1", "unrelated reply")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Len(t, inbox.conversationLookups, 1)
	require.Empty(t, queue.payloads)
}

func TestHandleEventsDropsBroadDuplicateOfAppMentionBeforePersistence(t *testing.T) {
	workspaceID := uuid.New()
	botUserID := "B1"
	repo := &mockRepo{slackWorkspace: slackrepository.SlackWorkspaceRecord{
		WorkspaceID:       workspaceID,
		SlackTeamID:       "T1",
		BotUserID:         &botUserID,
		InstallGeneration: uuid.New(),
		AuthorizedAt:      time.Unix(1_700_000_000, 0).UTC(),
	}}
	queue := &eventQueueStub{}
	inbox := &eventInboxCapture{}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})
	WithEventRuntime(queue, inbox)(service)

	response, err := service.HandleEvents(context.Background(), []byte(channelThreadEvent("Ev-mention-copy", "U1", "<@B1> show my work")))

	require.NoError(t, err)
	require.Empty(t, response.Challenge)
	require.Empty(t, inbox.conversationLookups)
	require.Empty(t, queue.payloads)
}
