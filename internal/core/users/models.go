package users

import (
	"time"

	"github.com/google/uuid"
)

// CoreUser represents a user in the application layer.
type CoreUser struct {
	ID          uuid.UUID
	Username    string
	Email       string
	Password    string
	FullName    string
	Role        string
	AvatarURL   string
	IsActive    bool
	LastLoginAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Token       *string
}
