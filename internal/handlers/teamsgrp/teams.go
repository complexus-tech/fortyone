package teamsgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/pkg/web"
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
	teams, err := h.teams.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppTeams(teams), http.StatusOK)
	return nil
}
