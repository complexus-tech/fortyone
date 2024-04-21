package sprintsgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	sprints *sprints.Service
	// audit  *audit.Service
}

// New constructs a new sprints handlers instance.
func New(sprints *sprints.Service) *Handlers {
	return &Handlers{
		sprints: sprints,
	}
}

// List returns a list of sprints.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	sprints, err := h.sprints.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppSprints(sprints), http.StatusOK)
	return nil
}
