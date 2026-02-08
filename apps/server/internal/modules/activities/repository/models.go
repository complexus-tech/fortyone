package activitiesrepository

import (
	"time"

	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	"github.com/google/uuid"
)

type dbActivity struct {
	ID           uuid.UUID `db:"activity_id"`
	StoryID      uuid.UUID `db:"story_id"`
	UserID       uuid.UUID `db:"user_id"`
	Type         string    `db:"activity_type"`
	Field        string    `db:"field_changed"`
	CurrentValue string    `db:"current_value"`
	CreatedAt    time.Time `db:"created_at"`
	WorkspaceID  uuid.UUID `db:"workspace_id"`

	// User details from JOIN
	Username  string `db:"username"`
	FullName  string `db:"full_name"`
	AvatarURL string `db:"avatar_url"`
	IsActive  bool   `db:"is_active"`
}

func toCoreActivity(a dbActivity) activities.CoreActivity {
	return activities.CoreActivity{
		ID:           a.ID,
		StoryID:      a.StoryID,
		UserID:       a.UserID,
		Type:         a.Type,
		Field:        a.Field,
		CurrentValue: a.CurrentValue,
		CreatedAt:    a.CreatedAt,
		WorkspaceID:  a.WorkspaceID,
		User: activities.UserDetails{
			ID:        a.UserID,
			Username:  a.Username,
			FullName:  a.FullName,
			AvatarURL: a.AvatarURL,
			IsActive:  a.IsActive,
		},
	}
}

func toCoreActivities(acts []dbActivity) []activities.CoreActivity {
	activities := make([]activities.CoreActivity, len(acts))
	for i, a := range acts {
		activities[i] = toCoreActivity(a)
	}
	return activities
}
