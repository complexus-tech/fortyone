package mayahttp

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

func (h *Handlers) executeListMyTasks(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListMyTasksArguments{Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_my_tasks arguments: %w", err)
		}
	}
	if args.Limit <= 0 {
		args.Limit = 10
	}
	if args.Limit > 25 {
		args.Limit = 25
	}

	statuses, err := h.states.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list statuses: %w", err)
	}
	statusesByID := make(map[uuid.UUID]states.CoreState, len(statuses))
	for _, status := range statuses {
		statusesByID[status.ID] = status
	}

	allStories, err := h.stories.MyStories(ctx, workspaceID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list my stories: %w", err)
	}

	voiceStories := make([]AppRealtimeVoiceStory, 0, min(args.Limit, len(allStories)))
	for _, story := range allStories {
		if !args.IncludeCompleted && story.CompletedAt != nil {
			continue
		}
		if story.DeletedAt != nil || story.ArchivedAt != nil {
			continue
		}
		voiceStories = append(voiceStories, toRealtimeVoiceStory(story, statusesByID))
		if len(voiceStories) >= args.Limit {
			break
		}
	}

	terminology := h.realtimeTerminology(ctx, workspaceID)
	message := fmt.Sprintf("No assigned %s matched the request.", terminology.Stories)
	if len(voiceStories) == 1 {
		message = fmt.Sprintf("Found 1 assigned %s.", terminology.Story)
	} else if len(voiceStories) > 1 {
		message = fmt.Sprintf("Found %d assigned %s.", len(voiceStories), terminology.Stories)
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Stories:     voiceStories,
		Count:       len(voiceStories),
		Message:     message,
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeGetContext(ctx context.Context, workspaceID, userID uuid.UUID) (AppRealtimeToolResponse, error) {
	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	currentUser, err := h.currentRealtimeUser(ctx, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	terminology := h.realtimeTerminology(ctx, workspaceID)

	return AppRealtimeToolResponse{
		Success:     true,
		Teams:       toRealtimeVoiceTeams(workspaceTeams),
		User:        &currentUser,
		Count:       len(workspaceTeams),
		Message:     teamContextMessage(workspaceTeams),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeListTeams(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListTeamsArguments{Limit: 25}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_teams arguments: %w", err)
		}
	}
	limit := clampLimit(args.Limit, 25)

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID, teams.CoreListTeamsFilter{
		Search: strings.TrimSpace(args.Search),
		Limit:  limit,
	})
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Teams:       toRealtimeVoiceTeams(workspaceTeams),
		Count:       len(workspaceTeams),
		Message:     fmt.Sprintf("Found %d team%s.", len(workspaceTeams), pluralSuffix(len(workspaceTeams))),
		Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
	}, nil
}

func (h *Handlers) executeListTeamMembers(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListTeamMembersArguments{Limit: 25}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_team_members arguments: %w", err)
		}
	}
	limit := clampLimit(args.Limit, 25)

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
	if team == nil {
		return AppRealtimeToolResponse{
			Success:     false,
			NeedsTeam:   true,
			Teams:       toRealtimeVoiceTeams(workspaceTeams),
			Message:     "Ask which team to list members for.",
			Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
		}, nil
	}

	members, err := h.users.List(ctx, workspaceID, users.CoreListUsersFilter{
		TeamID: &team.ID,
		Search: strings.TrimSpace(args.Search),
		Limit:  limit,
	})
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list team members: %w", err)
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Members:     toRealtimeVoiceMembers(members),
		Count:       len(members),
		Message:     fmt.Sprintf("Found %d member%s in %s.", len(members), pluralSuffix(len(members)), team.Name),
		Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
	}, nil
}

