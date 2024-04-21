package statesgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/pkg/web"
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

// List returns a list of states.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	states, err := h.states.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStates(states), http.StatusOK)
	return nil
}
