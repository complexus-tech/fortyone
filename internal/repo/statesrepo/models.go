package statesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/google/uuid"
)

type dbState struct {
	ID         uuid.UUID  `db:"status_id"`
	Name       string     `db:"name"`
	Category   string     `db:"category"`
	OrderIndex int        `db:"order_index"`
	Team       uuid.UUID  `db:"team_id"`
	Workspace  uuid.UUID  `db:"workspace_id"`
	IsDefault  bool       `db:"is_default"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

// group can be backlog, not started, in progress, done, closed
func toCoreState(p dbState) states.CoreState {
	return states.CoreState{
		ID:         p.ID,
		Name:       p.Name,
		Category:   p.Category,
		OrderIndex: p.OrderIndex,
		Team:       p.Team,
		Workspace:  p.Workspace,
		IsDefault:  p.IsDefault,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func toCoreStates(do []dbState) []states.CoreState {
	states := make([]states.CoreState, len(do))
	for i, o := range do {
		states[i] = toCoreState(o)
	}
	return states
}
