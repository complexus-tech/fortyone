package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/google/uuid"
)

const (
	defaultTeamWorkLimit     = 50
	defaultAssigneeWorkLimit = 10
	maxSelectedAssignees     = 50
	maxTeamWorkAssignees     = 50

	teamWorkAssigneeMe       = "me"
	teamWorkAssigneeSelected = "selected"
	teamWorkAssigneeAll      = "all"

	teamWorkModeInProgress = "in_progress"
	teamWorkModeActive     = "active"
	teamWorkModeCompleted  = "completed"
	teamWorkModeDue        = "due"

	teamWorkGroupNone     = "none"
	teamWorkGroupAssignee = "assignee"

	teamWorkAccessGranted = "granted"
	teamWorkAccessDenied  = "denied"
)

var activeTeamWorkCategories = []string{"backlog", "unstarted", "started", "paused"}

func (e *FortyOneToolExecutor) listTeamWork(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamID           string   `json:"team_id"`
		AssigneeScope    string   `json:"assignee_scope"`
		AssigneeIDs      []string `json:"assignee_ids"`
		Mode             string   `json:"mode"`
		StartDate        *string  `json:"start_date"`
		EndDate          *string  `json:"end_date"`
		GroupBy          string   `json:"group_by"`
		Limit            *int     `json:"limit"`
		LimitPerAssignee *int     `json:"limit_per_assignee"`
	}
	if err := decodeToolArguments(
		raw,
		&args,
		"team_id",
		"assignee_scope",
		"assignee_ids",
		"mode",
		"start_date",
		"end_date",
		"group_by",
		"limit",
		"limit_per_assignee",
	); err != nil {
		return nil, err
	}

	args.AssigneeScope = strings.ToLower(strings.TrimSpace(args.AssigneeScope))
	args.Mode = strings.ToLower(strings.TrimSpace(args.Mode))
	args.GroupBy = strings.ToLower(strings.TrimSpace(args.GroupBy))
	if err := validateTeamWorkArguments(args.AssigneeScope, args.AssigneeIDs, args.Mode, args.StartDate, args.EndDate, args.GroupBy, args.LimitPerAssignee); err != nil {
		return nil, err
	}
	limit, err := normalizedTeamWorkLimit(args.Limit)
	if err != nil {
		return nil, err
	}
	perAssigneeLimit := 0
	if args.GroupBy == teamWorkGroupAssignee {
		perAssigneeLimit, err = normalizedAssigneeWorkLimit(args.LimitPerAssignee)
		if err != nil {
			return nil, err
		}
		perAssigneeLimit = min(perAssigneeLimit, limit)
	}

	teamID, err := parseRequiredUUID(args.TeamID, "team_id")
	if err != nil {
		return nil, err
	}
	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	team, allowed := joinedByID[teamID]
	if !allowed {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, teamID)
	}
	sharedWorkEnabled := teamWorkSharedTeamAllowed(scope.SharedTeamIDs, teamID)
	if args.AssigneeScope != teamWorkAssigneeMe && !sharedWorkEnabled {
		result := buildTeamWorkResult(team, sharedWorkEnabled, args.AssigneeScope, args.Mode, args.GroupBy, nil, limit, perAssigneeLimit, nil, nil, nil, 0, true, false, false)
		result.Access = teamWorkAccessDenied
		result.AccessReason = "shared_team_scope_required"
		return marshalToolResult(result)
	}

	dateRange, err := teamWorkDateRangeForMode(args.Mode, args.StartDate, args.EndDate, scope.Timezone, time.Now())
	if err != nil {
		return nil, err
	}
	memberLimit := 0
	if args.AssigneeScope == teamWorkAssigneeAll {
		memberLimit = maxTeamWorkAssignees
	}
	members, membersByID, assigneesTruncated, err := e.activeTeamWorkMembers(ctx, scope.WorkspaceID, teamID, memberLimit)
	if err != nil {
		return nil, err
	}
	selectedMembers, err := resolveTeamWorkMembers(args.AssigneeScope, args.AssigneeIDs, scope.UserID, members, membersByID)
	if err != nil {
		return nil, err
	}
	if len(selectedMembers) == 0 {
		return marshalToolResult(buildTeamWorkResult(team, sharedWorkEnabled, args.AssigneeScope, args.Mode, args.GroupBy, dateRange, limit, perAssigneeLimit, selectedMembers, nil, nil, 0, !assigneesTruncated, assigneesTruncated, assigneesTruncated))
	}

	_, statusesByID, err := e.scopedStatuses(ctx, scope, joinedByID)
	if err != nil {
		return nil, err
	}
	assigneeIDs := make([]uuid.UUID, 0, len(selectedMembers))
	selectedByID := make(map[uuid.UUID]usersdomain.User, len(selectedMembers))
	for _, member := range selectedMembers {
		assigneeIDs = append(assigneeIDs, member.ID)
		selectedByID[member.ID] = member
	}
	categories := teamWorkCategories(args.Mode)
	showSubStories := false
	filters := storydomain.StoryFilters{
		TeamIDs:        []uuid.UUID{teamID},
		AssigneeIDs:    assigneeIDs,
		Categories:     append([]string(nil), categories...),
		CurrentUserID:  scope.UserID,
		WorkspaceID:    scope.WorkspaceID,
		ShowSubStories: &showSubStories,
	}
	if args.Mode == teamWorkModeCompleted {
		isCompleted := true
		filters.IsCompleted = &isCompleted
	} else {
		isNotCompleted := true
		filters.IsNotCompleted = &isNotCompleted
	}
	if dateRange != nil {
		switch args.Mode {
		case teamWorkModeCompleted:
			filters.CompletedAfter = &dateRange.CompletedAfter
			filters.CompletedBefore = &dateRange.CompletedBefore
		case teamWorkModeDue:
			filters.DeadlineAfter = &dateRange.DueAfter
			filters.DeadlineBefore = &dateRange.DueBefore
		}
	}

	orderBy := "updated"
	orderDirection := "desc"
	if args.Mode == teamWorkModeCompleted {
		orderBy = "completed"
	} else if args.Mode == teamWorkModeDue {
		orderBy = "deadline"
		orderDirection = "asc"
	}
	queryGroupBy := teamWorkGroupNone
	storiesPerGroup := limit
	if args.GroupBy == teamWorkGroupAssignee {
		queryGroupBy = teamWorkGroupAssignee
		storiesPerGroup = perAssigneeLimit
	}
	groups, err := e.teamWork.ListGroupedStories(ctx, storydomain.StoryQuery{
		Filters:         filters,
		GroupBy:         storydomain.StoryGroupBy(queryGroupBy),
		OrderBy:         storydomain.StoryOrderBy(orderBy),
		OrderDirection:  storydomain.SortDirection(orderDirection),
		StoriesPerGroup: storiesPerGroup,
		Page:            1,
		PageSize:        limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list team work: %w", err)
	}
	items := make([]teamWorkCandidate, 0, limit)
	groupMetadata := make(map[uuid.UUID]teamWorkGroupMetadata, len(selectedMembers))
	total := 0
	totalExact := !assigneesTruncated
	queryTruncated := assigneesTruncated
	flatLoaded := 0
	for _, group := range groups {
		if queryGroupBy == teamWorkGroupNone {
			if group.Key != teamWorkGroupNone {
				continue
			}
			if group.TotalCount > total {
				total = group.TotalCount
			}
			if group.HasMore || group.TotalCount > group.LoadedCount {
				queryTruncated = true
			}
			remaining := limit - len(items)
			if remaining <= 0 {
				queryTruncated = true
				continue
			}
			loaded := min(len(group.Stories), remaining)
			flatLoaded += loaded
			if group.TotalCount < loaded || group.LoadedCount != len(group.Stories) {
				totalExact = false
			}
			for _, story := range group.Stories[:loaded] {
				items = append(items, teamWorkCandidate{Story: story})
			}
			if loaded < len(group.Stories) {
				queryTruncated = true
			}
			continue
		}

		assigneeID, parseErr := uuid.Parse(strings.TrimSpace(group.Key))
		if parseErr != nil || assigneeID == uuid.Nil {
			continue
		}
		if _, selected := selectedByID[assigneeID]; !selected {
			continue
		}
		if _, duplicate := groupMetadata[assigneeID]; duplicate {
			continue
		}
		loaded := min(len(group.Stories), storiesPerGroup)
		groupTotal := max(group.TotalCount, loaded)
		metadata := teamWorkGroupMetadata{
			Total:      groupTotal,
			Loaded:     loaded,
			TotalExact: group.TotalCount >= loaded && group.LoadedCount == len(group.Stories),
			Truncated:  group.HasMore || groupTotal > loaded,
		}
		groupMetadata[assigneeID] = metadata
		total += groupTotal
		queryTruncated = queryTruncated || metadata.Truncated
		for _, story := range group.Stories[:loaded] {
			items = append(items, teamWorkCandidate{Story: story, AssigneeID: assigneeID, Grouped: true})
		}
		if loaded < len(group.Stories) {
			queryTruncated = true
		}
	}
	categorySet := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		categorySet[category] = struct{}{}
	}
	tasks := make([]taskResult, 0, len(items))
	validLoadedByGroup := make(map[uuid.UUID]int, len(groupMetadata))
	for _, candidate := range items {
		story := candidate.Story
		if story.Workspace != scope.WorkspaceID || story.Team != teamID || story.Parent != nil || story.Assignee == nil {
			continue
		}
		if candidate.Grouped && *story.Assignee != candidate.AssigneeID {
			continue
		}
		member, selected := selectedByID[*story.Assignee]
		if !selected || story.DeletedAt != nil || story.ArchivedAt != nil || story.Status == nil {
			continue
		}
		status, visible := statusesByID[*story.Status]
		if !visible || status.Team != teamID || status.Workspace != scope.WorkspaceID {
			continue
		}
		statusCategory := strings.ToLower(strings.TrimSpace(status.Category))
		if _, matchesMode := categorySet[statusCategory]; !matchesMode {
			continue
		}
		if !teamWorkStoryMatchesDate(story, args.Mode, dateRange) {
			continue
		}

		reference := storyReference(team.Code, story.SequenceID)
		tasks = append(tasks, taskResult{
			ID:               story.ID,
			Reference:        reference,
			URL:              storyURL(scope, reference),
			Title:            story.Title,
			TeamID:           team.ID,
			TeamName:         team.Name,
			TeamCode:         strings.ToUpper(strings.TrimSpace(team.Code)),
			StatusID:         story.Status,
			StatusName:       status.Name,
			StatusCategory:   statusCategory,
			AssigneeID:       story.Assignee,
			AssigneeName:     memberDisplayName(member),
			AssigneeUsername: strings.TrimSpace(member.Username),
			Priority:         story.Priority,
			EndDate:          story.EndDate,
			CompletedAt:      story.CompletedAt,
			UpdatedAt:        story.UpdatedAt,
		})
		if candidate.Grouped {
			validLoadedByGroup[candidate.AssigneeID]++
		}
	}
	sortTeamWorkTasks(tasks, args.Mode)
	if queryGroupBy == teamWorkGroupNone {
		dropped := flatLoaded - len(tasks)
		if dropped > 0 {
			total = max(len(tasks), total-dropped)
			totalExact = false
			queryTruncated = true
		}
	} else {
		total = 0
		for assigneeID, metadata := range groupMetadata {
			validLoaded := validLoadedByGroup[assigneeID]
			dropped := metadata.Loaded - validLoaded
			if dropped > 0 {
				metadata.Total = max(validLoaded, metadata.Total-dropped)
				metadata.TotalExact = false
				metadata.Truncated = true
			}
			groupMetadata[assigneeID] = metadata
			total += metadata.Total
			totalExact = totalExact && metadata.TotalExact
			queryTruncated = queryTruncated || metadata.Truncated
		}
	}
	if total < len(tasks) {
		total = len(tasks)
	}

	return marshalToolResult(buildTeamWorkResult(team, sharedWorkEnabled, args.AssigneeScope, args.Mode, args.GroupBy, dateRange, limit, perAssigneeLimit, selectedMembers, tasks, groupMetadata, total, totalExact, queryTruncated, assigneesTruncated))
}

