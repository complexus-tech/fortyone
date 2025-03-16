package usersgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/google/uuid"
)

type AppUser struct {
	ID                  uuid.UUID  `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	FullName            string     `json:"fullName"`
	AvatarURL           string     `json:"avatarUrl"`
	IsActive            bool       `json:"isActive"`
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
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	AvatarURL string `json:"avatarUrl"`
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
	Email string `json:"email"`
}

// VerifyEmailRequest represents a request to verify an email token
type VerifyEmailRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func toAppUser(user users.CoreUser) AppUser {
	return AppUser{
		ID:                  user.ID,
		Username:            user.Username,
		Email:               user.Email,
		FullName:            user.FullName,
		AvatarURL:           user.AvatarURL,
		IsActive:            user.IsActive,
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
