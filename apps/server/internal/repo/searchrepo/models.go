package searchrepo

import (
	"encoding/json"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/search"
	"github.com/google/uuid"
)

// dbStory represents the database model for a story in search results.
type dbStory struct {
	ID         uuid.UUID        `db:"id"`
	SequenceID int              `db:"sequence_id"`
	Title      string           `db:"title"`
	Parent     *uuid.UUID       `db:"parent_id"`
	Objective  *uuid.UUID       `db:"objective_id"`
	Status     *uuid.UUID       `db:"status_id"`
	Assignee   *uuid.UUID       `db:"assignee_id"`
	Reporter   *uuid.UUID       `db:"reporter_id"`
	Priority   string           `db:"priority"`
	Sprint     *uuid.UUID       `db:"sprint_id"`
	KeyResult  *uuid.UUID       `db:"key_result_id"`
	Team       uuid.UUID        `db:"team_id"`
	Workspace  uuid.UUID        `db:"workspace_id"`
	StartDate  *time.Time       `db:"start_date"`
	EndDate    *time.Time       `db:"end_date"`
	CreatedAt  time.Time        `db:"created_at"`
	UpdatedAt  time.Time        `db:"updated_at"`
	Labels     *json.RawMessage `db:"labels"`
}

// dbObjective represents the database model for an objective in search results.
type dbObjective struct {
	ID          uuid.UUID  `db:"objective_id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	LeadUser    *uuid.UUID `db:"lead_user_id"`
	Team        uuid.UUID  `db:"team_id"`
	Workspace   uuid.UUID  `db:"workspace_id"`
	StartDate   *time.Time `db:"start_date"`
	EndDate     *time.Time `db:"end_date"`
	Status      uuid.UUID  `db:"status_id"`
	Priority    *string    `db:"priority"`
	Health      *string    `db:"health"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// toCoreSearchStory converts a dbStory to a CoreSearchStory.
func toCoreSearchStory(story dbStory) search.CoreSearchStory {
	var labels []uuid.UUID

	if story.Labels != nil {
		_ = json.Unmarshal(*story.Labels, &labels)
	}

	return search.CoreSearchStory{
		ID:         story.ID,
		SequenceID: story.SequenceID,
		Title:      story.Title,
		Parent:     story.Parent,
		Objective:  story.Objective,
		Status:     story.Status,
		Assignee:   story.Assignee,
		Reporter:   story.Reporter,
		Priority:   story.Priority,
		Sprint:     story.Sprint,
		KeyResult:  story.KeyResult,
		Team:       story.Team,
		Workspace:  story.Workspace,
		StartDate:  story.StartDate,
		EndDate:    story.EndDate,
		CreatedAt:  story.CreatedAt,
		UpdatedAt:  story.UpdatedAt,
		Labels:     labels,
	}
}

// toCoreSearchStories converts multiple dbStories to CoreSearchStories.
func toCoreSearchStories(stories []dbStory) []search.CoreSearchStory {
	result := make([]search.CoreSearchStory, len(stories))
	for i, story := range stories {
		result[i] = toCoreSearchStory(story)
	}
	return result
}

// toCoreSearchObjective converts a dbObjective to a CoreSearchObjective.
func toCoreSearchObjective(objective dbObjective) search.CoreSearchObjective {
	return search.CoreSearchObjective{
		ID:          objective.ID,
		Name:        objective.Name,
		Description: objective.Description,
		LeadUser:    objective.LeadUser,
		Team:        objective.Team,
		Workspace:   objective.Workspace,
		StartDate:   objective.StartDate,
		EndDate:     objective.EndDate,
		Status:      objective.Status,
		Priority:    objective.Priority,
		Health:      objective.Health,
		CreatedAt:   objective.CreatedAt,
		UpdatedAt:   objective.UpdatedAt,
	}
}

// toCoreSearchObjectives converts multiple dbObjectives to CoreSearchObjectives.
func toCoreSearchObjectives(objectives []dbObjective) []search.CoreSearchObjective {
	result := make([]search.CoreSearchObjective, len(objectives))
	for i, objective := range objectives {
		result[i] = toCoreSearchObjective(objective)
	}
	return result
}
