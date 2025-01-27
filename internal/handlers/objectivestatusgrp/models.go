package objectivestatusgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/google/uuid"
)

// AppObjectiveStatusList represents an objective status in the application layer.
type AppObjectiveStatusList struct {
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

// toAppObjectiveStatuses converts a list of core objective statuses to a list of application objective statuses.
func toAppObjectiveStatuses(states []objectivestatus.CoreObjectiveStatus) []AppObjectiveStatusList {
	appStates := make([]AppObjectiveStatusList, len(states))
	for i, state := range states {
		appStates[i] = AppObjectiveStatusList{
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
