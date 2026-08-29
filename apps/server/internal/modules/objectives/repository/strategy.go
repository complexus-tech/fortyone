package objectivesrepository

import (
	"context"
	"errors"
	"fmt"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectivessql "github.com/complexus-tech/projects-api/internal/modules/objectives/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) GetStrategyMap(
	ctx context.Context,
	query objectivesdomain.StrategyQuery,
) (objectivesdomain.StrategyMap, error) {
	if err := query.Validate(); err != nil {
		return objectivesdomain.StrategyMap{}, err
	}
	canRead, err := repository.queries.CanReadObjectiveStrategy(ctx, objectivessql.CanReadObjectiveStrategyParams{
		WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
	})
	if err != nil {
		return objectivesdomain.StrategyMap{}, fmt.Errorf("authorize strategy read: %w", err)
	}
	if !canRead {
		return objectivesdomain.StrategyMap{}, objectivesdomain.ErrForbidden
	}

	strategy := objectivessql.GetWorkspaceStrategyRow{}
	strategy, err = repository.queries.GetWorkspaceStrategy(ctx, objectivessql.GetWorkspaceStrategyParams{
		WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return objectivesdomain.StrategyMap{}, fmt.Errorf("get workspace strategy: %w", err)
	}
	pillars, err := repository.queries.ListStrategicPillars(ctx, objectivessql.ListStrategicPillarsParams{
		WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
	})
	if err != nil {
		return objectivesdomain.StrategyMap{}, fmt.Errorf("list strategic pillars: %w", err)
	}
	alignments, err := repository.queries.ListVisibleStrategyAlignments(ctx, objectivessql.ListVisibleStrategyAlignmentsParams{
		ActorID: query.ActorID, WorkspaceID: query.WorkspaceID,
	})
	if err != nil {
		return objectivesdomain.StrategyMap{}, fmt.Errorf("list strategy alignments: %w", err)
	}

	objectiveIDs := make(map[uuid.UUID][]uuid.UUID, len(pillars))
	for _, pillar := range pillars {
		objectiveIDs[pillar.PillarID] = []uuid.UUID{}
	}
	for _, alignment := range alignments {
		objectiveIDs[alignment.PillarID] = append(objectiveIDs[alignment.PillarID], alignment.ObjectiveID)
	}
	result := objectivesdomain.StrategyMap{
		UltimateGoal: strategy.UltimateGoal, Description: strategy.Description,
		Pillars: make([]objectivesdomain.StrategicPillar, 0, len(pillars)),
	}
	for _, pillar := range pillars {
		result.Pillars = append(result.Pillars, objectivesdomain.StrategicPillar{
			ID: pillar.PillarID, Name: pillar.Name, Description: pillar.Description,
			OrderIndex: int(pillar.OrderIndex), ObjectiveIDs: objectiveIDs[pillar.PillarID],
		})
	}
	return result, nil
}

func (repository *Repository) UpdateStrategy(
	ctx context.Context,
	query objectivesdomain.StrategyQuery,
	strategy objectivesdomain.StrategyUpdate,
) error {
	if err := query.Validate(); err != nil {
		return err
	}
	if err := strategy.Validate(); err != nil {
		return err
	}
	_, err := repository.queries.UpdateWorkspaceStrategy(ctx, objectivessql.UpdateWorkspaceStrategyParams{
		WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
		UltimateGoal: strategy.UltimateGoal, Description: strategy.Description,
	})
	return mapStrategyError("update workspace strategy", err)
}

func (repository *Repository) CreateStrategicPillar(
	ctx context.Context,
	query objectivesdomain.StrategyQuery,
	pillar objectivesdomain.NewStrategicPillar,
) (objectivesdomain.StrategicPillar, error) {
	if err := query.Validate(); err != nil {
		return objectivesdomain.StrategicPillar{}, err
	}
	if err := pillar.Validate(); err != nil {
		return objectivesdomain.StrategicPillar{}, err
	}
	orderIndex, err := safecast.Int32(pillar.OrderIndex)
	if err != nil {
		return objectivesdomain.StrategicPillar{}, fmt.Errorf("%w: pillar order: %v", objectivesdomain.ErrInvalid, err)
	}
	row, err := repository.queries.CreateStrategicPillar(ctx, objectivessql.CreateStrategicPillarParams{
		WorkspaceID: query.WorkspaceID, ActorID: query.ActorID, Name: pillar.Name,
		Description: pillar.Description, OrderIndex: orderIndex,
	})
	if err != nil {
		return objectivesdomain.StrategicPillar{}, mapStrategyError("create strategic pillar", err)
	}
	return strategicPillarFromValues(row.PillarID, row.Name, row.Description, row.OrderIndex), nil
}

func (repository *Repository) UpdateStrategicPillar(
	ctx context.Context,
	query objectivesdomain.StrategyQuery,
	pillarID uuid.UUID,
	patch objectivesdomain.UpdateStrategicPillar,
) (objectivesdomain.StrategicPillar, error) {
	if err := query.Validate(); err != nil {
		return objectivesdomain.StrategicPillar{}, err
	}
	if pillarID == uuid.Nil {
		return objectivesdomain.StrategicPillar{}, fmt.Errorf("%w: pillar is required", objectivesdomain.ErrInvalid)
	}
	if err := patch.Validate(); err != nil {
		return objectivesdomain.StrategicPillar{}, err
	}
	params := objectivessql.UpdateStrategicPillarParams{
		WorkspaceID: query.WorkspaceID, ActorID: query.ActorID, PillarID: pillarID,
	}
	if value, specified := patch.Name.Value(); specified {
		params.SetName, params.Name = true, valueOrZero(value)
	}
	if value, specified := patch.Description.Value(); specified {
		params.SetDescription, params.Description = true, value
	}
	if value, specified := patch.OrderIndex.Value(); specified {
		orderIndex, err := safecast.Int32(valueOrZero(value))
		if err != nil {
			return objectivesdomain.StrategicPillar{}, fmt.Errorf("%w: pillar order: %v", objectivesdomain.ErrInvalid, err)
		}
		params.SetOrderIndex, params.OrderIndex = true, orderIndex
	}
	row, err := repository.queries.UpdateStrategicPillar(ctx, params)
	if err != nil {
		return objectivesdomain.StrategicPillar{}, mapStrategyError("update strategic pillar", err)
	}
	return strategicPillarFromValues(row.PillarID, row.Name, row.Description, row.OrderIndex), nil
}

func (repository *Repository) DeleteStrategicPillar(
	ctx context.Context,
	query objectivesdomain.StrategyQuery,
	pillarID uuid.UUID,
) error {
	if err := query.Validate(); err != nil {
		return err
	}
	if pillarID == uuid.Nil {
		return fmt.Errorf("%w: pillar is required", objectivesdomain.ErrInvalid)
	}
	_, err := repository.queries.DeleteStrategicPillar(ctx, objectivessql.DeleteStrategicPillarParams{
		WorkspaceID: query.WorkspaceID, ActorID: query.ActorID, PillarID: pillarID,
	})
	return mapStrategyError("delete strategic pillar", err)
}

func (repository *Repository) AlignObjective(
	ctx context.Context,
	query objectivesdomain.StrategyQuery,
	objectiveID uuid.UUID,
	pillarID *uuid.UUID,
) error {
	if err := query.Validate(); err != nil {
		return err
	}
	if objectiveID == uuid.Nil || (pillarID != nil && *pillarID == uuid.Nil) {
		return fmt.Errorf("%w: objective and pillar references must be valid", objectivesdomain.ErrInvalid)
	}
	var err error
	if pillarID == nil {
		_, err = repository.queries.DeleteObjectiveAlignment(ctx, objectivessql.DeleteObjectiveAlignmentParams{
			WorkspaceID: uuidPointer(query.WorkspaceID), ActorID: query.ActorID, ObjectiveID: objectiveID,
		})
	} else {
		_, err = repository.queries.AlignObjective(ctx, objectivessql.AlignObjectiveParams{
			WorkspaceID: query.WorkspaceID, ActorID: query.ActorID,
			ObjectiveID: objectiveID, PillarID: *pillarID,
		})
	}
	return mapStrategyError("align objective", err)
}

func strategicPillarFromValues(id uuid.UUID, name string, description *string, orderIndex int32) objectivesdomain.StrategicPillar {
	return objectivesdomain.StrategicPillar{
		ID: id, Name: name, Description: description, OrderIndex: int(orderIndex), ObjectiveIDs: []uuid.UUID{},
	}
}

func mapStrategyError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return objectivesdomain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: %w", operation, mapDatabaseError(err))
	}
	return nil
}
