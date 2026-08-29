package keyresultsrepository

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresultssql "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) Update(
	ctx context.Context,
	command keyresultsdomain.UpdateCommand,
) (keyresultsdomain.MutationResult, error) {
	normalized, err := command.Normalize()
	if err != nil {
		return keyresultsdomain.MutationResult{}, err
	}
	var result keyresultsdomain.MutationResult
	err = repository.withinTransaction(ctx, func(queries keyresultssql.Querier) error {
		row, err := queries.GetKeyResultForMutation(ctx, keyresultssql.GetKeyResultForMutationParams{
			ActorID: normalized.Access.ActorID, KeyResultID: normalized.KeyResultID,
			WorkspaceID: uuidPointer(normalized.Access.WorkspaceID), AllTeams: normalized.Access.AllTeams,
			AllowedTeamIds: normalized.Access.TeamIDs,
		})
		if err != nil {
			if err == pgx.ErrNoRows {
				return keyresultsdomain.ErrNotFound
			}
			return fmt.Errorf("lock key result: %w", err)
		}
		before := keyResultFromMutationRow(row)
		result.Before, result.After = before, before
		if normalized.ExpectedUpdatedAt != nil && !before.UpdatedAt.Equal(*normalized.ExpectedUpdatedAt) {
			return keyresultsdomain.ErrVersionConflict
		}

		changed := changedFields(before, normalized.Patch)
		if len(changed) == 0 {
			return nil
		}
		if err := validateEffectiveDates(before, normalized.Patch); err != nil {
			return err
		}
		teamID := uuid.Nil
		if row.TeamID != nil {
			teamID = *row.TeamID
		}
		if teamID == uuid.Nil {
			return keyresultsdomain.ErrInvalidReference
		}
		if err := validatePatchAssignees(ctx, queries, normalized, teamID); err != nil {
			return err
		}

		parameters := updateParameters(before, normalized)
		updatedRow, err := queries.UpdateKeyResult(ctx, parameters)
		if err != nil {
			if err == pgx.ErrNoRows {
				return keyresultsdomain.ErrNotFound
			}
			return fmt.Errorf("update key result: %w", err)
		}
		contributors := before.Contributors
		if normalized.Patch.Contributors.Set {
			contributors = normalized.Patch.Contributors.Value
			if err := replaceContributors(ctx, queries, normalized, contributors); err != nil {
				return err
			}
		}
		after := keyResultFromUpdateRow(updatedRow, contributors)
		if err := recordUpdateActivities(ctx, queries, normalized, after, changed); err != nil {
			return err
		}
		result.After = after
		result.ChangedFields = changed
		return nil
	})
	if err != nil {
		return keyresultsdomain.MutationResult{}, fmt.Errorf("update key result: %w", err)
	}
	return result, nil
}

func changedFields(current keyresultsdomain.KeyResult, patch keyresultsdomain.Patch) []string {
	changed := make([]string, 0, 9)
	if patch.Name.Set && patch.Name.Value != current.Name {
		changed = append(changed, "name")
	}
	if patch.MeasurementType.Set && patch.MeasurementType.Value != current.MeasurementType {
		changed = append(changed, "measurement_type")
	}
	if patch.StartValue.Set && patch.StartValue.Value != current.StartValue {
		changed = append(changed, "start_value")
	}
	if patch.CurrentValue.Set && patch.CurrentValue.Value != current.CurrentValue {
		changed = append(changed, "current_value")
	}
	if patch.TargetValue.Set && patch.TargetValue.Value != current.TargetValue {
		changed = append(changed, "target_value")
	}
	if patch.Lead.Set && !equalUUIDPointer(patch.Lead.Value, current.Lead) {
		changed = append(changed, "lead")
	}
	if patch.Contributors.Set && !slices.Equal(patch.Contributors.Value, current.Contributors) {
		changed = append(changed, "contributors")
	}
	if patch.StartDate.Set && !equalTimePointer(patch.StartDate.Value, current.StartDate) {
		changed = append(changed, "start_date")
	}
	if patch.EndDate.Set && !equalTimePointer(patch.EndDate.Value, current.EndDate) {
		changed = append(changed, "end_date")
	}
	return changed
}

func validateEffectiveDates(current keyresultsdomain.KeyResult, patch keyresultsdomain.Patch) error {
	startDate, endDate := current.StartDate, current.EndDate
	if patch.StartDate.Set {
		startDate = patch.StartDate.Value
	}
	if patch.EndDate.Set {
		endDate = patch.EndDate.Value
	}
	if startDate == nil || endDate == nil || endDate.Before(*startDate) {
		return fmt.Errorf("%w: end date cannot be before start date", keyresultsdomain.ErrInvalid)
	}
	return nil
}

func validatePatchAssignees(
	ctx context.Context,
	queries keyresultssql.Querier,
	command keyresultsdomain.UpdateCommand,
	teamID uuid.UUID,
) error {
	candidates := make([]uuid.UUID, 0)
	if command.Patch.Lead.Set && command.Patch.Lead.Value != nil {
		candidates = append(candidates, *command.Patch.Lead.Value)
	}
	if command.Patch.Contributors.Set {
		candidates = append(candidates, command.Patch.Contributors.Value...)
	}
	if len(candidates) == 0 {
		return nil
	}
	valid, err := queries.ValidateKeyResultAssignees(ctx, keyresultssql.ValidateKeyResultAssigneesParams{
		UserIds: candidates, WorkspaceID: command.Access.WorkspaceID, TeamID: teamID,
	})
	if err != nil {
		return fmt.Errorf("validate key result assignees: %w", err)
	}
	if !valid {
		return keyresultsdomain.ErrInvalidReference
	}
	return nil
}

