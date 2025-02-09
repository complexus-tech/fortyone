package teamsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/google/uuid"
)

type dbTeam struct {
	ID          uuid.UUID `db:"team_id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Code        string    `db:"code"`
	Color       string    `db:"color"`
	Workspace   uuid.UUID `db:"workspace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toCoreTeam(t dbTeam) teams.CoreTeam {
	return teams.CoreTeam{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Code:        t.Code,
		Color:       t.Color,
		Workspace:   t.Workspace,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func toCoreTeams(t []dbTeam) []teams.CoreTeam {
	result := make([]teams.CoreTeam, len(t))
	for i, team := range t {
		result[i] = toCoreTeam(team)
	}
	return result
}
