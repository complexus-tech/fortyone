package statesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/google/uuid"
)

// AppStateList represents a state in the application layer.
type AppStatesList struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Color      string     `json:"color"`
	Category   string     `json:"category" validate:"oneof=backlog unstarted started paused completed cancelled"`
	OrderIndex int        `json:"orderIndex"`
	Team       uuid.UUID  `json:"teamId"`
	Workspace  uuid.UUID  `json:"workspaceId"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt"`
}

// toAppStates converts a list of core states to a list of application states.
func toAppStates(states []states.CoreState) []AppStatesList {
	appStates := make([]AppStatesList, len(states))
	for i, state := range states {
		appStates[i] = AppStatesList{
			ID:         state.ID,
			Name:       state.Name,
			Color:      state.Color,
			Category:   state.Category,
			OrderIndex: state.OrderIndex,
			Team:       state.Team,
			Workspace:  state.Workspace,
			CreatedAt:  state.CreatedAt,
			UpdatedAt:  state.UpdatedAt,
		}
	}
	return appStates
}
