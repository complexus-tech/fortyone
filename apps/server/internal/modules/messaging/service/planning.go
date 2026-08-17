package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

// SprintsService is the membership-aware sprint surface used by planning
// assistant tools.
type SprintsService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters map[string]any) ([]sprints.CoreSprint, error)
	GetAnalytics(ctx context.Context, sprintID uuid.UUID, workspaceID uuid.UUID) (sprints.CoreSprintAnalytics, error)
}

// ObjectiveAnalyticsService is optional because the objective list already
// carries aggregate counts. Production uses it to provide the richer status
// breakdown when the repository supports analytics.
type ObjectiveAnalyticsService interface {
	GetAnalytics(ctx context.Context, objectiveID uuid.UUID, workspaceID uuid.UUID) (objectives.CoreObjectiveAnalytics, error)
}

func planningToolDefinitions() []ToolDefinition {
	nullableString := func(description string) map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": description,
			"maxLength":   maxSearchRunes,
		}
	}
	limit := map[string]any{
		"type":        []string{"integer", "null"},
		"description": "Maximum number of associated work items, or null for the default.",
		"minimum":     1,
		"maximum":     maxToolLimit,
	}

	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListSprints,
			Description: "List sprints the current user can access, optionally narrowed by sprint name and team.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"name":      nullableString("Optional sprint-name search text, or null for all sprints."),
				"team_name": nullableString("Optional team name or code, or null for all joined teams."),
				"limit":     limit,
			}, []string{"name", "team_name", "limit"}),
		},
		{
			Type:        "function",
			Name:        toolGetSprintSummary,
			Description: "Resolve one accessible sprint by name and return its dates, progress percentage, status, counts, and associated completed, remaining, and cancelled work.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"name":      map[string]any{"type": "string", "description": "Sprint name, such as Sprint 1.", "minLength": 1, "maxLength": maxSearchRunes},
				"team_name": nullableString("Optional team name or code when sprint names are ambiguous, or null."),
				"limit":     limit,
			}, []string{"name", "team_name", "limit"}),
		},
		{
			Type:        "function",
			Name:        toolGetObjectiveSummary,
			Description: "Resolve one accessible objective by name and return its dates, health, progress percentage, counts, and associated completed, remaining, and cancelled work.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"name":      map[string]any{"type": "string", "description": "Objective name.", "minLength": 1, "maxLength": maxSearchRunes},
				"team_name": nullableString("Optional team name or code when objective names are ambiguous, or null."),
				"limit":     limit,
			}, []string{"name", "team_name", "limit"}),
		},
	}
}

