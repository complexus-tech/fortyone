package objectivestatusrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/google/uuid"
)

type dbObjectiveStatus struct {
	ID          uuid.UUID `db:"status_id"`
	Name        string    `db:"name"`
	Color       string    `db:"color"`
	Category    string    `db:"category"`
	OrderIndex  int       `db:"order_index"`
	TeamID      uuid.UUID `db:"team_id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toCoreObjectiveStatus(s dbObjectiveStatus) objectivestatus.CoreObjectiveStatus {
	return objectivestatus.CoreObjectiveStatus{
		ID:         s.ID,
		Name:       s.Name,
		Color:      s.Color,
		Category:   s.Category,
		OrderIndex: s.OrderIndex,
		Team:       s.TeamID,
		Workspace:  s.WorkspaceID,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func toCoreObjectiveStatuses(statuses []dbObjectiveStatus) []objectivestatus.CoreObjectiveStatus {
	result := make([]objectivestatus.CoreObjectiveStatus, len(statuses))
	for i, s := range statuses {
		result[i] = toCoreObjectiveStatus(s)
	}
	return result
}
