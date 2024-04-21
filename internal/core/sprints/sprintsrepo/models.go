package sprintsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/google/uuid"
)

type dbSprint struct {
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

func toCoreSprint(p dbSprint) sprints.CoreSprint {
	return sprints.CoreSprint{
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

func toCoreSprints(do []dbSprint) []sprints.CoreSprint {
	sprints := make([]sprints.CoreSprint, len(do))
	for i, o := range do {
		sprints[i] = toCoreSprint(o)
	}
	return sprints
}
