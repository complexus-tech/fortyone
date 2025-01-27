package objectivestatusgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	objectivestatus *objectivestatus.Service
	// audit  *audit.Service
}

// New constructs a new objective statuses handlers instance.
func New(objectivestatus *objectivestatus.Service) *Handlers {
	return &Handlers{
		objectivestatus: objectivestatus,
	}
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

// List returns a list of objective statuses.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}
	states, err := h.objectivestatus.List(ctx, workspaceId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppObjectiveStatuses(states), http.StatusOK)
	return nil
}
