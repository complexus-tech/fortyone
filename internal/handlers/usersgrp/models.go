package usersgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/google/uuid"
)

type AppUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	FullName    string    `json:"fullName"`
	AvatarURL   string    `json:"avatarUrl"`
	IsActive    bool      `json:"isActive"`
	LastLoginAt time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Token       *string   `json:"token,omitempty"`
}

func toAppUser(user users.CoreUser) AppUser {
	return AppUser{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		FullName:    user.FullName,
		AvatarURL:   user.AvatarURL,
		IsActive:    user.IsActive,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Token:       user.Token,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
