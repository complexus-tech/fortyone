package usersgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	users *users.Service
	// audit  *audit.Service
}

// New constructs a new users handlers instance.
func New(users *users.Service) *Handlers {
	return &Handlers{
		users: users,
	}
}

// Login returns a user.
func (h *Handlers) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	email := web.Params(r, "email")
	password := web.Params(r, "password")

	user, err := h.users.Login(ctx, email, password)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppUser(user), http.StatusOK)
	return nil
}
