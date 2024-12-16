package activitiesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/google/uuid"
)

// dbActivity represents the database model for an dbActivity.
type dbActivity struct {
	ID            uuid.UUID `db:"activity_id"`
	StoryID       uuid.UUID `db:"story_id"`
	Type          string    `db:"activity_type"`
	Field         string    `db:"field_changed"`
	PreviousValue *string   `db:"previous_value"`
	CurrentValue  *string   `db:"current_value"`
	CreatedAt     time.Time `db:"created_at"`
}

// toCoreActivity converts a dbActivity to a CoreActivity.
func toCoreActivity(i dbActivity) activities.CoreActivity {
	return activities.CoreActivity{
		ID:            i.ID,
		StoryID:       i.StoryID,
		Type:          i.Type,
		Field:         i.Field,
		PreviousValue: i.PreviousValue,
		CurrentValue:  i.CurrentValue,
		CreatedAt:     i.CreatedAt,
	}
}

// toDBActivity converts a CoreActivity to a dbActivity.
func toDBActivity(i activities.CoreActivity) dbActivity {
	return dbActivity{
		ID:            i.ID,
		StoryID:       i.StoryID,
		Type:          i.Type,
		Field:         i.Field,
		PreviousValue: i.PreviousValue,
		CurrentValue:  i.CurrentValue,
		CreatedAt:     i.CreatedAt,
	}
}

// toCoreActivities converts a slice of dbActivities to a slice of CoreActivity.
func toCoreActivities(is []dbActivity) []activities.CoreActivity {
	ca := make([]activities.CoreActivity, len(is))
	for i, activity := range is {
		ca[i] = toCoreActivity(activity)
	}
	return ca
}
