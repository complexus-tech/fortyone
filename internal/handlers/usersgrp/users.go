package usersgrp

import (
	"context"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/golang-jwt/jwt/v5"
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

	expiresAt := time.Now().Add(time.Hour * 24)
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
