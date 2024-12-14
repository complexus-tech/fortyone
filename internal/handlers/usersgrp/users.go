package usersgrp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/users"
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

	email := req.Email
	password := req.Password

	user, err := h.users.Login(ctx, email, password)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	expiresAt := time.Now().Add(time.Hour * 24 * 30)
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	user.Token = &tokenString
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
