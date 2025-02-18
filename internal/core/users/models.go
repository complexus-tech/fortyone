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
	LastLoginAt         time.Time
	LastUsedWorkspaceID *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Token               *string
}

// CoreUpdateUser represents the fields that can be updated for a user.
type CoreUpdateUser struct {
	Username  string
	FullName  string
	AvatarURL string
}

// CoreNewUser represents a new user to be created.
type CoreNewUser struct {
	Email     string
	FullName  string
	AvatarURL string
}
