package teamsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/google/uuid"
)

type dbTeam struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Owner       *uuid.UUID `db:"owner"`
	StartDate   *time.Time `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

func toCoreTeam(p dbTeam) teams.CoreTeam {
	return teams.CoreTeam{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Owner:       p.Owner,
		StartDate:   p.StartDate,
		EndDate:     p.EndDate,
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
