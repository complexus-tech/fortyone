package mayahttp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

func (h *Handlers) executeWorkspaceBriefing(ctx context.Context, workspaceID, userID uuid.UUID, rawArgs json.RawMessage) (AppRealtimeToolResponse, error) {
	args := AppRealtimeWorkspaceBriefingArguments{Days: 30}
	if err := decodeRealtimeArguments(rawArgs, &args, "workspace_briefing"); err != nil {
		return AppRealtimeToolResponse{}, err
	}
	if args.Days <= 0 || args.Days > 365 {
		args.Days = 30
	}
	end := h.now().UTC()
	start := end.AddDate(0, 0, -args.Days)
	filters := reports.ReportFilters{StartDate: &start, EndDate: &end}
	overview, err := h.reports.GetWorkspaceOverview(ctx, workspaceID, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get workspace overview: %w", err)
	}
	pulse, err := h.reports.GetPulseReport(ctx, workspaceID, filters)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get workspace pulse: %w", err)
	}
	feedbackSummaries, err := h.feedback.ListTeamSummaries(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeToolResponse{}, fmt.Errorf("get feedback summary: %w", err)
	}
	feedbackItems := 0
	unreadFeedback := 0
	for _, summary := range feedbackSummaries {
		feedbackItems += summary.TotalCount
		unreadFeedback += summary.UnreadCount
	}
	briefing := AppRealtimeVoiceBriefing{
		TotalStories:      overview.Metrics.TotalStories,
		CompletedStories:  overview.Metrics.CompletedStories,
		ActiveObjectives:  overview.Metrics.ActiveObjectives,
		ActiveSprints:     overview.Metrics.ActiveSprints,
		TeamMembers:       overview.Metrics.TotalTeamMembers,
		OverdueStories:    pulse.Summary.OverdueStories,
		BlockedStories:    pulse.Summary.BlockedStories,
		AtRiskSprints:     pulse.Summary.AtRiskSprints,
		AtRiskObjectives:  pulse.Summary.AtRiskObjectives,
		FeedbackItems:     feedbackItems,
		UnreadFeedback:    unreadFeedback,
		OverloadedMembers: pulse.Summary.OverloadedMembers,
	}
	return AppRealtimeToolResponse{
		Success: true, Briefing: &briefing,
		Message: fmt.Sprintf("In the last %d days: %d stories completed, %d overdue, %d blocked, %d at-risk objectives, and %d unread feedback items.", args.Days, briefing.CompletedStories, briefing.OverdueStories, briefing.BlockedStories, briefing.AtRiskObjectives, briefing.UnreadFeedback),
	}, nil
}

func (h *Handlers) resolveRealtimeStory(ctx context.Context, workspaceID, userID uuid.UUID, reference string) (stories.CoreSingleStory, *AppRealtimeToolResponse, error) {
	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return stories.CoreSingleStory{}, nil, fmt.Errorf("list teams while resolving story: %w", err)
	}
	id, _, response, err := h.resolveRealtimeStoryLink(ctx, workspaceID, userID, teamList, nil, reference, "work item")
	if err != nil || response != nil {
		return stories.CoreSingleStory{}, response, err
	}
	if id == nil {
		return stories.CoreSingleStory{}, &AppRealtimeToolResponse{Success: false, NeedsStoryReference: true, Message: "Ask which work item the user meant."}, nil
	}
	story, err := h.stories.Get(ctx, *id, workspaceID)
	if err != nil {
		return stories.CoreSingleStory{}, nil, fmt.Errorf("get story: %w", err)
	}
	return story, nil, nil
}

