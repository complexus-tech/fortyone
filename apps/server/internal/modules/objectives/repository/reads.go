package objectivesrepository

import (
	"context"
	"fmt"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) List(
	ctx context.Context,
	query objectivesdomain.ListQuery,
) ([]objectivesdomain.Objective, error) {
	normalized, err := query.Normalize()
	if err != nil {
		return nil, err
	}
	resultLimit, err := safecast.Int32(normalized.Limit)
	if err != nil {
		return nil, fmt.Errorf("%w: objective result limit: %v", objectivesdomain.ErrInvalid, err)
	}
	resultOffset, err := safecast.Int32(normalized.Offset)
	if err != nil {
		return nil, fmt.Errorf("%w: objective result offset: %v", objectivesdomain.ErrInvalid, err)
	}
	rows, err := repository.queries.ListObjectives(ctx, objectivessql.ListObjectivesParams{
		WorkspaceID: uuidPointer(normalized.WorkspaceID),
		ActorID:     normalized.ActorID, ObjectiveID: normalized.ObjectiveID, TeamID: normalized.TeamID,
		Search: normalized.Search, ResultLimit: resultLimit, ResultOffset: resultOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("list objectives: %w", mapDatabaseError(err))
	}
	objectives := make([]objectivesdomain.Objective, 0, len(rows))
	for _, row := range rows {
		objectives = append(objectives, objectiveFromListRow(row))
	}
	return objectives, nil
}

func (repository *Repository) Get(
	ctx context.Context,
	query objectivesdomain.GetQuery,
) (objectivesdomain.Objective, error) {
	if err := query.Validate(); err != nil {
		return objectivesdomain.Objective{}, err
	}
	row, err := repository.queries.GetObjective(ctx, objectivessql.GetObjectiveParams{
		ObjectiveID: query.ObjectiveID, WorkspaceID: uuidPointer(query.WorkspaceID),
		InternalAccess: query.Internal, ActorID: query.ActorID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return objectivesdomain.Objective{}, objectivesdomain.ErrNotFound
		}
		return objectivesdomain.Objective{}, fmt.Errorf("get objective: %w", mapDatabaseError(err))
	}
	return objectiveFromGetRow(row), nil
}
