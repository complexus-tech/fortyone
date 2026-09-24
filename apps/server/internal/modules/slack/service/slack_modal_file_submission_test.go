package slack

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type modalFileImportQueueStub struct {
	intents []SlackFileImportIntent
	err     error
}

func (stub *modalFileImportQueueStub) QueueSlackFileImport(_ context.Context, intent SlackFileImportIntent) error {
	if stub.err != nil {
		return stub.err
	}
	stub.intents = append(stub.intents, intent)
	return nil
}

func TestCreateStoryFromSlackModalQueuesSelectedAndUploadedFiles(t *testing.T) {
	service, stories, requests, queue, workspaceID, teamID, statusID := newModalFileSubmissionFixture()
	payload := modalFileSubmissionPayload(t, teamID, statusID.String())

	response, err := service.handleViewSubmission(context.Background(), payload)
	require.NoError(t, err)
	require.Contains(t, string(response.Body), `"response_action":"clear"`)
	require.Equal(t, 1, stories.createCalls)
	require.Zero(t, requests.calls)
	require.Len(t, queue.intents, 2)

	source := queue.intents[0]
	require.Equal(t, "F123", source.FileID)
	require.Equal(t, workspaceID, source.WorkspaceID)
	require.Equal(t, teamID, stories.lastStory.Team)
	require.Equal(t, "T123", source.SlackTeamID)
	require.Equal(t, "U123", source.SlackUserID)
	require.Equal(t, "C123", source.ChannelID)
	require.Equal(t, "171.100", source.ThreadTS)
	require.Equal(t, "171.101", source.MessageTS)
	require.NotEqual(t, uuid.Nil, source.StoryID)
	require.NotEmpty(t, source.IdempotencyKey)

	uploaded := queue.intents[1]
	require.Equal(t, "F456", uploaded.FileID)
	require.Empty(t, uploaded.ChannelID)
	require.Empty(t, uploaded.ThreadTS)
	require.Empty(t, uploaded.MessageTS)
	require.Equal(t, source.StoryID, uploaded.StoryID)
	require.NotEqual(t, source.IdempotencyKey, uploaded.IdempotencyKey)
	require.Len(t, service.outbound.(*eventStoreStub).outboundInputs, 1)
	require.Contains(t, service.outbound.(*eventStoreStub).outboundInputs[0].Content, "Files are being attached in the background.")
}

func TestCreateRequestFromSlackModalRejectsFiles(t *testing.T) {
	service, stories, requests, queue, _, teamID, _ := newModalFileSubmissionFixture()
	payload := modalFileSubmissionPayload(t, teamID, slackRequestStatusValue)

	response, err := service.handleViewSubmission(context.Background(), payload)
	require.NoError(t, err)
	require.Contains(t, string(response.Body), `"response_action":"errors"`)
	require.Contains(t, string(response.Body), "Choose a story status to attach files")
	require.Zero(t, stories.createCalls)
	require.Zero(t, requests.calls)
	require.Empty(t, queue.intents)
}

func TestCreateStoryFromSlackModalDoesNotAcknowledgeUnqueuedFiles(t *testing.T) {
	service, stories, _, queue, _, teamID, statusID := newModalFileSubmissionFixture()
	payload := modalFileSubmissionPayload(t, teamID, statusID.String())
	queue.err = errors.New("queue unavailable")

	response, err := service.handleViewSubmission(context.Background(), payload)
	require.NoError(t, err)
	require.Contains(t, string(response.Body), `"response_action":"errors"`)
	require.Contains(t, string(response.Body), "Story saved, but its files could not be attached")
	require.Equal(t, 1, stories.createCalls)
	require.Empty(t, queue.intents)
	require.Empty(t, service.outbound.(*eventStoreStub).outboundInputs)
}

func TestCreateStoryFromSlackModalRequiresFileImportQueue(t *testing.T) {
	service, stories, _, _, _, teamID, statusID := newModalFileSubmissionFixture()
	service.fileImportQueue = nil
	payload := modalFileSubmissionPayload(t, teamID, statusID.String())

	response, err := service.handleViewSubmission(context.Background(), payload)
	require.NoError(t, err)
	require.Contains(t, string(response.Body), "Files cannot be attached right now")
	require.Zero(t, stories.createCalls)
}

func newModalFileSubmissionFixture() (*Service, *mockStoryService, *mockRequestStore, *modalFileImportQueueStub, uuid.UUID, uuid.UUID, uuid.UUID) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	statusID := uuid.New()
	actorID := uuid.New()
	grantedScopes := slackBotOAuthScopeValue()
	repo := &mockRepo{
		workspace:   slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:        slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:       []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:    []slackrepository.StatusRecord{{ID: statusID, Name: "Todo", Category: "unstarted"}},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID, Username: "actor", FullName: "Actor"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID: uuid.New(), WorkspaceID: workspaceID, SlackTeamID: "T123",
			InstallGeneration: uuid.New(), BotAccessToken: "xoxb-token", Scope: &grantedScopes, IsActive: true,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	stories := &mockStoryService{}
	requests := &mockRequestStore{}
	queue := &modalFileImportQueueStub{}
	service := newTestService(repo, requests, stories, Config{WebsiteURL: "https://fortyone.app"})
	service.outbound = newEventStoreStub()
	WithSlackFileImportQueue(queue)(service)
	return service, stories, requests, queue, workspaceID, teamID, statusID
}

func modalFileSubmissionPayload(t *testing.T, teamID uuid.UUID, statusValue string) interactionPayload {
	t.Helper()
	metadata, err := json.Marshal(slackModalPrivateMetadata{
		Source: requestSourceContext{
			SlackTeamID: "T123", SlackChannelID: "C123", SlackMessageTS: "171.101",
			SlackThreadTS: "171.100", SlackUserID: "U123",
		},
		SelectedTeamID: teamID.String(),
		SourceFiles:    []slackSourceFile{{ID: "F123", Name: "plan.pdf"}},
	})
	require.NoError(t, err)
	interaction := map[string]any{
		"type": "view_submission", "team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123"},
		"view": map[string]any{
			"id": "V123", "callback_id": "fortyone_create_task", "private_metadata": string(metadata),
			"state": map[string]any{"values": map[string]any{
				modalBlockTeam:  map[string]any{modalActionTeamSelect: map[string]any{"selected_option": map[string]any{"value": teamID.String()}}},
				modalBlockTitle: map[string]any{modalActionTitleInput: map[string]any{"value": "Ship feature"}},
				modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
					modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"selected_option": map[string]any{"value": statusValue}},
				},
				modalBlockSourceFiles: map[string]any{modalActionSourceFilesSelect: map[string]any{
					"selected_options": []map[string]any{{"value": "F123"}},
				}},
				modalBlockUploadFiles: map[string]any{modalActionUploadFilesInput: map[string]any{
					"type": "file_input", "files": []map[string]any{{"id": "F456", "name": "fresh.png"}},
				}},
			}},
		},
	}
	encoded, err := json.Marshal(interaction)
	require.NoError(t, err)
	var payload interactionPayload
	require.NoError(t, json.Unmarshal(encoded, &payload))
	return payload
}
