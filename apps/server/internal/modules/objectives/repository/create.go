package objectivesrepository

import (
	"context"
	"fmt"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

func createObjectiveParams(command objectivesdomain.CreateCommand, sequenceID int32) objectivessql.CreateObjectiveParams {
	objective := command.Objective
	return objectivessql.CreateObjectiveParams{
		SequenceID: sequenceID, Name: objective.Name, Description: objective.Description,
		ShortSummary: objective.ShortSummary, LeadUserID: objective.LeadUser,
		TeamID: uuidPointer(objective.Team), WorkspaceID: uuidPointer(command.WorkspaceID),
		StartDate: utcPointer(objective.StartDate), EndDate: utcPointer(objective.EndDate),
		IsPrivate: objective.IsPrivate, StatusID: uuidPointer(objective.Status),
		Priority: objective.Priority, Color: objective.Color, ActorID: uuidPointer(objective.CreatedBy),
	}
}

func (repository *Repository) createKeyResults(
	ctx context.Context,
	queries objectivessql.Querier,
	command objectivesdomain.CreateCommand,
	objectiveID uuid.UUID,
) ([]objectivesdomain.KeyResult, error) {
	if len(command.KeyResults) == 0 {
		return []objectivesdomain.KeyResult{}, nil
	}
	sequenceCount, _, err := keyResultSequenceRange(0, len(command.KeyResults))
	if err != nil {
		return nil, fmt.Errorf("validate key result sequence range: %w", err)
	}
	finalSequence, err := queries.AllocateKeyResultSequences(ctx, objectivessql.AllocateKeyResultSequencesParams{
		WorkspaceID: command.WorkspaceID, TeamID: command.Objective.Team,
		SequenceCount: sequenceCount,
	})
	if err != nil {
		return nil, fmt.Errorf("allocate key result sequences: %w", err)
	}
	_, firstSequence, err := keyResultSequenceRange(finalSequence, len(command.KeyResults))
	if err != nil {
		return nil, fmt.Errorf("calculate key result sequence range: %w", err)
	}
	created := make([]objectivesdomain.KeyResult, 0, len(command.KeyResults))
	for index, keyResult := range command.KeyResults {
		sequenceID, err := keyResultSequenceAt(firstSequence, index)
		if err != nil {
			return nil, fmt.Errorf("calculate key result %d sequence: %w", index, err)
		}
		row, err := queries.CreateObjectiveKeyResult(ctx, objectivessql.CreateObjectiveKeyResultParams{
			ObjectiveID: objectiveID, TeamID: command.Objective.Team,
			WorkspaceID: uuidPointer(command.WorkspaceID),
			SequenceID:  sequenceID, Name: keyResult.Name,
			MeasurementType: objectivessql.MeasurementType(keyResult.MeasurementType),
			StartValue:      keyResult.StartValue, CurrentValue: keyResult.CurrentValue,
			TargetValue: keyResult.TargetValue, LeadUserID: keyResult.Lead,
			StartDate: keyResult.StartDate.UTC(), EndDate: keyResult.EndDate.UTC(),
			ActorID: uuidPointer(command.Objective.CreatedBy),
		})
		if err != nil {
			return nil, fmt.Errorf("create key result %d: %w", index, err)
		}
		contributors := deduplicateUUIDs(keyResult.Contributors)
		for _, contributorID := range contributors {
			if _, err := queries.AddObjectiveKeyResultContributor(ctx, objectivessql.AddObjectiveKeyResultContributorParams{
				KeyResultID: row.ID, UserID: contributorID,
				TeamID: uuidPointer(command.Objective.Team), WorkspaceID: uuidPointer(command.WorkspaceID),
				ActorID: command.Objective.CreatedBy,
			}); err != nil {
				return nil, fmt.Errorf("add key result contributor: %w", err)
			}
		}
		created = append(created, keyResultFromCreateRow(row, contributors))
	}
	return created, nil
}

func keyResultSequenceRange(finalSequence int32, count int) (int32, int32, error) {
	sequenceCount, err := safecast.Int32(count)
	if err != nil || sequenceCount <= 0 {
		return 0, 0, fmt.Errorf("invalid sequence count %d: %w", count, objectivesdomain.ErrInvalid)
	}
	if finalSequence == 0 {
		return sequenceCount, 0, nil
	}
	firstSequence, err := safecast.Int64ToInt32(int64(finalSequence) - int64(sequenceCount) + 1)
	if err != nil || firstSequence <= 0 {
		return 0, 0, fmt.Errorf("invalid final sequence %d for count %d: %w", finalSequence, count, objectivesdomain.ErrInvalid)
	}
	return sequenceCount, firstSequence, nil
}

func keyResultSequenceAt(firstSequence int32, index int) (int32, error) {
	indexValue, err := safecast.Int32(index)
	if err != nil || indexValue < 0 {
		return 0, fmt.Errorf("invalid sequence index %d: %w", index, objectivesdomain.ErrInvalid)
	}
	sequenceID, err := safecast.Int64ToInt32(int64(firstSequence) + int64(indexValue))
	if err != nil || sequenceID <= 0 {
		return 0, fmt.Errorf("invalid sequence value: %w", objectivesdomain.ErrInvalid)
	}
	return sequenceID, nil
}

func createObjectiveActivities(
	ctx context.Context,
	queries objectivessql.Querier,
	command objectivesdomain.CreateCommand,
	result objectivesdomain.CreateResult,
) error {
	field, value, comment := "all", result.Objective.Name, ""
	if err := queries.CreateObjectiveActivity(ctx, objectivessql.CreateObjectiveActivityParams{
		ObjectiveID: result.Objective.ID, ActorID: command.Objective.CreatedBy,
		ActivityType: objectivessql.OkrActivityTypeCreate,
		UpdateType:   objectivessql.OkrUpdateTypeObjective,
		FieldChanged: &field, CurrentValue: &value, Comment: &comment,
		WorkspaceID: command.WorkspaceID,
	}); err != nil {
		return fmt.Errorf("record objective create activity: %w", err)
	}
	for _, keyResult := range result.KeyResults {
		name, keyResultID := keyResult.Name, keyResult.ID
		if err := queries.CreateObjectiveActivity(ctx, objectivessql.CreateObjectiveActivityParams{
			ObjectiveID: result.Objective.ID, KeyResultID: &keyResultID,
			ActorID:      command.Objective.CreatedBy,
			ActivityType: objectivessql.OkrActivityTypeCreate,
			UpdateType:   objectivessql.OkrUpdateTypeKeyResult,
			CurrentValue: &name, Comment: &comment, WorkspaceID: command.WorkspaceID,
		}); err != nil {
			return fmt.Errorf("record key result create activity: %w", err)
		}
	}
	return nil
}

func collectAssignableUsers(command objectivesdomain.CreateCommand) []uuid.UUID {
	values := make([]uuid.UUID, 0)
	if command.Objective.LeadUser != nil {
		values = append(values, *command.Objective.LeadUser)
	}
	for _, keyResult := range command.KeyResults {
		if keyResult.Lead != nil {
			values = append(values, *keyResult.Lead)
		}
		values = append(values, keyResult.Contributors...)
	}
	return deduplicateUUIDs(values)
}

func deduplicateUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	unique := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}
	return unique
}

func nullableBool(value *bool) bool { return value != nil && *value }
