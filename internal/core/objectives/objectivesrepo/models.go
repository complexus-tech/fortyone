package objectivesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/google/uuid"
)

type dbObjective struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Owner       *uuid.UUID `db:"owner"`
	StartDate   *time.Time `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
	IsPublic    bool       `db:"is_public"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

func toCoreObjective(p dbObjective) objectives.CoreObjective {
	return objectives.CoreObjective{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Owner:       p.Owner,
		StartDate:   p.StartDate,
		EndDate:     p.EndDate,
		IsPublic:    p.IsPublic,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toCoreObjectives(do []dbObjective) []objectives.CoreObjective {
	objectives := make([]objectives.CoreObjective, len(do))
	for i, o := range do {
		objectives[i] = toCoreObjective(o)
	}
	return objectives
}
