package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHandleInteractivityBlockSuggestionSearchesAllAuthorizedTeams(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	blockedTeamID := uuid.New()
	teams := make([]slackrepository.TeamRecord, 0, slackSelectMaxOptions+22)
	for index := 0; index < slackSelectMaxOptions+21; index++ {
		teams = append(teams, slackrepository.TeamRecord{
			ID:   uuid.New(),
			Code: fmt.Sprintf("T%03d", index),
			Name: fmt.Sprintf("Authorized Team %03d", index),
		})
	}
	targetTeam := teams[len(teams)-1]
	teams = append(teams, slackrepository.TeamRecord{ID: blockedTeamID, Code: "PRIVATE", Name: "Restricted Team"})
	repo := &mockRepo{
		teams:       teams,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			blockedTeamID: {{UserID: uuid.New()}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	search := func(query string) []map[string]any {
		t.Helper()
		interaction := map[string]any{
			"type":      "block_suggestion",
			"action_id": modalActionTeamSelect,
			"block_id":  modalBlockTeam,
			"value":     query,
			"team":      map[string]any{"id": "T123"},
			"user":      map[string]any{"id": "U123"},
			"view": map[string]any{
				"callback_id":      "fortyone_create_task",
				"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
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
			Options []map[string]any `json:"options"`
		}
		require.NoError(t, json.Unmarshal(response.Body, &responseBody))
		return responseBody.Options
	}

	require.Len(t, search(""), slackSelectMaxOptions)
	targetOptions := search(targetTeam.Code)
	require.Len(t, targetOptions, 1)
	require.Equal(t, targetTeam.ID.String(), selectedOptionValue(t, targetOptions[0]))
	require.Empty(t, search("Restricted Team"))
}

func TestHandleInteractivityBlockSuggestionSearchesOverflowStatuses(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statuses := make([]slackrepository.StatusRecord, 0, slackSelectMaxOptions+20)
	for index := 0; index < slackSelectMaxOptions+20; index++ {
		statuses = append(statuses, slackrepository.StatusRecord{
			ID:       uuid.New(),
			Name:     fmt.Sprintf("Workflow Status %03d", index),
			Category: "unstarted",
		})
	}
	targetStatus := statuses[len(statuses)-1]
	repo := &mockRepo{
		teams:    []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses: statuses,
		teamMembers: []slackrepository.TeamMemberRecord{
			{UserID: actorID},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	search := func(query string) []map[string]any {
		t.Helper()
		interaction := map[string]any{
			"type":      "block_suggestion",
			"action_id": modalTeamScopedID(modalActionStatusSelect, teamID),
			"block_id":  modalTeamScopedID(modalBlockStatus, teamID),
			"value":     query,
			"team":      map[string]any{"id": "T123"},
			"user":      map[string]any{"id": "U123"},
			"view": map[string]any{
				"callback_id": "fortyone_create_task",
				"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` +
					teamID.String() + `"}`,
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
			Options []map[string]any `json:"options"`
		}
		require.NoError(t, json.Unmarshal(response.Body, &responseBody))
		return responseBody.Options
	}

	initialOptions := search("")
	require.Len(t, initialOptions, slackSelectMaxOptions)
	require.Equal(t, slackRequestStatusValue, selectedOptionValue(t, initialOptions[0]))
	targetOptions := search(targetStatus.Name)
	require.Len(t, targetOptions, 1)
	require.Equal(t, targetStatus.ID.String(), selectedOptionValue(t, targetOptions[0]))
}

func TestHandleInteractivityBlockSuggestionReturnsTeamScopedAssigneeOptions(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installedBy := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalTeamScopedID(modalActionAssigneeSelect, teamID),
		"block_id":  modalTeamScopedID(modalBlockAssignee, teamID),
		"value":     "jo",
		"team":      map[string]any{"id": "T123", "domain": "acme"},
		"user":      map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team": map[string]any{
						"value": map[string]any{
							"type":            "static_select",
							"selected_option": map[string]any{"value": teamID.String()},
						},
					},
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
	require.Equal(t, "application/json", resp.ContentType)
	require.Contains(t, string(resp.Body), memberID.String())
}

func TestHandleInteractivityBlockSuggestionUsesScopedActionTeamWhenMetadataIsStale(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	oldTeamID := uuid.New()
	selectedTeamID := uuid.New()
	selectedMemberID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: oldTeamID, Code: "OLD", Name: "Old Team"},
			{ID: selectedTeamID, Code: "NEW", Name: "Selected Team"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			oldTeamID: {{UserID: actorID, FullName: "Slack Actor"}},
			selectedTeamID: {
				{UserID: actorID, FullName: "Slack Actor"},
				{UserID: selectedMemberID, FullName: "Joseph Selected"},
			},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalTeamScopedID(modalActionAssigneeSelect, selectedTeamID),
		"block_id":  modalTeamScopedID(modalBlockAssignee, selectedTeamID),
		"value":     "jo",
		"team":      map[string]any{"id": "T123"},
		"user":      map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id": "fortyone_create_task",
			// Slack can preserve metadata from the prior view while dispatching
			// from the newly rendered, team-scoped input.
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"},"selected_team_id":"` +
				oldTeamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{}},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	response, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Contains(t, string(response.Body), selectedMemberID.String())
}

func TestHandleInteractivityBlockSuggestionRejectsTeamOutsideActorsMembership(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()
	blockedMemberID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID, FullName: "Slack Actor"}},
			blockedTeamID: {{UserID: blockedMemberID, FullName: "Joseph Blocked"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID: workspaceID,
			SlackTeamID: "T123",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"value":     "jo",
		"team":      map[string]any{"id": "T123"},
		"user":      map[string]any{"id": "U123"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_user_id":"U123"}}`,
			"state": map[string]any{"values": map[string]any{
				"team": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": blockedTeamID.String()}}},
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
	require.JSONEq(t, `{"options":[]}`, string(resp.Body))
	require.NotContains(t, string(resp.Body), blockedMemberID.String())
}

func TestHandleInteractivityBlockSuggestionUsesViewInitialTeamWhenStateEmpty(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	installedBy := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			SlackTeamDomain:   "acme",
			BotAccessToken:    "xoxb-token",
			InstalledByUserID: &installedBy,
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"block_id":  modalBlockAssignee,
		"value":     "jo",
		"team":      map[string]any{"id": "T123", "domain": "acme"},
		"user":      map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"blocks": []map[string]any{
				{
					"block_id": modalBlockTeam,
					"element": map[string]any{
						"type":      "static_select",
						"action_id": modalActionTeamSelect,
						"initial_option": map[string]any{
							"value": teamID.String(),
						},
					},
				},
			},
			"state": map[string]any{
				"values": map[string]any{},
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
	require.Contains(t, string(resp.Body), memberID.String())
}

func TestHandleInteractivityBlockSuggestionReturnsNoOptionsBeforeTwoCharacters(t *testing.T) {
	teamID := uuid.New()
	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"value":     "j",
		"view": map[string]any{
			"callback_id":      "fortyone_create_task",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"state": map[string]any{
				"values": map[string]any{
					"team": map[string]any{
						"value": map[string]any{
							"type":            "static_select",
							"selected_option": map[string]any{"value": teamID.String()},
						},
					},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	service := newTestService(&mockRepo{}, &mockRequestStore{}, &mockStoryService{}, Config{})
	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.JSONEq(t, `{"options":[]}`, string(resp.Body))
}

func TestHandleInteractivityBlockSuggestionUsesActionFallbackFromActionsArray(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:     workspaceID,
			SlackTeamID:     "T123",
			SlackTeamDomain: "acme",
			BotAccessToken:  "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type": "block_suggestion",
		"team": map[string]any{"id": "T123", "domain": "acme"},
		"user": map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id":      "",
			"private_metadata": `{"slack_team_id":"T123","slack_team_domain":"acme"}`,
			"blocks": []map[string]any{
				{
					"block_id": modalBlockTeam,
					"element": map[string]any{
						"type":      "static_select",
						"action_id": modalActionTeamSelect,
						"initial_option": map[string]any{
							"value": teamID.String(),
						},
					},
				},
			},
			"state": map[string]any{"values": map[string]any{}},
		},
		"actions": []map[string]any{
			{
				"action_id": modalActionAssigneeSelect,
				"value":     "jo",
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
	require.Contains(t, string(resp.Body), memberID.String())
}

func TestHandleInteractivityBlockSuggestionUsesSelectedTeamFromPrivateMetadata(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	memberID := uuid.New()

	repo := &mockRepo{
		workspace: slackrepository.WorkspaceRecord{ID: workspaceID, Slug: "acme", Name: "Acme"},
		teams:     []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamID: {{UserID: memberID, Username: "joseph", FullName: "Joseph Mukorivo", Email: "joseph@example.com"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:     workspaceID,
			SlackTeamID:     "T123",
			SlackTeamDomain: "acme",
			BotAccessToken:  "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": memberID},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	interaction := map[string]any{
		"type":      "block_suggestion",
		"action_id": modalActionAssigneeSelect,
		"value":     "jo",
		"team":      map[string]any{"id": "T123", "domain": "acme"},
		"user":      map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"callback_id": "fortyone_create_task",
			"private_metadata": `{"source":{"slack_team_id":"T123","slack_team_domain":"acme"},"selected_team_id":"` +
				teamID.String() + `"}`,
			"state": map[string]any{"values": map[string]any{}},
		},
	}

	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	resp, err := service.HandleInteractivity(context.Background(), []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, string(resp.Body), memberID.String())
}
