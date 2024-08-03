package storiesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/google/uuid"
)

// dbStory represents the database model for an dbStory.
type dbStory struct {
	ID              uuid.UUID  `db:"id"`
	SequenceID      int        `db:"sequence_id"`
	Title           string     `db:"title"`
	Description     *string    `db:"description"`
	DescriptionHTML *string    `db:"description_html"`
	Parent          *uuid.UUID `db:"parent_id"`
	Objective       *uuid.UUID `db:"objective_id"`
	Epic            *uuid.UUID `db:"epic_id"`
	Workspace       uuid.UUID  `db:"workspace_id"`
	Team            uuid.UUID  `db:"team_id"`
	Status          *uuid.UUID `db:"status_id"`
	Assignee        *uuid.UUID `db:"assignee_id"`
	Estimate        *float32   `db:"estimate"`
	IsDraft         bool       `db:"is_draft"`
	BlockedBy       *uuid.UUID `db:"blocked_by_id"`
	Blocking        *uuid.UUID `db:"blocking_id"`
	Related         *uuid.UUID `db:"related_id"`
	Reporter        *uuid.UUID `db:"reporter_id"`
	Priority        string     `db:"priority"`
	Sprint          *uuid.UUID `db:"sprint_id"`
	StartDate       *time.Time `db:"start_date"`
	EndDate         *time.Time `db:"end_date"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
	ArchivedAt      *time.Time `db:"archived_at"`
}

// toCoreStory converts a dbStory to a CoreSingleStory.
func toCoreStory(i dbStory) stories.CoreSingleStory {
	return stories.CoreSingleStory{
		ID:              i.ID,
		SequenceID:      i.SequenceID,
		Title:           i.Title,
		Description:     i.Description,
		Parent:          i.Parent,
		Objective:       i.Objective,
		Team:            i.Team,
		Workspace:       i.Workspace,
		Status:          i.Status,
		Assignee:        i.Assignee,
		BlockedBy:       i.BlockedBy,
		Blocking:        i.Blocking,
		Related:         i.Related,
		Reporter:        i.Reporter,
		DescriptionHTML: i.DescriptionHTML,
		Priority:        i.Priority,
		Sprint:          i.Sprint,
		StartDate:       i.StartDate,
		EndDate:         i.EndDate,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		DeletedAt:       i.DeletedAt,
	}
}

// toCoreStories converts a slice of dbStories to a slice of CoreStoryList.
func toCoreStories(is []dbStory) []stories.CoreStoryList {
	cl := make([]stories.CoreStoryList, len(is))
	for i, story := range is {
		cl[i] = stories.CoreStoryList{
			ID:         story.ID,
			SequenceID: story.SequenceID,
			Title:      story.Title,
			Parent:     story.Parent,
			Objective:  story.Objective,
			Sprint:     story.Sprint,
			Epic:       story.Epic,
			Team:       story.Team,
			Workspace:  story.Workspace,
			Status:     story.Status,
			Assignee:   story.Assignee,
			Reporter:   story.Reporter,
			StartDate:  story.StartDate,
			EndDate:    story.EndDate,
			Priority:   story.Priority,
			CreatedAt:  story.CreatedAt,
			UpdatedAt:  story.UpdatedAt,
		}
	}
	return cl
}

// toDBStory converts a CoreSingleStory to a dbStory.
func toDBStory(i stories.CoreSingleStory) dbStory {
	return dbStory{
		SequenceID:      i.SequenceID,
		Title:           i.Title,
		Description:     i.Description,
		Parent:          i.Parent,
		Objective:       i.Objective,
		Workspace:       i.Workspace,
		Team:            i.Team,
		Status:          i.Status,
		Assignee:        i.Assignee,
		BlockedBy:       i.BlockedBy,
		Blocking:        i.Blocking,
		Related:         i.Related,
		Reporter:        i.Reporter,
		DescriptionHTML: i.DescriptionHTML,
		Priority:        i.Priority,
		Sprint:          i.Sprint,
		StartDate:       i.StartDate,
		EndDate:         i.EndDate,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		DeletedAt:       i.DeletedAt,
	}
}
