package teamsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/google/uuid"
)

// AppTeamList represents a team in the application layer.
type AppTeamsList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"title"`
	Description string    `json:"description"`
}

// toAppTeams converts a list of core teams to a list of application teams.
func toAppTeams(teams []teams.CoreTeam) []AppTeamsList {
	appTeams := make([]AppTeamsList, len(teams))
	for i, team := range teams {
		appTeams[i] = AppTeamsList{
			ID:          team.ID,
			Name:        team.Name,
			Description: team.Description,
		}
	}
	return appTeams
}
