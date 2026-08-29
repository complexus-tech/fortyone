package storieshttp

import (
	"fmt"
	"net/http"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func parseStoryQuery(r *http.Request) (StoryQuery, error) {
	query := StoryQuery{Filters: StoryFilters{}}
	var err error
	if query.GroupBy, err = parseStringParam(r, "groupBy", "status", 32); err != nil {
		return StoryQuery{}, err
	}
	if query.OrderBy, err = parseStringParam(r, "orderBy", "created", 32); err != nil {
		return StoryQuery{}, err
	}
	if query.OrderDirection, err = parseStringParam(r, "orderDirection", "desc", 8); err != nil {
		return StoryQuery{}, err
	}
	if groupKey, parseErr := parseOptionalStringParam(r, "groupKey", maxStoryScalarQueryBytes); parseErr != nil {
		return StoryQuery{}, parseErr
	} else if groupKey != nil {
		query.GroupKey = *groupKey
	}

	if !isValidGroupBy(query.GroupBy) {
		return StoryQuery{}, fmt.Errorf("invalid groupBy value: %s", query.GroupBy)
	}
	if !isValidOrderBy(query.OrderBy) {
		return StoryQuery{}, fmt.Errorf("invalid orderBy value: %s", query.OrderBy)
	}
	if !isValidOrderDirection(query.OrderDirection) {
		return StoryQuery{}, fmt.Errorf("invalid orderDirection value: %s", query.OrderDirection)
	}

	if query.StoriesPerGroup, err = parseStrictIntParam(r, "storiesPerGroup", 0, 0, maxStoryPageSize); err != nil {
		return StoryQuery{}, err
	}
	if query.Page, err = parseStrictIntParam(r, "page", 1, 1, maxStoryPage); err != nil {
		return StoryQuery{}, err
	}
	if query.PageSize, err = parseStrictIntParam(r, "pageSize", 0, 0, maxStoryPageSize); err != nil {
		return StoryQuery{}, err
	}

	for _, filter := range []struct {
		key    string
		values *[]uuid.UUID
	}{
		{key: "statusIds", values: &query.Filters.StatusIDs},
		{key: "excludedStatusIds", values: &query.Filters.ExcludedStatusIDs},
		{key: "assigneeIds", values: &query.Filters.AssigneeIDs},
		{key: "excludedAssigneeIds", values: &query.Filters.ExcludedAssigneeIDs},
		{key: "collaboratorIds", values: &query.Filters.CollaboratorIDs},
		{key: "reporterIds", values: &query.Filters.ReporterIDs},
		{key: "excludedReporterIds", values: &query.Filters.ExcludedReporterIDs},
		{key: "teamIds", values: &query.Filters.TeamIDs},
		{key: "excludedTeamIds", values: &query.Filters.ExcludedTeamIDs},
		{key: "sprintIds", values: &query.Filters.SprintIDs},
		{key: "excludedSprintIds", values: &query.Filters.ExcludedSprintIDs},
		{key: "labelIds", values: &query.Filters.LabelIDs},
		{key: "excludedLabelIds", values: &query.Filters.ExcludedLabelIDs},
	} {
		*filter.values, err = parseStrictUUIDArray(r, filter.key)
		if err != nil {
			return StoryQuery{}, err
		}
	}

	if query.Filters.TitleContains, err = parseOptionalStringParam(r, "titleContains", maxStoryScalarQueryBytes); err != nil {
		return StoryQuery{}, err
	}
	if query.Filters.TitleNotContains, err = parseOptionalStringParam(r, "titleNotContains", maxStoryScalarQueryBytes); err != nil {
		return StoryQuery{}, err
	}
	if query.Filters.EstimateValues, err = parseStrictInt16Array(r, "estimateValues"); err != nil {
		return StoryQuery{}, err
	}
	if query.Filters.ExcludedEstimateValues, err = parseStrictInt16Array(r, "excludedEstimateValues"); err != nil {
		return StoryQuery{}, err
	}
	if query.Filters.Priorities, err = parseStrictStringArray(r, "priorities"); err != nil {
		return StoryQuery{}, err
	}
	if query.Filters.ExcludedPriorities, err = parseStrictStringArray(r, "excludedPriorities"); err != nil {
		return StoryQuery{}, err
	}
	if query.Filters.Categories, err = parseStrictStringArray(r, "categories"); err != nil {
		return StoryQuery{}, err
	}
	for _, priority := range append(append([]string(nil), query.Filters.Priorities...), query.Filters.ExcludedPriorities...) {
		if !isValidPriority(priority) {
			return StoryQuery{}, fmt.Errorf("priorities must contain supported values")
		}
	}
	for _, category := range query.Filters.Categories {
		if !isValidStoryCategory(category) {
			return StoryQuery{}, fmt.Errorf("categories must contain supported values")
		}
	}

	for _, filter := range []struct {
		key   string
		value **uuid.UUID
	}{
		{key: "parentId", value: &query.Filters.Parent},
		{key: "objectiveId", value: &query.Filters.Objective},
		{key: "excludedObjectiveId", value: &query.Filters.ExcludedObjective},
		{key: "epicId", value: &query.Filters.Epic},
		{key: "keyResultId", value: &query.Filters.KeyResult},
	} {
		*filter.value, err = parseStrictOptionalUUID(r, filter.key)
		if err != nil {
			return StoryQuery{}, err
		}
	}

	for _, filter := range []struct {
		key   string
		value **bool
	}{
		{key: "hasNoAssignee", value: &query.Filters.HasNoAssignee},
		{key: "hasAssignee", value: &query.Filters.HasAssignee},
		{key: "hasBlockedBy", value: &query.Filters.HasBlockedBy},
		{key: "assignedToMe", value: &query.Filters.AssignedToMe},
		{key: "collaboratingWithMe", value: &query.Filters.CollaboratingWithMe},
		{key: "createdByMe", value: &query.Filters.CreatedByMe},
		{key: "showSubStories", value: &query.Filters.ShowSubStories},
		{key: "includeArchived", value: &query.Filters.IncludeArchived},
		{key: "includeDeleted", value: &query.Filters.IncludeDeleted},
	} {
		*filter.value, err = parseStrictOptionalBool(r, filter.key)
		if err != nil {
			return StoryQuery{}, err
		}
	}

	for _, filter := range storyDateFilters(&query.Filters) {
		*filter.value, err = parseStrictOptionalDate(r, filter.key)
		if err != nil {
			return StoryQuery{}, err
		}
	}

	return query, nil
}

func toCoreStoryQuery(query StoryQuery) stories.CoreStoryQuery {
	return stories.CoreStoryQuery{
		Filters: stories.CoreStoryFilters{
			StatusIDs:              query.Filters.StatusIDs,
			ExcludedStatusIDs:      query.Filters.ExcludedStatusIDs,
			AssigneeIDs:            query.Filters.AssigneeIDs,
			ExcludedAssigneeIDs:    query.Filters.ExcludedAssigneeIDs,
			CollaboratorIDs:        query.Filters.CollaboratorIDs,
			ReporterIDs:            query.Filters.ReporterIDs,
			ExcludedReporterIDs:    query.Filters.ExcludedReporterIDs,
			TitleContains:          query.Filters.TitleContains,
			TitleNotContains:       query.Filters.TitleNotContains,
			Priorities:             query.Filters.Priorities,
			ExcludedPriorities:     query.Filters.ExcludedPriorities,
			Categories:             query.Filters.Categories,
			TeamIDs:                query.Filters.TeamIDs,
			ExcludedTeamIDs:        query.Filters.ExcludedTeamIDs,
			SprintIDs:              query.Filters.SprintIDs,
			ExcludedSprintIDs:      query.Filters.ExcludedSprintIDs,
			LabelIDs:               query.Filters.LabelIDs,
			ExcludedLabelIDs:       query.Filters.ExcludedLabelIDs,
			EstimateValues:         query.Filters.EstimateValues,
			ExcludedEstimateValues: query.Filters.ExcludedEstimateValues,
			Parent:                 query.Filters.Parent,
			Objective:              query.Filters.Objective,
			ExcludedObjective:      query.Filters.ExcludedObjective,
			Epic:                   query.Filters.Epic,
			KeyResult:              query.Filters.KeyResult,
			HasNoAssignee:          query.Filters.HasNoAssignee,
			HasAssignee:            query.Filters.HasAssignee,
			HasBlockedBy:           query.Filters.HasBlockedBy,
			AssignedToMe:           query.Filters.AssignedToMe,
			CollaboratingWithMe:    query.Filters.CollaboratingWithMe,
			CreatedByMe:            query.Filters.CreatedByMe,
			ShowSubStories:         query.Filters.ShowSubStories,
			IncludeArchived:        query.Filters.IncludeArchived,
			IncludeDeleted:         query.Filters.IncludeDeleted,
			CreatedAfter:           query.Filters.CreatedAfter,
			CreatedBefore:          query.Filters.CreatedBefore,
			UpdatedAfter:           query.Filters.UpdatedAfter,
			UpdatedBefore:          query.Filters.UpdatedBefore,
			StartDateAfter:         query.Filters.StartDateAfter,
			StartDateBefore:        query.Filters.StartDateBefore,
			StartDateNot:           query.Filters.StartDateNot,
			DeadlineAfter:          query.Filters.DeadlineAfter,
			DeadlineBefore:         query.Filters.DeadlineBefore,
			DeadlineNot:            query.Filters.DeadlineNot,
			CompletedAfter:         query.Filters.CompletedAfter,
			CompletedBefore:        query.Filters.CompletedBefore,
		},
		GroupBy:         stories.StoryGroupBy(query.GroupBy),
		OrderBy:         stories.StoryOrderBy(query.OrderBy),
		OrderDirection:  stories.SortDirection(query.OrderDirection),
		StoriesPerGroup: query.StoriesPerGroup,
		GroupKey:        query.GroupKey,
		Page:            query.Page,
		PageSize:        query.PageSize,
	}
}
