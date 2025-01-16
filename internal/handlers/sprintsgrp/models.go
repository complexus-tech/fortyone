package sprintsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/google/uuid"
)

// AppSprintList represents a sprint in the application layer.
type AppSprintsList struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	Goal      *string     `json:"goal"`
	Objective *uuid.UUID  `json:"objectiveId"`
	Team      uuid.UUID   `json:"teamId"`
	Workspace uuid.UUID   `json:"workspaceId"`
	StartDate time.Time   `json:"startDate"`
	EndDate   time.Time   `json:"endDate"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
	Stats     SprintStats `json:"stats"`
}

type SprintStats struct {
	Total     int `json:"total"`
	Cancelled int `json:"cancelled"`
	Completed int `json:"completed"`
	Started   int `json:"started"`
	Unstarted int `json:"unstarted"`
	Backlog   int `json:"backlog"`
}

type AppFilters struct {
	Objective *uuid.UUID `json:"objectiveId" db:"objective_id"`
	Team      *uuid.UUID `json:"teamId" db:"team_id"`
}

// toAppSprints converts a list of core sprints to a list of application sprints.
func toAppSprints(sprints []sprints.CoreSprint) []AppSprintsList {
	appSprints := make([]AppSprintsList, len(sprints))
	for i, sprint := range sprints {
		appSprints[i] = AppSprintsList{
			ID:        sprint.ID,
			Name:      sprint.Name,
			Goal:      sprint.Goal,
			Objective: sprint.Objective,
			Team:      sprint.Team,
			Workspace: sprint.Workspace,
			StartDate: sprint.StartDate,
			EndDate:   sprint.EndDate,
			CreatedAt: sprint.CreatedAt,
			UpdatedAt: sprint.UpdatedAt,
			Stats: SprintStats{
				Total:     sprint.TotalStories,
				Cancelled: sprint.CancelledStories,
				Completed: sprint.CompletedStories,
				Started:   sprint.StartedStories,
				Unstarted: sprint.UnstartedStories,
				Backlog:   sprint.BacklogStories,
			},
		}
	}
	return appSprints
}
