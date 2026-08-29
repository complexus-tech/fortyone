package teamsrepository

import (
	"time"

	teamsdomain "github.com/complexus-tech/projects-api/internal/modules/teams/domain"
	teamsql "github.com/complexus-tech/projects-api/internal/modules/teams/repository/sqlc"
	"github.com/google/uuid"
)

type teamRow struct {
	id             uuid.UUID
	name           string
	code           string
	color          string
	isPrivate      bool
	workspaceID    uuid.UUID
	createdAt      time.Time
	updatedAt      time.Time
	memberCount    int32
	sprintsEnabled bool
}

func toCoreTeam(row teamRow) teamsdomain.Team {
	return teamsdomain.Team{
		ID:             row.id,
		Name:           row.name,
		Code:           row.code,
		Color:          row.color,
		IsPrivate:      row.isPrivate,
		Workspace:      row.workspaceID,
		CreatedAt:      row.createdAt,
		UpdatedAt:      row.updatedAt,
		MemberCount:    int(row.memberCount),
		SprintsEnabled: row.sprintsEnabled,
	}
}

func toCoreListTeam(row teamsql.ListTeamsForActorRow) teamsdomain.Team {
	return toCoreTeam(teamRow{
		id:             row.TeamID,
		name:           row.Name,
		code:           row.Code,
		color:          row.Color,
		isPrivate:      row.IsPrivate,
		workspaceID:    row.WorkspaceID,
		createdAt:      row.CreatedAt,
		updatedAt:      row.UpdatedAt,
		memberCount:    row.MemberCount,
		sprintsEnabled: row.SprintsEnabled,
	})
}

func toCoreListTeams(rows []teamsql.ListTeamsForActorRow) []teamsdomain.Team {
	result := make([]teamsdomain.Team, len(rows))
	for index, row := range rows {
		result[index] = toCoreListTeam(row)
	}
	return result
}

func toCorePublicTeam(row teamsql.ListPublicTeamsForActorRow) teamsdomain.Team {
	return toCoreTeam(teamRow{
		id:             row.TeamID,
		name:           row.Name,
		code:           row.Code,
		color:          row.Color,
		isPrivate:      row.IsPrivate,
		workspaceID:    row.WorkspaceID,
		createdAt:      row.CreatedAt,
		updatedAt:      row.UpdatedAt,
		memberCount:    row.MemberCount,
		sprintsEnabled: row.SprintsEnabled,
	})
}

func toCorePublicTeams(rows []teamsql.ListPublicTeamsForActorRow) []teamsdomain.Team {
	result := make([]teamsdomain.Team, len(rows))
	for index, row := range rows {
		result[index] = toCorePublicTeam(row)
	}
	return result
}

func toCoreGetTeam(row teamsql.GetTeamForActorRow) teamsdomain.Team {
	return toCoreTeam(teamRow{
		id:             row.TeamID,
		name:           row.Name,
		code:           row.Code,
		color:          row.Color,
		isPrivate:      row.IsPrivate,
		workspaceID:    row.WorkspaceID,
		createdAt:      row.CreatedAt,
		updatedAt:      row.UpdatedAt,
		memberCount:    row.MemberCount,
		sprintsEnabled: row.SprintsEnabled,
	})
}

func toCoreCreatedTeam(row teamsql.CreateTeamRow) teamsdomain.Team {
	return toCoreTeam(teamRow{
		id:          row.TeamID,
		name:        row.Name,
		code:        row.Code,
		color:       row.Color,
		isPrivate:   row.IsPrivate,
		workspaceID: row.WorkspaceID,
		createdAt:   row.CreatedAt,
		updatedAt:   row.UpdatedAt,
		memberCount: 1,
	})
}

func toCoreUpdatedTeam(row teamsql.UpdateTeamForWorkspaceRow) teamsdomain.Team {
	return toCoreTeam(teamRow{
		id:          row.TeamID,
		name:        row.Name,
		code:        row.Code,
		color:       row.Color,
		isPrivate:   row.IsPrivate,
		workspaceID: row.WorkspaceID,
		createdAt:   row.CreatedAt,
		updatedAt:   row.UpdatedAt,
	})
}
