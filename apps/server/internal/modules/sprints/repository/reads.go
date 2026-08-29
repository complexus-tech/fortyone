package sprintsrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	sprintssql "github.com/complexus-tech/projects-api/internal/modules/sprints/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) List(ctx context.Context, query sprintdomain.ListQuery) ([]sprintdomain.Sprint, error) {
	query, err := query.Normalize()
	if err != nil {
		return nil, err
	}
	limit, err := safecast.Int32(query.Filter.Limit)
	if err != nil {
		return nil, fmt.Errorf("map sprint list limit: %w", err)
	}
	offset, err := safecast.Int32(query.Filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("map sprint list offset: %w", err)
	}
	var search *string
	if query.Filter.Search != "" {
		search = &query.Filter.Search
	}
	rows, err := repository.queries.ListSprints(ctx, sprintssql.ListSprintsParams{
		ActorID: query.ActorID, WorkspaceID: query.WorkspaceID,
		SprintID: query.Filter.SprintID, ObjectiveID: query.Filter.ObjectiveID,
		TeamID: query.Filter.TeamID, Search: search, RowLimit: limit, RowOffset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}
	return mapListRows(rows)
}

func (repository *Repository) Running(
	ctx context.Context,
	workspaceID, actorID uuid.UUID,
	today time.Time,
) ([]sprintdomain.Sprint, error) {
	rows, err := repository.queries.ListRunningSprints(ctx, sprintssql.ListRunningSprintsParams{
		ActorID: actorID, WorkspaceID: workspaceID, Today: today,
	})
	if err != nil {
		return nil, fmt.Errorf("list running sprints: %w", err)
	}
	return mapRunningRows(rows)
}

func (repository *Repository) GetByID(
	ctx context.Context,
	sprintID, workspaceID, actorID uuid.UUID,
) (sprintdomain.Sprint, error) {
	row, err := repository.queries.GetSprintByID(ctx, sprintssql.GetSprintByIDParams{
		SprintID: sprintID, WorkspaceID: workspaceID, ActorID: actorID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return sprintdomain.Sprint{}, sprintdomain.ErrNotFound
	}
	if err != nil {
		return sprintdomain.Sprint{}, fmt.Errorf("get sprint: %w", err)
	}
	return sprintFromGetRow(row)
}
