// Package storytransport owns shared request contracts for web and native story
// mutations. It depends on domain values, never HTTP handlers or services.
package storytransport

import (
	"fmt"
	"strings"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/google/uuid"
)

// NewStoryRequest is the finite create payload shared by both transports.
type NewStoryRequest struct {
	Title                    string                         `json:"title" validate:"required"`
	EstimateValue            *int16                         `json:"estimateValue"`
	EstimatedDurationMinutes *int                           `json:"estimatedDurationMinutes"`
	MinimumFocusBlockMinutes *int                           `json:"minimumFocusBlockMinutes"`
	AutoSchedulingEnabled    bool                           `json:"autoSchedulingEnabled"`
	Description              *string                        `json:"description"`
	DescriptionHTML          *string                        `json:"descriptionHTML"`
	Parent                   *uuid.UUID                     `json:"parentId"`
	Objective                *uuid.UUID                     `json:"objectiveId"`
	Status                   *uuid.UUID                     `json:"statusId"`
	Assignee                 *uuid.UUID                     `json:"assigneeId"`
	Priority                 string                         `json:"priority" validate:"oneof='No Priority' Low Medium High Urgent"`
	Sprint                   *uuid.UUID                     `json:"sprintId"`
	KeyResult                *uuid.UUID                     `json:"keyResultId"`
	LabelIDs                 []uuid.UUID                    `json:"labelIds"`
	CustomFieldValues        []storydomain.CustomFieldValue `json:"customFieldValues" validate:"max=50,dive"`
	Team                     uuid.UUID                      `json:"teamId" validate:"required"`
	StartDate                *date.Date                     `json:"startDate"`
	EndDate                  *date.Date                     `json:"endDate"`
	IdempotencyKey           *string                        `json:"idempotencyKey" validate:"omitempty,max=128"`
}

// ToNewStory maps a validated request and an authenticated actor to a command.
func ToNewStory(a NewStoryRequest, userID uuid.UUID) storydomain.NewStory {
	var creationKey *string
	if a.IdempotencyKey != nil {
		trimmedKey := strings.TrimSpace(*a.IdempotencyKey)
		if trimmedKey != "" {
			key := fmt.Sprintf("app:%s:%s", userID, trimmedKey)
			creationKey = &key
		}
	}

	return storydomain.NewStory{
		Title:                    a.Title,
		EstimateValue:            a.EstimateValue,
		EstimatedDurationMinutes: a.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: a.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    a.AutoSchedulingEnabled,
		Description:              a.Description,
		DescriptionHTML:          a.DescriptionHTML,
		Parent:                   a.Parent,
		Objective:                a.Objective,
		Status:                   a.Status,
		Assignee:                 a.Assignee,
		Reporter:                 &userID,
		Priority:                 a.Priority,
		Sprint:                   a.Sprint,
		KeyResult:                a.KeyResult,
		LabelIDs:                 a.LabelIDs,
		CustomFieldValues:        a.CustomFieldValues,
		StartDate:                a.StartDate.TimePtr(),
		EndDate:                  a.EndDate.TimePtr(),
		Team:                     a.Team,
		CreationKey:              creationKey,
	}
}
