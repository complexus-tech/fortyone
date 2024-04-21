package objectives

import (
	"time"

	"github.com/google/uuid"
)

type CoreObjective struct {
	ID          uuid.UUID
	Name        string
	Description string
	Owner       *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