func (h *Handlers) toRealtimeVoiceSingleStory(ctx context.Context, workspaceID, userID uuid.UUID, story stories.CoreSingleStory) (AppRealtimeVoiceStory, error) {
	team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
	if err != nil {
		return AppRealtimeVoiceStory{}, fmt.Errorf("get story team: %w", err)
	}
	statuses, err := h.states.TeamList(ctx, workspaceID, story.Team)
	if err != nil {
		return AppRealtimeVoiceStory{}, fmt.Errorf("get story statuses: %w", err)
	}
	var statusName *AppRealtimeVoiceStatus
	if story.Status != nil {
		for _, status := range statuses {
			if status.ID == *story.Status {
				statusName = toRealtimeVoiceStatus(status)
				break
			}
		}
	}
	assignee := ""
	if story.Assignee != nil {
		if user, err := h.users.GetUser(ctx, *story.Assignee); err == nil {
			assignee = displayUserName(user)
		}
	}
	description := ""
	if story.Description != nil {
		description = strings.TrimSpace(*story.Description)
	}
	sprintName := ""
	if story.SprintSummary != nil {
		sprintName = story.SprintSummary.Name
	}
	objectiveName := ""
	if story.Objective != nil {
		if objective, err := h.objectives.Get(ctx, *story.Objective, workspaceID); err == nil {
			objectiveName = objective.Name
		}
	}
	return AppRealtimeVoiceStory{
		Reference: storyReference(team.Code, story.SequenceID), Title: story.Title, Description: description,
		Priority: story.Priority, EstimateLabel: story.EstimateLabel,
		EstimateValue: story.EstimateValue, Team: team.Name, Assignee: assignee,
		Sprint: sprintName, Objective: objectiveName, Status: statusName,
		StartDate: story.StartDate, EndDate: story.EndDate,
		CompletedAt: story.CompletedAt,
	}, nil
}

func (h *Handlers) resolveRealtimeStoryUpdates(ctx context.Context, workspaceID, userID uuid.UUID, story stories.CoreSingleStory, args AppRealtimeUpdateStoryArguments) (map[string]any, []string, *AppRealtimeToolResponse, error) {
	updates := make(map[string]any)
	summary := make([]string, 0, 7)
	if args.Title != "" {
		updates["title"] = args.Title
		summary = append(summary, fmt.Sprintf("title to %q", args.Title))
	}
	if args.Priority != "" {
		priority := normalizePriority(args.Priority)
		if _, ok := realtimeStoryPriorities[priority]; !ok {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask for a priority of No Priority, Low, Medium, High, or Urgent."}, nil
		}
		updates["priority"] = priority
		summary = append(summary, "priority to "+priority)
	}
	if args.Status != "" {
		statuses, err := h.states.TeamList(ctx, workspaceID, story.Team)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("list story statuses: %w", err)
		}
		matched := resolveRealtimeStatus(statuses, args.Status)
		if matched == nil {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask which exact team status to use."}, nil
		}
		updates["status_id"] = matched.ID
		summary = append(summary, "status to "+matched.Name)
	}
	if args.Unassign && (args.AssignToMe || args.AssigneeName != "") {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to unassign the story or assign it to someone."}, nil
	}
	if args.Unassign {
		updates["assignee_id"] = nil
		summary = append(summary, "remove the assignee")
	} else if args.AssignToMe || args.AssigneeName != "" {
		team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get story team: %w", err)
		}
		assigneeID, assigneeName, response, err := h.resolveRealtimeAssignee(ctx, workspaceID, userID, &team, AppRealtimeCreateTaskArguments{
			AssigneeName: args.AssigneeName, AssignToMe: args.AssignToMe,
		})
		if err != nil || response != nil {
			return nil, nil, response, err
		}
		updates["assignee_id"] = assigneeID
		summary = append(summary, "assignee to "+assigneeName)
	}
	if args.ClearEstimate && args.EstimateValue != nil {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to clear the estimate or set a new one."}, nil
	}
	if args.ClearEstimate {
		updates["estimate_unit"] = nil
		summary = append(summary, "clear the estimate")
	} else if args.EstimateValue != nil {
		updates["estimate_unit"] = args.EstimateValue
		summary = append(summary, fmt.Sprintf("estimate to %d", *args.EstimateValue))
	}
	currentUser, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get current user for dates: %w", err)
	}
	loc := userLocation(currentUser)
	now := h.now().In(loc)
	if args.ClearStartDate && args.StartDate != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to clear the start date or set a new one."}, nil
	}
	if args.ClearStartDate {
		updates["start_date"] = nil
		summary = append(summary, "clear the start date")
	} else if args.StartDate != "" {
		value, err := parseRealtimeDate(args.StartDate, loc, now)
		if err != nil {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask the user to clarify the start date."}, nil
		}
		updates["start_date"] = value
		summary = append(summary, "start date to "+args.StartDate)
	}
	if args.ClearEndDate && args.EndDate != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to clear the due date or set a new one."}, nil
	}
	if args.ClearEndDate {
		updates["end_date"] = nil
		summary = append(summary, "clear the due date")
	} else if args.EndDate != "" {
		value, err := parseRealtimeDate(args.EndDate, loc, now)
		if err != nil {
			return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask the user to clarify the due date."}, nil
		}
		updates["end_date"] = value
		summary = append(summary, "due date to "+args.EndDate)
	}
	if args.ClearSprint && args.SprintName != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to remove the sprint or set a new one."}, nil
	}
	if args.ClearSprint {
		updates["sprint_id"] = nil
		summary = append(summary, "remove the sprint")
	} else if args.SprintName != "" {
		team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get story team for sprint: %w", err)
		}
		sprint, _, response, err := h.resolveRealtimeSprint(ctx, workspaceID, userID, args.SprintName, team.Name)
		if err != nil || response != nil {
			return nil, nil, response, err
		}
		updates["sprint_id"] = sprint.ID
		summary = append(summary, "sprint to "+sprint.Name)
	}
	if args.ClearObjective && args.ObjectiveName != "" {
		return nil, nil, &AppRealtimeToolResponse{Success: false, Message: "Ask whether to remove the objective or set a new one."}, nil
	}
	if args.ClearObjective {
		updates["objective_id"] = nil
		summary = append(summary, "remove the objective")
	} else if args.ObjectiveName != "" {
		team, err := h.teams.GetByID(ctx, story.Team, workspaceID, userID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get story team for objective: %w", err)
		}
		objective, _, response, err := h.resolveRealtimeObjective(ctx, workspaceID, userID, args.ObjectiveName, team.Name)
		if err != nil || response != nil {
			return nil, nil, response, err
		}
		updates["objective_id"] = objective.ID
		summary = append(summary, "objective to "+objective.Name)
	}
	return updates, summary, nil, nil
}

