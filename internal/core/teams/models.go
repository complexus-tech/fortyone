package teams

import (
	"time"

	"github.com/google/uuid"
)

type CoreTeam struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Code        string
	Color       string
	Icon        string
	Workspace   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
