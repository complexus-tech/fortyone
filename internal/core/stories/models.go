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
	Epic       *uuid.UUID
	Status     *uuid.UUID
	Assignee   *uuid.UUID
	Reporter   *uuid.UUID
	Priority   string
	Sprint     *uuid.UUID
	Team       uuid.UUID
	Workspace  uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
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
	Priority        string
	Sprint          *uuid.UUID
	Epic            *uuid.UUID
	Team            uuid.UUID
	Workspace       uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	SubStories      []CoreStoryList
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
	Priority        string     `json:"priority"`
	Sprint          *uuid.UUID `json:"sprintId"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
	Team            uuid.UUID  `json:"teamId"`
}

type CoreUpdateStory struct {
	Title           *string
	Description     *string
	DescriptionHTML *string
	Parent          *uuid.UUID
	Objective       *uuid.UUID
	Status          *uuid.UUID
	Assignee        *uuid.UUID
	Priority        *string
	Sprint          *uuid.UUID
	StartDate       *time.Time
	EndDate         *time.Time
}

func toCoreSingleStory(ns CoreNewStory, workspaceId uuid.UUID) CoreSingleStory {
	now := time.Now()
	if ns.Priority == "" {
		ns.Priority = "No Priority"
	}

	return CoreSingleStory{
		Workspace:       workspaceId,
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

// CoreActivity represents the core model for an activity.
type CoreActivity struct {
	ID           uuid.UUID `json:"id"`
	StoryID      uuid.UUID `json:"storyId"`
	UserID       uuid.UUID `json:"userId"`
	Type         string    `json:"type"`
	Field        string    `json:"field"`
	CurrentValue string    `json:"currentValue"`
	CreatedAt    time.Time `json:"createdAt"`
}
