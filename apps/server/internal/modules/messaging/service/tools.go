package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	toolListTeams      = "list_teams"
	toolListMyTasks    = "list_my_tasks"
	toolSearchWork     = "search_work"
	toolListObjectives = "list_objectives"

	defaultToolLimit = 20
	maxToolLimit     = 50
	maxSearchRunes   = 200
)

// TeamsService is the subset of the teams domain used by assistant tools.
type TeamsService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error)
}

// StoriesService is the subset of the stories domain used by assistant tools.
type StoriesService interface {
	MyStories(ctx context.Context, workspaceID uuid.UUID) ([]stories.CoreStoryList, error)
}

// SearchService is the subset of the search domain used by assistant tools.
type SearchService interface {
	Search(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, params search.SearchParams) (search.CoreSearchResult, error)
}

// ObjectivesService is the subset of the objectives domain used by assistant tools.
type ObjectivesService interface {
	List(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, filters map[string]any) ([]objectives.CoreObjective, error)
}

// FortyOneToolExecutor exposes the deliberately small, read-only FortyOne tool
// catalog used by messaging assistants.
type FortyOneToolExecutor struct {
	teams       TeamsService
	stories     StoriesService
	search      SearchService
	objectives  ObjectivesService
	definitions []ToolDefinition
}

// NewFortyOneToolExecutor constructs a read-only executor backed by the existing
// domain services. Every execution re-establishes authoritative user context and
// resolves joined teams before returning data.
func NewFortyOneToolExecutor(
	teamsService TeamsService,
	storiesService StoriesService,
	searchService SearchService,
	objectivesService ObjectivesService,
) (*FortyOneToolExecutor, error) {
	if teamsService == nil || storiesService == nil || searchService == nil || objectivesService == nil {
		return nil, errors.New("all FortyOne assistant tool services are required")
	}

	return &FortyOneToolExecutor{
		teams:       teamsService,
		stories:     storiesService,
		search:      searchService,
		objectives:  objectivesService,
		definitions: fortyOneToolDefinitions(),
	}, nil
}

// Definitions returns a defensive copy of the fixed read-only catalog.
func (e *FortyOneToolExecutor) Definitions() []ToolDefinition {
	return cloneToolDefinitions(e.definitions)
}

// Execute runs one read-only tool in the supplied authoritative scope.
func (e *FortyOneToolExecutor) Execute(ctx context.Context, scope ToolScope, call ToolCall) (json.RawMessage, error) {
	if scope.WorkspaceID == uuid.Nil || scope.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace and user are required", ErrInvalidRequest)
	}

	ctx = platformauth.SetUserID(ctx, scope.UserID)
	switch call.Name {
	case toolListTeams:
		return e.listTeams(ctx, scope, call.Arguments)
	case toolListMyTasks:
		return e.listMyTasks(ctx, scope, call.Arguments)
	case toolSearchWork:
		return e.searchWork(ctx, scope, call.Arguments)
	case toolListObjectives:
		return e.listObjectives(ctx, scope, call.Arguments)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
	}
}

