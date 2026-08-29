package domain

import (
	"time"

	"github.com/google/uuid"
)

// OverdueGuidanceCursor identifies the last recipient processed by the story
// guidance worker. Recipients are ordered by assignee and workspace.
type OverdueGuidanceCursor struct {
	AssigneeID  uuid.UUID
	WorkspaceID uuid.UUID
}

// OverdueGuidanceRecipient is the persistence-neutral recipient context needed
// to load overdue stories for one workspace.
type OverdueGuidanceRecipient struct {
	AssigneeID    uuid.UUID
	AssigneeEmail string
	AssigneeName  string
	WorkspaceID   uuid.UUID
	WorkspaceName string
	WorkspaceSlug string
	EmailEnabled  bool
}

// OverdueGuidanceStory is an assigned story deadline signal sent to an
// eligible workspace member.
type OverdueGuidanceStory struct {
	ID             uuid.UUID
	Title          string
	EndDate        time.Time
	AssigneeID     uuid.UUID
	AssigneeEmail  string
	AssigneeName   string
	WorkspaceID    uuid.UUID
	WorkspaceName  string
	WorkspaceSlug  string
	TeamID         uuid.UUID
	TeamName       string
	TeamCode       string
	SequenceID     int
	StatusName     string
	StatusCategory string
	DeadlineStatus string
	DaysDifference int
	EmailEnabled   bool
}