func validateTeamWorkArguments(
	assigneeScope string,
	assigneeIDs []string,
	mode string,
	startDate, endDate *string,
	groupBy string,
	limitPerAssignee *int,
) error {
	switch assigneeScope {
	case teamWorkAssigneeMe, teamWorkAssigneeAll:
		if len(assigneeIDs) != 0 {
			return fmt.Errorf("%w: assignee_ids must be null or empty when assignee_scope is %s", ErrInvalidToolArguments, assigneeScope)
		}
	case teamWorkAssigneeSelected:
		if len(assigneeIDs) == 0 {
			return fmt.Errorf("%w: assignee_ids are required when assignee_scope is selected", ErrInvalidToolArguments)
		}
		if len(assigneeIDs) > maxSelectedAssignees {
			return fmt.Errorf("%w: assignee_ids cannot contain more than %d members", ErrInvalidToolArguments, maxSelectedAssignees)
		}
	default:
		return fmt.Errorf("%w: unsupported assignee_scope %q", ErrInvalidToolArguments, assigneeScope)
	}

	switch mode {
	case teamWorkModeInProgress, teamWorkModeActive:
		if startDate != nil || endDate != nil {
			return fmt.Errorf("%w: start_date and end_date must be null for %s work", ErrInvalidToolArguments, mode)
		}
	case teamWorkModeCompleted, teamWorkModeDue:
	default:
		return fmt.Errorf("%w: unsupported team work mode %q", ErrInvalidToolArguments, mode)
	}

	switch groupBy {
	case teamWorkGroupNone:
		if limitPerAssignee != nil {
			return fmt.Errorf("%w: limit_per_assignee must be null when group_by is none", ErrInvalidToolArguments)
		}
	case teamWorkGroupAssignee:
	default:
		return fmt.Errorf("%w: unsupported group_by %q", ErrInvalidToolArguments, groupBy)
	}
	return nil
}