func (h *Handlers) resolveRealtimeSprint(ctx context.Context, workspaceID, userID uuid.UUID, name, teamName string) (sprints.CoreSprint, teams.CoreTeam, *AppRealtimeToolResponse, error) {
	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return sprints.CoreSprint{}, teams.CoreTeam{}, nil, fmt.Errorf("list sprint teams: %w", err)
	}
	var teamID *uuid.UUID
	if strings.TrimSpace(teamName) != "" {
		team := resolveRealtimeTeam(teamList, teamName)
		if team == nil {
			return sprints.CoreSprint{}, teams.CoreTeam{}, ptr(realtimeTeamClarification(teamList, "Ask which team's sprint the user meant.")), nil
		}
		teamID = &team.ID
	}
	filters := map[string]any{}
	if teamID != nil {
		filters["team_id"] = *teamID
	}
	sprintList, err := h.sprints.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return sprints.CoreSprint{}, teams.CoreTeam{}, nil, fmt.Errorf("list sprints: %w", err)
	}
	normalized := normalizeName(name)
	matches := make([]int, 0, len(sprintList))
	for i, sprint := range sprintList {
		if normalizeName(sprint.Name) == normalized {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		for i, sprint := range sprintList {
			if strings.Contains(normalizeName(sprint.Name), normalized) {
				matches = append(matches, i)
			}
		}
	}
	if len(matches) != 1 {
		return sprints.CoreSprint{}, teams.CoreTeam{}, &AppRealtimeToolResponse{
			Success: false, Sprints: toRealtimeVoiceSprints(sprintList, indexTeamsByID(teamList), 10),
			Message: "Ask which sprint the user meant.",
		}, nil
	}
	sprint := sprintList[matches[0]]
	team := indexTeamsByID(teamList)[sprint.TeamID]
	return sprint, team, nil, nil
}

func (h *Handlers) resolveRealtimeObjective(ctx context.Context, workspaceID, userID uuid.UUID, name, teamName string) (objectives.CoreObjective, teams.CoreTeam, *AppRealtimeToolResponse, error) {
	teamList, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return objectives.CoreObjective{}, teams.CoreTeam{}, nil, fmt.Errorf("list objective teams: %w", err)
	}
	filters := map[string]any{}
	if strings.TrimSpace(teamName) != "" {
		team := resolveRealtimeTeam(teamList, teamName)
		if team == nil {
			return objectives.CoreObjective{}, teams.CoreTeam{}, ptr(realtimeTeamClarification(teamList, "Ask which team's objective the user meant.")), nil
		}
		filters["team_id"] = team.ID
	}
	objectiveList, err := h.objectives.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return objectives.CoreObjective{}, teams.CoreTeam{}, nil, fmt.Errorf("list objectives: %w", err)
	}
	normalized := normalizeName(name)
	matches := make([]int, 0, len(objectiveList))
	for i, objective := range objectiveList {
		if normalizeName(objective.Name) == normalized {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		for i, objective := range objectiveList {
			if strings.Contains(normalizeName(objective.Name), normalized) {
				matches = append(matches, i)
			}
		}
	}
	if len(matches) != 1 {
		teamsByID := indexTeamsByID(teamList)
		voiceObjectives := make([]AppRealtimeVoiceObjective, 0, len(objectiveList))
		for _, objective := range objectiveList {
			voiceObjectives = append(voiceObjectives, toRealtimeVoiceObjective(objective, teamsByID))
		}
		return objectives.CoreObjective{}, teams.CoreTeam{}, &AppRealtimeToolResponse{
			Success: false, Objectives: voiceObjectives,
			Message: "Ask which objective the user meant.",
		}, nil
	}
	objective := objectiveList[matches[0]]
	return objective, indexTeamsByID(teamList)[objective.Team], nil, nil
}

