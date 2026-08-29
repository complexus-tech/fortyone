package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHandleViewSubmissionCreatesSlackRequestWhenRequestStatusSelected(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installeBy := uuid.New()
	actorID := uuid.New()
	labelID := uuid.New()
	objectiveID := uuid.New()
	installGeneration := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:  []slackrepository.StatusRecord{{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: actorID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"},
		},
		labels:     []slackrepository.LabelRecord{{ID: labelID, Name: "Bug"}},
		objectives: []slackrepository.ObjectiveRecord{{ID: objectiveID, Name: "Improve reliability"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installeBy,
			InstallGeneration: installGeneration,
			IsActive:          true,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	requests := &mockRequestStore{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})
	store := newEventStoreStub()
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"user": map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": teamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Fix login bug"}},
					"description": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "User cannot log in from iOS"}},
					modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
						modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"type": "static_select", "selected_option": map[string]any{"value": slackRequestStatusValue}},
					},
					modalTeamScopedID(modalBlockAssignee, teamID): map[string]any{
						modalTeamScopedID(modalActionAssigneeSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": actorID.String()}},
					},
					modalTeamScopedID(modalBlockLabels, teamID): map[string]any{
						modalTeamScopedID(modalActionLabelsMultiSelect, teamID): map[string]any{"type": "multi_external_select", "selected_options": []map[string]any{{"value": labelID.String()}}},
					},
					modalTeamScopedID(modalBlockObjective, teamID): map[string]any{
						modalTeamScopedID(modalActionObjectiveSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": objectiveID.String()}},
					},
					"priority": map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": "High"}}},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), `"response_action":"clear"`)

	require.Equal(t, providerSlack, requests.last.Provider)
	require.Equal(t, SourceTypeSlackMessage, requests.last.SourceType)
	require.Equal(t, "Fix login bug", requests.last.Title)
	require.Equal(t, workspaceID, requests.last.WorkspaceID)
	require.Equal(t, teamID, requests.last.TeamID)
	require.Equal(t, "High", requests.last.Priority)
	require.Equal(t, &actorID, requests.last.AssigneeID)
	require.Equal(t, &objectiveID, requests.last.ObjectiveID)
	require.Equal(t, []uuid.UUID{labelID}, requests.last.LabelIDs)
	require.Equal(t, &actorID, requests.last.CreatedByUserID)
	require.Equal(t, []string{labelID.String()}, requests.last.Metadata["label_ids"])
	require.NotNil(t, requests.last.SourceURL)
	require.True(t, strings.Contains(*requests.last.SourceURL, "acme.slack.com/archives/C123"))
	require.Equal(t, 1, requests.bindCalls)
	require.Equal(t, "T123", requests.lastThread.ExternalWorkspaceID)
	require.Equal(t, "C123", requests.lastThread.ExternalChannelID)
	require.Equal(t, "171.200", requests.lastThread.ExternalThreadID)
	require.Len(t, store.outboundInputs, 1)
	require.Equal(
		t,
		fmt.Sprintf("Joseph Mukorivo <https://acme.fortyone.app/teams/%s/requests/%s|opened a request>", teamID, requests.lastCreated.ID),
		store.outboundInputs[0].Content,
	)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.NotNil(t, providerPayload.Metadata)
	require.Len(t, providerPayload.Metadata.Entities, 1)
	requestEntity := providerPayload.Metadata.Entities[0]
	require.Equal(t, slackRequestExternalRefType, requestEntity.ExternalRef.Type)
	require.Equal(t, fmt.Sprintf("https://acme.fortyone.app/teams/%s/requests/%s", teamID, requests.lastCreated.ID), requestEntity.URL)
	require.Equal(t, "Fix login bug", requestEntity.EntityPayload.Attributes.Title.Text)
	require.Equal(t, "U123", requestEntity.EntityPayload.Fields["created_by"].User.UserID)
	require.Equal(t, "Pending", requestEntity.EntityPayload.Fields["status"].Value)
	require.Nil(t, requestEntity.EntityPayload.Attributes.Title.Edit)
	require.NotNil(t, providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlLinks)
	require.NotNil(t, providerPayload.UnfurlMedia)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestHandleViewSubmissionCreatesStoryWhenNonTriageStatusSelected(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	triageStatusID := uuid.New()
	doneStatusID := uuid.New()
	installedBy := uuid.New()
	mappedActorID := uuid.New()
	assigneeID := uuid.New()
	labelID := uuid.New()
	objectiveID := uuid.New()
	installGeneration := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses: []slackrepository.StatusRecord{
			{ID: triageStatusID, Name: "Triage", Category: "unstarted"},
			{ID: doneStatusID, Name: "Done", Category: "completed"},
		},
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: mappedActorID, Username: "actor", FullName: "Slack Actor", Email: "actor@example.com"},
			{UserID: assigneeID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"},
		},
		labels:     []slackrepository.LabelRecord{{ID: labelID, Name: "Bug"}},
		objectives: []slackrepository.ObjectiveRecord{{ID: objectiveID, Name: "Improve reliability"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
			InstallGeneration: installGeneration,
			IsActive:          true,
		},
		slackUserLinks: map[string]uuid.UUID{
			"T123:U999": mappedActorID,
		},
	}
	requests := &mockRequestStore{}
	storyService := &mockStoryService{sequenceID: 123}
	service := newTestService(repo, requests, storyService, Config{WebsiteURL: "https://fortyone.app"})
	store := newEventStoreStub()
	service.outbound = store
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://slack.com/api/chat.postMessage", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true,"ts":"171.200"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"user": map[string]any{"id": "U999", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme","slack_channel_id":"C123","slack_message_ts":"171234.000100","slack_user_id":"U999","slack_username":"joseph"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": teamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Ship release"}},
					"description": map[string]any{"value": map[string]any{"type": "plain_text_input", "value": "Ready to ship"}},
					modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
						modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"type": "static_select", "selected_option": map[string]any{"value": doneStatusID.String()}},
					},
					modalTeamScopedID(modalBlockAssignee, teamID): map[string]any{
						modalTeamScopedID(modalActionAssigneeSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": assigneeID.String()}},
					},
					modalTeamScopedID(modalBlockLabels, teamID): map[string]any{
						modalTeamScopedID(modalActionLabelsMultiSelect, teamID): map[string]any{"type": "multi_external_select", "selected_options": []map[string]any{{"value": labelID.String()}}},
					},
					modalTeamScopedID(modalBlockObjective, teamID): map[string]any{
						modalTeamScopedID(modalActionObjectiveSelect, teamID): map[string]any{"type": "external_select", "selected_option": map[string]any{"value": objectiveID.String()}},
					},
					"priority": map[string]any{"value": map[string]any{"type": "static_select", "selected_option": map[string]any{"value": "Urgent"}}},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)

	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), `"response_action":"clear"`)

	require.Equal(t, "", requests.last.Provider)
	require.Equal(t, mappedActorID, storyService.lastActorID)
	require.Equal(t, teamID, storyService.lastStory.Team)
	require.NotNil(t, storyService.lastStory.Status)
	require.Equal(t, doneStatusID, *storyService.lastStory.Status)
	require.NotNil(t, storyService.lastStory.Assignee)
	require.Equal(t, assigneeID, *storyService.lastStory.Assignee)
	require.NotNil(t, storyService.lastStory.Objective)
	require.Equal(t, objectiveID, *storyService.lastStory.Objective)
	require.Equal(t, "Urgent", storyService.lastStory.Priority)
	require.Equal(t, []uuid.UUID{labelID}, storyService.lastStory.LabelIDs)
	require.Empty(t, repo.lastStoryLink.url, "direct stories must not create a Slack source association")
	require.Len(t, store.outboundInputs, 1)
	require.Equal(t, "Slack Actor created <https://acme.fortyone.app/work/ENG-123|ENG-123>", store.outboundInputs[0].Content)
	require.Equal(t, "C123", store.outboundInputs[0].ExternalChannelID)
	require.Empty(t, store.outboundInputs[0].ExternalThreadID)
	providerPayload, err := DecodeSlackProviderPayload(store.outboundInputs[0].ProviderPayload)
	require.NoError(t, err)
	require.Len(t, providerPayload.Metadata.Entities, 1)
	require.Equal(t, "https://acme.fortyone.app/work/ENG-123", providerPayload.Metadata.Entities[0].URL)
	require.False(t, *providerPayload.UnfurlLinks)
	require.False(t, *providerPayload.UnfurlMedia)
}

