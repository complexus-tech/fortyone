package teams

import (
	"time"

	"github.com/google/uuid"
)

type CoreTeam struct {
	ID          uuid.UUID
	Name        string
	Description string
	Owner       *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
