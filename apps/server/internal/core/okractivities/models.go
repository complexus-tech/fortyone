package okractivities

import (
	"time"

	"github.com/google/uuid"
)

// OKRUpdateType represents whether the activity is for an objective or key result
type OKRUpdateType string

const (
	UpdateTypeObjective OKRUpdateType = "objective"
	UpdateTypeKeyResult OKRUpdateType = "key_result"
)

// OKRActivityType represents the type of activity performed
type OKRActivityType string

const (
	ActivityTypeCreate OKRActivityType = "create"
	ActivityTypeUpdate OKRActivityType = "update"
	ActivityTypeDelete OKRActivityType = "delete"
)

// CoreActivity represents an OKR activity in the system
type CoreActivity struct {
	ID           uuid.UUID       `json:"id"`
	ObjectiveID  uuid.UUID       `json:"objectiveId"`
	KeyResultID  *uuid.UUID      `json:"keyResultId"`
	UserID       uuid.UUID       `json:"userId"`
	Type         OKRActivityType `json:"type"`
	UpdateType   OKRUpdateType   `json:"updateType"`
	Field        string          `json:"field"`
	CurrentValue string          `json:"currentValue"`
	Comment      string          `json:"comment"`
	CreatedAt    time.Time       `json:"createdAt"`
	WorkspaceID  uuid.UUID       `json:"workspaceId"`

	// User details (populated when fetching activities)
	User UserDetails `json:"user"`
}

// UserDetails represents basic user information for activities
type UserDetails struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	AvatarURL string    `json:"avatarUrl"`
	IsActive  bool      `json:"isActive"`
}

// CoreNewActivity represents the data needed to create a new activity
type CoreNewActivity struct {
	ObjectiveID  uuid.UUID
	KeyResultID  *uuid.UUID
	UserID       uuid.UUID
	Type         OKRActivityType
	UpdateType   OKRUpdateType
	Field        string
	CurrentValue string
	Comment      string
	WorkspaceID  uuid.UUID
}
