package objectivesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/google/uuid"
)

type dbObjective struct {
	ID               uuid.UUID  `db:"objective_id"`
	Name             string     `db:"name"`
	Description      *string    `db:"description"`
	LeadUser         *uuid.UUID `db:"lead_user_id"`
	Team             *uuid.UUID `db:"team_id"`
	Workspace        uuid.UUID  `db:"workspace_id"`
	StartDate        *time.Time `db:"start_date"`
	EndDate          *time.Time `db:"end_date"`
	IsPrivate        bool       `db:"is_private"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	Status           uuid.UUID  `db:"status_id"`
	Priority         *string    `db:"priority"`
	TotalStories     int        `db:"total_stories"`
	CancelledStories int        `db:"cancelled_stories"`
	CompletedStories int        `db:"completed_stories"`
	StartedStories   int        `db:"started_stories"`
	UnstartedStories int        `db:"unstarted_stories"`
	BacklogStories   int        `db:"backlog_stories"`
}

func toCoreObjective(p dbObjective) objectives.CoreObjective {
	return objectives.CoreObjective{
		ID:               p.ID,
		Name:             p.Name,
		Description:      p.Description,
		LeadUser:         p.LeadUser,
		Team:             p.Team,
		Workspace:        p.Workspace,
		StartDate:        p.StartDate,
		EndDate:          p.EndDate,
		IsPrivate:        p.IsPrivate,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		Status:           p.Status,
		Priority:         p.Priority,
		TotalStories:     p.TotalStories,
		CancelledStories: p.CancelledStories,
		CompletedStories: p.CompletedStories,
		StartedStories:   p.StartedStories,
		UnstartedStories: p.UnstartedStories,
		BacklogStories:   p.BacklogStories,
	}
}

func toCoreObjectives(do []dbObjective) []objectives.CoreObjective {
	objectives := make([]objectives.CoreObjective, len(do))
	for i, o := range do {
		objectives[i] = toCoreObjective(o)
	}
	return objectives
}
