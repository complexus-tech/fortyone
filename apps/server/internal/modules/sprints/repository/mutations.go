package sprintsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	sprintssql "github.com/complexus-tech/projects-api/internal/modules/sprints/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) Create(ctx context.Context, command sprintdomain.CreateCommand) (sprintdomain.Sprint, error) {
	command, err := command.Normalize()
	if err != nil {
		return sprintdomain.Sprint{}, err
	}

	var created sprintdomain.Sprint
	err = repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := repository.queries.WithTx(tx)
		_, err := queries.AuthorizeSprintCreate(ctx, sprintssql.AuthorizeSprintCreateParams{
			ActorID: command.ActorID, TeamID: command.Sprint.TeamID,
			WorkspaceID: command.Sprint.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return sprintdomain.ErrForbidden
		}
		if err != nil {
			return fmt.Errorf("authorize sprint create: %w", err)
		}
		if command.Sprint.ObjectiveID != nil {
			valid, err := queries.SprintObjectiveBelongsToTeam(ctx, sprintssql.SprintObjectiveBelongsToTeamParams{
				ObjectiveID: *command.Sprint.ObjectiveID,
				WorkspaceID: &command.Sprint.WorkspaceID,
				TeamID:      &command.Sprint.TeamID,
			})
			if err != nil {
				return fmt.Errorf("validate sprint objective: %w", err)
			}
			if !valid {
				return sprintdomain.ErrInvalidReference
			}
		}

		row, err := queries.CreateSprint(ctx, sprintssql.CreateSprintParams{
			Name: command.Sprint.Name, Goal: command.Sprint.Goal,
			ObjectiveID: command.Sprint.ObjectiveID, TeamID: command.Sprint.TeamID,
			WorkspaceID: command.Sprint.WorkspaceID,
			StartDate:   command.Sprint.StartDate, EndDate: command.Sprint.EndDate,
		})
		if err != nil {
			return mapMutationError("create sprint", err)
		}
		created, err = sprintFromCreateRow(row)
		if err != nil {
			return err
		}
		return insertAuditEvent(ctx, queries, created.WorkspaceID, created.TeamID, command.ActorID, created.ID, "sprint.created", map[string]any{
			"name": created.Name, "start_date": created.StartDate.Format(time.DateOnly), "end_date": created.EndDate.Format(time.DateOnly),
		})
	})
	if err != nil {
		return sprintdomain.Sprint{}, err
	}
	return created, nil
}

func (repository *Repository) Update(ctx context.Context, command sprintdomain.UpdateCommand) (sprintdomain.Sprint, error) {
	command, err := command.Normalize()
	if err != nil {
		return sprintdomain.Sprint{}, err
	}

	var updated sprintdomain.Sprint
	err = repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := repository.queries.WithTx(tx)
		lockedRow, err := queries.LockSprintForMutation(ctx, sprintssql.LockSprintForMutationParams{
			ActorID: command.ActorID, SprintID: command.SprintID, WorkspaceID: command.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return sprintdomain.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("lock sprint for update: %w", err)
		}
		current, err := sprintFromLockRow(lockedRow)
		if err != nil {
			return err
		}
		if command.ExpectedUpdatedAt != nil && !current.UpdatedAt.Equal(*command.ExpectedUpdatedAt) {
			return sprintdomain.ErrVersionConflict
		}
		if err := validateResultingReferences(ctx, queries, current, command.Patch); err != nil {
			return err
		}

		params := updateParams(command, current)
		row, err := queries.UpdateSprint(ctx, params)
		if errors.Is(err, pgx.ErrNoRows) && command.ExpectedUpdatedAt != nil {
			return sprintdomain.ErrVersionConflict
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return sprintdomain.ErrNotFound
		}
		if err != nil {
			return mapMutationError("update sprint", err)
		}
		updated, err = sprintFromUpdateRow(row)
		if err != nil {
			return err
		}
		return insertAuditEvent(ctx, queries, updated.WorkspaceID, updated.TeamID, command.ActorID, updated.ID, "sprint.updated", map[string]any{
			"name_changed": command.Patch.Name.Specified(), "goal_changed": command.Patch.Goal.Specified(),
			"objective_changed":  command.Patch.ObjectiveID.Specified(),
			"start_date_changed": command.Patch.StartDate.Specified(), "end_date_changed": command.Patch.EndDate.Specified(),
		})
	})
	if err != nil {
		return sprintdomain.Sprint{}, err
	}
	return updated, nil
}

