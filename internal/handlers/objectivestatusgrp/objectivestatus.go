package objectivestatusgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidStatusID    = errors.New("objective status id is not in its proper form")
)

type Handlers struct {
	objectiveStatus *objectivestatus.Service
	// audit  *audit.Service
}

// New constructs a new objective status handlers instance.
func New(objectiveStatus *objectivestatus.Service) *Handlers {
	return &Handlers{
		objectiveStatus: objectiveStatus,
	}
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	var req NewObjectiveStatus
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	status, err := h.objectiveStatus.Create(ctx, workspaceId, toCoreNewObjectiveStatus(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppObjectiveStatus(status), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	statusIdParam := web.Params(r, "statusId")

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	statusId, err := uuid.Parse(statusIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStatusID, http.StatusBadRequest)
	}

	var req UpdateObjectiveStatus
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	status, err := h.objectiveStatus.Update(ctx, workspaceId, statusId, toCoreUpdateObjectiveStatus(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppObjectiveStatus(status), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	statusIdParam := web.Params(r, "statusId")

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	statusId, err := uuid.Parse(statusIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStatusID, http.StatusBadRequest)
	}

	if err := h.objectiveStatus.Delete(ctx, workspaceId, statusId); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// List returns a list of objective statuses.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	statuses, err := h.objectiveStatus.List(ctx, workspaceId)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppObjectiveStatuses(statuses), http.StatusOK)
}
