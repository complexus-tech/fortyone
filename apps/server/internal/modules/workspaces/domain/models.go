package workspacedomain

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID
	Slug        string
	Name        string
	Color       string
	TeamSize    string
	AvatarURL   *string
	IsActive    bool
	UserRole    string
	CreatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TrialEndsOn *time.Time
	DeletedAt   *time.Time
	DeletedBy   *uuid.UUID
}

type CurrentMembership struct {
	WorkspaceID uuid.UUID
	Name        string
	Slug        string
	Role        string
}

type DefaultStatus struct {
	Name       string
	Category   string
	OrderIndex int
	Color      string
}

var DefaultObjectiveStatuses = []DefaultStatus{
	{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
	{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
	{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
	{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
}

type WorkspaceSettings struct {
	WorkspaceID        uuid.UUID
	StoryTerm          string
	SprintTerm         string
	ObjectiveTerm      string
	KeyResultTerm      string
	ObjectiveEnabled   bool
	KeyResultEnabled   bool
	WorkingDays        []int
	WorkingStartMinute int
	WorkingEndMinute   int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