func (e *FortyOneToolExecutor) listTeams(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct{}
	if err := decodeToolArguments(raw, &args); err != nil {
		return nil, err
	}

	joined, _, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	result := listTeamsResult{
		Total: len(joined),
		Teams: make([]teamResult, 0, min(len(joined), maxToolLimit)),
	}
	for _, team := range joined {
		if len(result.Teams) == maxToolLimit {
			result.Truncated = true
			break
		}
		result.Teams = append(result.Teams, teamResult{
			ID:             team.ID,
			Name:           team.Name,
			Code:           team.Code,
			IsPrivate:      team.IsPrivate,
			MemberCount:    team.MemberCount,
			SprintsEnabled: team.SprintsEnabled,
		})
	}
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) listMyTasks(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		Limit *int `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	items, err := e.stories.MyStories(ctx, scope.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("list my tasks: %w", err)
	}

	filtered := make([]taskResult, 0, min(len(items), limit))
	total := 0
	for _, story := range items {
		team, allowed := joinedByID[story.Team]
		if !allowed || story.Workspace != scope.WorkspaceID {
			continue
		}
		total++
		if len(filtered) == limit {
			continue
		}
		filtered = append(filtered, taskResult{
			ID:         story.ID,
			Reference:  storyReference(team.Code, story.SequenceID),
			Title:      story.Title,
			TeamID:     story.Team,
			StatusID:   story.Status,
			AssigneeID: story.Assignee,
			Priority:   story.Priority,
			EndDate:    story.EndDate,
			UpdatedAt:  story.UpdatedAt,
		})
	}

	return marshalToolResult(listTasksResult{
		Total:     total,
		Truncated: total > len(filtered),
		Tasks:     filtered,
	})
}

func (e *FortyOneToolExecutor) searchWork(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		Query  string  `json:"query"`
		TeamID *string `json:"team_id"`
		Kind   *string `json:"kind"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "query", "team_id", "kind", "limit"); err != nil {
		return nil, err
	}
	query := strings.TrimSpace(args.Query)
	if query == "" || len([]rune(query)) > maxSearchRunes {
		return nil, fmt.Errorf("%w: query must contain 1-%d characters", ErrInvalidToolArguments, maxSearchRunes)
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}

	searchType := search.SearchTypeAll
	if args.Kind != nil {
		switch *args.Kind {
		case "all":
			searchType = search.SearchTypeAll
		case "stories":
			searchType = search.SearchTypeStories
		case "objectives":
			searchType = search.SearchTypeObjectives
		default:
			return nil, fmt.Errorf("%w: unsupported search kind %q", ErrInvalidToolArguments, *args.Kind)
		}
	}

	result, err := e.search.Search(ctx, scope.WorkspaceID, scope.UserID, search.SearchParams{
		Type:     searchType,
		Query:    query,
		TeamID:   teamID,
		SortBy:   search.SortByRelevance,
		Page:     1,
		PageSize: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("search work: %w", err)
	}

	storiesResult := make([]searchStoryResult, 0, min(len(result.Stories), limit))
	for _, story := range result.Stories {
		team, allowed := joinedByID[story.Team]
		if !allowed || story.Workspace != scope.WorkspaceID || len(storiesResult) == limit {
			continue
		}
		storiesResult = append(storiesResult, searchStoryResult{
			ID:        story.ID,
			Reference: storyReference(team.Code, story.SequenceID),
			Title:     story.Title,
			TeamID:    story.Team,
			StatusID:  story.Status,
			Priority:  story.Priority,
			UpdatedAt: story.UpdatedAt,
		})
	}

	objectivesResult := make([]searchObjectiveResult, 0, min(len(result.Objectives), limit))
	for _, objective := range result.Objectives {
		if _, allowed := joinedByID[objective.Team]; !allowed || objective.Workspace != scope.WorkspaceID || len(objectivesResult) == limit {
			continue
		}
		objectivesResult = append(objectivesResult, searchObjectiveResult{
			ID:           objective.ID,
			Name:         objective.Name,
			ShortSummary: objective.ShortSummary,
			TeamID:       objective.Team,
			StatusID:     objective.Status,
			Priority:     objective.Priority,
			Health:       objective.Health,
			UpdatedAt:    objective.UpdatedAt,
		})
	}

	return marshalToolResult(searchWorkResult{
		Stories:    storiesResult,
		Objectives: objectivesResult,
	})
}

