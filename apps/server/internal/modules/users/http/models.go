package usershttp

import (
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

type AppUser struct {
	ID                  uuid.UUID  `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	FullName            string     `json:"fullName"`
	AvatarURL           string     `json:"avatarUrl"`
	IsActive            bool       `json:"isActive"`
	HasSeenWalkthrough  bool       `json:"hasSeenWalkthrough"`
	Timezone            string     `json:"timezone"`
	LastLoginAt         time.Time  `json:"-"`
	LastUsedWorkspaceID *uuid.UUID `json:"lastUsedWorkspaceId"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	Token               *string    `json:"token,omitempty"`
	Role                *string    `json:"role,omitempty"`
}

type AppFilter struct {
	TeamID *uuid.UUID `json:"teamId"`
}

type UpdateProfileRequest struct {
	Username           string  `json:"username"`
	FullName           *string `json:"fullName,omitempty"`
	AvatarURL          *string `json:"avatarUrl,omitempty"`
	HasSeenWalkthrough *bool   `json:"hasSeenWalkthrough,omitempty"`
	Timezone           *string `json:"timezone,omitempty"`
}

type SwitchWorkspaceRequest struct {
	WorkspaceID uuid.UUID `json:"workspaceId"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	FullName  string `json:"fullName"`
	AvatarURL string `json:"avatarUrl"`
}

type GoogleAuthRequest struct {
	Token string `json:"token"` // ID token from Google
}

// EmailVerificationRequest represents a request to send a verification email
type EmailVerificationRequest struct {
	Email    string `json:"email"`
	IsMobile bool   `json:"isMobile"` // Whether the request is coming from a mobile app login
}

// VerifyEmailRequest represents a request to verify an email token
type VerifyEmailRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

// GenerateSessionCodeResponse represents the response with session code
type GenerateSessionCodeResponse struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

// AppAutomationPreferences represents the user's automation preferences
type AppAutomationPreferences struct {
	UserID                     uuid.UUID `json:"userId"`
	WorkspaceID                uuid.UUID `json:"workspaceId"`
	AutoAssignSelf             bool      `json:"autoAssignSelf"`
	AssignSelfOnBranchCopy     bool      `json:"assignSelfOnBranchCopy"`
	MoveStoryToStartedOnBranch bool      `json:"moveStoryToStartedOnBranch"`
	OpenStoryInDialog          bool      `json:"openStoryInDialog"`
	CreatedAt                  time.Time `json:"createdAt"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// UpdateAutomationPreferencesRequest represents a request to update automation preferences
type UpdateAutomationPreferencesRequest struct {
	AutoAssignSelf             *bool `json:"autoAssignSelf,omitempty"`
	AssignSelfOnBranchCopy     *bool `json:"assignSelfOnBranchCopy,omitempty"`
	MoveStoryToStartedOnBranch *bool `json:"moveStoryToStartedOnBranch,omitempty"`
	OpenStoryInDialog          *bool `json:"openStoryInDialog,omitempty"`
}

func toAppUser(user users.CoreUser) AppUser {
	return AppUser{
		ID:                  user.ID,
		Username:            user.Username,
		Email:               user.Email,
		FullName:            user.FullName,
		AvatarURL:           user.AvatarURL,
		IsActive:            user.IsActive,
		HasSeenWalkthrough:  user.HasSeenWalkthrough,
		Timezone:            user.Timezone,
		LastLoginAt:         user.LastLoginAt,
		LastUsedWorkspaceID: user.LastUsedWorkspaceID,
		CreatedAt:           user.CreatedAt,
		UpdatedAt:           user.UpdatedAt,
		Token:               user.Token,
		Role:                user.Role,
	}
}

func toAppUsers(users []users.CoreUser) []AppUser {
	appUsers := make([]AppUser, len(users))
	for i, user := range users {
		appUsers[i] = toAppUser(user)
	}
	return appUsers
}

// Convert core automation preferences to app model
func toAppAutomationPreferences(prefs users.CoreAutomationPreferences) AppAutomationPreferences {
	return AppAutomationPreferences{
		UserID:                     prefs.UserID,
		WorkspaceID:                prefs.WorkspaceID,
		AutoAssignSelf:             prefs.AutoAssignSelf,
		AssignSelfOnBranchCopy:     prefs.AssignSelfOnBranchCopy,
		MoveStoryToStartedOnBranch: prefs.MoveStoryToStartedOnBranch,
		OpenStoryInDialog:          prefs.OpenStoryInDialog,
		CreatedAt:                  prefs.CreatedAt,
		UpdatedAt:                  prefs.UpdatedAt,
	}
}

// Convert update request to core update model
func toCoreUpdateAutomationPreferences(req UpdateAutomationPreferencesRequest) users.CoreUpdateAutomationPreferences {
	return users.CoreUpdateAutomationPreferences{
		AutoAssignSelf:             req.AutoAssignSelf,
		AssignSelfOnBranchCopy:     req.AssignSelfOnBranchCopy,
		MoveStoryToStartedOnBranch: req.MoveStoryToStartedOnBranch,
		OpenStoryInDialog:          req.OpenStoryInDialog,
	}
}

type AddUserMemoryRequest struct {
	Content string `json:"content" validate:"required"`
}

type UpdateUserMemoryRequest struct {
	Content string `json:"content" validate:"required"`
}

type AppUserMemoryItem struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toAppUserMemoryItem(mem users.CoreUserMemoryItem) AppUserMemoryItem {
	return AppUserMemoryItem{
		ID:          mem.ID,
		UserID:      mem.UserID,
		WorkspaceID: mem.WorkspaceID,
		Content:     mem.Content,
		CreatedAt:   mem.CreatedAt,
		UpdatedAt:   mem.UpdatedAt,
	}
}

func toAppUserMemoryItems(items []users.CoreUserMemoryItem) []AppUserMemoryItem {
	result := make([]AppUserMemoryItem, len(items))
	for i, item := range items {
		result[i] = toAppUserMemoryItem(item)
	}
	return result
}
