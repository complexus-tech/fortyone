package statesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	states *states.Service
	// audit  *audit.Service
}

// New constructs a new states handlers instance.
func New(states *states.Service) *Handlers {
	return &Handlers{
		states: states,
	}
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

// List returns a list of states.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}
	states, err := h.states.List(ctx, workspaceId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStates(states), http.StatusOK)
	return nil
}
