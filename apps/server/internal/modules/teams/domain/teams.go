package teamsdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCodeExists       = errors.New("team code already exists")
	ErrMemberExists     = errors.New("user is already a member of this team")
	ErrNotFound         = errors.New("team not found")
	ErrMemberNotFound   = errors.New("team member not found")
	ErrInvalidStoryTerm = errors.New("work naming must be story, task, issue, ticket, work item, deal, or the workspace default")
)

type Team struct {
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
	StoryTerm      *string
	// Patch flags distinguish omitted values from an explicit reset.
	StoryTermSet bool
	IsPrivateSet bool
}

func ValidStoryTerm(term *string) bool {
	if term == nil {
		return true
	}
	switch *term {
	case "story", "task", "issue", "ticket", "work item", "deal":
		return true
	default:
		return false
	}
}

type ListFilter struct {
	Search     string
	Limit      int
	Offset     int
	JoinedOnly bool
}

type MemberAIContext struct {
	RoleTitle       string
	RoleDescription string
}

type PublicTeamJoin struct {
	TeamID      uuid.UUID
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
}

type TeamSelfLeave struct {
	TeamID      uuid.UUID
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
}

type DefaultStatus struct {
	Name       string
	Category   string
	OrderIndex int
	Color      string
}

var DefaultStoryStatuses = [...]DefaultStatus{
	{Name: "Backlog", Category: "backlog", OrderIndex: 1000, Color: "#6b665c"},
	{Name: "To Do", Category: "unstarted", OrderIndex: 2000, Color: "#6b665c"},
	{Name: "In Progress", Category: "started", OrderIndex: 3000, Color: "#eab308"},
	{Name: "Done", Category: "completed", OrderIndex: 4000, Color: "#22c55e"},
	{Name: "Blocked", Category: "paused", OrderIndex: 5000, Color: "#6b665c"},
	{Name: "Cancelled", Category: "cancelled", OrderIndex: 6000, Color: "#f43f5e"},
}
