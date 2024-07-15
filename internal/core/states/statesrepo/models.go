package statesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/google/uuid"
)

type dbState struct {
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

// group can be backlog, not started, in progress, done, closed
func toCoreState(p dbState) states.CoreState {
	return states.CoreState{
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

func toCoreStates(do []dbState) []states.CoreState {
	states := make([]states.CoreState, len(do))
	for i, o := range do {
		states[i] = toCoreState(o)
	}
	return states
}
