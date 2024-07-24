package objectivesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/google/uuid"
)

// AppObjectiveList represents a list of objectives in the application.
type AppObjectiveList struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	LeadUser    *uuid.UUID `json:"leadUser"`
	Team        uuid.UUID  `json:"teamId"`
	Workspace   uuid.UUID  `json:"workspaceId"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	IsPrivate   bool       `json:"isPrivate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// toAppObjectives converts a list of core objectives to a list of application objectives.
func toAppObjectives(objectives []objectives.CoreObjective) []AppObjectiveList {
	appObjectives := make([]AppObjectiveList, len(objectives))
	for i, objective := range objectives {
		appObjectives[i] = AppObjectiveList{
			ID:          objective.ID,
			Name:        objective.Name,
			Description: objective.Description,
			LeadUser:    objective.LeadUser,
			Team:        objective.Team,
			Workspace:   objective.Workspace,
			StartDate:   objective.StartDate,
			EndDate:     objective.EndDate,
			IsPrivate:   objective.IsPrivate,
			CreatedAt:   objective.CreatedAt,
			UpdatedAt:   objective.UpdatedAt,
		}
	}
	return appObjectives
}
