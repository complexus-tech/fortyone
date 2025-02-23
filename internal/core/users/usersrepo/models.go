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
	FullName            string     `db:"full_name"`
	AvatarURL           string     `db:"avatar_url"`
	IsActive            bool       `db:"is_active"`
	LastLoginAt         time.Time  `db:"last_login_at"`
	LastUsedWorkspaceID *uuid.UUID `db:"last_used_workspace_id"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
	WorkspaceRole       string     `db:"workspace_role"`
	TeamRole            *string    `db:"team_role"`
}

// dbVerificationToken represents a verification token in the database
type dbVerificationToken struct {
	ID        uuid.UUID  `db:"id"`
	Token     string     `db:"token"`
	Email     string     `db:"email"`
	UserID    *uuid.UUID `db:"user_id"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	TokenType string     `db:"token_type"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

func toCoreUser(p dbUser) users.CoreUser {
	return users.CoreUser{
		ID:                  p.ID,
		Username:            p.Username,
		Email:               p.Email,
		FullName:            p.FullName,
		AvatarURL:           p.AvatarURL,
		IsActive:            p.IsActive,
		LastLoginAt:         p.LastLoginAt,
		LastUsedWorkspaceID: p.LastUsedWorkspaceID,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
		WorkspaceRole:       p.WorkspaceRole,
		TeamRole:            p.TeamRole,
	}
}

func toCoreUsers(du []dbUser) []users.CoreUser {
	coreUsers := make([]users.CoreUser, len(du))
	for i, user := range du {
		coreUsers[i] = toCoreUser(user)
	}
	return coreUsers
}

func toCoreVerificationToken(dbToken dbVerificationToken) users.CoreVerificationToken {
	return users.CoreVerificationToken{
		ID:        dbToken.ID,
		Token:     dbToken.Token,
		Email:     dbToken.Email,
		UserID:    dbToken.UserID,
		ExpiresAt: dbToken.ExpiresAt,
		UsedAt:    dbToken.UsedAt,
		TokenType: dbToken.TokenType,
		CreatedAt: dbToken.CreatedAt,
		UpdatedAt: dbToken.UpdatedAt,
	}
}
