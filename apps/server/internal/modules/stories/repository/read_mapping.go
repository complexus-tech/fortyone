package storiesrepository

import (
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

func mapVisibleStory(row storyreadsql.GetVisibleStoryRow) (storydomain.Story, error) {
	sequenceID, err := requiredSequenceID(row.SequenceID, row.ID)
	if err != nil {
		return storydomain.Story{}, err
	}

	return storydomain.Story{
		ID:                       row.ID,
		SequenceID:               sequenceID,
		Title:                    row.Title,
		TeamCode:                 row.TeamCode,
		Description:              row.Description,
		DescriptionHTML:          row.DescriptionHtml,
		Parent:                   row.ParentID,
		Objective:                row.ObjectiveID,
		Workspace:                row.WorkspaceID,
		Team:                     row.TeamID,
		Status:                   row.StatusID,
		Assignee:                 row.AssigneeID,
		BlockedBy:                row.BlockedByID,
		Blocking:                 row.BlockingID,
		Related:                  row.RelatedID,
		Reporter:                 row.ReporterID,
		Priority:                 row.Priority,
		Sprint:                   row.SprintID,
		SprintSummary:            sprintSummary(row.SprintID, row.SprintName, row.SprintGoal, row.SprintStartDate, row.SprintEndDate),
		KeyResult:                row.KeyResultID,
		EstimateValue:            row.EstimateUnit,
		EstimateScheme:           row.EstimateScheme,
		EstimatedDurationMinutes: int32PointerToInt(row.EstimatedDurationMinutes),
		MinimumFocusBlockMinutes: int32PointerToInt(row.MinimumFocusBlockMinutes),
		AutoSchedulingEnabled:    row.AutoSchedulingEnabled,
		AutoSchedulingLocked:     row.AutoSchedulingLocked,
		AutoSchedulingStatus:     row.AutoSchedulingStatus,
		AutoSchedulingReason:     row.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  row.AutoSchedulingUpdatedAt,
		StartDate:                row.StartDate,
		EndDate:                  row.EndDate,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		DeletedAt:                row.DeletedAt,
		ArchivedAt:               row.ArchivedAt,
		CompletedAt:              row.CompletedAt,
		CreationKey:              row.ExternalCreationKey,
		Labels:                   append([]uuid.UUID(nil), row.LabelIds...),
		Collaborators:            append([]uuid.UUID(nil), row.CollaboratorIds...),
		WatcherIDs:               append([]uuid.UUID(nil), row.WatcherIds...),
		WatcherCount:             int(row.WatcherCount),
		IsWatching:               boolValue(row.IsWatching),
		WatchingReason:           nonEmptyStringPointer(row.WatchingReason),
		SubStories:               []storydomain.StoryList{},
		Associations:             []storydomain.StoryAssociation{},
	}, nil
}

func requiredSequenceID(value *int32, storyID uuid.UUID) (int, error) {
	if value == nil || *value < 1 {
		return 0, fmt.Errorf("map story %s: valid sequence id is required", storyID)
	}
	return int(*value), nil
}

func int32PointerToInt(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func nonEmptyStringPointer(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func sprintSummary(
	id *uuid.UUID,
	name *string,
	goal *string,
	startDate *time.Time,
	endDate *time.Time,
) *storydomain.SprintSummary {
	if id == nil || name == nil || startDate == nil || endDate == nil {
		return nil
	}
	return &storydomain.SprintSummary{
		ID:        *id,
		Name:      *name,
		Goal:      goal,
		StartDate: *startDate,
		EndDate:   *endDate,
	}
}

func objectiveSummary(id *uuid.UUID, name *string, description *string) *storydomain.ObjectiveSummary {
	if id == nil || name == nil {
		return nil
	}
	return &storydomain.ObjectiveSummary{ID: *id, Name: *name, Description: description}
}
