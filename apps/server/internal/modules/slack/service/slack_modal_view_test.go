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
	"time"

	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestBuildCreateTaskModalViewRefreshesTeamDependentFields(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	teamOneID := uuid.New()
	teamTwoID := uuid.New()
	teamTwoStatusID := uuid.New()
	teamTwoAssigneeID := uuid.New()
	teamTwoLabelID := uuid.New()
	teamTwoObjectiveID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamOneID, Code: "ENG", Name: "Engineering"},
			{ID: teamTwoID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamOneID: {{ID: uuid.New(), Name: "Triage", Category: "unstarted"}},
			teamTwoID: {{ID: teamTwoStatusID, Name: "Done", Category: "completed"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			teamOneID: {{UserID: actorID, Username: "actor", FullName: "Slack Actor", Email: "actor@example.com"}},
			teamTwoID: {
				{UserID: actorID, Username: "actor", FullName: "Slack Actor", Email: "actor@example.com"},
				{UserID: teamTwoAssigneeID, Username: "ops-user", FullName: "Ops User", Email: "ops@example.com"},
			},
		},
		labelsByTeam: map[uuid.UUID][]slackrepository.LabelRecord{
			teamTwoID: {{ID: teamTwoLabelID, Name: "Operations"}},
		},
		objectivesByTeam: map[uuid.UUID][]slackrepository.ObjectiveRecord{
			teamTwoID: {{ID: teamTwoObjectiveID, Name: "Ship reliability"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			ID:                uuid.New(),
			WorkspaceID:       workspaceID,
			SlackTeamID:       "T123",
			InstallGeneration: uuid.New(),
			IsActive:          true,
		},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Ship release",
		Description: "Ready to ship",
		Source: requestSourceContext{
			SlackTeamID:     "T123",
			SlackTeamDomain: "acme",
			SlackChannelID:  "C123",
			SlackMessageTS:  "171234.000100",
		},
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID:      teamTwoID,
			Priority:    "High",
			AssigneeID:  &teamTwoAssigneeID,
			LabelIDs:    []uuid.UUID{teamTwoLabelID},
			ObjectiveID: &teamTwoObjectiveID,
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	statusElement := findBlockElement(blocks, modalBlockStatus)
	statusOptions := statusElement["options"].([]map[string]any)
	require.Len(t, statusOptions, 2)
	require.Equal(t, slackRequestStatusValue, selectedOptionValue(t, statusOptions[0]))
	require.Equal(t, teamTwoStatusID.String(), selectedOptionValue(t, statusOptions[1]))

	assigneeElement := findBlockElement(blocks, modalBlockAssignee)
	require.Equal(t, "external_select", fmt.Sprint(assigneeElement["type"]))
	require.Equal(t, "2", fmt.Sprint(assigneeElement["min_query_length"]))
	initialAssignee := assigneeElement["initial_option"].(map[string]any)
	require.Equal(t, teamTwoAssigneeID.String(), selectedOptionValue(t, initialAssignee))

	labelsElement := findBlockElement(blocks, modalBlockLabels)
	require.Equal(t, "multi_external_select", fmt.Sprint(labelsElement["type"]))
	require.Equal(t, "2", fmt.Sprint(labelsElement["min_query_length"]))
	initialLabels := labelsElement["initial_options"].([]map[string]any)
	require.Len(t, initialLabels, 1)
	require.Equal(t, teamTwoLabelID.String(), selectedOptionValue(t, initialLabels[0]))

	objectiveElement := findBlockElement(blocks, modalBlockObjective)
	require.Equal(t, "external_select", fmt.Sprint(objectiveElement["type"]))
	initialObjective := objectiveElement["initial_option"].(map[string]any)
	require.Equal(t, teamTwoObjectiveID.String(), selectedOptionValue(t, initialObjective))
}

func TestBuildCreateTaskModalViewVersionsOnlyTeamDependentIDs(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	teamOneID := uuid.New()
	teamTwoID := uuid.New()
	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamOneID, Code: "ENG", Name: "Engineering"},
			{ID: teamTwoID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamOneID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
			teamTwoID: {{ID: uuid.New(), Name: "Triage", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	build := func(teamID uuid.UUID) []map[string]any {
		view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
			Title:       "Preserved title",
			Description: "Preserved description",
			WorkspaceID: workspaceID,
			ActorID:     actorID,
			Selection:   createTaskModalSelection{TeamID: teamID, Priority: "High"},
		})
		require.NoError(t, err)
		return view["blocks"].([]map[string]any)
	}

	teamOneBlocks := build(teamOneID)
	teamTwoBlocks := build(teamTwoID)
	teamTwoBlocksAgain := build(teamTwoID)

	for _, stableBlockID := range []string{
		modalBlockTeam,
		modalBlockTitle,
		modalBlockDescription,
		modalBlockPriority,
	} {
		teamOneBlock := findBlock(teamOneBlocks, stableBlockID)
		teamTwoBlock := findBlock(teamTwoBlocks, stableBlockID)
		require.Equal(t, stableBlockID, fmt.Sprint(teamOneBlock["block_id"]))
		require.Equal(t, fmt.Sprint(teamOneBlock["block_id"]), fmt.Sprint(teamTwoBlock["block_id"]))
		require.Equal(t, fmt.Sprint(teamOneBlock["element"].(map[string]any)["action_id"]), fmt.Sprint(teamTwoBlock["element"].(map[string]any)["action_id"]))
	}

	dependentIDs := []struct {
		blockID  string
		actionID string
	}{
		{blockID: modalBlockStatus, actionID: modalActionStatusSelect},
		{blockID: modalBlockAssignee, actionID: modalActionAssigneeSelect},
		{blockID: modalBlockLabels, actionID: modalActionLabelsMultiSelect},
		{blockID: modalBlockObjective, actionID: modalActionObjectiveSelect},
	}
	for _, dependent := range dependentIDs {
		teamOneBlock := findBlock(teamOneBlocks, dependent.blockID)
		teamTwoBlock := findBlock(teamTwoBlocks, dependent.blockID)
		teamTwoBlockAgain := findBlock(teamTwoBlocksAgain, dependent.blockID)

		teamOneBlockID := fmt.Sprint(teamOneBlock["block_id"])
		teamTwoBlockID := fmt.Sprint(teamTwoBlock["block_id"])
		teamOneActionID := fmt.Sprint(teamOneBlock["element"].(map[string]any)["action_id"])
		teamTwoActionID := fmt.Sprint(teamTwoBlock["element"].(map[string]any)["action_id"])

		require.Equal(t, modalTeamScopedID(dependent.blockID, teamOneID), teamOneBlockID)
		require.Equal(t, modalTeamScopedID(dependent.blockID, teamTwoID), teamTwoBlockID)
		require.Equal(t, modalTeamScopedID(dependent.actionID, teamOneID), teamOneActionID)
		require.Equal(t, modalTeamScopedID(dependent.actionID, teamTwoID), teamTwoActionID)
		require.NotEqual(t, teamOneBlockID, teamTwoBlockID)
		require.NotEqual(t, teamOneActionID, teamTwoActionID)
		require.Equal(t, teamTwoBlockID, fmt.Sprint(teamTwoBlockAgain["block_id"]))
		require.Equal(t, teamTwoActionID, fmt.Sprint(teamTwoBlockAgain["element"].(map[string]any)["action_id"]))
		require.LessOrEqual(t, len(teamTwoBlockID), modalElementIDMaxBytes)
		require.LessOrEqual(t, len(teamTwoActionID), modalElementIDMaxBytes)
	}
}

func TestModalTeamScopedIDIsDeterministicAndWithinSlackLimit(t *testing.T) {
	teamID := uuid.New()
	base := strings.Repeat("a", modalElementIDMaxBytes+50)

	first := modalTeamScopedID(base, teamID)
	second := modalTeamScopedID(base, teamID)

	require.Equal(t, first, second)
	require.Len(t, first, modalElementIDMaxBytes)
	require.True(t, strings.HasSuffix(first, modalTeamScopedIDSeparator+strings.ReplaceAll(teamID.String(), "-", "")))
}

func TestSlackOptionsRespectProviderLimits(t *testing.T) {
	option := toSlackOption(
		strings.Repeat("界", slackOptionTextMaxRunes+10),
		strings.Repeat("v", slackOptionValueMaxRunes+10),
	)
	require.Len(t, []rune(optionText(t, option)), slackOptionTextMaxRunes)
	require.Len(t, []rune(selectedOptionValue(t, option)), slackOptionValueMaxRunes)

	options := make([]map[string]any, 0, slackSelectMaxOptions+20)
	for index := 0; index < slackSelectMaxOptions+20; index++ {
		options = append(options, toSlackOption(fmt.Sprintf("Option %d", index), fmt.Sprintf("value-%d", index)))
	}
	block := selectInputBlock("provider_limit", "provider_limit_select", "Provider limit", options, nil, false, false)
	renderedOptions := block["element"].(map[string]any)["options"].([]map[string]any)
	require.Len(t, renderedOptions, slackSelectMaxOptions)

	response, err := interactionOptionsResponse(options)
	require.NoError(t, err)
	var responseBody struct {
		Options []map[string]any `json:"options"`
	}
	require.NoError(t, json.Unmarshal(response.Body, &responseBody))
	require.Len(t, responseBody.Options, slackSelectMaxOptions)
}

func TestBuildCreateTaskModalViewListsOnlyActorsTeams(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	allowedTeamID := uuid.New()
	blockedTeamID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: allowedTeamID, Code: "ENG", Name: "Engineering"},
			{ID: blockedTeamID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			allowedTeamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			allowedTeamID: {{UserID: actorID}},
			blockedTeamID: {{UserID: uuid.New()}},
		},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: blockedTeamID},
	})
	require.NoError(t, err)

	teamElement := findBlockElement(view["blocks"].([]map[string]any), modalBlockTeam)
	require.Equal(t, "external_select", teamElement["type"])
	require.NotContains(t, teamElement, "options")
	require.Equal(t, slackExternalSearchMinRunes, teamElement["min_query_length"])
	require.Equal(t, allowedTeamID.String(), selectedOptionValue(t, teamElement["initial_option"]))
}

