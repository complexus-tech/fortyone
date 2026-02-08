package objectivestatusrepository

import (
	"time"

	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	"github.com/google/uuid"
)

type dbObjectiveStatus struct {
	ID         uuid.UUID  `db:"status_id"`
	Name       string     `db:"name"`
	Category   string     `db:"category"`
	OrderIndex int        `db:"order_index"`
	Workspace  uuid.UUID  `db:"workspace_id"`
	IsDefault  bool       `db:"is_default"`
	Color      string     `db:"color"`
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
		Workspace:  s.Workspace,
		IsDefault:  s.IsDefault,
		Color:      s.Color,
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
