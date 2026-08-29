package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

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

	searchType := workSearchTypeAll
	if args.Kind != nil {
		switch *args.Kind {
		case "all":
			searchType = workSearchTypeAll
		case "stories":
			searchType = workSearchTypeStories
		case "objectives":
			searchType = workSearchTypeObjectives
		default:
			return nil, fmt.Errorf("%w: unsupported search kind %q", ErrInvalidToolArguments, *args.Kind)
		}
	}

	result, err := e.search.Search(ctx, scope.WorkspaceID, scope.UserID, workSearchParams{
		Type:     searchType,
		Query:    query,
		TeamID:   teamID,
		SortBy:   workSearchSortByRelevance,
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
			ID:                       story.ID,
			Reference:                storyReference(team.Code, story.SequenceID),
			URL:                      storyURL(scope, storyReference(team.Code, story.SequenceID)),
			Title:                    story.Title,
			TeamID:                   story.Team,
			StatusID:                 story.Status,
			Priority:                 story.Priority,
			EstimateLabel:            story.EstimateLabel,
			EstimateValue:            story.EstimateValue,
			EstimateScheme:           story.EstimateScheme,
			EstimatedDurationMinutes: story.EstimatedDurationMinutes,
			MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
			AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
			AutoSchedulingLocked:     story.AutoSchedulingLocked,
			AutoSchedulingStatus:     story.AutoSchedulingStatus,
			AutoSchedulingReason:     story.AutoSchedulingReason,
			AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
			UpdatedAt:                story.UpdatedAt,
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
