package workspaces

import (
	"time"

	"github.com/google/uuid"
)

// CoreWorkspace represents a workspace in the application layer.
type CoreWorkspace struct {
	ID        uuid.UUID
	Slug      string
	Name      string
	Color     string
	TeamSize  string
	IsActive  bool
	UserRole  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DefaultStatus defines a default status that gets created with each workspace
type DefaultStatus struct {
	Name       string
	Category   string
	OrderIndex int
}

// DefaultObjectiveStatuses defines the default statuses created for objectives in a workspace
var DefaultObjectiveStatuses = []DefaultStatus{
	{Name: "To Do", Category: "unstarted", OrderIndex: 0},
	{Name: "In Progress", Category: "started", OrderIndex: 1},
	{Name: "Blocked", Category: "paused", OrderIndex: 2},
	{Name: "Done", Category: "completed", OrderIndex: 3},
}
