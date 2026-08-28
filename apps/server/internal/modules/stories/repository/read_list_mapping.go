package storiesrepository

import (
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

type storyListRecord struct {
	id                       uuid.UUID
	sequenceID               *int32
	title                    string
	priority                 string
	estimateUnit             *int16
	estimateScheme           string
	estimatedDurationMinutes *int32
	minimumFocusBlockMinutes *int32
	autoSchedulingEnabled    bool
	autoSchedulingLocked     bool
	autoSchedulingStatus     string
	autoSchedulingReason     *string
	autoSchedulingUpdatedAt  *time.Time
	parentID                 *uuid.UUID
	objectiveID              *uuid.UUID
	objectiveName            *string
	objectiveDescription     *string
	sprintID                 *uuid.UUID
	sprintName               *string
	sprintGoal               *string
	sprintStartDate          *time.Time
	sprintEndDate            *time.Time
	teamID                   uuid.UUID
	teamCode                 string
	teamName                 string
	workspaceID              uuid.UUID
	statusID                 *uuid.UUID
	assigneeID               *uuid.UUID
	collaboratorCount        int32
	reporterID               *uuid.UUID
	keyResultID              *uuid.UUID
	startDate                *time.Time
	endDate                  *time.Time
	createdAt                time.Time
	updatedAt                time.Time
	completedAt              *time.Time
	deletedAt                *time.Time
	archivedAt               *time.Time
	labelIDs                 []uuid.UUID
}

func mapMyVisibleStories(rows []storyreadsql.ListMyVisibleStoriesRow) ([]storydomain.StoryList, error) {
	result := make([]storydomain.StoryList, len(rows))
	for index, row := range rows {
		story, err := mapStoryListRecord(storyListRecordFromMyStory(row))
		if err != nil {
			return nil, err
		}
		result[index] = story
	}
	return result, nil
}

func mapVisibleCategoryStories(rows []storyreadsql.ListVisibleStoriesByCategoryRow) ([]storydomain.StoryList, error) {
	result := make([]storydomain.StoryList, len(rows))
	for index, row := range rows {
		story, err := mapStoryListRecord(storyListRecordFromCategory(row))
		if err != nil {
			return nil, err
		}
		result[index] = story
	}
	return result, nil
}

func mapVisibleSubStory(row storyreadsql.ListVisibleSubStoriesRow) (storydomain.StoryList, error) {
	return mapStoryListRecord(storyListRecordFromSubStory(row))
}

func mapStoryListRecord(record storyListRecord) (storydomain.StoryList, error) {
	sequenceID, err := requiredSequenceID(record.sequenceID, record.id)
	if err != nil {
		return storydomain.StoryList{}, err
	}
	return storydomain.StoryList{
		ID:                       record.id,
		SequenceID:               sequenceID,
		Title:                    record.title,
		Priority:                 record.priority,
		EstimateValue:            record.estimateUnit,
		EstimateScheme:           record.estimateScheme,
		EstimatedDurationMinutes: int32PointerToInt(record.estimatedDurationMinutes),
		MinimumFocusBlockMinutes: int32PointerToInt(record.minimumFocusBlockMinutes),
		AutoSchedulingEnabled:    record.autoSchedulingEnabled,
		AutoSchedulingLocked:     record.autoSchedulingLocked,
		AutoSchedulingStatus:     record.autoSchedulingStatus,
		AutoSchedulingReason:     record.autoSchedulingReason,
		AutoSchedulingUpdatedAt:  record.autoSchedulingUpdatedAt,
		Parent:                   record.parentID,
		Objective:                record.objectiveID,
		ObjectiveSummary:         objectiveSummary(record.objectiveID, record.objectiveName, record.objectiveDescription),
		Sprint:                   record.sprintID,
		SprintSummary:            sprintSummary(record.sprintID, record.sprintName, record.sprintGoal, record.sprintStartDate, record.sprintEndDate),
		Team:                     record.teamID,
		TeamSummary:              &storydomain.TeamSummary{ID: record.teamID, Code: record.teamCode, Name: record.teamName},
		Workspace:                record.workspaceID,
		Status:                   record.statusID,
		Assignee:                 record.assigneeID,
		CollaboratorCount:        int(record.collaboratorCount),
		Reporter:                 record.reporterID,
		KeyResult:                record.keyResultID,
		StartDate:                record.startDate,
		EndDate:                  record.endDate,
		CreatedAt:                record.createdAt,
		UpdatedAt:                record.updatedAt,
		CompletedAt:              record.completedAt,
		DeletedAt:                record.deletedAt,
		ArchivedAt:               record.archivedAt,
		Labels:                   append([]uuid.UUID(nil), record.labelIDs...),
		SubStories:               []storydomain.StoryList{},
	}, nil
}

func storyListRecordFromMyStory(row storyreadsql.ListMyVisibleStoriesRow) storyListRecord {
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

func storyListRecordFromCategory(row storyreadsql.ListVisibleStoriesByCategoryRow) storyListRecord {
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

func storyListRecordFromSubStory(row storyreadsql.ListVisibleSubStoriesRow) storyListRecord {
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
