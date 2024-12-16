package activities

import (
	"time"

	"github.com/google/uuid"
)

// CoreActivity represents the core model for an activity.
type CoreActivity struct {
	ID            uuid.UUID `json:"id"`
	StoryID       uuid.UUID `json:"storyId"`
	Type          string    `json:"type"`
	Field         string    `json:"field"`
	PreviousValue *string   `json:"previousValue"`
	CurrentValue  *string   `json:"currentValue"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CoreNewActivity struct {
	StoryID       uuid.UUID `json:"storyId"`
	Type          string    `json:"type"`
	Field         string    `json:"field"`
	PreviousValue *string   `json:"previousValue"`
	CurrentValue  *string   `json:"currentValue"`
}

func toCoreActivity(i CoreNewActivity) CoreActivity {
	return CoreActivity{
		StoryID:       i.StoryID,
		Type:          i.Type,
		Field:         i.Field,
		PreviousValue: i.PreviousValue,
		CurrentValue:  i.CurrentValue,
	}
}
