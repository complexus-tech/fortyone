package sprints

import (
	"time"

	"github.com/google/uuid"
)

type CoreSprint struct {
	ID          uuid.UUID
	Name        string
	Description string
	Owner       *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
