package storiesrepository

import (
	"fmt"
	"strings"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const (
	maxFilteredStoryResults = 1_000
	maxStoryFilterValues    = 100
	maxStoryGroupCatalog    = 1_000
	maxGroupedPageSize      = 100
	maxGroupedPage          = 10_000
)

func normalizeStoryFilters(filters storydomain.StoryFilters) (storydomain.StoryFilters, error) {
	if filters.Epic != nil {
		return filters, fmt.Errorf("%w: epic filtering is not supported by the story schema", storydomain.ErrInvalidReadQuery)
	}
	if err := validateUUIDFilterSlices(filters); err != nil {
		return filters, err
	}
	if err := validateStringFilter("priority", filters.Priorities, validStoryPriority); err != nil {
		return filters, err
	}
	if err := validateStringFilter("excluded priority", filters.ExcludedPriorities, validStoryPriority); err != nil {
		return filters, err
	}
	if err := validateStringFilter("category", filters.Categories, validStoryCategory); err != nil {
		return filters, err
	}
	if err := validateEstimateFilters(filters.EstimateValues, filters.ExcludedEstimateValues); err != nil {
		return filters, err
	}
	if enabled(filters.HasNoAssignee) && enabled(filters.HasAssignee) {
		return filters, fmt.Errorf("%w: assignee filters are contradictory", storydomain.ErrInvalidReadQuery)
	}
	if enabled(filters.IsCompleted) && enabled(filters.IsNotCompleted) {
		return filters, fmt.Errorf("%w: completion filters are contradictory", storydomain.ErrInvalidReadQuery)
	}
	if invalidRange(filters.CreatedAfter, filters.CreatedBefore) ||
		invalidRange(filters.UpdatedAfter, filters.UpdatedBefore) ||
		invalidRange(filters.StartDateAfter, filters.StartDateBefore) ||
		invalidRange(filters.DeadlineAfter, filters.DeadlineBefore) ||
		invalidRange(filters.CompletedAfter, filters.CompletedBefore) {
		return filters, fmt.Errorf("%w: a filter range starts after it ends", storydomain.ErrInvalidReadQuery)
	}
	if filters.Limit < 0 || filters.Limit > maxFilteredStoryResults+1 {
		return filters, fmt.Errorf("%w: limit must be between 0 and %d", storydomain.ErrInvalidReadQuery, maxFilteredStoryResults+1)
	}
	if filters.Offset < 0 || filters.Offset > maxGroupedPage*maxGroupedPageSize {
		return filters, fmt.Errorf("%w: offset is outside the supported window", storydomain.ErrInvalidReadQuery)
	}

	filters.TitleContains = normalizedOptionalText(filters.TitleContains)
	filters.TitleNotContains = normalizedOptionalText(filters.TitleNotContains)
	filters.CurrentUserID = uuid.Nil
	filters.WorkspaceID = uuid.Nil
	return filters, nil
}

func validateStoryQuery(query storydomain.StoryQuery) (storydomain.StoryQuery, error) {
	filters, err := normalizeStoryFilters(query.Filters)
	if err != nil {
		return query, err
	}
	query.Filters = filters
	if !validStoryGroup(query.GroupBy) {
		return query, fmt.Errorf("%w: unsupported group", storydomain.ErrInvalidReadQuery)
	}
	if !validStoryOrder(query.OrderBy) {
		return query, fmt.Errorf("%w: unsupported order", storydomain.ErrInvalidReadQuery)
	}
	if query.OrderDirection != storydomain.SortAscending && query.OrderDirection != storydomain.SortDescending {
		return query, fmt.Errorf("%w: unsupported order direction", storydomain.ErrInvalidReadQuery)
	}
	return query, nil
}

func validateInitialStoryGroupQuery(query storydomain.StoryQuery) (storydomain.StoryQuery, error) {
	query, err := validateStoryQuery(query)
	if err != nil {
		return query, err
	}
	if query.StoriesPerGroup < 1 || query.StoriesPerGroup > maxGroupedPageSize {
		return query, fmt.Errorf("%w: stories per group must be between 1 and %d", storydomain.ErrInvalidReadQuery, maxGroupedPageSize)
	}
	return query, nil
}

func validateGroupPage(query storydomain.StoryQuery, groupKey string) (int32, int32, error) {
	if query.Page < 1 || query.Page > maxGroupedPage {
		return 0, 0, fmt.Errorf("%w: page must be between 1 and %d", storydomain.ErrInvalidReadQuery, maxGroupedPage)
	}
	if query.PageSize < 1 || query.PageSize > maxGroupedPageSize {
		return 0, 0, fmt.Errorf("%w: page size must be between 1 and %d", storydomain.ErrInvalidReadQuery, maxGroupedPageSize)
	}
	if err := validateGroupKey(query.GroupBy, groupKey); err != nil {
		return 0, 0, err
	}
	offset, err := safecast.Int64ToInt32(int64(query.Page-1) * int64(query.PageSize))
	if err != nil {
		return 0, 0, fmt.Errorf("%w: page offset is outside the supported window", storydomain.ErrInvalidReadQuery)
	}
	limit, err := safecast.Int64ToInt32(int64(query.PageSize) + 1)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: page size is outside the supported window", storydomain.ErrInvalidReadQuery)
	}
	return offset, limit, nil
}

