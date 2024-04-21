package stories

import (
	"time"

	"github.com/google/uuid"
)

// CoreStoryList represents a list of stories.
type CoreStoryList struct {
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

// CoreSingleStory represents a single story.
type CoreSingleStory struct {
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
