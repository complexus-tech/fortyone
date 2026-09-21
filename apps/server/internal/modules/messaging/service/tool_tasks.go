package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/google/uuid"
)

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
			ID:                team.ID,
			Name:              team.Name,
			Code:              team.Code,
			IsPrivate:         team.IsPrivate,
			MemberCount:       team.MemberCount,
			SprintsEnabled:    team.SprintsEnabled,
			SharedWorkEnabled: teamWorkSharedTeamAllowed(scope.SharedTeamIDs, team.ID),
		})
	}
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) listMyTasks(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		Limit      *int    `json:"limit"`
		AssignedBy *string `json:"assigned_by"`
	}
	if err := decodeToolArguments(raw, &args, "limit", "assigned_by"); err != nil {
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

	var statusesByID map[uuid.UUID]messagingState
	if e.states != nil {
		_, statusesByID, err = e.scopedStatuses(ctx, scope, joinedByID)
		if err != nil {
			return nil, err
		}
	}

	var assigneeName string
	var assigneeUsername string
	if e.users != nil {
		currentUser, getUserErr := e.users.GetUser(ctx, scope.UserID)
		if getUserErr != nil {
			return nil, fmt.Errorf("load current user for task enrichment: %w", getUserErr)
		}
		if currentUser.ID != scope.UserID || !currentUser.IsActive || currentUser.IsSystem {
			return nil, errors.New("load current user for task enrichment: unexpected inactive or mismatched user")
		}
		assigneeName = memberDisplayName(currentUser)
		assigneeUsername = strings.TrimSpace(currentUser.Username)
	}

	eligible := make([]storydomain.StoryList, 0, len(items))
	for _, story := range items {
		_, allowed := joinedByID[story.Team]
		if !allowed || story.Workspace != scope.WorkspaceID || story.Assignee == nil || *story.Assignee != scope.UserID {
			continue
		}
		if story.CompletedAt != nil || story.DeletedAt != nil || story.ArchivedAt != nil {
			continue
		}
		if story.Status != nil && statusesByID != nil {
			if status, visible := statusesByID[*story.Status]; visible {
				if statusIsClosed(status.Category) {
					continue
				}
			}
		}
		eligible = append(eligible, story)
	}

	var assignmentFilter *assignmentFilterResult
	if args.AssignedBy != nil {
		query := strings.TrimSpace(*args.AssignedBy)
		if query == "" {
			return nil, errors.New("assigned_by must not be empty")
		}
		if utf8.RuneCountInString(query) > 100 {
			return nil, errors.New("assigned_by must be 100 characters or fewer")
		}
		matched := storydomain.FilterStoriesAssignedBy(eligible, query)
		assignmentFilter = assignmentFilterToolResult(query, matched)
		eligible = matched.Stories
	}

	total := len(eligible)
	filtered := make([]taskResult, 0, min(total, limit))
	for _, story := range eligible {
		if len(filtered) == limit {
			break
		}
		team := joinedByID[story.Team]
		var statusName string
		var statusCategory string
		if story.Status != nil && statusesByID != nil {
			if status, visible := statusesByID[*story.Status]; visible {
				statusName = status.Name
				statusCategory = status.Category
			}
		}
		assignedByID, assignedByName, assignedByUsername := assignmentActorFields(story.AssignedBy)
		filtered = append(filtered, taskResult{
			ID:                       story.ID,
			Reference:                storyReference(team.Code, story.SequenceID),
			URL:                      storyURL(scope, storyReference(team.Code, story.SequenceID)),
			Title:                    story.Title,
			TeamID:                   story.Team,
			TeamName:                 team.Name,
			TeamCode:                 strings.ToUpper(strings.TrimSpace(team.Code)),
			StatusID:                 story.Status,
			StatusName:               statusName,
			StatusCategory:           statusCategory,
			AssigneeID:               story.Assignee,
			AssigneeName:             assigneeName,
			AssigneeUsername:         assigneeUsername,
			AssignedByID:             assignedByID,
			AssignedByName:           assignedByName,
			AssignedByUsername:       assignedByUsername,
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
			EndDate:                  story.EndDate,
			UpdatedAt:                story.UpdatedAt,
		})
	}

	return marshalToolResult(listTasksResult{
		Total:            total,
		Truncated:        total > len(filtered),
		AssignmentFilter: assignmentFilter,
		Tasks:            filtered,
	})
}

func assignmentFilterToolResult(query string, result storydomain.AssignmentFilterResult) *assignmentFilterResult {
	filter := &assignmentFilterResult{Query: query, Status: "not_found"}
	if len(result.Candidates) > 1 {
		filter.Status = "ambiguous"
	}
	filter.Candidates = make([]assignmentActorResult, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		filter.Candidates = append(filter.Candidates, assignmentActorToolResult(candidate))
	}
	if result.Resolved != nil {
		resolved := assignmentActorToolResult(*result.Resolved)
		filter.Status = "matched"
		filter.Resolved = &resolved
	}
	return filter
}

func assignmentActorToolResult(actor storydomain.AssignmentActor) assignmentActorResult {
	return assignmentActorResult{Name: assignmentActorDisplayName(actor), Username: strings.TrimSpace(actor.Username)}
}