func validateGroupKey(groupBy storydomain.StoryGroupBy, key string) error {
	key = strings.TrimSpace(key)
	switch groupBy {
	case storydomain.StoryGroupNone:
		if key != "none" {
			return fmt.Errorf("%w: the ungrouped key must be none", storydomain.ErrInvalidReadQuery)
		}
	case storydomain.StoryGroupPriority:
		if !validStoryPriority(key) {
			return fmt.Errorf("%w: unsupported priority group", storydomain.ErrInvalidReadQuery)
		}
	case storydomain.StoryGroupAssignee, storydomain.StoryGroupSprint:
		if key != "null" {
			if _, err := uuid.Parse(key); err != nil {
				return fmt.Errorf("%w: invalid group identifier", storydomain.ErrInvalidReadQuery)
			}
		}
	case storydomain.StoryGroupStatus, storydomain.StoryGroupTeam:
		if _, err := uuid.Parse(key); err != nil {
			return fmt.Errorf("%w: invalid group identifier", storydomain.ErrInvalidReadQuery)
		}
	default:
		return fmt.Errorf("%w: unsupported group", storydomain.ErrInvalidReadQuery)
	}
	return nil
}

func validateUUIDFilterSlices(filters storydomain.StoryFilters) error {
	for _, filter := range []struct {
		name   string
		values []uuid.UUID
	}{
		{name: "status", values: filters.StatusIDs},
		{name: "excluded status", values: filters.ExcludedStatusIDs},
		{name: "assignee", values: filters.AssigneeIDs},
		{name: "excluded assignee", values: filters.ExcludedAssigneeIDs},
		{name: "collaborator", values: filters.CollaboratorIDs},
		{name: "reporter", values: filters.ReporterIDs},
		{name: "excluded reporter", values: filters.ExcludedReporterIDs},
		{name: "team", values: filters.TeamIDs},
		{name: "excluded team", values: filters.ExcludedTeamIDs},
		{name: "sprint", values: filters.SprintIDs},
		{name: "excluded sprint", values: filters.ExcludedSprintIDs},
		{name: "label", values: filters.LabelIDs},
		{name: "excluded label", values: filters.ExcludedLabelIDs},
	} {
		if len(filter.values) > maxStoryFilterValues {
			return fmt.Errorf("%w: %s filter exceeds %d values", storydomain.ErrInvalidReadQuery, filter.name, maxStoryFilterValues)
		}
		for _, value := range filter.values {
			if value == uuid.Nil {
				return fmt.Errorf("%w: %s filter contains a zero id", storydomain.ErrInvalidReadQuery, filter.name)
			}
		}
	}
	return nil
}

func validateStringFilter(name string, values []string, valid func(string) bool) error {
	if len(values) > maxStoryFilterValues {
		return fmt.Errorf("%w: %s filter exceeds %d values", storydomain.ErrInvalidReadQuery, name, maxStoryFilterValues)
	}
	for _, value := range values {
		if !valid(value) {
			return fmt.Errorf("%w: unsupported %s", storydomain.ErrInvalidReadQuery, name)
		}
	}
	return nil
}

func validateEstimateFilters(filters ...[]int16) error {
	for _, values := range filters {
		if len(values) > maxStoryFilterValues {
			return fmt.Errorf("%w: estimate filter exceeds %d values", storydomain.ErrInvalidReadQuery, maxStoryFilterValues)
		}
		for _, value := range values {
			switch value {
			case 1, 2, 3, 5, 8:
			default:
				return fmt.Errorf("%w: unsupported estimate value", storydomain.ErrInvalidReadQuery)
			}
		}
	}
	return nil
}

func validStoryPriority(value string) bool {
	switch value {
	case "Urgent", "High", "Medium", "Low", "No Priority":
		return true
	default:
		return false
	}
}

func validStoryCategory(value string) bool {
	_, err := parseStoryCategory(value)
	return err == nil
}

func validStoryGroup(value storydomain.StoryGroupBy) bool {
	switch value {
	case storydomain.StoryGroupNone, storydomain.StoryGroupStatus, storydomain.StoryGroupAssignee,
		storydomain.StoryGroupPriority, storydomain.StoryGroupTeam, storydomain.StoryGroupSprint:
		return true
	default:
		return false
	}
}

func validStoryOrder(value storydomain.StoryOrderBy) bool {
	switch value {
	case storydomain.StoryOrderCreated, storydomain.StoryOrderUpdated, storydomain.StoryOrderPriority,
		storydomain.StoryOrderDeadline, storydomain.StoryOrderCompleted:
		return true
	default:
		return false
	}
}

func enabled(value *bool) bool { return value != nil && *value }

func normalizedOptionalText(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func invalidRange(start, end *time.Time) bool {
	return start != nil && end != nil && start.After(*end)
}
