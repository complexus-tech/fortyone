package statesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	statesdomain "github.com/complexus-tech/projects-api/internal/modules/states/domain"
	statessql "github.com/complexus-tech/projects-api/internal/modules/states/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	errRepositoryNotConfigured = errors.New("states repository is not configured")
	errTransactionsUnavailable = errors.New("states repository transactions are unavailable")
)

type Repository struct {
	queries        statessql.Querier
	runTransaction func(context.Context, func(statessql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	queries := statessql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(statessql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries statessql.Querier) *Repository {
	return &Repository{queries: queries}
}

func (repository *Repository) configured() error {
	if repository == nil || repository.queries == nil {
		return errRepositoryNotConfigured
	}
	return nil
}

func (repository *Repository) withinTransaction(ctx context.Context, operation func(statessql.Querier) error) error {
	if err := repository.configured(); err != nil {
		return err
	}
	if repository.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return repository.runTransaction(ctx, operation)
}

func (repository *Repository) Create(ctx context.Context, actorID, workspaceID uuid.UUID, input statesdomain.NewState) (statesdomain.State, error) {
	var created statesdomain.State
	err := repository.withinTransaction(ctx, func(queries statessql.Querier) error {
		exists, err := queries.StateTeamExistsForMember(ctx, statessql.StateTeamExistsForMemberParams{
			ActorID: actorID, TeamID: input.Team, WorkspaceID: workspaceID,
		})
		if err != nil {
			return fmt.Errorf("authorize state team: %w", err)
		}
		if !exists {
			return statesdomain.ErrNotFound
		}
		if err := queries.LockStateOrdering(ctx, statessql.LockStateOrderingParams{
			WorkspaceID: workspaceID, Category: input.Category,
		}); err != nil {
			return fmt.Errorf("lock state ordering: %w", err)
		}
		if input.IsDefault {
			if err := queries.LockStateDefaults(ctx, statessql.LockStateDefaultsParams{TeamID: input.Team}); err != nil {
				return fmt.Errorf("lock state defaults: %w", err)
			}
			if err := queries.ResetStateDefaults(ctx, statessql.ResetStateDefaultsParams{
				WorkspaceID: workspaceID, TeamID: input.Team,
			}); err != nil {
				return fmt.Errorf("reset state defaults: %w", err)
			}
		}
		orderIndex, err := queries.NextStateOrderIndex(ctx, statessql.NextStateOrderIndexParams{
			WorkspaceID: workspaceID, Category: input.Category,
		})
		if err != nil {
			return fmt.Errorf("allocate state order: %w", err)
		}
		row, err := queries.InsertState(ctx, statessql.InsertStateParams{
			Name: input.Name, Category: input.Category, OrderIndex: orderIndex,
			Color: input.Color, TeamID: input.Team, WorkspaceID: workspaceID,
			IsDefault: input.IsDefault,
		})
		if err != nil {
			return fmt.Errorf("insert state: %w", err)
		}
		created = stateFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex, row.TeamID,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt)
		return nil
	})
	return created, err
}

func (repository *Repository) Update(ctx context.Context, actorID, workspaceID, stateID uuid.UUID, input statesdomain.UpdateState) (statesdomain.State, error) {
	if input.Name == nil && input.OrderIndex == nil && input.IsDefault == nil && input.Color == nil {
		return statesdomain.State{}, statesdomain.ErrNoFields
	}
	if err := repository.configured(); err != nil {
		return statesdomain.State{}, err
	}

	update := func(queries statessql.Querier) (statesdomain.State, error) {
		params := statessql.UpdateStateForMemberParams{
			SetName: input.Name != nil, SetOrderIndex: input.OrderIndex != nil,
			SetIsDefault: input.IsDefault != nil, SetColor: input.Color != nil,
			StatusID: stateID, WorkspaceID: workspaceID, ActorID: actorID,
		}
		if input.Name != nil {
			params.Name = *input.Name
		}
		if input.OrderIndex != nil {
			orderIndex, err := stateOrderIndex(*input.OrderIndex)
			if err != nil {
				return statesdomain.State{}, err
			}
			params.OrderIndex = orderIndex
		}
		if input.IsDefault != nil {
			params.IsDefault = *input.IsDefault
		}
		if input.Color != nil {
			params.Color = *input.Color
		}
		row, err := queries.UpdateStateForMember(ctx, params)
		if err != nil {
			return statesdomain.State{}, mapError("update state", err)
		}
		return stateFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex, row.TeamID,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt), nil
	}

	if input.IsDefault == nil || !*input.IsDefault {
		return update(repository.queries)
	}

	var updated statesdomain.State
	err := repository.withinTransaction(ctx, func(queries statessql.Querier) error {
		teamID, err := queries.GetStateTeamForMember(ctx, statessql.GetStateTeamForMemberParams{
			ActorID: actorID, StatusID: stateID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return mapError("get state team", err)
		}
		if err := queries.LockStateDefaults(ctx, statessql.LockStateDefaultsParams{TeamID: teamID}); err != nil {
			return fmt.Errorf("lock state defaults: %w", err)
		}
		if err := queries.ResetStateDefaults(ctx, statessql.ResetStateDefaultsParams{
			WorkspaceID: workspaceID, TeamID: teamID,
		}); err != nil {
			return fmt.Errorf("reset state defaults: %w", err)
		}
		updated, err = update(queries)
		return err
	})
	return updated, err
}

func stateOrderIndex(value int) (int32, error) {
	orderIndex, err := safecast.Int32(value)
	if err != nil {
		return 0, statesdomain.ErrInvalidOrder
	}
	return orderIndex, nil
}

func (repository *Repository) Delete(ctx context.Context, actorID, workspaceID, stateID uuid.UUID) error {
	return repository.withinTransaction(ctx, func(queries statessql.Querier) error {
		target, err := queries.GetStateForDelete(ctx, statessql.GetStateForDeleteParams{
			ActorID: actorID, StatusID: stateID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return mapError("get state for delete", err)
		}
		category := stringValue(target.Category)
		if err := queries.LockStateCategory(ctx, statessql.LockStateCategoryParams{
			TeamID: target.TeamID, Category: category,
		}); err != nil {
			return fmt.Errorf("lock state category: %w", err)
		}
		stories, err := queries.CountWorkspaceStoriesWithState(ctx, statessql.CountWorkspaceStoriesWithStateParams{
			WorkspaceID: workspaceID, StatusID: stateID,
		})
		if err != nil {
			return fmt.Errorf("count state stories: %w", err)
		}
		if stories > 0 {
			return statesdomain.ErrStatusHasStories
		}
		states, err := queries.CountTeamStatesInCategory(ctx, statessql.CountTeamStatesInCategoryParams{
			WorkspaceID: workspaceID, TeamID: target.TeamID, Category: category,
		})
		if err != nil {
			return fmt.Errorf("count category states: %w", err)
		}
		if states <= 1 {
			return statesdomain.ErrLastInCategory
		}
		rows, err := queries.DeleteState(ctx, statessql.DeleteStateParams{StatusID: stateID, WorkspaceID: workspaceID})
		if err != nil {
			return fmt.Errorf("delete state: %w", err)
		}
		if rows != 1 {
			return statesdomain.ErrNotFound
		}
		return nil
	})
}

func (repository *Repository) Get(ctx context.Context, workspaceID, stateID uuid.UUID) (statesdomain.State, error) {
	if err := repository.configured(); err != nil {
		return statesdomain.State{}, err
	}
	row, err := repository.queries.GetState(ctx, statessql.GetStateParams{StatusID: stateID, WorkspaceID: workspaceID})
	if err != nil {
		return statesdomain.State{}, mapError("get state", err)
	}
	return stateFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex, row.TeamID,
		row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt), nil
}

func (repository *Repository) List(ctx context.Context, workspaceID, actorID uuid.UUID) ([]statesdomain.State, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListStatesForMember(ctx, statessql.ListStatesForMemberParams{
		ActorID: actorID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list member states: %w", err)
	}
	states := make([]statesdomain.State, len(rows))
	for index, row := range rows {
		states[index] = stateFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex, row.TeamID,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt)
	}
	return states, nil
}

func (repository *Repository) TeamList(ctx context.Context, workspaceID, teamID uuid.UUID) ([]statesdomain.State, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListTeamStates(ctx, statessql.ListTeamStatesParams{WorkspaceID: workspaceID, TeamID: teamID})
	if err != nil {
		return nil, fmt.Errorf("list team states: %w", err)
	}
	states := make([]statesdomain.State, len(rows))
	for index, row := range rows {
		states[index] = stateFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex, row.TeamID,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt)
	}
	return states, nil
}

func (repository *Repository) TeamListForMember(ctx context.Context, workspaceID, teamID, actorID uuid.UUID) ([]statesdomain.State, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListTeamStatesForMember(ctx, statessql.ListTeamStatesForMemberParams{
		ActorID: actorID, WorkspaceID: workspaceID, TeamID: teamID,
	})
	if err != nil {
		return nil, fmt.Errorf("list team states for member: %w", err)
	}
	states := make([]statesdomain.State, len(rows))
	for index, row := range rows {
		states[index] = stateFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex, row.TeamID,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt)
	}
	return states, nil
}

func stateFromValues(
	id uuid.UUID,
	name string,
	category *string,
	orderIndex *int32,
	teamID uuid.UUID,
	workspaceID uuid.UUID,
	isDefault bool,
	color *string,
	createdAt time.Time,
	updatedAt time.Time,
) statesdomain.State {
	return statesdomain.State{
		ID: id, Name: name, Category: stringValue(category), OrderIndex: intValue(orderIndex),
		Team: teamID, Workspace: workspaceID, IsDefault: isDefault, Color: stringValue(color),
		CreatedAt: createdAt, UpdatedAt: updatedAt,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int32) int {
	if value == nil {
		return 0
	}
	return int(*value)
}

func mapError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return statesdomain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
