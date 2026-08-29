package storiesrepository

import (
	"fmt"
	"math"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

func mutationSnapshotToStory(row storyreadsql.GetStoryMutationSnapshotRow) storydomain.Story {
	return storydomain.Story{
		ID:                       row.ID,
		SequenceID:               intValue(row.SequenceID),
		Title:                    row.Title,
		EstimateValue:            row.EstimateUnit,
		EstimateScheme:           row.EstimateScheme,
		EstimatedDurationMinutes: intPointer(row.EstimatedDurationMinutes),
		MinimumFocusBlockMinutes: intPointer(row.MinimumFocusBlockMinutes),
		AutoSchedulingEnabled:    row.AutoSchedulingEnabled,
		AutoSchedulingLocked:     row.AutoSchedulingLocked,
		AutoSchedulingStatus:     row.AutoSchedulingStatus,
		AutoSchedulingReason:     row.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  row.AutoSchedulingUpdatedAt,
		Description:              row.Description,
		DescriptionHTML:          row.DescriptionHtml,
		Parent:                   row.ParentID,
		Objective:                row.ObjectiveID,
		Status:                   row.StatusID,
		Assignee:                 row.AssigneeID,
		Reporter:                 row.ReporterID,
		Priority:                 row.Priority,
		Sprint:                   row.SprintID,
		KeyResult:                row.KeyResultID,
		Team:                     row.TeamID,
		Workspace:                row.WorkspaceID,
		StartDate:                row.StartDate,
		EndDate:                  row.EndDate,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		DeletedAt:                row.DeletedAt,
		ArchivedAt:               row.ArchivedAt,
		CompletedAt:              row.CompletedAt,
		CreationKey:              row.ExternalCreationKey,
	}
}

func oauthApplicationCreationReplayToStory(
	row storyreadsql.GetOAuthApplicationStoryCreationReplayRow,
) storydomain.Story {
	return storydomain.Story{
		ID:                       row.ID,
		SequenceID:               intValue(row.SequenceID),
		Title:                    row.Title,
		EstimateValue:            row.EstimateUnit,
		EstimateScheme:           row.EstimateScheme,
		EstimatedDurationMinutes: intPointer(row.EstimatedDurationMinutes),
		MinimumFocusBlockMinutes: intPointer(row.MinimumFocusBlockMinutes),
		AutoSchedulingEnabled:    row.AutoSchedulingEnabled,
		AutoSchedulingLocked:     row.AutoSchedulingLocked,
		AutoSchedulingStatus:     row.AutoSchedulingStatus,
		AutoSchedulingReason:     row.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  row.AutoSchedulingUpdatedAt,
		Description:              row.Description,
		DescriptionHTML:          row.DescriptionHtml,
		Parent:                   row.ParentID,
		Objective:                row.ObjectiveID,
		Status:                   row.StatusID,
		Assignee:                 row.AssigneeID,
		Reporter:                 row.ReporterID,
		Priority:                 row.Priority,
		Sprint:                   row.SprintID,
		KeyResult:                row.KeyResultID,
		Team:                     row.TeamID,
		Workspace:                row.WorkspaceID,
		StartDate:                row.StartDate,
		EndDate:                  row.EndDate,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		DeletedAt:                row.DeletedAt,
		ArchivedAt:               row.ArchivedAt,
		CompletedAt:              row.CompletedAt,
		CreationKey:              row.ExternalCreationKey,
	}
}

func createdMutationToStory(row storyreadsql.CreateStoryMutationRow, estimateScheme string, labels []uuid.UUID) storydomain.Story {
	return storydomain.Story{
		ID:                       row.ID,
		SequenceID:               intValue(row.SequenceID),
		Title:                    row.Title,
		EstimateValue:            row.EstimateUnit,
		EstimateScheme:           estimateScheme,
		EstimatedDurationMinutes: intPointer(row.EstimatedDurationMinutes),
		MinimumFocusBlockMinutes: intPointer(row.MinimumFocusBlockMinutes),
		AutoSchedulingEnabled:    row.AutoSchedulingEnabled,
		AutoSchedulingLocked:     row.AutoSchedulingLocked,
		AutoSchedulingStatus:     row.AutoSchedulingStatus,
		AutoSchedulingReason:     row.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  row.AutoSchedulingUpdatedAt,
		Description:              row.Description,
		DescriptionHTML:          row.DescriptionHtml,
		Parent:                   row.ParentID,
		Objective:                row.ObjectiveID,
		Status:                   row.StatusID,
		Assignee:                 row.AssigneeID,
		BlockedBy:                row.BlockedByID,
		Blocking:                 row.BlockingID,
		Related:                  row.RelatedID,
		Reporter:                 row.ReporterID,
		Priority:                 row.Priority,
		Sprint:                   row.SprintID,
		KeyResult:                row.KeyResultID,
		Team:                     row.TeamID,
		Workspace:                row.WorkspaceID,
		StartDate:                row.StartDate,
		EndDate:                  row.EndDate,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		Labels:                   append([]uuid.UUID(nil), labels...),
		CreationKey:              row.ExternalCreationKey,
		CreatedNow:               true,
	}
}

func storyCreateParams(story storydomain.Story, sequence int32) (storyreadsql.CreateStoryMutationParams, error) {
	estimatedDuration, err := int32Pointer(story.EstimatedDurationMinutes)
	if err != nil {
		return storyreadsql.CreateStoryMutationParams{}, fmt.Errorf("map estimated duration: %w", err)
	}
	minimumFocusBlock, err := int32Pointer(story.MinimumFocusBlockMinutes)
	if err != nil {
		return storyreadsql.CreateStoryMutationParams{}, fmt.Errorf("map minimum focus block: %w", err)
	}
	priority := story.Priority
	return storyreadsql.CreateStoryMutationParams{
		StoryID: story.ID, SequenceID: &sequence, Title: story.Title,
		Description: story.Description, DescriptionHtml: story.DescriptionHTML,
		ParentID: story.Parent, ObjectiveID: story.Objective, StatusID: story.Status,
		AssigneeID: story.Assignee, BlockedByID: story.BlockedBy, BlockingID: story.Blocking,
		RelatedID: story.Related, ReporterID: story.Reporter, Priority: &priority,
		EstimateUnit: story.EstimateValue, EstimatedDurationMinutes: estimatedDuration,
		MinimumFocusBlockMinutes: minimumFocusBlock, AutoSchedulingEnabled: story.AutoSchedulingEnabled,
		AutoSchedulingLocked: story.AutoSchedulingLocked, AutoSchedulingStatus: story.AutoSchedulingStatus,
		AutoSchedulingReason: story.AutoSchedulingReason, AutoSchedulingUpdatedAt: story.AutoSchedulingUpdatedAt,
		SprintID: story.Sprint, KeyResultID: story.KeyResult, TeamID: story.Team,
		WorkspaceID: story.Workspace, StartDate: story.StartDate, EndDate: story.EndDate,
		ExternalCreationKey: story.CreationKey, CreatedAt: story.CreatedAt.UTC(), UpdatedAt: story.UpdatedAt.UTC(),
	}, nil
}

func storyPatchParams(
	command storydomain.UpdateStoryCommand,
	now time.Time,
) (storyreadsql.ApplyStoryPatchParams, error) {
	patch := command.Patch
	params := storyreadsql.ApplyStoryPatchParams{
		StoryID: command.StoryID, WorkspaceID: command.Scope.WorkspaceID,
		ActorKind: string(command.Scope.Actor.Kind), ActorID: command.Scope.Actor.PrincipalID,
		ActorCredentialID: command.Scope.Actor.CredentialID, Now: now.UTC(),
		UpdatedAt: now.UTC(), ExpectedUpdatedAt: command.ExpectedUpdatedAt.UTC(),
	}
	params.SetTitle, params.Title = requiredStringField(patch.Title)
	params.SetEstimateUnit, params.EstimateUnit = nullableField(patch.EstimateValue)
	params.SetAutoSchedulingEnabled, params.AutoSchedulingEnabled = requiredBoolField(patch.AutoSchedulingEnabled)
	params.SetAutoSchedulingLocked, params.AutoSchedulingLocked = requiredBoolField(patch.AutoSchedulingLocked)
	params.SetAutoSchedulingStatus, params.AutoSchedulingStatus = requiredStringField(patch.AutoSchedulingStatus)
	params.SetAutoSchedulingReason, params.AutoSchedulingReason = nullableField(patch.AutoSchedulingReason)
	params.SetAutoSchedulingUpdatedAt, params.AutoSchedulingUpdatedAt = nullableField(patch.AutoSchedulingUpdatedAt)
	params.SetDescription, params.Description = nullableField(patch.Description)
	params.SetDescriptionHtml, params.DescriptionHtml = nullableField(patch.DescriptionHTML)
	params.SetParentID, params.ParentID = nullableField(patch.ParentID)
	params.SetObjectiveID, params.ObjectiveID = nullableField(patch.ObjectiveID)
	params.SetStatusID, params.StatusID = nullableField(patch.StatusID)
	params.SetAssigneeID, params.AssigneeID = nullableField(patch.AssigneeID)
	params.SetPriority, params.Priority = requiredStringField(patch.Priority)
	params.SetSprintID, params.SprintID = nullableField(patch.SprintID)
	params.SetKeyResultID, params.KeyResultID = nullableField(patch.KeyResultID)
	params.SetStartDate, params.StartDate = nullableField(patch.StartDate)
	params.SetEndDate, params.EndDate = nullableField(patch.EndDate)
	params.SetCompletedAt, params.CompletedAt = nullableField(patch.CompletedAt)

	var err error
	params.SetEstimatedDurationMinutes, params.EstimatedDurationMinutes, err = nullableInt32Field(patch.EstimatedDurationMinutes)
	if err != nil {
		return storyreadsql.ApplyStoryPatchParams{}, fmt.Errorf("map estimated duration: %w", err)
	}
	params.SetMinimumFocusBlockMinutes, params.MinimumFocusBlockMinutes, err = nullableInt32Field(patch.MinimumFocusBlockMinutes)
	if err != nil {
		return storyreadsql.ApplyStoryPatchParams{}, fmt.Errorf("map minimum focus block: %w", err)
	}
	return params, nil
}

func requiredStringField(field storydomain.Field[string]) (bool, string) {
	value, specified := field.Value()
	if !specified || value == nil {
		return specified, ""
	}
	return true, *value
}

func requiredBoolField(field storydomain.Field[bool]) (bool, bool) {
	value, specified := field.Value()
	if !specified || value == nil {
		return specified, false
	}
	return true, *value
}

func nullableField[T any](field storydomain.Field[T]) (bool, *T) {
	value, specified := field.Value()
	return specified, value
}

func nullableInt32Field(field storydomain.Field[int]) (bool, *int32, error) {
	value, specified := field.Value()
	if !specified || value == nil {
		return specified, nil, nil
	}
	converted, err := int32Value(*value)
	if err != nil {
		return false, nil, err
	}
	return true, &converted, nil
}

func int32Pointer(value *int) (*int32, error) {
	if value == nil {
		return nil, nil
	}
	converted, err := int32Value(*value)
	if err != nil {
		return nil, err
	}
	return &converted, nil
}

func int32Value(value int) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("value %d is outside PostgreSQL integer range", value)
	}
	return int32(value), nil
}

func intPointer(value *int32) *int {
	if value == nil {
		return nil
	}
	converted := int(*value)
	return &converted
}

func intValue(value *int32) int {
	if value == nil {
		return 0
	}
	return int(*value)
}