func TestBuildCreateTaskModalViewKeepsAuthorizedTeamsReachableBeyondStaticLimit(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	teams := make([]slackrepository.TeamRecord, 0, slackSelectMaxOptions+25)
	for index := 0; index < slackSelectMaxOptions+25; index++ {
		teams = append(teams, slackrepository.TeamRecord{
			ID:   uuid.New(),
			Code: fmt.Sprintf("T%03d", index),
			Name: fmt.Sprintf("Authorized Team %03d", index),
		})
	}
	selectedTeam := teams[len(teams)-1]
	repo := &mockRepo{
		teams:       teams,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: selectedTeam.ID},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	teamBlock := findBlock(blocks, modalBlockTeam)
	teamElement := teamBlock["element"].(map[string]any)
	require.Equal(t, "external_select", teamElement["type"])
	require.Equal(t, slackExternalSearchMinRunes, teamElement["min_query_length"])
	require.NotContains(t, teamElement, "options")
	require.Equal(t, true, teamBlock["dispatch_action"])
	require.Equal(t, selectedTeam.ID.String(), selectedOptionValue(t, teamElement["initial_option"]))
}

func TestBuildCreateTaskModalViewShowsRequestAsFirstSyntheticStatus(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	toDoStatusID := uuid.New()
	actorID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamID, Code: "ENG", Name: "Engineering"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: toDoStatusID, Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID: teamID,
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	statusElement := findBlockElement(blocks, modalBlockStatus)
	statusOptions := statusElement["options"].([]map[string]any)
	require.Len(t, statusOptions, 2)
	require.Equal(t, slackRequestStatusValue, selectedOptionValue(t, statusOptions[0]))
	require.Equal(t, "Request", optionText(t, statusOptions[0]))
	require.Equal(t, toDoStatusID.String(), selectedOptionValue(t, statusOptions[1]))
	require.Equal(t, "To Do", optionText(t, statusOptions[1]))
	require.Equal(t, toDoStatusID.String(), selectedOptionValue(t, statusElement["initial_option"]))
}

