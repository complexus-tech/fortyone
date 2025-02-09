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
	Workspace   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DefaultStatus struct {
	Name       string
	Category   string
	OrderIndex int
}

var DefaultStoryStatuses = []DefaultStatus{
	{Name: "Backlog", Category: "backlog", OrderIndex: 0},
	{Name: "To Do", Category: "unstarted", OrderIndex: 1},
	{Name: "In Progress", Category: "started", OrderIndex: 2},
	{Name: "Blocked", Category: "paused", OrderIndex: 3},
	{Name: "Done", Category: "completed", OrderIndex: 4},
}

var DefaultObjectiveStatuses = []DefaultStatus{
	{Name: "To Do", Category: "unstarted", OrderIndex: 0},
	{Name: "In Progress", Category: "started", OrderIndex: 1},
	{Name: "Blocked", Category: "paused", OrderIndex: 2},
	{Name: "Done", Category: "completed", OrderIndex: 3},
}