func (e *FortyOneToolExecutor) listSprints(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		Name     *string `json:"name"`
		TeamName *string `json:"team_name"`
		Limit    *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "name", "team_name", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	joined, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	team, err := teamByName(args.TeamName, joined)
	if err != nil {
		return nil, err
	}
	filters := map[string]any{"limit": limit}
	if team != nil {
		filters["team_id"] = team.ID
	}
	if args.Name != nil && strings.TrimSpace(*args.Name) != "" {
		filters["search"] = strings.TrimSpace(*args.Name)
	}
	items, err := e.sprints.List(ctx, scope.WorkspaceID, scope.UserID, filters)
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}

	result := listSprintsResult{Sprints: make([]sprintResult, 0, min(len(items), limit))}
	for _, sprint := range items {
		if sprint.Workspace != scope.WorkspaceID || joinedByID[sprint.Team].ID == uuid.Nil {
			continue
		}
		result.Total++
		if len(result.Sprints) == limit {
			continue
		}
		result.Sprints = append(result.Sprints, newSprintResult(sprint, joinedByID[sprint.Team]))
	}
	result.Truncated = result.Total > len(result.Sprints)
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) getSprintSummary(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args planningSummaryArguments
	if err := decodeToolArguments(raw, &args, "name", "team_name", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	joined, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	team, err := teamByName(args.TeamName, joined)
	if err != nil {
		return nil, err
	}
	sprint, err := e.resolveSprint(ctx, scope, args.Name, team)
	if err != nil {
		return nil, err
	}
	analytics, err := e.sprints.GetAnalytics(ctx, sprint.ID, scope.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("get sprint analytics: %w", err)
	}
	work, err := e.listPlanningWork(ctx, scope, joinedByID, planningWorkFilter{TeamID: sprint.Team, SprintID: &sprint.ID}, limit)
	if err != nil {
		return nil, fmt.Errorf("list sprint work: %w", err)
	}

	return marshalToolResult(sprintSummaryResult{
		Sprint: sprintResult{
			ID:                sprint.ID,
			Name:              sprint.Name,
			Goal:              sprint.Goal,
			TeamID:            sprint.Team,
			TeamName:          joinedByID[sprint.Team].Name,
			TeamCode:          strings.ToUpper(strings.TrimSpace(joinedByID[sprint.Team].Code)),
			StartDate:         formatPlanningDate(sprint.StartDate),
			EndDate:           formatPlanningDate(sprint.EndDate),
			Status:            analytics.Overview.Status,
			ProgressPercent:   analytics.Overview.CompletionPercentage,
			TotalStories:      analytics.StoryBreakdown.Total,
			CompletedStories:  analytics.StoryBreakdown.Completed,
			InProgressStories: analytics.StoryBreakdown.InProgress,
			TodoStories:       analytics.StoryBreakdown.Todo,
			BlockedStories:    analytics.StoryBreakdown.Blocked,
			CancelledStories:  analytics.StoryBreakdown.Cancelled,
			DaysElapsed:       &analytics.Overview.DaysElapsed,
			DaysRemaining:     &analytics.Overview.DaysRemaining,
		},
		Work: work,
	})
}

func (e *FortyOneToolExecutor) getObjectiveSummary(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args planningSummaryArguments
	if err := decodeToolArguments(raw, &args, "name", "team_name", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	joined, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	team, err := teamByName(args.TeamName, joined)
	if err != nil {
		return nil, err
	}
	objective, err := e.resolveObjective(ctx, scope, args.Name, team)
	if err != nil {
		return nil, err
	}
	work, err := e.listPlanningWork(ctx, scope, joinedByID, planningWorkFilter{TeamID: objective.Team, ObjectiveID: &objective.ID}, limit)
	if err != nil {
		return nil, fmt.Errorf("list objective work: %w", err)
	}

	progress := objectiveProgress{
		Total:      objective.TotalStories,
		Completed:  objective.CompletedStories,
		InProgress: objective.StartedStories,
		Todo:       objective.UnstartedStories,
		Cancelled:  objective.CancelledStories,
		Backlog:    objective.BacklogStories,
	}
	if analyticsService, ok := e.objectives.(ObjectiveAnalyticsService); ok {
		analytics, analyticsErr := analyticsService.GetAnalytics(ctx, objective.ID, scope.WorkspaceID)
		if analyticsErr != nil {
			return nil, fmt.Errorf("get objective analytics: %w", analyticsErr)
		}
		progress = objectiveProgress{
			Total:      analytics.ProgressBreakdown.Total,
			Completed:  analytics.ProgressBreakdown.Completed,
			InProgress: analytics.ProgressBreakdown.InProgress,
			Todo:       analytics.ProgressBreakdown.Todo,
			Blocked:    analytics.ProgressBreakdown.Blocked,
			Cancelled:  analytics.ProgressBreakdown.Cancelled,
		}
	}

	health := ""
	if objective.Health != nil {
		health = string(*objective.Health)
	}
	return marshalToolResult(objectiveSummaryResult{
		Objective: planningObjectiveResult{
			ID:                objective.ID,
			SequenceID:        objective.SequenceID,
			Name:              objective.Name,
			ShortSummary:      objective.ShortSummary,
			TeamID:            objective.Team,
			TeamName:          joinedByID[objective.Team].Name,
			TeamCode:          strings.ToUpper(strings.TrimSpace(joinedByID[objective.Team].Code)),
			Health:            health,
			Priority:          objective.Priority,
			StartDate:         formatPlanningDatePtr(objective.StartDate),
			EndDate:           formatPlanningDatePtr(objective.EndDate),
			KeyResultCount:    objective.KeyResultCount,
			ProgressPercent:   completionPercent(progress.Completed, progress.Total),
			TotalStories:      progress.Total,
			CompletedStories:  progress.Completed,
			InProgressStories: progress.InProgress,
			TodoStories:       progress.Todo,
			BlockedStories:    progress.Blocked,
			CancelledStories:  progress.Cancelled,
		},
		Work: work,
	})
}

func (e *FortyOneToolExecutor) resolveSprint(ctx context.Context, scope ToolScope, name string, team *teams.CoreTeam) (sprints.CoreSprint, error) {
	filters := map[string]any{"search": strings.TrimSpace(name), "limit": maxToolLimit}
	if team != nil {
		filters["team_id"] = team.ID
	}
	items, err := e.sprints.List(ctx, scope.WorkspaceID, scope.UserID, filters)
	if err != nil {
		return sprints.CoreSprint{}, fmt.Errorf("resolve sprint: %w", err)
	}
	matching := make([]sprints.CoreSprint, 0, len(items))
	wanted := normalizePlanningName(name)
	for _, item := range items {
		if item.Workspace == scope.WorkspaceID && normalizePlanningName(item.Name) == wanted {
			matching = append(matching, item)
		}
	}
	if len(matching) == 0 {
		for _, item := range items {
			if item.Workspace == scope.WorkspaceID && strings.Contains(normalizePlanningName(item.Name), wanted) {
				matching = append(matching, item)
			}
		}
	}
	if len(matching) == 0 {
		return sprints.CoreSprint{}, fmt.Errorf("sprint %q was not found in the accessible teams", strings.TrimSpace(name))
	}
	if len(matching) > 1 {
		return sprints.CoreSprint{}, fmt.Errorf("sprint %q is ambiguous; include the team name", strings.TrimSpace(name))
	}
	return matching[0], nil
}

func (e *FortyOneToolExecutor) resolveObjective(ctx context.Context, scope ToolScope, name string, team *teams.CoreTeam) (objectives.CoreObjective, error) {
	filters := map[string]any{"search": strings.TrimSpace(name), "limit": maxToolLimit}
	if team != nil {
		filters["team_id"] = team.ID
	}
	items, err := e.objectives.List(ctx, scope.WorkspaceID, scope.UserID, filters)
	if err != nil {
		return objectives.CoreObjective{}, fmt.Errorf("resolve objective: %w", err)
	}
	matching := make([]objectives.CoreObjective, 0, len(items))
	wanted := normalizePlanningName(name)
	for _, item := range items {
		if item.Workspace == scope.WorkspaceID && normalizePlanningName(item.Name) == wanted {
			matching = append(matching, item)
		}
	}
	if len(matching) == 0 {
		for _, item := range items {
			if item.Workspace == scope.WorkspaceID && strings.Contains(normalizePlanningName(item.Name), wanted) {
				matching = append(matching, item)
			}
		}
	}
	if len(matching) == 0 {
		return objectives.CoreObjective{}, fmt.Errorf("objective %q was not found in the accessible teams", strings.TrimSpace(name))
	}
	if len(matching) > 1 {
		return objectives.CoreObjective{}, fmt.Errorf("objective %q is ambiguous; include the team name", strings.TrimSpace(name))
	}
	return matching[0], nil
}

func teamByName(value *string, joined []teams.CoreTeam) (*teams.CoreTeam, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	wanted := normalizePlanningName(*value)
	for index := range joined {
		if normalizePlanningName(joined[index].Name) == wanted || normalizePlanningName(joined[index].Code) == wanted {
			return &joined[index], nil
		}
	}
	return nil, fmt.Errorf("team %q is not accessible to the user", strings.TrimSpace(*value))
}

func normalizePlanningName(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}

type planningSummaryArguments struct {
	Name     string  `json:"name"`
	TeamName *string `json:"team_name"`
	Limit    *int    `json:"limit"`
}

type planningWorkFilter struct {
	TeamID      uuid.UUID
	SprintID    *uuid.UUID
	ObjectiveID *uuid.UUID
}

func (e *FortyOneToolExecutor) listPlanningWork(ctx context.Context, scope ToolScope, joinedByID map[uuid.UUID]teams.CoreTeam, filter planningWorkFilter, limit int) (planningWorkResult, error) {
	filters := map[string]any{"show_sub_stories": false}
	if filter.SprintID != nil {
		filters["sprint_ids"] = []uuid.UUID{*filter.SprintID}
	}
	if filter.ObjectiveID != nil {
		filters["objective_id"] = *filter.ObjectiveID
	}
	items, err := e.completed.List(ctx, scope.WorkspaceID, filters)
	if err != nil {
		return planningWorkResult{}, err
	}

	var statusesByID map[uuid.UUID]states.CoreState
	if e.states != nil {
		_, statusesByID, err = e.scopedStatuses(ctx, scope, joinedByID)
		if err != nil {
			return planningWorkResult{}, err
		}
	}
	result := planningWorkResult{
		Access:    "team",
		Completed: make([]planningStoryResult, 0, min(len(items), limit)),
		Remaining: make([]planningStoryResult, 0, min(len(items), limit)),
		Cancelled: make([]planningStoryResult, 0, min(len(items), limit)),
	}
	sharedWorkEnabled := teamWorkSharedTeamAllowed(scope.SharedTeamIDs, filter.TeamID)
	if !sharedWorkEnabled {
		result.Access = "personal"
	}
	for _, story := range items {
		team, allowed := joinedByID[story.Team]
		if !allowed || story.Workspace != scope.WorkspaceID || story.DeletedAt != nil || story.ArchivedAt != nil {
			continue
		}
		if filter.SprintID != nil && (story.Sprint == nil || *story.Sprint != *filter.SprintID) {
			continue
		}
		if filter.ObjectiveID != nil && (story.Objective == nil || *story.Objective != *filter.ObjectiveID) {
			continue
		}
		if !sharedWorkEnabled && (story.Assignee == nil || *story.Assignee != scope.UserID) {
			continue
		}
		result.Total++
		item := planningStoryResult{
			ID:          story.ID,
			Reference:   storyReference(team.Code, story.SequenceID),
			URL:         storyURL(scope, storyReference(team.Code, story.SequenceID)),
			Title:       story.Title,
			TeamName:    team.Name,
			StatusName:  "",
			CompletedAt: story.CompletedAt,
			UpdatedAt:   story.UpdatedAt,
		}
		category := ""
		if story.Status != nil && statusesByID != nil {
			if status, visible := statusesByID[*story.Status]; visible {
				item.StatusName = status.Name
				category = strings.ToLower(strings.TrimSpace(status.Category))
			}
		}
		switch {
		case category == "cancelled":
			result.Cancelled = append(result.Cancelled, item)
		case story.CompletedAt != nil || category == "completed":
			result.Completed = append(result.Completed, item)
		default:
			result.Remaining = append(result.Remaining, item)
		}
	}

	sort.SliceStable(result.Completed, func(i, j int) bool {
		if result.Completed[i].CompletedAt == nil {
			return false
		}
		if result.Completed[j].CompletedAt == nil {
			return true
		}
		return result.Completed[i].CompletedAt.After(*result.Completed[j].CompletedAt)
	})
	sort.SliceStable(result.Remaining, func(i, j int) bool { return result.Remaining[i].UpdatedAt.After(result.Remaining[j].UpdatedAt) })
	sort.SliceStable(result.Cancelled, func(i, j int) bool { return result.Cancelled[i].UpdatedAt.After(result.Cancelled[j].UpdatedAt) })
	result.Completed, result.completedTruncated = capPlanningStories(result.Completed, limit)
	result.Remaining, result.remainingTruncated = capPlanningStories(result.Remaining, limit)
	result.Cancelled, result.cancelledTruncated = capPlanningStories(result.Cancelled, limit)
	result.Truncated = result.completedTruncated || result.remainingTruncated || result.cancelledTruncated
	return result, nil
}

func capPlanningStories(items []planningStoryResult, limit int) ([]planningStoryResult, bool) {
	if len(items) <= limit {
		return items, false
	}
	return items[:limit], true
}

func newSprintResult(sprint sprints.CoreSprint, team teams.CoreTeam) sprintResult {
	return sprintResult{
		ID:                sprint.ID,
		Name:              sprint.Name,
		Goal:              sprint.Goal,
		TeamID:            sprint.Team,
		TeamName:          team.Name,
		TeamCode:          strings.ToUpper(strings.TrimSpace(team.Code)),
		StartDate:         formatPlanningDate(sprint.StartDate),
		EndDate:           formatPlanningDate(sprint.EndDate),
		ProgressPercent:   completionPercent(sprint.CompletedStories, sprint.TotalStories),
		TotalStories:      sprint.TotalStories,
		CompletedStories:  sprint.CompletedStories,
		InProgressStories: sprint.StartedStories,
		TodoStories:       sprint.UnstartedStories,
		CancelledStories:  sprint.CancelledStories,
	}
}

func formatPlanningDate(value time.Time) string {
	return value.UTC().Format(time.DateOnly)
}

func formatPlanningDatePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := formatPlanningDate(*value)
	return &formatted
}

func completionPercent(completed, total int) int {
	if total <= 0 || completed <= 0 {
		return 0
	}
	if completed >= total {
		return 100
	}
	return (completed*100 + total/2) / total
}

type sprintResult struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Goal              *string   `json:"goal,omitempty"`
	TeamID            uuid.UUID `json:"team_id"`
	TeamName          string    `json:"team_name"`
	TeamCode          string    `json:"team_code"`
	StartDate         string    `json:"start_date"`
	EndDate           string    `json:"end_date"`
	Status            string    `json:"status,omitempty"`
	ProgressPercent   int       `json:"progress_percent"`
	TotalStories      int       `json:"total_stories"`
	CompletedStories  int       `json:"completed_stories"`
	InProgressStories int       `json:"in_progress_stories"`
	TodoStories       int       `json:"todo_stories"`
	BlockedStories    int       `json:"blocked_stories,omitempty"`
	CancelledStories  int       `json:"cancelled_stories"`
	DaysElapsed       *int      `json:"days_elapsed,omitempty"`
	DaysRemaining     *int      `json:"days_remaining,omitempty"`
}

