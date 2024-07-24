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
	Icon        string    `db:"icon"`
	Workspace   uuid.UUID `db:"workspace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toCoreTeam(p dbTeam) teams.CoreTeam {
	return teams.CoreTeam{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Code:        p.Code,
		Color:       p.Color,
		Icon:        p.Icon,
		Workspace:   p.Workspace,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toCoreTeams(do []dbTeam) []teams.CoreTeam {
	teams := make([]teams.CoreTeam, len(do))
	for i, o := range do {
		teams[i] = toCoreTeam(o)
	}
	return teams
}
