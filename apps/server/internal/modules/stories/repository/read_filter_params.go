package storiesrepository

import (
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

type filteredReadOptions struct {
	groupBy    storydomain.StoryGroupBy
	orderBy    storydomain.StoryOrderBy
	direction  storydomain.SortDirection
	applyGroup bool
	groupKey   string
	mode       string
	limit      int32
	offset     int32
}

func filteredStoryParams(
	scope storydomain.ReadScope,
	filters storydomain.StoryFilters,
	options filteredReadOptions,
) storyreadsql.ListVisibleFilteredStoryRowsParams {
	parentID, hasParent := optionalUUID(filters.Parent)
	objectiveID, hasObjective := optionalUUID(filters.Objective)
	excludedObjectiveID, hasExcludedObjective := optionalUUID(filters.ExcludedObjective)
	keyResultID, hasKeyResult := optionalUUID(filters.KeyResult)
	titleContains, hasTitleContains := optionalString(filters.TitleContains)
	titleNotContains, hasTitleNotContains := optionalString(filters.TitleNotContains)
	createdAfter, hasCreatedAfter := optionalTime(filters.CreatedAfter)
	createdBefore, hasCreatedBefore := optionalTime(filters.CreatedBefore)
	updatedAfter, hasUpdatedAfter := optionalTime(filters.UpdatedAfter)
	updatedBefore, hasUpdatedBefore := optionalTime(filters.UpdatedBefore)
	startDateAfter, hasStartDateAfter := optionalTime(filters.StartDateAfter)
	startDateBefore, hasStartDateBefore := optionalTime(filters.StartDateBefore)
	startDateNot, hasStartDateNot := optionalTime(filters.StartDateNot)
	deadlineAfter, hasDeadlineAfter := optionalTime(filters.DeadlineAfter)
	deadlineBefore, hasDeadlineBefore := optionalTime(filters.DeadlineBefore)
	deadlineNot, hasDeadlineNot := optionalTime(filters.DeadlineNot)
	completedAfter, hasCompletedAfter := optionalTime(filters.CompletedAfter)
	completedBefore, hasCompletedBefore := optionalTime(filters.CompletedBefore)

	return storyreadsql.ListVisibleFilteredStoryRowsParams{
		ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
		StatusIds:              cloneUUIDs(filters.StatusIDs),
		ExcludedStatusIds:      cloneUUIDs(filters.ExcludedStatusIDs),
		AssigneeIds:            cloneUUIDs(filters.AssigneeIDs),
		ExcludedAssigneeIds:    cloneUUIDs(filters.ExcludedAssigneeIDs),
		CollaboratorIds:        cloneUUIDs(filters.CollaboratorIDs),
		ReporterIds:            cloneUUIDs(filters.ReporterIDs),
		ExcludedReporterIds:    cloneUUIDs(filters.ExcludedReporterIDs),
		HasTitleContains:       hasTitleContains,
		TitleContains:          titleContains,
		HasTitleNotContains:    hasTitleNotContains,
		TitleNotContains:       titleNotContains,
		Priorities:             cloneStrings(filters.Priorities),
		ExcludedPriorities:     cloneStrings(filters.ExcludedPriorities),
		Categories:             cloneStrings(filters.Categories),
		TeamIds:                cloneUUIDs(filters.TeamIDs),
		ExcludedTeamIds:        cloneUUIDs(filters.ExcludedTeamIDs),
		SprintIds:              cloneUUIDs(filters.SprintIDs),
		ExcludedSprintIds:      cloneUUIDs(filters.ExcludedSprintIDs),
		LabelIds:               cloneUUIDs(filters.LabelIDs),
		ExcludedLabelIds:       cloneUUIDs(filters.ExcludedLabelIDs),
		EstimateValues:         cloneInt16s(filters.EstimateValues),
		ExcludedEstimateValues: cloneInt16s(filters.ExcludedEstimateValues),
		HasParent:              hasParent,
		ParentID:               parentID,
		HasObjective:           hasObjective,
		ObjectiveID:            objectiveID,
		HasExcludedObjective:   hasExcludedObjective,
		ExcludedObjectiveID:    excludedObjectiveID,
		HasKeyResult:           hasKeyResult,
		KeyResultID:            keyResultID,
		HasNoAssignee:          enabled(filters.HasNoAssignee),
		HasAssignee:            enabled(filters.HasAssignee),
		HasBlockedBy:           enabled(filters.HasBlockedBy),
		AssignedToMe:           enabled(filters.AssignedToMe),
		CollaboratingWithMe:    enabled(filters.CollaboratingWithMe),
		CreatedByMe:            enabled(filters.CreatedByMe),
		ShowSubStories:         enabled(filters.ShowSubStories),
		HasCreatedAfter:        hasCreatedAfter,
		CreatedAfter:           createdAfter,
		HasCreatedBefore:       hasCreatedBefore,
		CreatedBefore:          createdBefore,
		HasUpdatedAfter:        hasUpdatedAfter,
		UpdatedAfter:           updatedAfter,
		HasUpdatedBefore:       hasUpdatedBefore,
		UpdatedBefore:          updatedBefore,
		HasStartDateAfter:      hasStartDateAfter,
		StartDateAfter:         startDateAfter,
		HasStartDateBefore:     hasStartDateBefore,
		StartDateBefore:        startDateBefore,
		HasStartDateNot:        hasStartDateNot,
		StartDateNot:           startDateNot,
		HasDeadlineAfter:       hasDeadlineAfter,
		DeadlineAfter:          deadlineAfter,
		HasDeadlineBefore:      hasDeadlineBefore,
		DeadlineBefore:         deadlineBefore,
		HasDeadlineNot:         hasDeadlineNot,
		DeadlineNot:            deadlineNot,
		HasCompletedAfter:      hasCompletedAfter,
		CompletedAfter:         completedAfter,
		HasCompletedBefore:     hasCompletedBefore,
		CompletedBefore:        completedBefore,
		IsCompleted:            enabled(filters.IsCompleted),
		IsNotCompleted:         enabled(filters.IsNotCompleted),
		IncludeArchived:        enabled(filters.IncludeArchived),
		IncludeDeleted:         enabled(filters.IncludeDeleted),
		GroupBy:                string(options.groupBy),
		OrderBy:                string(options.orderBy),
		OrderDirection:         string(options.direction),
		ApplyGroupFilter:       options.applyGroup,
		GroupKey:               options.groupKey,
		ReadMode:               options.mode,
		ResultLimit:            options.limit,
		ResultOffset:           options.offset,
	}
}

func groupCatalogParams(scope storydomain.ReadScope, query storydomain.StoryQuery) storyreadsql.ListVisibleStoryGroupCatalogParams {
	return storyreadsql.ListVisibleStoryGroupCatalogParams{
		CatalogLimit:           int32(maxStoryGroupCatalog + 1),
		ActorID:                scope.ActorID,
		WorkspaceID:            scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
		TeamIds:                cloneUUIDs(query.Filters.TeamIDs),
		ExcludedTeamIds:        cloneUUIDs(query.Filters.ExcludedTeamIDs),
		AssigneeIds:            cloneUUIDs(query.Filters.AssigneeIDs),
		GroupBy:                string(query.GroupBy),
	}
}

func optionalUUID(value *uuid.UUID) (uuid.UUID, bool) {
	if value == nil {
		return uuid.Nil, false
	}
	return *value, true
}

func optionalString(value *string) (string, bool) {
	if value == nil {
		return "", false
	}
	return *value, true
}

func optionalTime(value *time.Time) (time.Time, bool) {
	if value == nil {
		return time.Time{}, false
	}
	return *value, true
}

func cloneUUIDs(values []uuid.UUID) []uuid.UUID {
	if len(values) == 0 {
		return []uuid.UUID{}
	}
	return append([]uuid.UUID(nil), values...)
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func cloneInt16s(values []int16) []int16 {
	if len(values) == 0 {
		return []int16{}
	}
	return append([]int16(nil), values...)
}
