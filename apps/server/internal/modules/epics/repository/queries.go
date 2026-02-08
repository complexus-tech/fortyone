package epicsrepository

import (
	"context"
	"time"

	epics "github.com/complexus-tech/projects-api/internal/modules/epics/service"
	"github.com/google/uuid"
)

func (r *repo) List(ctx context.Context) ([]epics.CoreEpic, error) {

	p := []dbEpic{
		{ID: uuid.New(), Name: "Epic 1", Description: "This is epic 1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Epic 2", Description: "This is epic 2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Name: "Epic 3", Description: "This is epic 3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	return toCoreEpics(p), nil
}
