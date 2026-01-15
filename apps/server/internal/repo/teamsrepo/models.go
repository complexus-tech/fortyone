package teamsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/google/uuid"
)

type dbTeam struct {
	ID             uuid.UUID `db:"team_id"`
	Name           string    `db:"name"`
	Code           string    `db:"code"`
	Color          string    `db:"color"`
	IsPrivate      bool      `db:"is_private"`
	Workspace      uuid.UUID `db:"workspace_id"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	MemberCount    int       `db:"member_count"`
	SprintsEnabled bool      `db:"sprints_enabled"`
}

func toCoreTeam(t dbTeam) teams.CoreTeam {
	return teams.CoreTeam{
		ID:             t.ID,
		Name:           t.Name,
		Code:           t.Code,
		Color:          t.Color,
		IsPrivate:      t.IsPrivate,
		Workspace:      t.Workspace,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		MemberCount:    t.MemberCount,
		SprintsEnabled: t.SprintsEnabled,
	}
}

func toCoreTeams(t []dbTeam) []teams.CoreTeam {
	result := make([]teams.CoreTeam, len(t))
	for i, team := range t {
		result[i] = toCoreTeam(team)
	}
	return result
}