func (h *Handlers) executeSearchWork(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeSearchArguments{Type: "all", Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid search_work arguments: %w", err)
		}
	}
	args.Query = strings.TrimSpace(args.Query)
	if args.Query == "" {
		return AppRealtimeToolResponse{
			Success: false,
			Message: "Ask what to search for.",
		}, nil
	}

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	var teamID *uuid.UUID
	if strings.TrimSpace(args.TeamName) != "" {
		team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
		if team == nil {
			return AppRealtimeToolResponse{
				Success:     false,
				NeedsTeam:   true,
				Teams:       toRealtimeVoiceTeams(workspaceTeams),
				Message:     "Ask which team to search in.",
				Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
			}, nil
		}
		teamID = &team.ID
	}

	searchType := search.SearchTypeAll
	switch strings.ToLower(strings.TrimSpace(args.Type)) {
	case "stories", "tasks", "issues", "work_items", "work-items", "work items":
		searchType = search.SearchTypeStories
	case "objectives", "goals", "projects":
		searchType = search.SearchTypeObjectives
	}
	searchResult, err := h.search.Search(ctx, workspaceID, userID, search.SearchParams{
		Type:     searchType,
		Query:    args.Query,
		TeamID:   teamID,
		SortBy:   search.SortByRelevance,
		Page:     1,
		PageSize: clampLimit(args.Limit, 10),
	})
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("search work: %w", err)
	}

	teamsByID := indexTeamsByID(workspaceTeams)
	statusesByID, err := h.statusesByID(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}

	stories := make([]AppRealtimeVoiceStory, 0, len(searchResult.Stories))
	for _, story := range searchResult.Stories {
		stories = append(stories, toRealtimeVoiceSearchStory(story, teamsByID, statusesByID))
	}
	objectives := make([]AppRealtimeVoiceObjective, 0, len(searchResult.Objectives))
	for _, objective := range searchResult.Objectives {
		objectives = append(objectives, toRealtimeVoiceSearchObjective(objective, teamsByID))
	}

	return AppRealtimeToolResponse{
		Success:     true,
		Stories:     stories,
		Objectives:  objectives,
		Count:       searchResult.TotalStories + searchResult.TotalObjectives,
		Message:     fmt.Sprintf("Found %d matching work items.", searchResult.TotalStories+searchResult.TotalObjectives),
		Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
	}, nil
}

func (h *Handlers) executeListObjectives(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListObjectivesArguments{Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_objectives arguments: %w", err)
		}
	}

	filters := map[string]any{
		"limit":  clampLimit(args.Limit, 10),
		"offset": 0,
	}
	if search := strings.TrimSpace(args.Search); search != "" {
		filters["search"] = search
	}

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	if strings.TrimSpace(args.TeamName) != "" {
		team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
		if team == nil {
			return AppRealtimeToolResponse{
				Success:     false,
				NeedsTeam:   true,
				Teams:       toRealtimeVoiceTeams(workspaceTeams),
				Message:     "Ask which team to list objectives for.",
				Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
			}, nil
		}
		filters["team_id"] = team.ID
	}

	objectiveList, err := h.objectives.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list objectives: %w", err)
	}
	teamsByID := indexTeamsByID(workspaceTeams)
	result := make([]AppRealtimeVoiceObjective, 0, len(objectiveList))
	for _, objective := range objectiveList {
		result = append(result, toRealtimeVoiceObjective(objective, teamsByID))
	}

	terminology := h.realtimeTerminology(ctx, workspaceID)
	return AppRealtimeToolResponse{
		Success:     true,
		Objectives:  result,
		Count:       len(result),
		Message:     fmt.Sprintf("Found %d %s.", len(result), termForCount(terminology.Objective, terminology.Objectives, len(result))),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeListKeyResults(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeListKeyResultsArguments{Limit: 10}
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid list_key_results arguments: %w", err)
		}
	}

	filters := keyresults.CoreKeyResultFilters{
		WorkspaceID:    workspaceID,
		CurrentUserID:  userID,
		Page:           1,
		PageSize:       clampLimit(args.Limit, 10),
		OrderBy:        "updated_at",
		OrderDirection: "desc",
	}

	if strings.TrimSpace(args.TeamName) != "" {
		workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
		if err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
		}
		team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
		if team == nil {
			return AppRealtimeToolResponse{
				Success:     false,
				NeedsTeam:   true,
				Teams:       toRealtimeVoiceTeams(workspaceTeams),
				Message:     "Ask which team to list key results for.",
				Terminology: ptr(h.realtimeTerminology(ctx, workspaceID)),
			}, nil
		}
		filters.TeamIDs = []uuid.UUID{team.ID}
	}

	response, err := h.keyResults.ListPaginated(ctx, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list key results: %w", err)
	}

	keyResults := make([]AppRealtimeVoiceKeyResult, 0, len(response.KeyResults))
	for _, keyResult := range response.KeyResults {
		keyResults = append(keyResults, toRealtimeVoiceKeyResult(keyResult))
	}

	terminology := h.realtimeTerminology(ctx, workspaceID)
	return AppRealtimeToolResponse{
		Success:     true,
		KeyResults:  keyResults,
		Count:       len(keyResults),
		Message:     fmt.Sprintf("Found %d %s.", len(keyResults), termForCount(terminology.KeyResult, terminology.KeyResults, len(keyResults))),
		Terminology: &terminology,
	}, nil
}