func (e *FortyOneToolExecutor) listObjectives(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamID *string `json:"team_id"`
		Query  *string `json:"query"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "query", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}

	filters := map[string]any{"limit": limit}
	if teamID != nil {
		filters["team_id"] = *teamID
	}
	if args.Query != nil {
		query := strings.TrimSpace(*args.Query)
		if len([]rune(query)) > maxSearchRunes {
			return nil, fmt.Errorf("%w: query must not exceed %d characters", ErrInvalidToolArguments, maxSearchRunes)
		}
		if query != "" {
			filters["search"] = query
		}
	}

	items, err := e.objectives.List(ctx, scope.WorkspaceID, scope.UserID, filters)
	if err != nil {
		return nil, fmt.Errorf("list objectives: %w", err)
	}
	result := make([]objectiveResult, 0, min(len(items), limit))
	for _, objective := range items {
		if _, allowed := joinedByID[objective.Team]; !allowed || objective.Workspace != scope.WorkspaceID || len(result) == limit {
			continue
		}
		var health *string
		if objective.Health != nil {
			value := string(*objective.Health)
			health = &value
		}
		result = append(result, objectiveResult{
			ID:               objective.ID,
			SequenceID:       objective.SequenceID,
			Name:             objective.Name,
			ShortSummary:     objective.ShortSummary,
			TeamID:           objective.Team,
			StatusID:         objective.Status,
			Priority:         objective.Priority,
			Health:           health,
			StartDate:        objective.StartDate,
			EndDate:          objective.EndDate,
			KeyResultCount:   objective.KeyResultCount,
			TotalStories:     objective.TotalStories,
			CompletedStories: objective.CompletedStories,
			UpdatedAt:        objective.UpdatedAt,
		})
	}
	return marshalToolResult(listObjectivesResult{
		Count:      len(result),
		Objectives: result,
	})
}

func (e *FortyOneToolExecutor) joinedTeams(ctx context.Context, scope ToolScope) ([]teams.CoreTeam, map[uuid.UUID]teams.CoreTeam, error) {
	items, err := e.teams.List(ctx, scope.WorkspaceID, scope.UserID, teams.CoreListTeamsFilter{JoinedOnly: true})
	if err != nil {
		return nil, nil, fmt.Errorf("list joined teams: %w", err)
	}
	joined := make([]teams.CoreTeam, 0, len(items))
	joinedByID := make(map[uuid.UUID]teams.CoreTeam, len(items))
	for _, team := range items {
		if team.ID == uuid.Nil || team.Workspace != scope.WorkspaceID {
			continue
		}
		if _, duplicate := joinedByID[team.ID]; duplicate {
			continue
		}
		joined = append(joined, team)
		joinedByID[team.ID] = team
	}
	return joined, joinedByID, nil
}

func accessibleTeamID(raw *string, joined map[uuid.UUID]teams.CoreTeam) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	teamID, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil || teamID == uuid.Nil {
		return nil, fmt.Errorf("%w: team_id must be a UUID or null", ErrInvalidToolArguments)
	}
	if _, ok := joined[teamID]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, teamID)
	}
	return &teamID, nil
}

func normalizedLimit(value *int) (int, error) {
	if value == nil {
		return defaultToolLimit, nil
	}
	if *value < 1 || *value > maxToolLimit {
		return 0, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidToolArguments, maxToolLimit)
	}
	return *value, nil
}

func decodeToolArguments(raw json.RawMessage, target any, required ...string) error {
	if len(raw) == 0 {
		return fmt.Errorf("%w: arguments are required", ErrInvalidToolArguments)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return fmt.Errorf("%w: arguments must be a JSON object", ErrInvalidToolArguments)
	}
	for _, key := range required {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("%w: missing %s", ErrInvalidToolArguments, key)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: arguments contain trailing data", ErrInvalidToolArguments)
	}
	return nil
}

func marshalToolResult(value any) (json.RawMessage, error) {
	result, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal tool result: %w", err)
	}
	return result, nil
}

func storyReference(teamCode string, sequenceID int) string {
	if teamCode == "" {
		return fmt.Sprintf("#%d", sequenceID)
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID)
}

func fortyOneToolDefinitions() []ToolDefinition {
	nullableLimit := func() map[string]any {
		return map[string]any{
			"type":        []string{"integer", "null"},
			"description": "Maximum number of results, or null for the default.",
			"minimum":     1,
			"maximum":     maxToolLimit,
		}
	}
	nullableTeamID := func() map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": "A team UUID returned by list_teams, or null for all joined teams.",
		}
	}

	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListTeams,
			Description: "List only the FortyOne teams the current user has joined.",
			Strict:      true,
			Parameters:  strictObjectSchema(map[string]any{}, []string{}),
		},
		{
			Type:        "function",
			Name:        toolListMyTasks,
			Description: "List tasks relevant to the current user across only their joined teams.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"limit": nullableLimit(),
			}, []string{"limit"}),
		},
		{
			Type:        "function",
			Name:        toolSearchWork,
			Description: "Search task and objective titles within only the current user's joined teams.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Plain-language text to search for.",
					"minLength":   1,
					"maxLength":   maxSearchRunes,
				},
				"team_id": nullableTeamID(),
				"kind": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Limit the search to stories or objectives, or null/all for both.",
					"enum":        []any{"all", "stories", "objectives", nil},
				},
				"limit": nullableLimit(),
			}, []string{"query", "team_id", "kind", "limit"}),
		},
		{
			Type:        "function",
			Name:        toolListObjectives,
			Description: "List objectives within only the current user's joined teams.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": nullableTeamID(),
				"query": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Optional objective-name search text, or null.",
					"maxLength":   maxSearchRunes,
				},
				"limit": nullableLimit(),
			}, []string{"team_id", "query", "limit"}),
		},
	}
}

func strictObjectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

type teamResult struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	IsPrivate      bool      `json:"is_private"`
	MemberCount    int       `json:"member_count"`
	SprintsEnabled bool      `json:"sprints_enabled"`
}

type listTeamsResult struct {
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated"`
	Teams     []teamResult `json:"teams"`
}

