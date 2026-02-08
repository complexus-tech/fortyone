package epicsrepository

import (
	"time"

	epics "github.com/complexus-tech/projects-api/internal/modules/epics/service"
	"github.com/google/uuid"
)

type dbEpic struct {
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

func toCoreEpic(p dbEpic) epics.CoreEpic {
	return epics.CoreEpic{
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

func toCoreEpics(do []dbEpic) []epics.CoreEpic {
	epics := make([]epics.CoreEpic, len(do))
	for i, o := range do {
		epics[i] = toCoreEpic(o)
	}
	return epics
}