func TestBuildCreateTaskModalViewUsesSearchableStatusWhenStaticLimitExceeded(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statuses := make([]slackrepository.StatusRecord, 0, slackSelectMaxOptions)
	for index := 0; index < slackSelectMaxOptions; index++ {
		statuses = append(statuses, slackrepository.StatusRecord{
			ID:       uuid.New(),
			Name:     fmt.Sprintf("Workflow Status %03d", index),
			Category: "unstarted",
		})
	}
	selectedStatus := statuses[len(statuses)-1]
	repo := &mockRepo{
		teams:       []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:    statuses,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID:     teamID,
			StatusKind: slackStatusKindStory,
			StatusID:   &selectedStatus.ID,
		},
	})
	require.NoError(t, err)

	statusElement := findBlockElement(view["blocks"].([]map[string]any), modalBlockStatus)
	require.Equal(t, "external_select", statusElement["type"])
	require.Equal(t, slackExternalSearchMinRunes, statusElement["min_query_length"])
	require.NotContains(t, statusElement, "options")
	require.Equal(t, selectedStatus.ID.String(), selectedOptionValue(t, statusElement["initial_option"]))
}

func TestBuildCreateTaskModalViewUsesStaticStatusAtProviderBoundary(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()
	statuses := make([]slackrepository.StatusRecord, 0, slackSelectMaxOptions-1)
	for index := 0; index < slackSelectMaxOptions-1; index++ {
		statuses = append(statuses, slackrepository.StatusRecord{
			ID:       uuid.New(),
			Name:     fmt.Sprintf("Workflow Status %03d", index),
			Category: "unstarted",
		})
	}
	repo := &mockRepo{
		teams:       []slackrepository.TeamRecord{{ID: teamID, Code: "ENG", Name: "Engineering"}},
		statuses:    statuses,
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection:   createTaskModalSelection{TeamID: teamID},
	})
	require.NoError(t, err)

	statusElement := findBlockElement(view["blocks"].([]map[string]any), modalBlockStatus)
	require.Equal(t, "static_select", statusElement["type"])
	require.Len(t, statusElement["options"], slackSelectMaxOptions)
	require.Equal(t, statuses[0].ID.String(), selectedOptionValue(t, statusElement["initial_option"]))
}

