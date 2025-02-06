package users

import (
	"time"

	"github.com/google/uuid"
)

// CoreUser represents a user in the application layer.
type CoreUser struct {
	ID                  uuid.UUID
	Username            string
	Email               string
	Password            string
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
	Username  string
	Email     string
	Password  string
	FullName  string
	AvatarURL string
}