func realtimeNavigationResponse(path, target string) AppRealtimeToolResponse {
	return AppRealtimeToolResponse{
		Success: true, Message: "Opening " + strings.ReplaceAll(target, "-", " ") + ".",
		ClientAction: &AppRealtimeClientAction{Type: "navigate", Path: path},
	}
}

func realtimeTeamClarification(teamList []teams.CoreTeam, message string) AppRealtimeToolResponse {
	return AppRealtimeToolResponse{
		Success: false, NeedsTeam: true, Teams: toRealtimeVoiceTeams(teamList), Message: message,
	}
}

func changedConfirmationResponse(token string) AppRealtimeToolResponse {
	return AppRealtimeToolResponse{
		Success: false, RequiresConfirmation: true, ConfirmationToken: token,
		Message: "The details changed after confirmation. Read back the current details and ask the user to confirm them again.",
	}
}

func decodeRealtimeArguments(rawArgs json.RawMessage, target any, toolName string) error {
	if len(rawArgs) == 0 {
		return nil
	}
	if err := json.Unmarshal(rawArgs, target); err != nil {
		return fmt.Errorf("invalid %s arguments: %w", toolName, err)
	}
	return nil
}

func resolveRealtimeStatus(statuses []states.CoreState, value string) *states.CoreState {
	normalized := normalizeName(value)
	for i := range statuses {
		if normalizeName(statuses[i].Name) == normalized || normalizeName(statuses[i].Category) == normalized {
			return &statuses[i]
		}
	}
	var match *states.CoreState
	for i := range statuses {
		if strings.Contains(normalizeName(statuses[i].Name), normalized) {
			if match != nil {
				return nil
			}
			match = &statuses[i]
		}
	}
	return match
}

func toRealtimeVoiceSprint(value sprints.CoreSprint, teamName string) AppRealtimeVoiceSprint {
	goal := ""
	if value.Goal != nil {
		goal = *value.Goal
	}
	completion := 0
	if value.TotalStories > 0 {
		completion = value.CompletedStories * 100 / value.TotalStories
	}
	return AppRealtimeVoiceSprint{
		Name: value.Name, Team: teamName, Goal: goal, StartDate: value.StartDate,
		EndDate: value.EndDate, TotalStories: value.TotalStories,
		CompletedStories: value.CompletedStories, StartedStories: value.StartedStories,
		CompletionPercentage: completion,
	}
}

func toRealtimeVoiceSprints(sprintList []sprints.CoreSprint, teamsByID map[uuid.UUID]teams.CoreTeam, limit int) []AppRealtimeVoiceSprint {
	result := make([]AppRealtimeVoiceSprint, 0, min(limit, len(sprintList)))
	for _, sprint := range sprintList {
		result = append(result, toRealtimeVoiceSprint(sprint, teamsByID[sprint.TeamID].Name))
		if len(result) >= limit {
			break
		}
	}
	return result
}

func feedbackItems(matches []realtimeFeedbackMatch) []AppRealtimeVoiceFeedbackItem {
	result := make([]AppRealtimeVoiceFeedbackItem, 0, len(matches))
	for _, match := range matches {
		result = append(result, match.item)
	}
	return result
}

func responseOrEmpty(response *AppRealtimeToolResponse) AppRealtimeToolResponse {
	if response == nil {
		return AppRealtimeToolResponse{}
	}
	return *response
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func displayWorkloadMemberName(fullName, username string) string {
	return firstNonEmpty(strings.TrimSpace(fullName), strings.TrimSpace(username), "Unknown member")
}

func activityPluralEnding(count int) string {
	if count == 1 {
		return "y"
	}
	return "ies"
}