func TestBuildCreateTaskModalViewRendersExternalOptionalSelects(t *testing.T) {
	workspaceID := uuid.New()
	teamID := uuid.New()
	actorID := uuid.New()

	repo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: teamID, Code: "ENG", Name: "Engineering"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			teamID: {{ID: uuid.New(), Name: "To Do", Category: "unstarted"}},
		},
		teamMembers: []slackrepository.TeamMemberRecord{{UserID: actorID}},
	}

	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{WebsiteURL: "https://app.example.com"})

	view, err := service.buildCreateTaskModalView(context.Background(), createTaskModalViewInput{
		Title:       "Title",
		Description: "Description",
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Selection: createTaskModalSelection{
			TeamID: teamID,
		},
	})
	require.NoError(t, err)

	blocks := view["blocks"].([]map[string]any)
	assigneeElement := findBlockElement(blocks, modalBlockAssignee)
	require.Equal(t, "external_select", fmt.Sprint(assigneeElement["type"]))

	labelsElement := findBlockElement(blocks, modalBlockLabels)
	require.Equal(t, "multi_external_select", fmt.Sprint(labelsElement["type"]))

	objectiveElement := findBlockElement(blocks, modalBlockObjective)
	require.Equal(t, "external_select", fmt.Sprint(objectiveElement["type"]))
}

