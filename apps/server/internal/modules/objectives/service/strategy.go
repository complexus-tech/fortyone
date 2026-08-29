package objectives

import (
	"context"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) strategyQuery(ctx context.Context, workspaceID uuid.UUID, scope platformauth.Scope) (objectivesdomain.StrategyQuery, error) {
	actor, err := actorFor(ctx, workspaceID, uuid.Nil, scope)
	if err != nil {
		return objectivesdomain.StrategyQuery{}, err
	}
	return objectivesdomain.StrategyQuery{WorkspaceID: workspaceID, ActorID: actor.PrincipalID}, nil
}

func (service *Service) GetStrategyMap(ctx context.Context, workspaceID uuid.UUID) (CoreStrategyMap, error) {
	query, err := service.strategyQuery(ctx, workspaceID, platformauth.ScopeObjectivesRead)
	if err != nil {
		return CoreStrategyMap{}, err
	}
	return service.repo.GetStrategyMap(ctx, query)
}

func (service *Service) UpdateStrategy(ctx context.Context, workspaceID uuid.UUID, strategy CoreStrategyUpdate) error {
	if err := strategy.Validate(); err != nil {
		return err
	}
	query, err := service.strategyQuery(ctx, workspaceID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return err
	}
	return service.repo.UpdateStrategy(ctx, query, strategy)
}

func (service *Service) CreateStrategicPillar(
	ctx context.Context,
	workspaceID uuid.UUID,
	pillar CoreNewStrategicPillar,
) (CoreStrategicPillar, error) {
	if err := pillar.Validate(); err != nil {
		return CoreStrategicPillar{}, err
	}
	query, err := service.strategyQuery(ctx, workspaceID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return CoreStrategicPillar{}, err
	}
	return service.repo.CreateStrategicPillar(ctx, query, pillar)
}

func (service *Service) UpdateStrategicPillar(
	ctx context.Context,
	workspaceID, pillarID uuid.UUID,
	pillar CoreUpdateStrategicPillar,
) (CoreStrategicPillar, error) {
	patch := objectivesdomain.UpdateStrategicPillar{}
	if pillar.Name != nil {
		patch.Name = objectivesdomain.SetField(*pillar.Name)
	}
	if pillar.Description != nil {
		patch.Description = objectivesdomain.SetField(*pillar.Description)
	}
	if pillar.OrderIndex != nil {
		patch.OrderIndex = objectivesdomain.SetField(*pillar.OrderIndex)
	}
	return service.UpdateStrategicPillarIntent(ctx, workspaceID, pillarID, patch)
}

func (service *Service) UpdateStrategicPillarIntent(
	ctx context.Context,
	workspaceID, pillarID uuid.UUID,
	patch objectivesdomain.UpdateStrategicPillar,
) (CoreStrategicPillar, error) {
	if err := patch.Validate(); err != nil {
		return CoreStrategicPillar{}, err
	}
	query, err := service.strategyQuery(ctx, workspaceID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return CoreStrategicPillar{}, err
	}
	return service.repo.UpdateStrategicPillar(ctx, query, pillarID, patch)
}

func (service *Service) DeleteStrategicPillar(ctx context.Context, workspaceID, pillarID uuid.UUID) error {
	query, err := service.strategyQuery(ctx, workspaceID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return err
	}
	return service.repo.DeleteStrategicPillar(ctx, query, pillarID)
}

func (service *Service) AlignObjective(
	ctx context.Context,
	workspaceID, objectiveID uuid.UUID,
	pillarID *uuid.UUID,
) error {
	query, err := service.strategyQuery(ctx, workspaceID, platformauth.ScopeObjectivesWrite)
	if err != nil {
		return err
	}
	return service.repo.AlignObjective(ctx, query, objectiveID, pillarID)
}
