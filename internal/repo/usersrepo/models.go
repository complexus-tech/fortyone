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
	HasSeenWalkthrough  bool       `db:"has_seen_walkthrough"`
	Timezone            string     `db:"timezone"`
	LastLoginAt         time.Time  `db:"last_login_at"`
	LastUsedWorkspaceID *uuid.UUID `db:"last_used_workspace_id"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
	Role                *string    `db:"role"`
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

// dbAutomationPreferences represents automation preferences in the database
type dbAutomationPreferences struct {
	UserID                     uuid.UUID `db:"user_id"`
	WorkspaceID                uuid.UUID `db:"workspace_id"`
	AutoAssignSelf             bool      `db:"auto_assign_self"`
	AssignSelfOnBranchCopy     bool      `db:"assign_self_on_branch_copy"`
	MoveStoryToStartedOnBranch bool      `db:"move_story_to_started_on_branch"`
	OpenStoryInDialog          bool      `db:"open_story_in_dialog"`
	CreatedAt                  time.Time `db:"created_at"`
	UpdatedAt                  time.Time `db:"updated_at"`
}

func toCoreUser(p dbUser) users.CoreUser {
	return users.CoreUser{
		ID:                  p.ID,
		Username:            p.Username,
		Email:               p.Email,
		FullName:            p.FullName,
		AvatarURL:           p.AvatarURL,
		IsActive:            p.IsActive,
		HasSeenWalkthrough:  p.HasSeenWalkthrough,
		Timezone:            p.Timezone,
		LastLoginAt:         p.LastLoginAt,
		LastUsedWorkspaceID: p.LastUsedWorkspaceID,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
		Role:                p.Role,
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

// Convert database automation preferences to core model
func toCoreAutomationPreferences(p dbAutomationPreferences) users.CoreAutomationPreferences {
	return users.CoreAutomationPreferences{
		UserID:                     p.UserID,
		WorkspaceID:                p.WorkspaceID,
		AutoAssignSelf:             p.AutoAssignSelf,
		AssignSelfOnBranchCopy:     p.AssignSelfOnBranchCopy,
		MoveStoryToStartedOnBranch: p.MoveStoryToStartedOnBranch,
		OpenStoryInDialog:          p.OpenStoryInDialog,
		CreatedAt:                  p.CreatedAt,
		UpdatedAt:                  p.UpdatedAt,
	}
}

type dbUserMemoryItem struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	UserID      uuid.UUID `db:"user_id"`
	Content     string    `db:"content"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func toCoreUserMemoryItem(db dbUserMemoryItem) users.CoreUserMemoryItem {
	return users.CoreUserMemoryItem{
		ID:          db.ID,
		WorkspaceID: db.WorkspaceID,
		UserID:      db.UserID,
		Content:     db.Content,
		CreatedAt:   db.CreatedAt,
		UpdatedAt:   db.UpdatedAt,
	}
}

func toCoreUserMemoryItems(dbItems []dbUserMemoryItem) []users.CoreUserMemoryItem {
	items := make([]users.CoreUserMemoryItem, len(dbItems))
	for i, db := range dbItems {
		items[i] = toCoreUserMemoryItem(db)
	}
	return items
}
