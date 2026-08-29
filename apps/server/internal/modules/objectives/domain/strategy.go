package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type StrategyMap struct {
	UltimateGoal string
	Description  *string
	Pillars      []StrategicPillar
}

type StrategicPillar struct {
	ID           uuid.UUID
	Name         string
	Description  *string
	OrderIndex   int
	ObjectiveIDs []uuid.UUID
}

type StrategyUpdate struct {
	UltimateGoal string
	Description  *string
}

const MaximumStrategyNameLength = 255

func (strategy StrategyUpdate) Validate() error {
	if strings.TrimSpace(strategy.UltimateGoal) == "" {
		return fmt.Errorf("%w: ultimate goal is required", ErrInvalid)
	}
	if len([]rune(strategy.UltimateGoal)) > MaximumStrategyNameLength {
		return fmt.Errorf("%w: ultimate goal cannot exceed %d characters", ErrInvalid, MaximumStrategyNameLength)
	}
	return nil
}

type NewStrategicPillar struct {
	Name        string
	Description *string
	OrderIndex  int
}

func (pillar NewStrategicPillar) Validate() error {
	if strings.TrimSpace(pillar.Name) == "" {
		return fmt.Errorf("%w: pillar name is required", ErrInvalid)
	}
	if len([]rune(pillar.Name)) > MaximumStrategyNameLength {
		return fmt.Errorf("%w: pillar name cannot exceed %d characters", ErrInvalid, MaximumStrategyNameLength)
	}
	return nil
}

type UpdateStrategicPillar struct {
	Name        Field[string]
	Description Field[string]
	OrderIndex  Field[int]
}

func (patch UpdateStrategicPillar) Empty() bool {
	return !patch.Name.Specified() && !patch.Description.Specified() && !patch.OrderIndex.Specified()
}

func (patch UpdateStrategicPillar) Validate() error {
	if patch.Empty() {
		return fmt.Errorf("%w: at least one pillar field is required", ErrInvalid)
	}
	if value, specified := patch.Name.Value(); specified && (value == nil || strings.TrimSpace(*value) == "") {
		return fmt.Errorf("%w: pillar name cannot be blank or null", ErrInvalid)
	}
	if value, specified := patch.Name.Value(); specified && value != nil && len([]rune(*value)) > MaximumStrategyNameLength {
		return fmt.Errorf("%w: pillar name cannot exceed %d characters", ErrInvalid, MaximumStrategyNameLength)
	}
	if value, specified := patch.OrderIndex.Value(); specified && value == nil {
		return fmt.Errorf("%w: pillar order cannot be null", ErrInvalid)
	}
	return nil
}

type StrategyQuery struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
}

func (query StrategyQuery) Validate() error {
	if query.WorkspaceID == uuid.Nil || query.ActorID == uuid.Nil {
		return fmt.Errorf("%w: workspace and actor are required", ErrInvalid)
	}
	return nil
}
