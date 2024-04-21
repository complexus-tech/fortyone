package objectivesgrp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	objectives *objectives.Service
	// audit  *audit.Service
}

// New constructs a new objectives handlers instance.
func New(objectives *objectives.Service) *Handlers {
	return &Handlers{
		objectives: objectives,
	}
}

// List returns a list of objectives.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectives, err := h.objectives.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppObjectives(objectives), http.StatusOK)
	return nil
}
