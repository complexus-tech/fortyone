package users

import (
	"time"

	"github.com/google/uuid"
)

// CoreVerificationToken represents a verification token in the application layer
type CoreVerificationToken struct {
	ID        uuid.UUID  `json:"id"`
	Token     string     `json:"token"`
	Email     string     `json:"email"`
	UserID    *uuid.UUID `json:"userId"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt"`
	TokenType string     `json:"tokenType"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// CoreUser represents a user in the application layer.
type CoreUser struct {
	ID                  uuid.UUID
	Username            string
	Email               string
	FullName            string
	AvatarURL           string
	IsActive            bool
	HasSeenWalkthrough  bool
	Timezone            string
	LastLoginAt         time.Time
	LastUsedWorkspaceID *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Token               *string
	Role                *string
}

// CoreUpdateUser represents the fields that can be updated for a user.
type CoreUpdateUser struct {
	Username           *string
	FullName           *string
	AvatarURL          *string
	HasSeenWalkthrough *bool
	Timezone           *string
}

// CoreNewUser represents a new user to be created.
type CoreNewUser struct {
	Email     string
	FullName  string
	AvatarURL string
	Timezone  string
}

// CoreAutomationPreferences represents the automation preferences for a user in a workspace
type CoreAutomationPreferences struct {
	UserID                     uuid.UUID
	WorkspaceID                uuid.UUID
	AutoAssignSelf             bool
	AssignSelfOnBranchCopy     bool
	MoveStoryToStartedOnBranch bool
	OpenStoryInDialog          bool
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

// CoreUpdateAutomationPreferences represents the fields that can be updated for automation preferences
type CoreUpdateAutomationPreferences struct {
	AutoAssignSelf             *bool
	AssignSelfOnBranchCopy     *bool
	MoveStoryToStartedOnBranch *bool
	OpenStoryInDialog          *bool
}

// CoreUserMemoryItem represents a single memory item for a user.
type CoreUserMemoryItem struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type NewUserMemoryItem struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Content     string
}

type UpdateUserMemoryItem struct {
	Content *string
}
