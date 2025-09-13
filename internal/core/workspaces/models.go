package workspaces

import (
	"time"

	"github.com/google/uuid"
)

// CoreWorkspace represents a workspace in the application layer.
type CoreWorkspace struct {
	ID          uuid.UUID
	Slug        string
	Name        string
	Color       string
	TeamSize    string
	AvatarURL   *string
	IsActive    bool
	UserRole    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TrialEndsOn *time.Time
}

// DefaultStatus defines a default status that gets created with each workspace
type DefaultStatus struct {
	Name       string
	Category   string
	OrderIndex int
	Color      string
}

// DefaultObjectiveStatuses defines the default statuses created for objectives in a workspace
var DefaultObjectiveStatuses = []DefaultStatus{
	{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
	{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
	{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
	{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
}

// CoreWorkspaceSettings represents workspace settings in the application layer
// including both terminology preferences and feature toggles.
type CoreWorkspaceSettings struct {
	WorkspaceID      uuid.UUID
	StoryTerm        string
	SprintTerm       string
	ObjectiveTerm    string
	KeyResultTerm    string
	ObjectiveEnabled bool
	KeyResultEnabled bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