type listSprintsResult struct {
	Total     int            `json:"total"`
	Truncated bool           `json:"truncated"`
	Sprints   []sprintResult `json:"sprints"`
}

type planningObjectiveResult struct {
	ID                uuid.UUID `json:"id"`
	SequenceID        int       `json:"sequence_id"`
	Name              string    `json:"name"`
	ShortSummary      *string   `json:"short_summary,omitempty"`
	TeamID            uuid.UUID `json:"team_id"`
	TeamName          string    `json:"team_name"`
	TeamCode          string    `json:"team_code"`
	Health            string    `json:"health,omitempty"`
	Priority          *string   `json:"priority,omitempty"`
	StartDate         *string   `json:"start_date,omitempty"`
	EndDate           *string   `json:"end_date,omitempty"`
	KeyResultCount    int       `json:"key_result_count"`
	ProgressPercent   int       `json:"progress_percent"`
	TotalStories      int       `json:"total_stories"`
	CompletedStories  int       `json:"completed_stories"`
	InProgressStories int       `json:"in_progress_stories"`
	TodoStories       int       `json:"todo_stories"`
	BlockedStories    int       `json:"blocked_stories,omitempty"`
	CancelledStories  int       `json:"cancelled_stories"`
}

type objectiveProgress struct {
	Total      int
	Completed  int
	InProgress int
	Todo       int
	Blocked    int
	Cancelled  int
	Backlog    int
}

type objectiveSummaryResult struct {
	Objective planningObjectiveResult `json:"objective"`
	Work      planningWorkResult      `json:"work"`
}

type sprintSummaryResult struct {
	Sprint sprintResult       `json:"sprint"`
	Work   planningWorkResult `json:"work"`
}

type planningStoryResult struct {
	ID          uuid.UUID  `json:"id"`
	Reference   string     `json:"reference"`
	URL         string     `json:"url,omitempty"`
	Title       string     `json:"title"`
	TeamName    string     `json:"team_name"`
	StatusName  string     `json:"status_name,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type planningWorkResult struct {
	Access             string                `json:"access"`
	Total              int                   `json:"total"`
	Truncated          bool                  `json:"truncated"`
	Completed          []planningStoryResult `json:"completed"`
	Remaining          []planningStoryResult `json:"remaining"`
	Cancelled          []planningStoryResult `json:"cancelled"`
	completedTruncated bool                  `json:"-"`
	remainingTruncated bool                  `json:"-"`
	cancelledTruncated bool                  `json:"-"`
}

var (
	_ SprintsService            = (*sprints.Service)(nil)
	_ ObjectiveAnalyticsService = (*objectives.Service)(nil)
	_ StoriesService            = (*stories.Service)(nil)
)
