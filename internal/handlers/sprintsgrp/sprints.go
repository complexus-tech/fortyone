package sprintsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/sprints"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	sprints *sprints.Service
	// audit  *audit.Service
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

// New constructs a new sprints handlers instance.
func New(sprints *sprints.Service) *Handlers {
	return &Handlers{
		sprints: sprints,
	}
}

// List returns a list of sprints.
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

	sprints, err := h.sprints.List(ctx, workspaceId, filters)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppSprints(sprints), http.StatusOK)
	return nil
}
