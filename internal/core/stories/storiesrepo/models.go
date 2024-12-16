package storiesrepo

import (
	"encoding/json"
	"log"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/google/uuid"
)

// dbStory represents the database model for an dbStory.
type dbStory struct {
	ID              uuid.UUID        `db:"id"`
	SequenceID      int              `db:"sequence_id"`
	Title           string           `db:"title"`
	Description     *string          `db:"description"`
	DescriptionHTML *string          `db:"description_html"`
	Parent          *uuid.UUID       `db:"parent_id"`
	Objective       *uuid.UUID       `db:"objective_id"`
	Epic            *uuid.UUID       `db:"epic_id"`
	Workspace       uuid.UUID        `db:"workspace_id"`
	Team            uuid.UUID        `db:"team_id"`
	Status          *uuid.UUID       `db:"status_id"`
	Assignee        *uuid.UUID       `db:"assignee_id"`
	Estimate        *float32         `db:"estimate"`
	IsDraft         bool             `db:"is_draft"`
	BlockedBy       *uuid.UUID       `db:"blocked_by_id"`
	Blocking        *uuid.UUID       `db:"blocking_id"`
	Related         *uuid.UUID       `db:"related_id"`
	Reporter        *uuid.UUID       `db:"reporter_id"`
	Priority        string           `db:"priority"`
	Sprint          *uuid.UUID       `db:"sprint_id"`
	StartDate       *time.Time       `db:"start_date"`
	EndDate         *time.Time       `db:"end_date"`
	CreatedAt       time.Time        `db:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at"`
	DeletedAt       *time.Time       `db:"deleted_at"`
	ArchivedAt      *time.Time       `db:"archived_at"`
	SubStories      *json.RawMessage `db:"sub_stories"`
}

// toCoreStory converts a dbStory to a CoreSingleStory.
func toCoreStory(i dbStory) stories.CoreSingleStory {
	var subStories []stories.CoreStoryList
	if i.SubStories != nil {
		err := json.Unmarshal(*i.SubStories, &subStories)
		if err != nil {
			log.Printf("Failed to unmarshal sub_stories: %s", err)
		}
	}

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
		SubStories:      subStories,
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

// dbActivity represents the database model for an dbActivity.
type dbActivity struct {
	ID           uuid.UUID `db:"activity_id"`
	StoryID      uuid.UUID `db:"story_id"`
	UserID       uuid.UUID `db:"user_id"`
	Type         string    `db:"activity_type"`
	Field        string    `db:"field_changed"`
	CurrentValue string    `db:"current_value"`
	CreatedAt    time.Time `db:"created_at"`
}

// toCoreActivity converts a dbActivity to a CoreActivity.
func toCoreActivity(i dbActivity) stories.CoreActivity {
	return stories.CoreActivity{
		ID:           i.ID,
		StoryID:      i.StoryID,
		UserID:       i.UserID,
		Type:         i.Type,
		Field:        i.Field,
		CurrentValue: i.CurrentValue,
		CreatedAt:    i.CreatedAt,
	}
}

// toDBActivity converts a CoreActivity to a dbActivity.
func toDBActivity(i stories.CoreActivity) dbActivity {
	return dbActivity{
		StoryID:      i.StoryID,
		UserID:       i.UserID,
		Type:         i.Type,
		Field:        i.Field,
		CurrentValue: i.CurrentValue,
	}
}

// toCoreActivities converts a slice of dbActivities to a slice of CoreActivity.
func toCoreActivities(is []dbActivity) []stories.CoreActivity {
	ca := make([]stories.CoreActivity, len(is))
	for i, activity := range is {
		ca[i] = toCoreActivity(activity)
	}
	return ca
}
