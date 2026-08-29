package objectivestatusrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	objectivestatusdomain "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/domain"
	objectivestatussql "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	errRepositoryNotConfigured = errors.New("objective statuses repository is not configured")
	errTransactionsUnavailable = errors.New("objective statuses repository transactions are unavailable")
)

type Repository struct {
	queries        objectivestatussql.Querier
	runTransaction func(context.Context, func(objectivestatussql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	queries := objectivestatussql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(objectivestatussql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{
			IsoLevel:   pgx.ReadCommitted,
			AccessMode: pgx.ReadWrite,
		}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries objectivestatussql.Querier) *Repository {
	return &Repository{queries: queries}
}

func (repository *Repository) configured() error {
	if repository == nil || repository.queries == nil {
		return errRepositoryNotConfigured
	}
	return nil
}

func (repository *Repository) withinTransaction(ctx context.Context, operation func(objectivestatussql.Querier) error) error {
	if err := repository.configured(); err != nil {
		return err
	}
	if repository.runTransaction == nil {
		return errTransactionsUnavailable
	}
	return repository.runTransaction(ctx, operation)
}

func (repository *Repository) Create(ctx context.Context, actorID, workspaceID uuid.UUID, input objectivestatusdomain.NewStatus) (objectivestatusdomain.Status, error) {
	var created objectivestatusdomain.Status
	err := repository.withinTransaction(ctx, func(queries objectivestatussql.Querier) error {
		authorized, err := queries.ObjectiveStatusAdminAuthorized(ctx, objectivestatussql.ObjectiveStatusAdminAuthorizedParams{
			WorkspaceID: workspaceID, ActorID: actorID,
		})
		if err != nil {
			return fmt.Errorf("authorize objective status: %w", err)
		}
		if !authorized {
			return objectivestatusdomain.ErrNotFound
		}
		if err := queries.LockObjectiveStatusOrdering(ctx, objectivestatussql.LockObjectiveStatusOrderingParams{
			WorkspaceID: workspaceID, Category: input.Category,
		}); err != nil {
			return fmt.Errorf("lock objective status ordering: %w", err)
		}
		if input.IsDefault {
			if err := queries.LockObjectiveStatusDefaults(ctx, objectivestatussql.LockObjectiveStatusDefaultsParams{
				WorkspaceID: workspaceID,
			}); err != nil {
				return fmt.Errorf("lock objective status defaults: %w", err)
			}
			if err := queries.ResetObjectiveStatusDefaults(ctx, objectivestatussql.ResetObjectiveStatusDefaultsParams{
				WorkspaceID: workspaceID,
			}); err != nil {
				return fmt.Errorf("reset objective status defaults: %w", err)
			}
		}
		orderIndex, err := queries.NextObjectiveStatusOrderIndex(ctx, objectivestatussql.NextObjectiveStatusOrderIndexParams{
			WorkspaceID: workspaceID, Category: input.Category,
		})
		if err != nil {
			return fmt.Errorf("allocate objective status order: %w", err)
		}
		row, err := queries.InsertObjectiveStatus(ctx, objectivestatussql.InsertObjectiveStatusParams{
			Name: input.Name, Category: input.Category, OrderIndex: orderIndex,
			WorkspaceID: workspaceID, IsDefault: input.IsDefault, Color: input.Color,
		})
		if err != nil {
			return fmt.Errorf("insert objective status: %w", err)
		}
		created = objectiveStatusFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt)
		return nil
	})
	return created, err
}

func (repository *Repository) Update(ctx context.Context, actorID, workspaceID, statusID uuid.UUID, input objectivestatusdomain.UpdateStatus) (objectivestatusdomain.Status, error) {
	if input.Name == nil && input.OrderIndex == nil && input.IsDefault == nil && input.Color == nil {
		return objectivestatusdomain.Status{}, objectivestatusdomain.ErrNoFields
	}
	if err := repository.configured(); err != nil {
		return objectivestatusdomain.Status{}, err
	}

	update := func(queries objectivestatussql.Querier) (objectivestatusdomain.Status, error) {
		params := objectivestatussql.UpdateObjectiveStatusForAdminParams{
			SetName: input.Name != nil, SetOrderIndex: input.OrderIndex != nil,
			SetIsDefault: input.IsDefault != nil, SetColor: input.Color != nil,
			StatusID: statusID, WorkspaceID: workspaceID, ActorID: actorID,
		}
		if input.Name != nil {
			params.Name = *input.Name
		}
		if input.OrderIndex != nil {
			orderIndex, err := objectiveStatusOrderIndex(*input.OrderIndex)
			if err != nil {
				return objectivestatusdomain.Status{}, err
			}
			params.OrderIndex = orderIndex
		}
		if input.IsDefault != nil {
			params.IsDefault = *input.IsDefault
		}
		if input.Color != nil {
			params.Color = *input.Color
		}
		row, err := queries.UpdateObjectiveStatusForAdmin(ctx, params)
		if err != nil {
			return objectivestatusdomain.Status{}, mapError("update objective status", err)
		}
		return objectiveStatusFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt), nil
	}

	if input.IsDefault == nil || !*input.IsDefault {
		return update(repository.queries)
	}

	var updated objectivestatusdomain.Status
	err := repository.withinTransaction(ctx, func(queries objectivestatussql.Querier) error {
		exists, err := queries.ObjectiveStatusExistsForAdmin(ctx, objectivestatussql.ObjectiveStatusExistsForAdminParams{
			ActorID: actorID, StatusID: statusID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return fmt.Errorf("authorize objective status update: %w", err)
		}
		if !exists {
			return objectivestatusdomain.ErrNotFound
		}
		if err := queries.LockObjectiveStatusDefaults(ctx, objectivestatussql.LockObjectiveStatusDefaultsParams{
			WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("lock objective status defaults: %w", err)
		}
		if err := queries.ResetObjectiveStatusDefaults(ctx, objectivestatussql.ResetObjectiveStatusDefaultsParams{
			WorkspaceID: workspaceID,
		}); err != nil {
			return fmt.Errorf("reset objective status defaults: %w", err)
		}
		updated, err = update(queries)
		return err
	})
	return updated, err
}

func objectiveStatusOrderIndex(value int) (int32, error) {
	orderIndex, err := safecast.Int32(value)
	if err != nil {
		return 0, objectivestatusdomain.ErrInvalidOrder
	}
	return orderIndex, nil
}

func (repository *Repository) Delete(ctx context.Context, actorID, workspaceID, statusID uuid.UUID) error {
	return repository.withinTransaction(ctx, func(queries objectivestatussql.Querier) error {
		categoryValue, err := queries.GetObjectiveStatusForDelete(ctx, objectivestatussql.GetObjectiveStatusForDeleteParams{
			ActorID: actorID, StatusID: statusID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return mapError("get objective status for delete", err)
		}
		category := valueOrEmpty(categoryValue)
		if err := queries.LockObjectiveStatusCategory(ctx, objectivestatussql.LockObjectiveStatusCategoryParams{
			WorkspaceID: workspaceID, Category: category,
		}); err != nil {
			return fmt.Errorf("lock objective status category: %w", err)
		}
		objectives, err := queries.CountObjectivesWithStatus(ctx, objectivestatussql.CountObjectivesWithStatusParams{
			WorkspaceID: workspaceID, StatusID: statusID,
		})
		if err != nil {
			return fmt.Errorf("count status objectives: %w", err)
		}
		if objectives > 0 {
			return objectivestatusdomain.ErrStatusHasObjectives
		}
		statuses, err := queries.CountObjectiveStatusesInCategory(ctx, objectivestatussql.CountObjectiveStatusesInCategoryParams{
			WorkspaceID: workspaceID, Category: category,
		})
		if err != nil {
			return fmt.Errorf("count objective statuses in category: %w", err)
		}
		if statuses <= 1 {
			return objectivestatusdomain.ErrLastInCategory
		}
		rows, err := queries.DeleteObjectiveStatus(ctx, objectivestatussql.DeleteObjectiveStatusParams{
			StatusID: statusID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return fmt.Errorf("delete objective status: %w", err)
		}
		if rows != 1 {
			return objectivestatusdomain.ErrNotFound
		}
		return nil
	})
}

func (repository *Repository) List(ctx context.Context, workspaceID uuid.UUID) ([]objectivestatusdomain.Status, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListObjectiveStatuses(ctx, objectivestatussql.ListObjectiveStatusesParams{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list objective statuses: %w", err)
	}
	statuses := make([]objectivestatusdomain.Status, len(rows))
	for index, row := range rows {
		statuses[index] = objectiveStatusFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt)
	}
	return statuses, nil
}

func (repository *Repository) ListForMember(ctx context.Context, actorID, workspaceID uuid.UUID) ([]objectivestatusdomain.Status, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	rows, err := repository.queries.ListObjectiveStatusesForMember(ctx, objectivestatussql.ListObjectiveStatusesForMemberParams{
		ActorID: actorID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list objective statuses: %w", err)
	}
	statuses := make([]objectivestatusdomain.Status, len(rows))
	for index, row := range rows {
		statuses[index] = objectiveStatusFromValues(row.StatusID, row.Name, row.Category, row.OrderIndex,
			row.WorkspaceID, row.IsDefault, row.Color, row.CreatedAt, row.UpdatedAt)
	}
	return statuses, nil
}

func objectiveStatusFromValues(
	id uuid.UUID,
	name string,
	category *string,
	orderIndex *int32,
	workspaceID uuid.UUID,
	isDefault bool,
	color *string,
	createdAt *time.Time,
	updatedAt *time.Time,
) objectivestatusdomain.Status {
	return objectivestatusdomain.Status{
		ID: id, Name: name, Category: valueOrEmpty(category), OrderIndex: intOrZero(orderIndex),
		WorkspaceID: workspaceID, IsDefault: isDefault, Color: valueOrEmpty(color),
		CreatedAt: timeOrZero(createdAt), UpdatedAt: timeOrZero(updatedAt),
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intOrZero(value *int32) int {
	if value == nil {
		return 0
	}
	return int(*value)
}

func timeOrZero(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func mapError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return objectivestatusdomain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}