func TestHandleBlockActionsResetsTeamDependentSelections(t *testing.T) {
	workspaceID := uuid.New()
	actorID := uuid.New()
	oldTeamID := uuid.New()
	newTeamID := uuid.New()
	oldStatusID := uuid.New()
	newStatusID := uuid.New()
	oldAssigneeID := uuid.New()
	oldLabelID := uuid.New()
	oldObjectiveID := uuid.New()

	baseRepo := &mockRepo{
		teams: []slackrepository.TeamRecord{
			{ID: oldTeamID, Code: "ENG", Name: "Engineering"},
			{ID: newTeamID, Code: "OPS", Name: "Operations"},
		},
		statusesByTeam: map[uuid.UUID][]slackrepository.StatusRecord{
			oldTeamID: {{ID: oldStatusID, Name: "In Progress", Category: "started"}},
			newTeamID: {{ID: newStatusID, Name: "To Do", Category: "unstarted"}},
		},
		membersByTeam: map[uuid.UUID][]slackrepository.TeamMemberRecord{
			oldTeamID: {
				{UserID: actorID},
				{UserID: oldAssigneeID},
			},
			newTeamID: {{UserID: actorID}},
		},
		labelsByTeam: map[uuid.UUID][]slackrepository.LabelRecord{
			oldTeamID: {{ID: oldLabelID, Name: "Bug"}},
		},
		objectivesByTeam: map[uuid.UUID][]slackrepository.ObjectiveRecord{
			oldTeamID: {{ID: oldObjectiveID, Name: "Reliability"}},
		},
		slackWorkspace: slackrepository.SlackWorkspaceRecord{
			WorkspaceID:    workspaceID,
			SlackTeamID:    "T123",
			BotAccessToken: "xoxb-token",
		},
		slackUserLinks: map[string]uuid.UUID{"T123:U123": actorID},
	}
	repo := &blockingSlackWorkspaceRepo{
		mockRepo: baseRepo,
		started:  make(chan struct{}, 1),
		release:  make(chan struct{}),
	}
	service := newTestService(repo, &mockRequestStore{}, &mockStoryService{}, Config{})

	var updateRequest struct {
		View struct {
			Blocks []map[string]any `json:"blocks"`
		} `json:"view"`
	}
	updateResult := make(chan error, 1)
	service.client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://slack.com/api/views.update" {
			updateResult <- fmt.Errorf("unexpected Slack endpoint %q", req.URL.String())
		} else {
			updateResult <- json.NewDecoder(req.Body).Decode(&updateRequest)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	metadata, err := json.Marshal(slackModalPrivateMetadata{
		Source: requestSourceContext{
			SlackTeamID: "T123",
			SlackUserID: "U123",
		},
		SelectedTeamID: oldTeamID.String(),
	})
	require.NoError(t, err)
	interaction := map[string]any{
		"type": "block_actions",
		"team": map[string]any{"id": "T123"},
		"user": map[string]any{"id": "U123", "username": "joseph"},
		"view": map[string]any{
			"id":               "V123",
			"hash":             "hash",
			"callback_id":      "fortyone_create_task",
			"private_metadata": string(metadata),
			"state": map[string]any{
				"values": map[string]any{
					"team":        map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": newTeamID.String()}}},
					"title":       map[string]any{"value": map[string]any{"value": "Preserved title"}},
					"description": map[string]any{"value": map[string]any{"value": "Preserved description"}},
					modalTeamScopedID(modalBlockStatus, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionStatusSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": oldStatusID.String()}},
					},
					"priority": map[string]any{"value": map[string]any{"selected_option": map[string]any{"value": "High"}}},
					modalTeamScopedID(modalBlockAssignee, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionAssigneeSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": oldAssigneeID.String()}},
					},
					modalTeamScopedID(modalBlockLabels, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionLabelsMultiSelect, oldTeamID): map[string]any{"selected_options": []map[string]any{{"value": oldLabelID.String()}}},
					},
					modalTeamScopedID(modalBlockObjective, oldTeamID): map[string]any{
						modalTeamScopedID(modalActionObjectiveSelect, oldTeamID): map[string]any{"selected_option": map[string]any{"value": oldObjectiveID.String()}},
					},
				},
			},
		},
		"actions": []map[string]any{{
			"block_id":  modalBlockTeam,
			"action_id": modalActionTeamSelect,
			"selected_option": map[string]any{
				"value": newTeamID.String(),
			},
		}},
	}
	payloadBytes, err := json.Marshal(interaction)
	require.NoError(t, err)
	form := url.Values{}
	form.Set("payload", string(payloadBytes))

	requestCtx, cancelRequest := context.WithCancel(context.Background())
	resp, err := service.HandleInteractivity(requestCtx, []byte(form.Encode()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	select {
	case <-repo.started:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous Slack block action")
	}
	cancelRequest()
	close(repo.release)
	select {
	case updateErr := <-updateResult:
		require.NoError(t, updateErr)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for asynchronous Slack view update")
	}

	require.Equal(t, "Preserved title", findBlockElement(updateRequest.View.Blocks, modalBlockTitle)["initial_value"])
	require.Equal(t, "Preserved description", findBlockElement(updateRequest.View.Blocks, modalBlockDescription)["initial_value"])
	require.Equal(t, "High", selectedOptionValue(t, findBlockElement(updateRequest.View.Blocks, modalBlockPriority)["initial_option"]))
	require.Equal(t, newTeamID.String(), selectedOptionValue(t, findBlockElement(updateRequest.View.Blocks, modalBlockTeam)["initial_option"]))
	require.Equal(t, newStatusID.String(), selectedOptionValue(t, findBlockElement(updateRequest.View.Blocks, modalBlockStatus)["initial_option"]))
	require.NotContains(t, findBlockElement(updateRequest.View.Blocks, modalBlockAssignee), "initial_option")
	require.NotContains(t, findBlockElement(updateRequest.View.Blocks, modalBlockLabels), "initial_options")
	require.NotContains(t, findBlockElement(updateRequest.View.Blocks, modalBlockObjective), "initial_option")
	for _, dependentBlockID := range []string{modalBlockStatus, modalBlockAssignee, modalBlockLabels, modalBlockObjective} {
		updatedBlockID := fmt.Sprint(findBlock(updateRequest.View.Blocks, dependentBlockID)["block_id"])
		require.Equal(t, modalTeamScopedID(dependentBlockID, newTeamID), updatedBlockID)
		require.NotEqual(t, modalTeamScopedID(dependentBlockID, oldTeamID), updatedBlockID)
	}
}
