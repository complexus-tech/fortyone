package storiesrepository

import (
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

func mapVisibleAssociations(rows []storyreadsql.ListVisibleStoryAssociationsRow) ([]storydomain.StoryAssociation, error) {
	result := make([]storydomain.StoryAssociation, len(rows))
	for index, row := range rows {
		sequenceID, err := requiredSequenceID(row.RelatedSequenceID, row.RelatedID)
		if err != nil {
			return nil, err
		}
		related := storydomain.StoryList{
			ID:                       row.RelatedID,
			SequenceID:               sequenceID,
			Title:                    row.RelatedTitle,
			Priority:                 row.RelatedPriority,
			EstimateValue:            row.RelatedEstimateUnit,
			EstimateScheme:           row.RelatedEstimateScheme,
			EstimatedDurationMinutes: int32PointerToInt(row.RelatedEstimatedDurationMinutes),
			MinimumFocusBlockMinutes: int32PointerToInt(row.RelatedMinimumFocusBlockMinutes),
			AutoSchedulingEnabled:    row.RelatedAutoSchedulingEnabled,
			AutoSchedulingLocked:     row.RelatedAutoSchedulingLocked,
			AutoSchedulingStatus:     row.RelatedAutoSchedulingStatus,
			AutoSchedulingReason:     row.RelatedAutoSchedulingReason,
			AutoSchedulingUpdatedAt:  row.RelatedAutoSchedulingUpdatedAt,
			Parent:                   row.RelatedParentID,
			Objective:                row.RelatedObjectiveID,
			Sprint:                   row.RelatedSprintID,
			Team:                     row.RelatedTeamID,
			Workspace:                row.RelatedWorkspaceID,
			Status:                   row.RelatedStatusID,
			Assignee:                 row.RelatedAssigneeID,
			CollaboratorCount:        int(row.RelatedCollaboratorCount),
			Reporter:                 row.RelatedReporterID,
			KeyResult:                row.RelatedKeyResultID,
			StartDate:                row.RelatedStartDate,
			EndDate:                  row.RelatedEndDate,
			CreatedAt:                row.RelatedCreatedAt,
			UpdatedAt:                row.RelatedUpdatedAt,
			CompletedAt:              row.RelatedCompletedAt,
			DeletedAt:                row.RelatedDeletedAt,
			ArchivedAt:               row.RelatedArchivedAt,
			Labels:                   append([]uuid.UUID(nil), row.RelatedLabelIds...),
			SubStories:               []storydomain.StoryList{},
		}
		result[index] = storydomain.StoryAssociation{
			ID:          row.ID,
			FromStoryID: row.FromStoryID,
			ToStoryID:   row.ToStoryID,
			Type:        row.AssociationType,
			Story:       related,
		}
	}
	return result, nil
}
