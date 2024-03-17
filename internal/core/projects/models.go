package projects

import (
	"time"

	"github.com/google/uuid"
)

type CoreProject struct {
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
