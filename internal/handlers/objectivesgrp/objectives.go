package objectivesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/objectives"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	objectives *objectives.Service
	// audit  *audit.Service
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

// New constructs a new objectives handlers instance.
func New(objectives *objectives.Service) *Handlers {
	return &Handlers{
		objectives: objectives,
	}
}

// List returns a list of objectives.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	objectives, err := h.objectives.List(ctx, workspaceId, filters)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppObjectives(objectives), http.StatusOK)
	return nil
}
