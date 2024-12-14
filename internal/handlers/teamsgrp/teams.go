package teamsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	teams *teams.Service
	// audit  *audit.Service
}

// New constructs a new teams handlers instance.
func New(teams *teams.Service) *Handlers {
	return &Handlers{
		teams: teams,
	}
}

// List returns a list of teams.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	teams, err := h.teams.List(ctx, workspaceId)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	web.Respond(ctx, w, toAppTeams(teams), http.StatusOK)
	return nil
}
