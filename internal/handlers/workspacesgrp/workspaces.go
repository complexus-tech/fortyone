package workspacesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	workspaces *workspaces.Service
	secretKey  string
	// audit  *audit.Service
}

// New constructs a new workspaces andlers instance.
func New(workspaces *workspaces.Service, secretKey string) *Handlers {
	return &Handlers{
		workspaces: workspaces,
		secretKey:  secretKey,
	}
}

// List returns a list of users for a workspace.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	workspaces, err := h.workspaces.List(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	web.Respond(ctx, w, toAppWorkspaces(workspaces), http.StatusOK)
	return nil
}
