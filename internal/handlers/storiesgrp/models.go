package storiesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// AppSingleStory represents a single story in the application.
type AppSingleStory struct {
	ID              uuid.UUID  `json:"id"`
	SequenceID      int        `json:"sequenceId"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	DescriptionHTML string     `json:"descriptionHTML"`
	Parent          *uuid.UUID `json:"parentId"`
	Project         *uuid.UUID `json:"projectId"`
	Status          *uuid.UUID `json:"statusId"`
	Assignee        *uuid.UUID `json:"assigneeId"`
	BlockedBy       *uuid.UUID `json:"blockedById"`
	Blocking        *uuid.UUID `json:"blockingId"`
	Related         *uuid.UUID `json:"relatedId"`
	Reporter        *uuid.UUID `json:"reporterId"`
	Priority        *uuid.UUID `json:"priorityId"`
	Sprint          *uuid.UUID `json:"sprintId"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt"`
}

// AppStoryList represents a single story in the list of stories in the application.
type AppStoryList struct {
	ID         uuid.UUID  `json:"id"`
	SequenceID int        `json:"sequenceId"`
	Title      string     `json:"title"`
	Project    *uuid.UUID `json:"project"`
	Status     *uuid.UUID `json:"status"`
	Assignee   *uuid.UUID `json:"assignee"`
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	Priority   *uuid.UUID `json:"priority"`
	Sprint     *uuid.UUID `json:"sprint"`
}

func toAppStory(i stories.CoreSingleStory) AppSingleStory {
	return AppSingleStory{
		ID:              i.ID,
		SequenceID:      i.SequenceID,
		Description:     i.Description,
		DescriptionHTML: i.DescriptionHTML,
		Parent:          i.Parent,
		Title:           i.Title,
		Project:         i.Project,
		Status:          i.Status,
		Assignee:        i.Assignee,
		Priority:        i.Priority,
		Sprint:          i.Sprint,
		StartDate:       i.StartDate,
		EndDate:         i.EndDate,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		DeletedAt:       i.DeletedAt,
		BlockedBy:       i.BlockedBy,
		Blocking:        i.Blocking,
		Related:         i.Related,
		Reporter:        i.Reporter,
	}
}

func toAppStories(stories []stories.CoreStoryList) []AppStoryList {
	appStories := make([]AppStoryList, len(stories))
	for i, story := range stories {
		appStories[i] = AppStoryList{
			ID:         story.ID,
			SequenceID: story.SequenceID,
			Title:      story.Title,
			Project:    story.Project,
			Status:     story.Status,
			Assignee:   story.Assignee,
			Priority:   story.Priority,
			Sprint:     story.Sprint,
			StartDate:  story.StartDate,
			EndDate:    story.EndDate,
		}
	}
	return appStories
}

type AppNewStory struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (a AppNewStory) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(a)
}
