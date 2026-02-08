package stateshttp

import (
	"time"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	"github.com/google/uuid"
)

// AppStatesList represents a state in the application layer.
type AppStatesList struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Category   string     `json:"category" validate:"oneof=backlog unstarted started paused completed cancelled"`
	OrderIndex int        `json:"orderIndex"`
	Team       uuid.UUID  `json:"teamId"`
	Workspace  uuid.UUID  `json:"workspaceId"`
	IsDefault  bool       `json:"isDefault"`
	Color      string     `json:"color"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
}

type NewState struct {
	Name      string    `json:"name" validate:"required"`
	Category  string    `json:"category" validate:"required,oneof=backlog unstarted started paused completed cancelled"`
	Team      uuid.UUID `json:"teamId" validate:"required"`
	IsDefault bool      `json:"isDefault"`
	Color     string    `json:"color" validate:"required"`
}

type UpdateState struct {
	Name       *string `json:"name,omitempty"`
	OrderIndex *int    `json:"orderIndex,omitempty"`
	IsDefault  *bool   `json:"isDefault,omitempty"`
	Color      *string `json:"color,omitempty"`
}

func toAppState(s states.CoreState) AppStatesList {
	return AppStatesList{
		ID:         s.ID,
		Name:       s.Name,
		Category:   s.Category,
		OrderIndex: s.OrderIndex,
		Team:       s.Team,
		Workspace:  s.Workspace,
		IsDefault:  s.IsDefault,
		Color:      s.Color,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func toAppStates(ss []states.CoreState) []AppStatesList {
	states := make([]AppStatesList, len(ss))
	for i, s := range ss {
		states[i] = toAppState(s)
	}
	return states
}

func toCoreNewState(ns NewState) states.CoreNewState {
	return states.CoreNewState{
		Name:      ns.Name,
		Category:  ns.Category,
		Team:      ns.Team,
		IsDefault: ns.IsDefault,
		Color:     ns.Color,
	}
}

func toCoreUpdateState(us UpdateState) states.CoreUpdateState {
	return states.CoreUpdateState{
		Name:       us.Name,
		OrderIndex: us.OrderIndex,
		IsDefault:  us.IsDefault,
		Color:      us.Color,
	}
}
