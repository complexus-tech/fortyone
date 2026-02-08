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
	ID             uuid.UUID
	Name           string
	Code           string
	Color          string
	IsPrivate      bool
	Workspace      uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	MemberCount    int
	SprintsEnabled bool
}

type DefaultStatus struct {
	Name       string
	Category   string
	OrderIndex int
	Color      string
}

var DefaultStoryStatuses = []DefaultStatus{
	{Name: "Backlog", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
	{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
	{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
	{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
	{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
	{Name: "Cancelled", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
}
