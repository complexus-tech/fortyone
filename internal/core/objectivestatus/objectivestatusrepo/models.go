package objectivestatusrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/google/uuid"
)

type dbObjectiveStatus struct {
	ID         uuid.UUID  `db:"status_id"`
	Name       string     `db:"name"`
	Category   string     `db:"category"`
	OrderIndex int        `db:"order_index"`
	Team       uuid.UUID  `db:"team_id"`
	Workspace  uuid.UUID  `db:"workspace_id"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

func toCoreObjectiveStatus(s dbObjectiveStatus) objectivestatus.CoreObjectiveStatus {
	return objectivestatus.CoreObjectiveStatus{
		ID:         s.ID,
		Name:       s.Name,
		Category:   s.Category,
		OrderIndex: s.OrderIndex,
		Team:       s.Team,
		Workspace:  s.Workspace,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func toCoreObjectiveStatuses(ss []dbObjectiveStatus) []objectivestatus.CoreObjectiveStatus {
	statuses := make([]objectivestatus.CoreObjectiveStatus, len(ss))
	for i, s := range ss {
		statuses[i] = toCoreObjectiveStatus(s)
	}
	return statuses
}
