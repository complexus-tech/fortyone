package teams

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrTeamCodeExists is returned when attempting to create or update a team with a code that already exists
var ErrTeamCodeExists = errors.New("team code already exists")

// ErrTeamMemberExists is returned when attempting to add a member that is already in the team
var ErrTeamMemberExists = errors.New("user is already a member of this team")

type CoreTeam struct {
	ID          uuid.UUID
	Name        string
	Code        string
	Color       string
	IsPrivate   bool
	Workspace   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	MemberCount int
}

type DefaultStatus struct {
	Name       string
	Category   string
	OrderIndex int
	Color      string
}

var DefaultStoryStatuses = []DefaultStatus{
	{Name: "Backlog", Category: "backlog", OrderIndex: 0, Color: "#6b665c"},
	{Name: "To Do", Category: "unstarted", OrderIndex: 1, Color: "#6b665c"},
	{Name: "In Progress", Category: "started", OrderIndex: 2, Color: "#eab308"},
	{Name: "Blocked", Category: "paused", OrderIndex: 3, Color: "#6b665c"},
	{Name: "Done", Category: "completed", OrderIndex: 4, Color: "#22c55e"},
	{Name: "Cancelled", Category: "cancelled", OrderIndex: 5, Color: "#f43f5e"},
}
