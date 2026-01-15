package sprintsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/google/uuid"
)

type dbSprint struct {
	ID               uuid.UUID  `db:"sprint_id"`
	Name             string     `db:"name"`
	Goal             *string    `db:"goal"`
	Objective        *uuid.UUID `db:"objective_id"`
	Team             uuid.UUID  `db:"team_id"`
	Workspace        uuid.UUID  `db:"workspace_id"`
	StartDate        time.Time  `db:"start_date"`
	EndDate          time.Time  `db:"end_date"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	TotalStories     int        `db:"total_stories"`
	CancelledStories int        `db:"cancelled_stories"`
	CompletedStories int        `db:"completed_stories"`
	StartedStories   int        `db:"started_stories"`
	UnstartedStories int        `db:"unstarted_stories"`
	BacklogStories   int        `db:"backlog_stories"`
}

func toCoreSprint(s dbSprint) sprints.CoreSprint {
	return sprints.CoreSprint{
		ID:               s.ID,
		Name:             s.Name,
		Goal:             s.Goal,
		Objective:        s.Objective,
		Team:             s.Team,
		Workspace:        s.Workspace,
		StartDate:        s.StartDate,
		EndDate:          s.EndDate,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		TotalStories:     s.TotalStories,
		CancelledStories: s.CancelledStories,
		CompletedStories: s.CompletedStories,
		StartedStories:   s.StartedStories,
		UnstartedStories: s.UnstartedStories,
		BacklogStories:   s.BacklogStories,
	}
}

func toCoreSprints(do []dbSprint) []sprints.CoreSprint {
	sprints := make([]sprints.CoreSprint, len(do))
	for i, o := range do {
		sprints[i] = toCoreSprint(o)
	}
	return sprints
}