func (h *Handlers) executeCreateTask(ctx context.Context, workspaceID, userID, sessionID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	var args AppRealtimeCreateTaskArguments
	if len(rawArgs) > 0 {
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return AppRealtimeToolResponse{}, fmt.Errorf("invalid create_task arguments: %w", err)
		}
	}

	args.Title = strings.TrimSpace(args.Title)
	args.Description = strings.TrimSpace(args.Description)
	args.TeamName = strings.TrimSpace(args.TeamName)
	args.AssigneeName = strings.TrimSpace(args.AssigneeName)
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	args.BlockedByRef = strings.TrimSpace(args.BlockedByRef)
	args.BlockingRef = strings.TrimSpace(args.BlockingRef)
	args.RelatedRef = strings.TrimSpace(args.RelatedRef)
	args.Priority = normalizePriority(args.Priority)
	terminology := h.realtimeTerminology(ctx, workspaceID)
	if args.Title == "" {
		return AppRealtimeToolResponse{
			Success:     false,
			Message:     fmt.Sprintf("Ask the user for the %s title before creating it.", terminology.Story),
			Terminology: &terminology,
		}, nil
	}
	isConfirmed := args.Confirmed
	providedConfirmationToken := strings.TrimSpace(args.ConfirmationToken)

	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list teams: %w", err)
	}
	team := resolveRealtimeTeam(workspaceTeams, args.TeamName)
	if team == nil {
		return AppRealtimeToolResponse{
			Success:     false,
			NeedsTeam:   true,
			Teams:       toRealtimeVoiceTeams(workspaceTeams),
			Message:     fmt.Sprintf("Ask the user which team this %s should be created in before creating it.", terminology.Story),
			Terminology: &terminology,
		}, nil
	}

	statuses, err := h.states.TeamList(ctx, workspaceID, team.ID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("list team statuses: %w", err)
	}
	status := findDefaultRealtimeStatus(statuses)
	if status == nil {
		return AppRealtimeToolResponse{
			Success:     false,
			Error:       fmt.Sprintf("No statuses are configured for %s.", team.Name),
			Terminology: &terminology,
		}, nil
	}

	currentUser, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get current user: %w", err)
	}
	loc := userLocation(currentUser)
	now := h.now().In(loc)
	assigneeID, assigneeName, assigneeResponse, err := h.resolveRealtimeAssignee(ctx, workspaceID, userID, team, args)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if assigneeResponse != nil {
		assigneeResponse.Terminology = &terminology
		return *assigneeResponse, nil
	}
	startDate, err := parseRealtimeDate(args.StartDate, loc, now)
	if err != nil {
		return AppRealtimeToolResponse{
			Success:     false,
			Message:     fmt.Sprintf("Ask the user to clarify the start date. %s", err.Error()),
			Terminology: &terminology,
		}, nil
	}
	endDate, err := parseRealtimeDate(args.EndDate, loc, now)
	if err != nil {
		return AppRealtimeToolResponse{
			Success:     false,
			Message:     fmt.Sprintf("Ask the user to clarify the due date. %s", err.Error()),
			Terminology: &terminology,
		}, nil
	}

	blockedByID, blockedByRef, linkResponse, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, workspaceTeams, team, args.BlockedByRef, "blocker")
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if linkResponse != nil {
		linkResponse.Terminology = &terminology
		return *linkResponse, nil
	}
	blockingID, blockingRef, linkResponse, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, workspaceTeams, team, args.BlockingRef, "blocked work item")
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if linkResponse != nil {
		linkResponse.Terminology = &terminology
		return *linkResponse, nil
	}
	relatedID, relatedRef, linkResponse, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, workspaceTeams, team, args.RelatedRef, "related work item")
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if linkResponse != nil {
		linkResponse.Terminology = &terminology
		return *linkResponse, nil
	}

	confirmationInput := struct {
		Title         string
		Description   string
		TeamID        uuid.UUID
		StatusID      uuid.UUID
		AssigneeID    *uuid.UUID
		Priority      string
		EstimateValue *int16
		StartDate     *time.Time
		EndDate       *time.Time
		BlockedByID   *uuid.UUID
		BlockingID    *uuid.UUID
		RelatedID     *uuid.UUID
	}{
		Title: args.Title, Description: args.Description, TeamID: team.ID,
		StatusID: status.ID, AssigneeID: assigneeID, Priority: args.Priority,
		EstimateValue: args.EstimateValue, StartDate: startDate, EndDate: endDate,
		BlockedByID: blockedByID, BlockingID: blockingID, RelatedID: relatedID,
	}
	expectedConfirmationToken, err := h.confirmationToken(sessionID, "create_task", confirmationInput)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if !isConfirmed {
		return AppRealtimeToolResponse{
			Success:              false,
			RequiresConfirmation: true,
			Message:              fmt.Sprintf("Ask the user to confirm before creating the %s %q in %s.", terminology.Story, args.Title, team.Name),
			Terminology:          &terminology,
			ConfirmationToken:    expectedConfirmationToken,
			Confirmation: &AppRealtimeConfirmation{
				Title:         args.Title,
				Description:   args.Description,
				TeamName:      team.Name,
				AssigneeName:  assigneeName,
				Priority:      args.Priority,
				EstimateValue: args.EstimateValue,
				StartDate:     args.StartDate,
				EndDate:       args.EndDate,
				BlockedByRef:  blockedByRef,
				BlockingRef:   blockingRef,
				RelatedRef:    relatedRef,
			},
		}, nil
	}
	confirmed, err := h.validateConfirmationToken(sessionID, "create_task", confirmationInput, providedConfirmationToken)
	if err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if !confirmed {
		response := changedConfirmationResponse(expectedConfirmationToken)
		response.Terminology = &terminology
		return response, nil
	}

	description := args.Description
	descriptionHTML := ""
	if description != "" {
		descriptionHTML = "<p>" + html.EscapeString(description) + "</p>"
	}

	story, err := h.stories.Create(ctx, stories.CoreNewStory{
		Title:           args.Title,
		Description:     optionalString(description),
		DescriptionHTML: optionalString(descriptionHTML),
		Status:          &status.ID,
		Assignee:        assigneeID,
		BlockedBy:       blockedByID,
		Blocking:        blockingID,
		Related:         relatedID,
		Reporter:        &userID,
		Priority:        args.Priority,
		EstimateValue:   args.EstimateValue,
		StartDate:       startDate,
		EndDate:         endDate,
		Team:            team.ID,
	}, workspaceID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("create story: %w", err)
	}

	h.invalidateStoryListCaches(ctx, workspaceID)

	voiceStory := toRealtimeVoiceCreatedStory(story, *team, *status, assigneeName)
	voiceStory.BlockedBy = blockedByRef
	voiceStory.Blocking = blockingRef
	voiceStory.Related = relatedRef
	return AppRealtimeToolResponse{
		Success:     true,
		Story:       &voiceStory,
		Message:     fmt.Sprintf("Created the %s %q in %s.", terminology.Story, story.Title, team.Name),
		Terminology: &terminology,
	}, nil
}
