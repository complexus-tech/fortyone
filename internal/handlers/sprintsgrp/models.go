package sprintsgrp

import (
	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/google/uuid"
)

// AppSprintList represents a sprint in the application layer.
type AppSprintsList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"title"`
	Description string    `json:"description"`
}

// toAppSprints converts a list of core sprints to a list of application sprints.
func toAppSprints(sprints []sprints.CoreSprint) []AppSprintsList {
	appSprints := make([]AppSprintsList, len(sprints))
	for i, sprint := range sprints {
		appSprints[i] = AppSprintsList{
			ID:          sprint.ID,
			Name:        sprint.Name,
			Description: sprint.Description,
		}
	}
	return appSprints
}