func updateParameters(current keyresultsdomain.KeyResult, command keyresultsdomain.UpdateCommand) keyresultssql.UpdateKeyResultParams {
	startDate, endDate := *current.StartDate, *current.EndDate
	if command.Patch.StartDate.Set {
		startDate = *command.Patch.StartDate.Value
	}
	if command.Patch.EndDate.Set {
		endDate = *command.Patch.EndDate.Value
	}
	name := current.Name
	if command.Patch.Name.Set {
		name = command.Patch.Name.Value
	}
	measurementType := current.MeasurementType
	if command.Patch.MeasurementType.Set {
		measurementType = command.Patch.MeasurementType.Value
	}
	startValue := current.StartValue
	if command.Patch.StartValue.Set {
		startValue = command.Patch.StartValue.Value
	}
	currentValue := current.CurrentValue
	if command.Patch.CurrentValue.Set {
		currentValue = command.Patch.CurrentValue.Value
	}
	targetValue := current.TargetValue
	if command.Patch.TargetValue.Set {
		targetValue = command.Patch.TargetValue.Value
	}
	lead := current.Lead
	if command.Patch.Lead.Set {
		lead = command.Patch.Lead.Value
	}
	return keyresultssql.UpdateKeyResultParams{
		SetName: command.Patch.Name.Set, Name: name,
		SetMeasurementType: command.Patch.MeasurementType.Set,
		MeasurementType:    keyresultssql.MeasurementType(measurementType),
		SetStartValue:      command.Patch.StartValue.Set, StartValue: startValue,
		SetCurrentValue: command.Patch.CurrentValue.Set, CurrentValue: currentValue,
		SetTargetValue: command.Patch.TargetValue.Set, TargetValue: targetValue,
		SetLead: command.Patch.Lead.Set, LeadUserID: lead,
		SetStartDate: command.Patch.StartDate.Set, StartDate: startDate,
		SetEndDate: command.Patch.EndDate.Set, EndDate: endDate,
		KeyResultID: command.KeyResultID, WorkspaceID: uuidPointer(command.Access.WorkspaceID),
		ActorID: command.Access.ActorID, AllTeams: command.Access.AllTeams,
		AllowedTeamIds: command.Access.TeamIDs,
	}
}

func replaceContributors(
	ctx context.Context,
	queries keyresultssql.Querier,
	command keyresultsdomain.UpdateCommand,
	contributors []uuid.UUID,
) error {
	if err := queries.DeleteKeyResultContributors(ctx, keyresultssql.DeleteKeyResultContributorsParams{
		KeyResultID: command.KeyResultID, WorkspaceID: uuidPointer(command.Access.WorkspaceID),
	}); err != nil {
		return fmt.Errorf("delete key result contributors: %w", err)
	}
	for _, contributorID := range contributors {
		rows, err := queries.AddKeyResultContributor(ctx, keyresultssql.AddKeyResultContributorParams{
			UserID: contributorID, KeyResultID: command.KeyResultID,
			WorkspaceID: uuidPointer(command.Access.WorkspaceID),
		})
		if err != nil {
			return fmt.Errorf("add key result contributor: %w", err)
		}
		if rows != 1 {
			return keyresultsdomain.ErrInvalidReference
		}
	}
	return nil
}

func recordUpdateActivities(
	ctx context.Context,
	queries keyresultssql.Querier,
	command keyresultsdomain.UpdateCommand,
	after keyresultsdomain.KeyResult,
	changed []string,
) error {
	for index, field := range changed {
		value := activityValue(after, field)
		comment := ""
		if index == len(changed)-1 {
			comment = command.Comment
		}
		rows, err := queries.CreateKeyResultActivity(ctx, keyresultssql.CreateKeyResultActivityParams{
			ActivityType: keyresultssql.OkrActivityTypeUpdate,
			FieldChanged: &field, CurrentValue: &value, Comment: &comment,
			ActorID: command.Access.ActorID, KeyResultID: command.KeyResultID,
			WorkspaceID: uuidPointer(command.Access.WorkspaceID),
		})
		if err != nil {
			return fmt.Errorf("record key result update activity: %w", err)
		}
		if rows != 1 {
			return keyresultsdomain.ErrForbidden
		}
	}
	return nil
}

func activityValue(keyResult keyresultsdomain.KeyResult, field string) string {
	switch field {
	case "name":
		return keyResult.Name
	case "measurement_type":
		return keyResult.MeasurementType
	case "start_value":
		return strconv.FormatFloat(keyResult.StartValue, 'f', -1, 64)
	case "current_value":
		return strconv.FormatFloat(keyResult.CurrentValue, 'f', -1, 64)
	case "target_value":
		return strconv.FormatFloat(keyResult.TargetValue, 'f', -1, 64)
	case "lead":
		if keyResult.Lead != nil {
			return keyResult.Lead.String()
		}
	case "contributors":
		values := make([]string, len(keyResult.Contributors))
		for index, contributorID := range keyResult.Contributors {
			values[index] = contributorID.String()
		}
		return strings.Join(values, ",")
	case "start_date":
		if keyResult.StartDate != nil {
			return keyResult.StartDate.Format(time.RFC3339)
		}
	case "end_date":
		if keyResult.EndDate != nil {
			return keyResult.EndDate.Format(time.RFC3339)
		}
	}
	return ""
}

func equalUUIDPointer(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func equalTimePointer(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}
