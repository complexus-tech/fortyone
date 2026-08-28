package objectivesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) Create(
	ctx context.Context,
	command objectivesdomain.CreateCommand,
) (objectivesdomain.CreateResult, error) {
	if err := command.Validate(); err != nil {
		return objectivesdomain.CreateResult{}, err
	}

	var result objectivesdomain.CreateResult
	err := repository.withinTransaction(ctx, func(queries objectivessql.Querier) error {
		authorization, err := queries.GetObjectiveCreateAuthorization(ctx, objectivessql.GetObjectiveCreateAuthorizationParams{
			TeamID: command.Objective.Team, WorkspaceID: command.WorkspaceID,
			ActorID: command.Objective.CreatedBy, StatusID: command.Objective.Status,
			LeadUserID: command.Objective.LeadUser,
		})
		if err != nil {
			return fmt.Errorf("authorize objective create: %w", err)
		}
		if !authorization.TeamExists || !authorization.StatusValid || !nullableBool(authorization.LeadValid) {
			return objectivesdomain.ErrInvalidReference
		}
		if !authorization.ActorAuthorized {
			return objectivesdomain.ErrForbidden
		}

		assignableUsers := collectAssignableUsers(command)
		if len(assignableUsers) > 0 {
			count, err := queries.CountAssignableObjectiveUsers(ctx, objectivessql.CountAssignableObjectiveUsersParams{
				WorkspaceID: command.WorkspaceID, TeamID: command.Objective.Team, UserIds: assignableUsers,
			})
			if err != nil {
				return fmt.Errorf("validate objective assignees: %w", err)
			}
			if int(count) != len(assignableUsers) {
				return objectivesdomain.ErrInvalidReference
			}
		}

		sequenceID, err := queries.AllocateObjectiveSequence(ctx, objectivessql.AllocateObjectiveSequenceParams{
			WorkspaceID: command.WorkspaceID, TeamID: command.Objective.Team,
		})
		if err != nil {
			return fmt.Errorf("allocate objective sequence: %w", err)
		}
		created, err := queries.CreateObjective(ctx, createObjectiveParams(command, sequenceID))
		if err != nil {
			return fmt.Errorf("create objective: %w", err)
		}
		result.Objective = objectiveFromCreateRow(created)

		result.KeyResults, err = repository.createKeyResults(ctx, queries, command, result.Objective.ID)
		if err != nil {
			return err
		}
		if err := createObjectiveActivities(ctx, queries, command, result); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return objectivesdomain.CreateResult{}, mapDatabaseError(err)
	}
	return result, nil
}

func (repository *Repository) Update(
	ctx context.Context,
	command objectivesdomain.UpdateCommand,
) (objectivesdomain.Objective, error) {
	if err := command.Validate(); err != nil {
		return objectivesdomain.Objective{}, err
	}

	var updated objectivesdomain.Objective
	err := repository.withinTransaction(ctx, func(queries objectivessql.Querier) error {
		current, err := queries.GetObjectiveForMutation(ctx, objectivessql.GetObjectiveForMutationParams{
			ActorID: command.ActorID, ObjectiveID: command.ObjectiveID,
			WorkspaceID: uuidPointer(command.WorkspaceID),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return objectivesdomain.ErrNotFound
			}
			return fmt.Errorf("authorize objective update: %w", err)
		}
		if command.ExpectedUpdatedAt != nil && !current.UpdatedAt.UTC().Equal(command.ExpectedUpdatedAt.UTC()) {
			return objectivesdomain.ErrVersionConflict
		}
		if current.TeamID == nil {
			return objectivesdomain.ErrInvalidReference
		}
		if err := validateObjectiveDatePatch(current.StartDate, current.EndDate, command.Patch); err != nil {
			return err
		}
		references, err := queries.ValidateObjectivePatchReferences(ctx, patchReferenceParams(command, *current.TeamID))
		if err != nil {
			return fmt.Errorf("validate objective update references: %w", err)
		}
		if !nullableBool(references.StatusValid) || !nullableBool(references.LeadValid) {
			return objectivesdomain.ErrInvalidReference
		}

		row, err := queries.UpdateObjective(ctx, objectivePatchParams(command))
		if err != nil {
			return fmt.Errorf("update objective: %w", err)
		}
		updated = objectiveFromUpdateRow(row)
		for _, change := range objectivePatchChanges(command.Patch) {
			field, value, comment := change.field, change.value, command.Comment
			if err := queries.CreateObjectiveActivity(ctx, objectivessql.CreateObjectiveActivityParams{
				ObjectiveID: command.ObjectiveID, ActorID: command.ActorID,
				ActivityType: objectivessql.OkrActivityTypeUpdate,
				UpdateType:   objectivessql.OkrUpdateTypeObjective,
				FieldChanged: &field, CurrentValue: &value, Comment: &comment,
				WorkspaceID: command.WorkspaceID,
			}); err != nil {
				return fmt.Errorf("record objective update activity: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return objectivesdomain.Objective{}, mapDatabaseError(err)
	}
	return updated, nil
}

func validateObjectiveDatePatch(
	currentStart, currentEnd *time.Time,
	patch objectivesdomain.ObjectivePatch,
) error {
	startDate, endDate := currentStart, currentEnd
	if value, specified := patch.StartDate.Value(); specified {
		startDate = value
	}
	if value, specified := patch.EndDate.Value(); specified {
		endDate = value
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return fmt.Errorf("%w: end date cannot be before start date", objectivesdomain.ErrInvalid)
	}
	return nil
}

func (repository *Repository) Delete(ctx context.Context, command objectivesdomain.DeleteCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}
	_, err := repository.queries.DeleteObjective(ctx, objectivessql.DeleteObjectiveParams{
		ObjectiveID: command.ObjectiveID, WorkspaceID: uuidPointer(command.WorkspaceID), ActorID: command.ActorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return objectivesdomain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("delete objective: %w", mapDatabaseError(err))
	}
	return nil
}
