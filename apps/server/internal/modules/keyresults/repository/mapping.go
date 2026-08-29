package keyresultsrepository

import (
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresultssql "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository/sqlc"
	"github.com/google/uuid"
)

func keyResultFromValues(
	id uuid.UUID,
	sequenceID int32,
	objectiveID uuid.UUID,
	name, measurementType string,
	startValue, currentValue, targetValue float64,
	lead *uuid.UUID,
	contributors []uuid.UUID,
	startDate, endDate, createdAt, updatedAt time.Time,
	createdBy *uuid.UUID,
) keyresultsdomain.KeyResult {
	creatorID := uuid.Nil
	if createdBy != nil {
		creatorID = *createdBy
	}
	start := startDate.UTC()
	end := endDate.UTC()
	return keyresultsdomain.KeyResult{
		ID: id, SequenceID: int(sequenceID), ObjectiveID: objectiveID,
		Name: name, MeasurementType: measurementType,
		StartValue: startValue, CurrentValue: currentValue, TargetValue: targetValue,
		Lead: lead, Contributors: append([]uuid.UUID(nil), contributors...),
		StartDate: &start, EndDate: &end, CreatedAt: createdAt.UTC(), UpdatedAt: updatedAt.UTC(),
		CreatedBy: creatorID,
	}
}

func keyResultFromGetRow(row keyresultssql.GetKeyResultRow) keyresultsdomain.KeyResult {
	return keyResultFromValues(
		row.ID, row.SequenceID, row.ObjectiveID, row.Name, row.MeasurementType,
		row.StartValue, row.CurrentValue, row.TargetValue, row.Lead, row.ContributorIds,
		row.StartDate, row.EndDate, row.CreatedAt, row.UpdatedAt, row.CreatedBy,
	)
}

func keyResultFromObjectiveListRow(row keyresultssql.ListObjectiveKeyResultsRow) keyresultsdomain.KeyResult {
	return keyResultFromValues(
		row.ID, row.SequenceID, row.ObjectiveID, row.Name, row.MeasurementType,
		row.StartValue, row.CurrentValue, row.TargetValue, row.Lead, row.ContributorIds,
		row.StartDate, row.EndDate, row.CreatedAt, row.UpdatedAt, row.CreatedBy,
	)
}

func keyResultFromMutationRow(row keyresultssql.GetKeyResultForMutationRow) keyresultsdomain.KeyResult {
	return keyResultFromValues(
		row.ID, row.SequenceID, row.ObjectiveID, row.Name, row.MeasurementType,
		row.StartValue, row.CurrentValue, row.TargetValue, row.Lead, row.ContributorIds,
		row.StartDate, row.EndDate, row.CreatedAt, row.UpdatedAt, row.CreatedBy,
	)
}

func keyResultFromCreateRow(row keyresultssql.CreateKeyResultRow, contributors []uuid.UUID) keyresultsdomain.KeyResult {
	return keyResultFromValues(
		row.ID, row.SequenceID, row.ObjectiveID, row.Name, row.MeasurementType,
		row.StartValue, row.CurrentValue, row.TargetValue, row.Lead, contributors,
		row.StartDate, row.EndDate, row.CreatedAt, row.UpdatedAt, row.CreatedBy,
	)
}

func keyResultFromUpdateRow(row keyresultssql.UpdateKeyResultRow, contributors []uuid.UUID) keyresultsdomain.KeyResult {
	return keyResultFromValues(
		row.ID, row.SequenceID, row.ObjectiveID, row.Name, row.MeasurementType,
		row.StartValue, row.CurrentValue, row.TargetValue, row.Lead, contributors,
		row.StartDate, row.EndDate, row.CreatedAt, row.UpdatedAt, row.CreatedBy,
	)
}

func keyResultWithObjectiveFromRow(row keyresultssql.ListKeyResultsRow) keyresultsdomain.KeyResultWithObjective {
	keyResult := keyResultFromValues(
		row.ID, row.SequenceID, row.ObjectiveID, row.Name, row.MeasurementType,
		row.StartValue, row.CurrentValue, row.TargetValue, row.Lead, row.ContributorIds,
		row.StartDate, row.EndDate, row.CreatedAt, row.UpdatedAt, row.CreatedBy,
	)
	teamID := uuid.Nil
	if row.TeamID != nil {
		teamID = *row.TeamID
	}
	workspaceID := uuid.Nil
	if row.WorkspaceID != nil {
		workspaceID = *row.WorkspaceID
	}
	return keyresultsdomain.KeyResultWithObjective{
		KeyResult: keyResult, ObjectiveName: row.ObjectiveName, ObjectiveID: row.ObjectiveID,
		TeamID: teamID, TeamName: row.TeamName, TeamCode: row.TeamCode, WorkspaceID: workspaceID,
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}
