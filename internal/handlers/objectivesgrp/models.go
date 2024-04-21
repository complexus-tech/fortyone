package objectivesgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/google/uuid"
)

// AppObjectiveList represents a list of objectives in the application.
type AppObjectiveList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"title"`
	Description string    `json:"description"`
}

// toAppObjectives converts a list of core objectives to a list of application objectives.
func toAppObjectives(objectives []objectives.CoreObjective) []AppObjectiveList {
	appObjectives := make([]AppObjectiveList, len(objectives))
	for i, objective := range objectives {
		appObjectives[i] = AppObjectiveList{
			ID:          objective.ID,
			Name:        objective.Name,
			Description: objective.Description,
		}
	}
	return appObjectives
}
