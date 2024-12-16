package usersrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/google/uuid"
)

type dbUser struct {
	ID                  uuid.UUID  `db:"user_id"`
	Username            string     `db:"username"`
	Email               string     `db:"email"`
	Password            string     `db:"password_hash"`
	FullName            string     `db:"full_name"`
	AvatarURL           string     `db:"avatar_url"`
	IsActive            bool       `db:"is_active"`
	LastLoginAt         time.Time  `db:"last_login_at"`
	LastUsedWorkspaceID *uuid.UUID `db:"last_used_workspace_id"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
}

func toCoreUser(p dbUser) users.CoreUser {
	return users.CoreUser{
		ID:                  p.ID,
		Username:            p.Username,
		Email:               p.Email,
		Password:            p.Password,
		FullName:            p.FullName,
		AvatarURL:           p.AvatarURL,
		IsActive:            p.IsActive,
		LastLoginAt:         p.LastLoginAt,
		LastUsedWorkspaceID: p.LastUsedWorkspaceID,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}

func toCoreUsers(du []dbUser) []users.CoreUser {
	coreUsers := make([]users.CoreUser, len(du))
	for i, user := range du {
		coreUsers[i] = toCoreUser(user)
	}
	return coreUsers
}
