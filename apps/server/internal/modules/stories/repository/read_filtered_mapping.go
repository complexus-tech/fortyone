package storiesrepository

import (
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
)

type filteredStoryRow struct {
	story      storydomain.StoryList
	groupKey   string
	totalCount int
}

func mapFilteredStoryRows(rows []storyreadsql.ListVisibleFilteredStoryRowsRow) ([]filteredStoryRow, error) {
	result := make([]filteredStoryRow, len(rows))
	for index, row := range rows {
		story, err := mapStoryListRecord(storyListRecordFromFilteredRow(row))
		if err != nil {
			return nil, err
		}
		result[index] = filteredStoryRow{
			story:      story,
			groupKey:   row.GroupKey,
			totalCount: int(row.TotalCount),
		}
	}
	return result, nil
}

func storyListRecordFromFilteredRow(row storyreadsql.ListVisibleFilteredStoryRowsRow) storyListRecord {
	return storyListRecord{
		id: row.ID, sequenceID: row.SequenceID, title: row.Title, priority: row.Priority,
		estimateUnit: row.EstimateUnit, estimateScheme: row.EstimateScheme,
		estimatedDurationMinutes: row.EstimatedDurationMinutes, minimumFocusBlockMinutes: row.MinimumFocusBlockMinutes,
		autoSchedulingEnabled: row.AutoSchedulingEnabled, autoSchedulingLocked: row.AutoSchedulingLocked,
		autoSchedulingStatus: row.AutoSchedulingStatus, autoSchedulingReason: row.AutoSchedulingReason,
		autoSchedulingUpdatedAt: row.AutoSchedulingUpdatedAt, parentID: row.ParentID,
		objectiveID: row.ObjectiveID, objectiveName: row.ObjectiveName, objectiveDescription: row.ObjectiveDescription,
		sprintID: row.SprintID, sprintName: row.SprintName, sprintGoal: row.SprintGoal,
		sprintStartDate: row.SprintStartDate, sprintEndDate: row.SprintEndDate,
		teamID: row.TeamID, teamCode: row.TeamCode, teamName: row.TeamName, workspaceID: row.WorkspaceID,
		statusID: row.StatusID, assigneeID: row.AssigneeID, collaboratorCount: row.CollaboratorCount,
		reporterID: row.ReporterID, keyResultID: row.KeyResultID, startDate: row.StartDate, endDate: row.EndDate,
		createdAt: row.CreatedAt, updatedAt: row.UpdatedAt, completedAt: row.CompletedAt,
		deletedAt: row.DeletedAt, archivedAt: row.ArchivedAt, labelIDs: row.LabelIds,
	}
}