func TestHandleViewSubmissionRejectsTeamOutsideActorsMembership(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()
	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID}},
			blockedTeamID: {{UserID: uuid.New()}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	requests := &mockRequestStore{}
	storyService := &mockStoryService{}
	service := newTestService(repo, requests, storyService, Config{})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
			"state": map[string]any{"values": map[string]any{
				"team":     map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": blockedTeamID.String()}}},
				"title":    map[string]any{"value": map[string]any{"value": "Unauthorized task"}},
				"status":   map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": slackRequestStatusValue}}},
				"priority": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": "High"}}},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), modalBlockTeam)
	require.Contains(t, string(resp.Body), "no longer available to you")
	require.Zero(t, requests.calls)
	require.Zero(t, storyService.createCalls)
}

func TestHandleViewSubmissionValidatesLabelsBeforeCreatingStory(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statusID := uuid.New()
	staleLabelID := uuid.New()
	repo := &mockRepo{
		workspace:   slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:       []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:    []slackrepository.StatusRecord{{ID: statusID, Name: "To Do", Category: "unstarted"}},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		labels:      []slackrepository.LabelRecord{{ID: uuid.New(), Name: "Current label"}},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	storyService := &mockStoryService{}
	service := newTestService(repo, &mockRequestStore{}, storyService, Config{})

	interaction := map[string]any{
		"type": "view_submission",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
			"state": map[string]any{"values": map[string]any{
				"team":     map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": teamID.String()}}},
				"title":    map[string]any{"value": map[string]any{"value": "Task with stale label"}},
				"status":   map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": statusID.String()}}},
				"labels":   map[string]any{"value": map[string]any{"selected_options": []map[string]any{{"value": staleLabelID.String()}}}},
				"priority": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": "High"}}},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), modalBlockLabels)
	require.Zero(t, storyService.createCalls)
}

func TestHandleViewSubmissionRejectsUnavailableScopedAssigneeAndObjective(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statusID := uuid.New()

	tests := []struct {
		name          string
		blockID       string
		actionID      string
		selectedValue uuid.UUID
		errorText     string
	}{
		{
			name:          "assignee",
			blockID:       modalBlockAssignee,
			actionID:      modalActionAssigneeSelect,
			selectedValue: uuid.New(),
			errorText:     "Selected assignee is no longer available",
		},
		{
			name:          "objective",
			blockID:       modalBlockObjective,
			actionID:      modalActionObjectiveSelect,
			selectedValue: uuid.New(),
			errorText:     "Selected objective is no longer available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
				teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
				statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
					teamID: {{ID: statusID, Name: "To Do", Category: "unstarted"}},
				},
				teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
				slackWorkspace: slackrepository.SlackWorkspaceRecord{
					WorkspaceID:    workspaceID,
					SlackTeamID:    "T123",
					BotAccessToken: "xoxb-token",
				},
				slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
			}
			storyService := &mockStoryService{}
			service := newTestService(repo, &mockRequestStore{}, storyService, Config{})
			selectedBlockID := modalTeamScopedID(tt.blockID, teamID)
			selectedActionID := modalTeamScopedID(tt.actionID, teamID)
			interaction := map[string]any{
				"type": "view_submission",
				"team": map[string]any{"id": "T123"},
				"user": map[string]any{"id": "U123"},
				"view": map[string]any{
					"callback_id":      "fortyone_create_task",
					"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` + teamID.String() + `"}`,
					"state": map[string]any{"values": map[string]any{
						modalBlockTeam: map[string]any{
							modalActionTeamSelect: map[string]any{"selected_option": map[string]any{"value": teamID.String()}},
						},
						modalBlockTitle: map[string]any{
							modalActionTitleInput: map[string]any{"value": "Do not coerce selection"},
						},
						modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
							modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"selected_option": map[string]any{"value": statusID.String()}},
						},
						selectedBlockID: map[string]any{
							selectedActionID: map[string]any{"selected_option": map[string]any{"value": tt.selectedValue.String()}},
						},
					}},
				},
			}
			payloadBytes, err := json.Marshal(interaction)
			require.NoError(t, err)
			form := url.Values{}
			form.Set("payload", string(payloadBytes))

			response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			var responseBody struct {
				ResponseAction string            `json:"response_action"`
				Errors         map[string]string `json:"errors"`
			}
			require.NoError(t, json.Unmarshal(response.Body, &responseBody))
			require.Equal(t, "errors", responseBody.ResponseAction)
			require.Equal(t, map[string]string{selectedBlockID: tt.errorText}, responseBody.Errors)
			require.Zero(t, storyService.createCalls)
		})
	}
}
