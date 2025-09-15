package activitiesgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/google/uuid"
)

type AppActivity struct {
	ID           uuid.UUID `json:"id"`
	StoryID      uuid.UUID `json:"storyId"`
	UserID       uuid.UUID `json:"userId"`
	Type         string    `json:"type"`
	Field        string    `json:"field"`
	CurrentValue string    `json:"currentValue"`
	CreatedAt    time.Time `json:"createdAt"`
	WorkspaceID  uuid.UUID `json:"workspaceId"`

	// User details
	User AppUserDetails `json:"user"`
}

// AppUserDetails represents basic user information for activities
type AppUserDetails struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
}

type AppFilters struct {
	Limit string `json:"limit"`
}

func toAppActivity(a activities.CoreActivity) AppActivity {
	return AppActivity{
		ID:           a.ID,
		StoryID:      a.StoryID,
		UserID:       a.UserID,
		Type:         a.Type,
		Field:        a.Field,
		CurrentValue: a.CurrentValue,
		CreatedAt:    a.CreatedAt,
		WorkspaceID:  a.WorkspaceID,
		User: AppUserDetails{
			ID:        a.User.ID,
			Username:  a.User.Username,
			FullName:  a.User.FullName,
			AvatarURL: a.User.AvatarURL,
		},
	}
}

func toAppActivities(acts []activities.CoreActivity) []AppActivity {
	activities := make([]AppActivity, len(acts))
	for i, a := range acts {
		activities[i] = toAppActivity(a)
	}
	return activities
}