func assignmentActorFields(actor *storydomain.AssignmentActor) (*uuid.UUID, string, string) {
	if actor == nil {
		return nil, "", ""
	}
	id := actor.ID
	return &id, assignmentActorDisplayName(*actor), strings.TrimSpace(actor.Username)
}

func assignmentActorDisplayName(actor storydomain.AssignmentActor) string {
	if name := strings.TrimSpace(actor.FullName); name != "" {
		return name
	}
	return strings.TrimSpace(actor.Username)
}

func (e *FortyOneToolExecutor) listCompletedTasks(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
		Limit     *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "start_date", "end_date", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}
	completedAfter, completedBefore, err := completedTaskDateRange(args.StartDate, args.EndDate, scope.Timezone, time.Now())
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	assignedToMe, showSubStories := true, false
	items, err := e.completed.List(ctx, scope.WorkspaceID, messagingStoryFilters{
		AssignedToMe:    &assignedToMe,
		Categories:      []string{"completed"},
		CompletedAfter:  &completedAfter,
		CompletedBefore: &completedBefore,
		ShowSubStories:  &showSubStories,
	})
	if err != nil {
		return nil, fmt.Errorf("list completed tasks: %w", err)
	}

	var statusesByID map[uuid.UUID]messagingState
	if e.states != nil {
		_, statusesByID, err = e.scopedStatuses(ctx, scope, joinedByID)
		if err != nil {
			return nil, err
		}
	}

	var assigneeName string
	var assigneeUsername string
	if e.users != nil {
		currentUser, getUserErr := e.users.GetUser(ctx, scope.UserID)
		if getUserErr != nil {
			return nil, fmt.Errorf("load current user for completed task enrichment: %w", getUserErr)
		}
		if currentUser.ID != scope.UserID || !currentUser.IsActive || currentUser.IsSystem {
			return nil, errors.New("load current user for completed task enrichment: unexpected inactive or mismatched user")
		}
		assigneeName = memberDisplayName(currentUser)
		assigneeUsername = strings.TrimSpace(currentUser.Username)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].CompletedAt == nil {
			return false
		}
		if items[j].CompletedAt == nil {
			return true
		}
		return items[i].CompletedAt.After(*items[j].CompletedAt)
	})

	tasks := make([]taskResult, 0, min(len(items), limit))
	total := 0
	for _, story := range items {
		team, allowed := joinedByID[story.Team]
		if !allowed || story.Workspace != scope.WorkspaceID || story.Assignee == nil || *story.Assignee != scope.UserID {
			continue
		}
		if story.CompletedAt == nil || story.CompletedAt.Before(completedAfter) || story.CompletedAt.After(completedBefore) || story.DeletedAt != nil || story.ArchivedAt != nil {
			continue
		}

		var statusName string
		var statusCategory string
		if story.Status != nil && statusesByID != nil {
			if status, visible := statusesByID[*story.Status]; visible {
				statusName = status.Name
				statusCategory = status.Category
				if !strings.EqualFold(strings.TrimSpace(statusCategory), "completed") {
					continue
				}
			}
		}
		total++
		if len(tasks) == limit {
			continue
		}
		tasks = append(tasks, taskResult{
			ID:                       story.ID,
			Reference:                storyReference(team.Code, story.SequenceID),
			URL:                      storyURL(scope, storyReference(team.Code, story.SequenceID)),
			Title:                    story.Title,
			TeamID:                   story.Team,
			TeamName:                 team.Name,
			TeamCode:                 strings.ToUpper(strings.TrimSpace(team.Code)),
			StatusID:                 story.Status,
			StatusName:               statusName,
			StatusCategory:           statusCategory,
			AssigneeID:               story.Assignee,
			AssigneeName:             assigneeName,
			AssigneeUsername:         assigneeUsername,
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
			EndDate:                  story.EndDate,
			CompletedAt:              story.CompletedAt,
			UpdatedAt:                story.UpdatedAt,
		})
	}

	return marshalToolResult(listTasksResult{
		Total:     total,
		Truncated: total > len(tasks),
		Tasks:     tasks,
	})
}

func completedTaskDateRange(startDate, endDate *string, timezone string, now time.Time) (time.Time, time.Time, error) {
	location := time.UTC
	if strings.TrimSpace(timezone) != "" {
		loaded, err := time.LoadLocation(strings.TrimSpace(timezone))
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid completed task timezone %q: %w", timezone, err)
		}
		location = loaded
	}

	today := now.In(location)
	start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	end := start
	if startDate != nil {
		parsed, err := parseCompletedTaskDate(*startDate, location)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start = parsed
		end = parsed
	}
	if endDate != nil {
		parsed, err := parseCompletedTaskDate(*endDate, location)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end = parsed
		if startDate == nil {
			start = parsed
		}
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("completed task end_date must be on or after start_date")
	}
	if end.AddDate(0, 0, -maxCompletedTaskDays).After(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("completed task date range cannot exceed %d days", maxCompletedTaskDays)
	}

	endExclusive := end.AddDate(0, 0, 1)
	return start.UTC(), endExclusive.Add(-time.Nanosecond).UTC(), nil
}

func parseCompletedTaskDate(value string, location *time.Location) (time.Time, error) {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), location)
	if err != nil {
		return time.Time{}, fmt.Errorf("completed task dates must use YYYY-MM-DD: %w", err)
	}
	return parsed, nil
}