func (repository *Repository) Delete(ctx context.Context, command sprintdomain.DeleteCommand) error {
	if err := command.Validate(); err != nil {
		return err
	}
	return repository.transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
		queries := repository.queries.WithTx(tx)
		if _, err := queries.LockSprintForMutation(ctx, sprintssql.LockSprintForMutationParams{
			ActorID: command.ActorID, SprintID: command.SprintID, WorkspaceID: command.WorkspaceID,
		}); errors.Is(err, pgx.ErrNoRows) {
			return sprintdomain.ErrNotFound
		} else if err != nil {
			return fmt.Errorf("lock sprint for delete: %w", err)
		}
		deleted, err := queries.DeleteSprint(ctx, sprintssql.DeleteSprintParams{
			SprintID: command.SprintID, WorkspaceID: command.WorkspaceID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return sprintdomain.ErrNotFound
		}
		if err != nil {
			return mapMutationError("delete sprint", err)
		}
		return insertAuditEvent(ctx, queries, command.WorkspaceID, deleted.TeamID, command.ActorID, command.SprintID, "sprint.deleted", map[string]any{
			"name": deleted.Name, "start_date": deleted.StartDate.Format(time.DateOnly), "end_date": deleted.EndDate.Format(time.DateOnly),
		})
	})
}

func updateParams(command sprintdomain.UpdateCommand, current sprintdomain.Sprint) sprintssql.UpdateSprintParams {
	name, nameSet := command.Patch.Name.Value()
	goal, goalSet := command.Patch.Goal.Value()
	objective, objectiveSet := command.Patch.ObjectiveID.Value()
	startDate, startDateSet := command.Patch.StartDate.Value()
	endDate, endDateSet := command.Patch.EndDate.Value()

	params := sprintssql.UpdateSprintParams{
		NameSet: nameSet, GoalSet: goalSet, Goal: goal,
		ObjectiveSet: objectiveSet, ObjectiveID: objective,
		StartDateSet: startDateSet, EndDateSet: endDateSet,
		SprintID: command.SprintID, WorkspaceID: command.WorkspaceID,
		ExpectedUpdatedAtSet: command.ExpectedUpdatedAt != nil,
	}
	if name != nil {
		params.Name = *name
	}
	if startDate != nil {
		params.StartDate = *startDate
	} else {
		params.StartDate = current.StartDate
	}
	if endDate != nil {
		params.EndDate = *endDate
	} else {
		params.EndDate = current.EndDate
	}
	if command.ExpectedUpdatedAt != nil {
		params.ExpectedUpdatedAt = *command.ExpectedUpdatedAt
	}
	return params
}

func validateResultingReferences(
	ctx context.Context,
	queries *sprintssql.Queries,
	current sprintdomain.Sprint,
	patch sprintdomain.Patch,
) error {
	startDate, endDate := current.StartDate, current.EndDate
	if value, specified := patch.StartDate.Value(); specified && value != nil {
		startDate = *value
	}
	if value, specified := patch.EndDate.Value(); specified && value != nil {
		endDate = *value
	}
	if endDate.Before(startDate) {
		return fmt.Errorf("%w: end date cannot be before start date", sprintdomain.ErrInvalid)
	}

	objectiveID, objectiveSet := patch.ObjectiveID.Value()
	if !objectiveSet || objectiveID == nil {
		return nil
	}
	valid, err := queries.SprintObjectiveBelongsToTeam(ctx, sprintssql.SprintObjectiveBelongsToTeamParams{
		ObjectiveID: *objectiveID, WorkspaceID: &current.WorkspaceID, TeamID: &current.TeamID,
	})
	if err != nil {
		return fmt.Errorf("validate sprint objective: %w", err)
	}
	if !valid {
		return sprintdomain.ErrInvalidReference
	}
	return nil
}

func insertAuditEvent(
	ctx context.Context,
	queries *sprintssql.Queries,
	workspaceID, teamID, actorID, sprintID uuid.UUID,
	eventType string,
	metadata map[string]any,
) error {
	raw, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal sprint audit metadata: %w", err)
	}
	if err := queries.InsertSprintAuditEvent(ctx, sprintssql.InsertSprintAuditEventParams{
		WorkspaceID: workspaceID, TeamID: &teamID, ActorID: &actorID,
		SprintID: &sprintID, EventType: eventType, Metadata: raw,
	}); err != nil {
		return fmt.Errorf("insert sprint audit event: %w", err)
	}
	return nil
}

func mapMutationError(operation string, err error) error {
	switch platformdatabase.Classify(err) {
	case platformdatabase.ErrorClassForeignKeyViolation:
		return fmt.Errorf("%s: %w", operation, sprintdomain.ErrInvalidReference)
	case platformdatabase.ErrorClassCheckViolation, platformdatabase.ErrorClassNotNullViolation:
		return fmt.Errorf("%s: %w", operation, sprintdomain.ErrInvalid)
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}
