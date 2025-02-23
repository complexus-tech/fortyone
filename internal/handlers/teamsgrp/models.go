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
	Code        string    `json:"code"`
	Color       string    `json:"color"`
	IsPrivate   bool      `json:"isPrivate"`
	Workspace   uuid.UUID `json:"workspaceId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	MemberCount int       `json:"memberCount"`
}

// toAppTeams converts a list of core teams to a list of application teams.
func toAppTeams(teams []teams.CoreTeam) []AppTeamsList {
	appTeams := make([]AppTeamsList, len(teams))
	for i, team := range teams {
		appTeams[i] = AppTeamsList{
			ID:          team.ID,
			Name:        team.Name,
			Code:        team.Code,
			Color:       team.Color,
			IsPrivate:   team.IsPrivate,
			Workspace:   team.Workspace,
			CreatedAt:   team.CreatedAt,
			UpdatedAt:   team.UpdatedAt,
			MemberCount: team.MemberCount,
		}
	}
	return appTeams
}

type AppNewTeam struct {
	Name      string `json:"name" validate:"required"`
	Code      string `json:"code" validate:"required"`
	Color     string `json:"color" validate:"required"`
	IsPrivate bool   `json:"isPrivate"`
}

type AppUpdateTeam struct {
	Name      string `json:"name,omitempty"`
	Code      string `json:"code,omitempty"`
	Color     string `json:"color,omitempty"`
	IsPrivate *bool  `json:"isPrivate,omitempty"`
}

type AppNewTeamMember struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=member guest admin"`
}

type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Color     string    `json:"color"`
	IsPrivate bool      `json:"isPrivate"`
	Workspace uuid.UUID `json:"workspace_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toTeam(team teams.CoreTeam) Team {
	return Team{
		ID:        team.ID,
		Name:      team.Name,
		Code:      team.Code,
		Color:     team.Color,
		IsPrivate: team.IsPrivate,
		Workspace: team.Workspace,
		CreatedAt: team.CreatedAt,
		UpdatedAt: team.UpdatedAt,
	}
}

type CreateTeamRequest struct {
	Name  string `json:"name" validate:"required"`
	Code  string `json:"code" validate:"required"`
	Color string `json:"color" validate:"required"`
}

type UpdateTeamRequest struct {
	Name  string `json:"name,omitempty"`
	Code  string `json:"code,omitempty"`
	Color string `json:"color,omitempty"`
}