func normalizedTeamWorkLimit(value *int) (int, error) {
	if value == nil {
		return defaultTeamWorkLimit, nil
	}
	if *value < 1 || *value > maxToolLimit {
		return 0, fmt.Errorf("%w: limit must be between 1 and %d", ErrInvalidToolArguments, maxToolLimit)
	}
	return *value, nil
}

func normalizedAssigneeWorkLimit(value *int) (int, error) {
	if value == nil {
		return defaultAssigneeWorkLimit, nil
	}
	if *value < 1 || *value > maxToolLimit {
		return 0, fmt.Errorf("%w: limit_per_assignee must be between 1 and %d", ErrInvalidToolArguments, maxToolLimit)
	}
	return *value, nil
}

func teamWorkSharedTeamAllowed(sharedTeamIDs []uuid.UUID, teamID uuid.UUID) bool {
	if len(sharedTeamIDs) == 0 {
		return false
	}
	for _, sharedTeamID := range sharedTeamIDs {
		if sharedTeamID == teamID {
			return true
		}
	}
	return false
}

func (e *FortyOneToolExecutor) activeTeamWorkMembers(
	ctx context.Context,
	workspaceID, teamID uuid.UUID,
	limit int,
) ([]usersdomain.User, map[uuid.UUID]usersdomain.User, bool, error) {
	queryLimit := 0
	if limit > 0 {
		queryLimit = limit + 1
	}
	items, err := e.users.List(ctx, workspaceID, usersdomain.ListUsersFilter{TeamID: &teamID, Limit: queryLimit})
	if err != nil {
		return nil, nil, false, fmt.Errorf("list team work members: %w", err)
	}
	members := make([]usersdomain.User, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, member := range items {
		if member.ID == uuid.Nil || !member.IsActive || member.IsSystem {
			continue
		}
		if _, duplicate := seen[member.ID]; duplicate {
			continue
		}
		members = append(members, member)
		seen[member.ID] = struct{}{}
	}
	sort.SliceStable(members, func(i, j int) bool {
		left := strings.ToLower(memberDisplayName(members[i]) + "\x00" + strings.TrimSpace(members[i].Username))
		right := strings.ToLower(memberDisplayName(members[j]) + "\x00" + strings.TrimSpace(members[j].Username))
		if left == right {
			return members[i].ID.String() < members[j].ID.String()
		}
		return left < right
	})
	truncated := limit > 0 && len(members) > limit
	if truncated {
		members = members[:limit]
	}
	membersByID := make(map[uuid.UUID]usersdomain.User, len(members))
	for _, member := range members {
		membersByID[member.ID] = member
	}
	return members, membersByID, truncated, nil
}

func resolveTeamWorkMembers(
	assigneeScope string,
	rawAssigneeIDs []string,
	actorID uuid.UUID,
	members []usersdomain.User,
	membersByID map[uuid.UUID]usersdomain.User,
) ([]usersdomain.User, error) {
	switch assigneeScope {
	case teamWorkAssigneeMe:
		member, ok := membersByID[actorID]
		if !ok {
			return nil, fmt.Errorf("%w: authenticated user is not an active member of the selected team", ErrTeamNotAccessible)
		}
		return []usersdomain.User{member}, nil
	case teamWorkAssigneeAll:
		return append([]usersdomain.User(nil), members...), nil
	case teamWorkAssigneeSelected:
		selectedIDs := make(map[uuid.UUID]struct{}, len(rawAssigneeIDs))
		for _, rawAssigneeID := range rawAssigneeIDs {
			assigneeID, err := parseRequiredUUID(rawAssigneeID, "assignee_ids")
			if err != nil {
				return nil, err
			}
			if _, ok := membersByID[assigneeID]; !ok {
				return nil, fmt.Errorf("%w: assignee %s is not an active member of the selected team", ErrInvalidToolArguments, assigneeID)
			}
			selectedIDs[assigneeID] = struct{}{}
		}
		selected := make([]usersdomain.User, 0, len(selectedIDs))
		for _, member := range members {
			if _, ok := selectedIDs[member.ID]; ok {
				selected = append(selected, member)
			}
		}
		return selected, nil
	default:
		return nil, fmt.Errorf("%w: unsupported assignee_scope %q", ErrInvalidToolArguments, assigneeScope)
	}
}

type teamWorkCandidate struct {
	Story      storydomain.StoryList
	AssigneeID uuid.UUID
	Grouped    bool
}

func teamWorkStoryMatchesDate(story storydomain.StoryList, mode string, dateRange *teamWorkDateRange) bool {
	switch mode {
	case teamWorkModeCompleted:
		return dateRange != nil && story.CompletedAt != nil &&
			!story.CompletedAt.Before(dateRange.CompletedAfter) &&
			!story.CompletedAt.After(dateRange.CompletedBefore)
	case teamWorkModeDue:
		if dateRange == nil || story.CompletedAt != nil || story.EndDate == nil {
			return false
		}
		deadline := time.Date(story.EndDate.Year(), story.EndDate.Month(), story.EndDate.Day(), 0, 0, 0, 0, time.UTC)
		return !deadline.Before(dateRange.DueAfter) && !deadline.After(dateRange.DueBefore)
	case teamWorkModeInProgress, teamWorkModeActive:
		return story.CompletedAt == nil
	default:
		return false
	}
}

func sortTeamWorkTasks(tasks []taskResult, mode string) {
	sort.SliceStable(tasks, func(i, j int) bool {
		left := tasks[i]
		right := tasks[j]
		switch mode {
		case teamWorkModeCompleted:
			if left.CompletedAt != nil && right.CompletedAt != nil && !left.CompletedAt.Equal(*right.CompletedAt) {
				return left.CompletedAt.After(*right.CompletedAt)
			}
		case teamWorkModeDue:
			if left.EndDate != nil && right.EndDate != nil && !left.EndDate.Equal(*right.EndDate) {
				return left.EndDate.Before(*right.EndDate)
			}
		default:
			if !left.UpdatedAt.Equal(right.UpdatedAt) {
				return left.UpdatedAt.After(right.UpdatedAt)
			}
		}
		if left.AssigneeName != right.AssigneeName {
			return strings.ToLower(left.AssigneeName) < strings.ToLower(right.AssigneeName)
		}
		return left.Reference < right.Reference
	})
}

func buildTeamWorkResult(
	team teamsdomain.Team,
	sharedWorkEnabled bool,
	assigneeScope, mode, groupBy string,
	dateRange *teamWorkDateRange,
	limit, perAssigneeLimit int,
	members []usersdomain.User,
	tasks []taskResult,
	groupMetadata map[uuid.UUID]teamWorkGroupMetadata,
	total int,
	totalExact bool,
	queryTruncated bool,
	assigneesTruncated bool,
) listTeamWorkResult {
	result := listTeamWorkResult{
		Team: teamResult{
			ID:                team.ID,
			Name:              team.Name,
			Code:              strings.ToUpper(strings.TrimSpace(team.Code)),
			IsPrivate:         team.IsPrivate,
			MemberCount:       team.MemberCount,
			SprintsEnabled:    team.SprintsEnabled,
			SharedWorkEnabled: sharedWorkEnabled,
		},
		Access:             teamWorkAccessGranted,
		AssigneeScope:      assigneeScope,
		Mode:               mode,
		GroupBy:            groupBy,
		Limit:              limit,
		LimitPerAssignee:   perAssigneeLimit,
		Total:              total,
		TotalIsExact:       totalExact,
		AssigneesTruncated: assigneesTruncated,
	}
	if dateRange != nil {
		startDate := dateRange.StartDate
		endDate := dateRange.EndDate
		result.StartDate = &startDate
		result.EndDate = &endDate
	}

	if groupBy == teamWorkGroupNone {
		returned := min(len(tasks), limit)
		result.Tasks = append([]taskResult(nil), tasks[:returned]...)
		result.Returned = returned
		result.Truncated = queryTruncated || returned < result.Total
		return result
	}

	tasksByAssignee := make(map[uuid.UUID][]taskResult, len(members))
	for _, task := range tasks {
		if task.AssigneeID != nil {
			tasksByAssignee[*task.AssigneeID] = append(tasksByAssignee[*task.AssigneeID], task)
		}
	}
	result.AssigneeTotal = len(members)
	allocatedByAssignee := make(map[uuid.UUID]int, len(groupMetadata))
	remaining := limit
	for remaining > 0 {
		allocatedThisRound := false
		for _, member := range members {
			memberTasks := tasksByAssignee[member.ID]
			allocated := allocatedByAssignee[member.ID]
			if allocated >= len(memberTasks) {
				continue
			}
			allocatedByAssignee[member.ID] = allocated + 1
			remaining--
			allocatedThisRound = true
			if remaining == 0 {
				break
			}
		}
		if !allocatedThisRound {
			break
		}
	}
	result.Groups = make([]teamWorkAssigneeGroup, 0, len(members))
	for _, member := range members {
		memberTasks := tasksByAssignee[member.ID]
		returned := allocatedByAssignee[member.ID]
		metadata, ok := groupMetadata[member.ID]
		if !ok {
			metadata.TotalExact = true
		}
		groupTotal := max(metadata.Total, len(memberTasks))
		result.Groups = append(result.Groups, teamWorkAssigneeGroup{
			AssigneeID:       member.ID,
			AssigneeName:     memberDisplayName(member),
			AssigneeUsername: strings.TrimSpace(member.Username),
			Total:            groupTotal,
			TotalIsExact:     metadata.TotalExact,
			Returned:         returned,
			Truncated:        metadata.Truncated || returned < groupTotal,
			Tasks:            append([]taskResult{}, memberTasks[:returned]...),
		})
		result.AssigneeReturned++
		result.Returned += returned
	}
	result.Truncated = queryTruncated || result.Returned < result.Total
	return result
}
