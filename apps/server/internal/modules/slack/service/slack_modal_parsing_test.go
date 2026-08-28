package slack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseCommandTitleSupportsCreateTaskPrefix(t *testing.T) {
	title := parseCommandTitle("create task Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("Improve onboarding")
	require.Equal(t, "Improve onboarding", title)

	title = parseCommandTitle("")
	require.Equal(t, "New task", title)
}

func TestParseViewSubmissionReadsTeamScopedDependentIDs(t *testing.T) {
	teamID := uuid.New()
	statusID := uuid.New()
	assigneeID := uuid.New()
	labelID := uuid.New()
	objectiveID := uuid.New()
	interaction := map[string]any{
		"view": map[string]any{
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` + teamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{
				modalBlockTeam: map[string]any{
					modalActionTeamSelect: map[string]any{"selected_option": map[string]any{"value": teamID.String()}},
				},
				modalBlockTitle: map[string]any{
					modalActionTitleInput: map[string]any{"value": "Scoped task"},
				},
				modalBlockDescription: map[string]any{
					modalActionDescriptionInput: map[string]any{"value": "Scoped description"},
				},
				modalTeamScopedID(modalBlockStatus, teamID): map[string]any{
					modalTeamScopedID(modalActionStatusSelect, teamID): map[string]any{"selected_option": map[string]any{"value": statusID.String()}},
				},
				modalTeamScopedID(modalBlockAssignee, teamID): map[string]any{
					modalTeamScopedID(modalActionAssigneeSelect, teamID): map[string]any{"selected_option": map[string]any{"value": assigneeID.String()}},
				},
				modalTeamScopedID(modalBlockLabels, teamID): map[string]any{
					modalTeamScopedID(modalActionLabelsMultiSelect, teamID): map[string]any{"selected_options": []map[string]any{{"value": labelID.String()}}},
				},
				modalTeamScopedID(modalBlockObjective, teamID): map[string]any{
					modalTeamScopedID(modalActionObjectiveSelect, teamID): map[string]any{"selected_option": map[string]any{"value": objectiveID.String()}},
				},
				modalBlockPriority: map[string]any{
					modalActionPrioritySelect: map[string]any{"selected_option": map[string]any{"value": "High"}},
				},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	var payload interactionPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))

	submission, err := parseViewSubmission(payload)
	require.NoError(t, err)
	require.Equal(t, "Scoped task", submission.Title)
	require.Equal(t, "Scoped description", submission.Description)
	require.Equal(t, teamID, submission.TeamID)
	require.Equal(t, slackStatusKindStory, submission.StatusKind)
	require.Equal(t, statusID, *submission.StatusID)
	require.Equal(t, assigneeID, *submission.AssigneeID)
	require.Equal(t, []uuid.UUID{labelID}, submission.LabelIDs)
	require.Equal(t, objectiveID, *submission.ObjectiveID)
	require.Equal(t, "High", submission.Priority)
	require.Equal(t, modalTeamScopedID(modalBlockStatus, teamID), submission.BlockIDs.Status)
	require.Equal(t, modalTeamScopedID(modalBlockAssignee, teamID), submission.BlockIDs.Assignee)
	require.Equal(t, modalTeamScopedID(modalBlockLabels, teamID), submission.BlockIDs.Labels)
	require.Equal(t, modalTeamScopedID(modalBlockObjective, teamID), submission.BlockIDs.Objective)
}

func TestParseViewSubmissionIgnoresStaleDependentIDsFromPreviousTeam(t *testing.T) {
	oldTeamID := uuid.New()
	newTeamID := uuid.New()
	interaction := map[string]any{
		"view": map[string]any{
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` + oldTeamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{
				modalBlockTeam: map[string]any{
					modalActionTeamSelect: map[string]any{"selected_option": map[string]any{"value": newTeamID.String()}},
				},
				modalBlockTitle: map[string]any{
					modalActionTitleInput: map[string]any{"value": "Preserved title"},
				},
				modalTeamScopedID(modalBlockStatus, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionStatusSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": uuid.NewString()}},
				},
				modalTeamScopedID(modalBlockAssignee, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionAssigneeSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": uuid.NewString()}},
				},
				modalTeamScopedID(modalBlockLabels, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionLabelsMultiSelect, oldTeamID): map[string]any{"selected_options": []map[string]any{{"value": uuid.NewString()}}},
				},
				modalTeamScopedID(modalBlockObjective, oldTeamID): map[string]any{
					modalTeamScopedID(modalActionObjectiveSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": uuid.NewString()}},
				},
				modalBlockPriority: map[string]any{
					modalActionPrioritySelect: map[string]any{"selected_option": map[string]any{"value": "Urgent"}},
				},
			}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	var payload interactionPayload
	require.NoError(t, json.Unmarshal(payloadBytes, &payload))

	submission, err := parseViewSubmission(payload)
	require.NoError(t, err)
	require.Equal(t, newTeamID, submission.TeamID)
	require.Equal(t, "Preserved title", submission.Title)
	require.Equal(t, "Urgent", submission.Priority)
	require.Equal(t, slackStatusKindRequest, submission.StatusKind)
	require.Nil(t, submission.StatusID)
	require.Nil(t, submission.AssigneeID)
	require.Empty(t, submission.LabelIDs)
	require.Nil(t, submission.ObjectiveID)
	require.Equal(t, modalTeamScopedID(modalBlockStatus, newTeamID), submission.BlockIDs.Status)
	require.Equal(t, modalTeamScopedID(modalBlockAssignee, newTeamID), submission.BlockIDs.Assignee)
	require.Equal(t, modalTeamScopedID(modalBlockLabels, newTeamID), submission.BlockIDs.Labels)
	require.Equal(t, modalTeamScopedID(modalBlockObjective, newTeamID), submission.BlockIDs.Objective)
}

func TestHandleViewSubmissionDefaultsClearedStatusToRequest(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installedBy := uuid.New()
	actorID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		team:      slackrepository.TeamRecord{ID: teamID, Code: "ENG", Name: "Engineering"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: actorID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}

	requests := &mockRequestStore{}
	service := newTestService(repo, requests, &mockStoryService{}, Config{WebsiteURL: "https://fortyone.app"})

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
	require.Equal(t, teamID, requests.last.TeamID)
}

func TestBuildWorkspaceURLSupportsSubdomainsAndLocalhost(t *testing.T) {
	t.Run("hosted_url_uses_workspace_subdomain", func(t *testing.T) {
		integrationURL := buildWorkspaceURL("https://fortyone.app", "acme", "settings", "workspace", "integrations", "slack")
		require.Equal(t, "https://acme.fortyone.app/settings/workspace/integrations/slack", integrationURL)

		accountIntegrationURL := buildWorkspaceURL("https://fortyone.app", "acme", "settings", "integrations", "slack")
		require.Equal(t, "https://acme.fortyone.app/settings/integrations/slack", accountIntegrationURL)

		taskURL := buildTaskURL("https://fortyone.app", "acme", "PRD-571")
		require.Equal(t, "https://acme.fortyone.app/work/PRD-571", taskURL)

		requestURL := buildRequestURL("https://fortyone.app", "acme", "team-1", "req-1")
		require.Equal(t, "https://acme.fortyone.app/teams/team-1/requests/req-1", requestURL)
	})

	t.Run("localhost_url_uses_workspace_path_prefix", func(t *testing.T) {
		integrationURL := buildWorkspaceURL("http://localhost:3000", "acme", "settings", "workspace", "integrations", "slack")
		require.Equal(t, "http://localhost:3000/acme/settings/workspace/integrations/slack", integrationURL)

		accountIntegrationURL := buildWorkspaceURL("http://localhost:3000", "acme", "settings", "integrations", "slack")
		require.Equal(t, "http://localhost:3000/acme/settings/integrations/slack", accountIntegrationURL)
	})
}

func TestBuildPrefilledDescriptionUsesLinearStyleFormat(t *testing.T) {
	description := buildPrefilledDescription(requestSourceContext{
		SlackUserID:   "U12345",
		SlackUsername: "joseph",
		SlackText:     "hey",
	})
	require.Equal(t, "@[joseph](U12345) said:\n> hey", description)
}

func TestBuildCreateTaskModalViewMarksDescriptionOptional(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: teamID},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	descriptionBlock := findBlock(blocks, modalBlockDescription)
	require.Equal(t, true, descriptionBlock["optional"])
}

func TestBuildCreateTaskModalViewPreservesPersonalTeamOrdering(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	firstTeam := slackrepository.TeamRecord{ID: uuid.New(), Code: "WEB", Name: "Web"}
	secondTeam := slackrepository.TeamRecord{ID: uuid.New(), Code: "ENG", Name: "Engineering"}
	repo := &mockRepo{
		// ListWorkspaceTeamsForUser returns the repository's personal order.
		// Keep it deliberately different from alphabetical order.
		teams:       []slackrepository.TeamRecord{firstTeam, secondTeam},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			firstTeam.ID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Ordered story",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Source: requestSourceContext{
			SlackTeamID:    "T123",
			SlackChannelID: "C123",
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	teamElement := findBlockElement(blocks, modalBlockTeam)
	require.Equal(t, firstTeam.ID.String(), selectedOptionValue(t, teamElement["initial_option"]))

	options := slackTeamSuggestionOptions([]slackrepository.TeamRecord{firstTeam, secondTeam}, "")
	require.Len(t, options, 2)
	require.Equal(t, firstTeam.ID.String(), selectedOptionValue(t, options[0]))
	require.Equal(t, secondTeam.ID.String(), selectedOptionValue(t, options[1]))
}
