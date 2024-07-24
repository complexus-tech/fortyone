package storiesgrp

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// AppBulkDeleteRequest represents a request to delete multiple stories.
type AppBulkDeleteRequest struct {
	StoryIDs []uuid.UUID `json:"storyIds"`
}

// AppBulkRestoreRequest represents a request to restore multiple stories.
type AppBulkRestoreRequest struct {
	StoryIDs []uuid.UUID `json:"storyIds"`
}

// AppSingleStory represents a single story in the application.
type AppSingleStory struct {
	ID              uuid.UUID  `json:"id"`
	SequenceID      int        `json:"sequenceId"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	DescriptionHTML *string    `json:"descriptionHTML"`
	Parent          *uuid.UUID `json:"parentId"`
	Status          *uuid.UUID `json:"statusId"`
	Assignee        *uuid.UUID `json:"assigneeId"`
	BlockedBy       *uuid.UUID `json:"blockedById"`
	Blocking        *uuid.UUID `json:"blockingId"`
	Related         *uuid.UUID `json:"relatedId"`
	Reporter        *uuid.UUID `json:"reporterId"`
	Priority        string     `json:"priority"`
	Sprint          *uuid.UUID `json:"sprintId"`
	Epic            *uuid.UUID `json:"epicId"`
	Objective       *uuid.UUID `json:"objectiveId"`
	Team            uuid.UUID  `json:"teamId"`
	Workspace       uuid.UUID  `json:"workspaceId"`
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
	Objective  *uuid.UUID `json:"objectiveId"`
	Status     *uuid.UUID `json:"statusId"`
	Assignee   *uuid.UUID `json:"assigneeId"`
	Priority   string     `json:"priority"`
	Sprint     *uuid.UUID `json:"sprintId"`
	Workspace  uuid.UUID  `json:"workspaceId"`
	Team       uuid.UUID  `json:"teamId"`
	StartDate  *time.Time `json:"startDate"`
	EndDate    *time.Time `json:"endDate"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func toAppStory(i stories.CoreSingleStory) AppSingleStory {
	return AppSingleStory{
		ID:              i.ID,
		SequenceID:      i.SequenceID,
		Description:     i.Description,
		DescriptionHTML: i.DescriptionHTML,
		Parent:          i.Parent,
		Title:           i.Title,
		Objective:       i.Objective,
		Status:          i.Status,
		Assignee:        i.Assignee,
		Priority:        i.Priority,
		Sprint:          i.Sprint,
		Epic:            i.Epic,
		Team:            i.Team,
		Workspace:       i.Workspace,
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
			Objective:  story.Objective,
			Team:       story.Team,
			Workspace:  story.Workspace,
			Status:     story.Status,
			Assignee:   story.Assignee,
			Priority:   story.Priority,
			Sprint:     story.Sprint,
			StartDate:  story.StartDate,
			EndDate:    story.EndDate,
			CreatedAt:  story.CreatedAt,
			UpdatedAt:  story.UpdatedAt,
		}
	}
	return appStories
}

type AppNewStory struct {
	Title           string     `json:"title" validate:"required"`
	Description     *string    `json:"description"`
	DescriptionHTML *string    `json:"descriptionHTML"`
	Parent          *uuid.UUID `json:"parentId"`
	Objective       *uuid.UUID `json:"objectiveId"`
	Status          *uuid.UUID `json:"statusId"`
	Assignee        *uuid.UUID `json:"assigneeId"`
	Priority        string     `json:"priority" validate:"oneof='No Priority' Low Medium High Urgent"`
	Sprint          *uuid.UUID `json:"sprintId"`
	Team            uuid.UUID  `json:"teamId" validate:"required"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
}

type AppUpdateStory struct {
	Title           *string    `json:"title"`
	Description     *string    `json:"description"`
	DescriptionHTML *string    `json:"descriptionHTML"`
	Parent          *uuid.UUID `json:"parentId"`
	Objective       *uuid.UUID `json:"objectiveId"`
	Status          *uuid.UUID `json:"statusId"`
	Assignee        *uuid.UUID `json:"assigneeId"`
	Priority        *string    `json:"priority" validate:"omitempty,oneof='No Priority' Low Medium High Urgent"`
	Sprint          *uuid.UUID `json:"sprintId"`
	StartDate       *time.Time `json:"startDate"`
	EndDate         *time.Time `json:"endDate"`
}

func toCoreNewStory(a AppNewStory) stories.CoreNewStory {
	return stories.CoreNewStory{
		Title:           a.Title,
		Description:     a.Description,
		DescriptionHTML: a.DescriptionHTML,
		Parent:          a.Parent,
		Objective:       a.Objective,
		Status:          a.Status,
		Assignee:        a.Assignee,
		Priority:        a.Priority,
		Sprint:          a.Sprint,
		StartDate:       a.StartDate,
		EndDate:         a.EndDate,
		Team:            a.Team,
	}
}

func toCoreUpdateStory(a AppUpdateStory) stories.CoreUpdateStory {
	return stories.CoreUpdateStory{
		Title:           a.Title,
		Description:     a.Description,
		DescriptionHTML: a.DescriptionHTML,
		Parent:          a.Parent,
		Objective:       a.Objective,
		Status:          a.Status,
		Assignee:        a.Assignee,
		Priority:        a.Priority,
		Sprint:          a.Sprint,
		StartDate:       a.StartDate,
		EndDate:         a.EndDate,
	}
}

func (a AppNewStory) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(a)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range validationErrors {
				fieldName := getJSONTagName(reflect.TypeOf(a), e.Field())
				switch e.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s is required", fieldName))
				case "oneof":
					options := strings.Split(e.Param(), " ")
					formattedOptions := formatOptions(options)
					errorMessages = append(errorMessages, fmt.Sprintf("%s should be one of: %s", fieldName, formattedOptions))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("%s failed validation: %s", fieldName, e.Tag()))
				}
			}
			return fmt.Errorf(strings.Join(errorMessages, "; "))
		}
	}
	return err
}

func formatOptions(options []string) string {
	if len(options) == 0 {
		return ""
	}
	if len(options) == 1 {
		return options[0]
	}
	return fmt.Sprintf("%s or %s", strings.Join(options[:len(options)-1], ", "), options[len(options)-1])
}

func getJSONTagName(t reflect.Type, fieldName string) string {
	field, found := t.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return fieldName // Return original field name if no JSON tag
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return fieldName // Return original field name if JSON tag is "-"
	}

	return parts[0] // Return the JSON tag name
}