type taskResult struct {
	ID         uuid.UUID  `json:"id"`
	Reference  string     `json:"reference"`
	Title      string     `json:"title"`
	TeamID     uuid.UUID  `json:"team_id"`
	StatusID   *uuid.UUID `json:"status_id"`
	AssigneeID *uuid.UUID `json:"assignee_id"`
	Priority   string     `json:"priority"`
	EndDate    *time.Time `json:"end_date"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type listTasksResult struct {
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated"`
	Tasks     []taskResult `json:"tasks"`
}

type searchStoryResult struct {
	ID        uuid.UUID  `json:"id"`
	Reference string     `json:"reference"`
	Title     string     `json:"title"`
	TeamID    uuid.UUID  `json:"team_id"`
	StatusID  *uuid.UUID `json:"status_id"`
	Priority  string     `json:"priority"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type searchObjectiveResult struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ShortSummary *string   `json:"short_summary"`
	TeamID       uuid.UUID `json:"team_id"`
	StatusID     uuid.UUID `json:"status_id"`
	Priority     *string   `json:"priority"`
	Health       *string   `json:"health"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type searchWorkResult struct {
	Stories    []searchStoryResult     `json:"stories"`
	Objectives []searchObjectiveResult `json:"objectives"`
}

type objectiveResult struct {
	ID               uuid.UUID  `json:"id"`
	SequenceID       int        `json:"sequence_id"`
	Name             string     `json:"name"`
	ShortSummary     *string    `json:"short_summary"`
	TeamID           uuid.UUID  `json:"team_id"`
	StatusID         uuid.UUID  `json:"status_id"`
	Priority         *string    `json:"priority"`
	Health           *string    `json:"health"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	KeyResultCount   int        `json:"key_result_count"`
	TotalStories     int        `json:"total_stories"`
	CompletedStories int        `json:"completed_stories"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type listObjectivesResult struct {
	Count      int               `json:"count"`
	Objectives []objectiveResult `json:"objectives"`
}
