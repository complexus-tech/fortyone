package workflows

import (
	"time"

	"github.com/google/uuid"
)

type CoreWorkflow struct {
	ID        uuid.UUID
	Name      string
	Team      uuid.UUID
	Workspace uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CoreNewWorkflow struct {
	Name string
	Team uuid.UUID
}

type CoreUpdateWorkflow struct {
	Name *string
}
