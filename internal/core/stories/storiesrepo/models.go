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
	Description     string     `db:"description"`
	DescriptionHTML string     `db:"description_html"`
	Parent          *uuid.UUID `db:"parent_id"`
	Project         *uuid.UUID `db:"project_id"`
	Status          *uuid.UUID `db:"status_id"`
	Assignee        *uuid.UUID `db:"assignee_id"`
	BlockedBy       *uuid.UUID `db:"blocked_by_id"`
	Blocking        *uuid.UUID `db:"blocking_id"`
	Related         *uuid.UUID `db:"related_id"`
	Reporter        *uuid.UUID `db:"reporter_id"`
	Priority        *uuid.UUID `db:"priority_id"`
	Attachments     string     `db:"attachments"`
	Sprint          *uuid.UUID `db:"sprint_id"`
	Watchers        string     `db:"watchers"`
	Labels          string     `db:"labels"`
	Comments        string     `db:"comments"`
	StartDate       *time.Time `db:"start_date"`
	EndDate         *time.Time `db:"end_date"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

// toCoreStory converts a dbStory to a CoreSingleStory.
func toCoreStory(i dbStory) stories.CoreSingleStory {
	return stories.CoreSingleStory{
		ID:              i.ID,
		SequenceID:      i.SequenceID,
		Title:           i.Title,
		Description:     i.Description,
		Parent:          i.Parent,
		Project:         i.Project,
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
			Project:    story.Project,
			Status:     story.Status,
			Assignee:   story.Assignee,
			StartDate:  story.StartDate,
			EndDate:    story.EndDate,
			Priority:   story.Priority,
			Sprint:     story.Sprint,
		}
	}
	return cl
}
