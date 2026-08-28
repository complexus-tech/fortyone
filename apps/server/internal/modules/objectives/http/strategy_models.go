package objectiveshttp

import (
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/google/uuid"
)

type AppStrategyMap struct {
	UltimateGoal string               `json:"ultimateGoal"`
	Description  *string              `json:"description"`
	Pillars      []AppStrategicPillar `json:"pillars"`
}

type AppStrategicPillar struct {
	ID           uuid.UUID   `json:"id"`
	Name         string      `json:"name"`
	Description  *string     `json:"description"`
	OrderIndex   int         `json:"orderIndex"`
	ObjectiveIDs []uuid.UUID `json:"objectiveIds"`
}

type AppStrategyUpdate struct {
	UltimateGoal string  `json:"ultimateGoal" validate:"required"`
	Description  *string `json:"description"`
}

type AppNewStrategicPillar struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	OrderIndex  int     `json:"orderIndex"`
}

type AppUpdateStrategicPillar struct {
	Name        PatchField[string] `json:"name"`
	Description PatchField[string] `json:"description"`
	OrderIndex  PatchField[int]    `json:"orderIndex"`
}

func (request AppUpdateStrategicPillar) Validate() error {
	return request.Patch().Validate()
}

func (request AppUpdateStrategicPillar) Patch() objectivesdomain.UpdateStrategicPillar {
	return objectivesdomain.UpdateStrategicPillar{
		Name: patchField(request.Name), Description: patchField(request.Description),
		OrderIndex: patchField(request.OrderIndex),
	}
}

type AppObjectiveAlignment struct {
	PillarID *uuid.UUID `json:"pillarId"`
}

func toAppStrategyMap(strategy objectives.CoreStrategyMap) AppStrategyMap {
	pillars := make([]AppStrategicPillar, 0, len(strategy.Pillars))
	for _, pillar := range strategy.Pillars {
		pillars = append(pillars, toAppStrategicPillar(pillar))
	}
	return AppStrategyMap{UltimateGoal: strategy.UltimateGoal, Description: strategy.Description, Pillars: pillars}
}

func toAppStrategicPillar(pillar objectives.CoreStrategicPillar) AppStrategicPillar {
	return AppStrategicPillar{
		ID: pillar.ID, Name: pillar.Name, Description: pillar.Description,
		OrderIndex: pillar.OrderIndex, ObjectiveIDs: nonNilObjectiveIDs(pillar.ObjectiveIDs),
	}
}

func nonNilObjectiveIDs(values []uuid.UUID) []uuid.UUID {
	if values == nil {
		return []uuid.UUID{}
	}
	return values
}
