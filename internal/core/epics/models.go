package epics

import (
	"time"

	"github.com/google/uuid"
)

type CoreEpic struct {
	ID          uuid.UUID
	Name        string
	Description string
	Owner       *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
