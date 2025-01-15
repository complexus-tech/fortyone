package teamsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/google/uuid"
)

// AppTeamList represents a team in the application layer.
type AppTeamsList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Code        string    `json:"code"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	Workspace   uuid.UUID `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// toAppTeams converts a list of core teams to a list of application teams.
func toAppTeams(teams []teams.CoreTeam) []AppTeamsList {
	appTeams := make([]AppTeamsList, len(teams))
	for i, team := range teams {
		appTeams[i] = AppTeamsList{
			ID:          team.ID,
			Name:        team.Name,
			Description: team.Description,
			Code:        team.Code,
			Color:       team.Color,
			Icon:        team.Icon,
			Workspace:   team.Workspace,
			CreatedAt:   team.CreatedAt,
			UpdatedAt:   team.UpdatedAt,
		}
	}
	return appTeams
}

type AppNewTeam struct {
	Name        string  `json:"name" validate:"required"`
	Description *string `json:"description"`
	Code        string  `json:"code" validate:"required"`
	Color       string  `json:"color" validate:"required"`
	Icon        string  `json:"icon" validate:"required"`
}

type AppUpdateTeam struct {
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Code        string  `json:"code,omitempty"`
	Color       string  `json:"color,omitempty"`
	Icon        string  `json:"icon,omitempty"`
}

type AppNewTeamMember struct {
	Role string `json:"role" validate:"required"`
}
