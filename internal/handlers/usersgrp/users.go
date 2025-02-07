package usersgrp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	users     *users.Service
	secretKey string
	// audit  *audit.Service
}

// New constructs a new users handlers instance.
func New(users *users.Service, secretKey string) *Handlers {
	return &Handlers{
		users:     users,
		secretKey: secretKey,
	}
}

// Login returns a user and a JWT token.
func (h *Handlers) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req LoginRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	user, err := h.users.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) {
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	expiresAt := time.Now().Add(time.Hour * 24 * 7)
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	user.Token = &tokenString
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// GetProfile returns the current user's profile
func (h *Handlers) GetProfile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// UpdateProfile updates the current user's profile
func (h *Handlers) UpdateProfile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req UpdateProfileRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := users.CoreUpdateUser{
		Username:  req.Username,
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
	}

	if err := h.users.UpdateUser(ctx, userID, updates); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Get updated user profile
	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// DeleteProfile soft deletes the current user's account
func (h *Handlers) DeleteProfile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.users.DeleteUser(ctx, userID); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// SwitchWorkspace updates the current user's active workspace
func (h *Handlers) SwitchWorkspace(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req SwitchWorkspaceRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if err := h.users.UpdateUserWorkspace(ctx, userID, req.WorkspaceID); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Get updated user profile
	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// List returns a list of users for a workspace.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	users, err := h.users.List(ctx, workspaceId)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	web.Respond(ctx, w, toAppUsers(users), http.StatusOK)
	return nil
}

// ResetPassword updates the current user's password
func (h *Handlers) ResetPassword(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req ResetPasswordRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if err := h.users.ResetPassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		if errors.Is(err, users.ErrInvalidPassword) {
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Register creates a new user account.
func (h *Handlers) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req RegisterRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	newUser := users.CoreNewUser{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
	}

	user, err := h.users.Register(ctx, newUser)
	if err != nil {
		if errors.Is(err, users.ErrEmailTaken) {
			return web.RespondError(ctx, w, err, http.StatusConflict)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusCreated)
}
