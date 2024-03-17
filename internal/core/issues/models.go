package issues

import (
	"time"

	"github.com/google/uuid"
)

// CoreIssueList represents a list of issues.
type CoreIssueList struct {
	ID         uuid.UUID
	SequenceID int
	Title      string
	Parent     *uuid.UUID
	Project    *uuid.UUID
	Status     *uuid.UUID
	Assignee   *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Priority   *uuid.UUID
	Sprint     *uuid.UUID
}

// CoreSingleIssue represents a single issue.
type CoreSingleIssue struct {
	ID              uuid.UUID
	SequenceID      int
	Title           string
	Description     string
	DescriptionHTML string
	Parent          *uuid.UUID
	Project         *uuid.UUID
	Status          *uuid.UUID
	Assignee        *uuid.UUID
	BlockedBy       *uuid.UUID
	Blocking        *uuid.UUID
	Related         *uuid.UUID
	Reporter        *uuid.UUID
	Type            *string
	Priority        *uuid.UUID
	Attachments     *string
	Sprint          *uuid.UUID
	Team            *uuid.UUID
	Watchers        *string
	Labels          *string
	Comments        *string
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}
