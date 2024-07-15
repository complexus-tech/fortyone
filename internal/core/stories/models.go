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
	Objective  *uuid.UUID
	Status     *uuid.UUID
	Assignee   *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Priority   *string
	Sprint     *uuid.UUID
}

// CoreSingleStory represents a single story.
type CoreSingleStory struct {
	ID              uuid.UUID
	SequenceID      int
	Title           string
	Description     *string
	DescriptionHTML *string
	Parent          *uuid.UUID
	Objective       *uuid.UUID
	Status          *uuid.UUID
	Assignee        *uuid.UUID
	BlockedBy       *uuid.UUID
	Blocking        *uuid.UUID
	Related         *uuid.UUID
	Reporter        *uuid.UUID
	Priority        *string
	Sprint          *uuid.UUID
	Team            *uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type CoreNewStory struct {
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	DescriptionHTML *string    `json:"descriptionHTML"`
	Parent          *uuid.UUID `json:"parentId"`
	Objective       *uuid.UUID `json:"objectiveId"`
	Status          *uuid.UUID `json:"statusId"`
	Assignee        *uuid.UUID `json:"assigneeId"`
	BlockedBy       *uuid.UUID `json:"blockedById"`
	Blocking        *uuid.UUID `json:"blockingId"`
	Related         *uuid.UUID `json:"relatedId"`
	Reporter        *uuid.UUID `json:"reporterId"`
	Priority        *string    `json:"priority"`
	Sprint          *uuid.UUID `json:"sprintId"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
	Team            *uuid.UUID `json:"teamId"`
}

func toCoreSingleStory(ns CoreNewStory) CoreSingleStory {
	now := time.Now()
	return CoreSingleStory{
		ID:              uuid.New(),
		SequenceID:      0,
		Title:           ns.Title,
		Description:     ns.Description,
		DescriptionHTML: ns.DescriptionHTML,
		Parent:          ns.Parent,
		Objective:       ns.Objective,
		Status:          ns.Status,
		Assignee:        ns.Assignee,
		BlockedBy:       ns.BlockedBy,
		Blocking:        ns.Blocking,
		Related:         ns.Related,
		Reporter:        ns.Reporter,
		Priority:        ns.Priority,
		Sprint:          ns.Sprint,
		StartDate:       ns.StartDate,
		EndDate:         ns.EndDate,
		Team:            ns.Team,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
