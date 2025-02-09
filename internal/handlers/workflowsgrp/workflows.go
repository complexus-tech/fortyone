package workflowsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/workflows"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidWorkflowID  = errors.New("workflow id is not in its proper form")
	ErrInvalidTeamID      = errors.New("team id is not in its proper form")
)

type Handlers struct {
	workflows *workflows.Service
}

func New(workflows *workflows.Service) *Handlers {
	return &Handlers{
		workflows: workflows,
	}
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	var req NewWorkflow
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workflow, err := h.workflows.Create(ctx, workspaceId, toCoreNewWorkflow(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppWorkflow(workflow), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workflowIdParam := web.Params(r, "workflowId")

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	workflowId, err := uuid.Parse(workflowIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkflowID, http.StatusBadRequest)
	}

	var req UpdateWorkflow
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workflow, err := h.workflows.Update(ctx, workspaceId, workflowId, toCoreUpdateWorkflow(req))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppWorkflow(workflow), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workflowIdParam := web.Params(r, "workflowId")

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	workflowId, err := uuid.Parse(workflowIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkflowID, http.StatusBadRequest)
	}

	if err := h.workflows.Delete(ctx, workspaceId, workflowId); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	workflows, err := h.workflows.List(ctx, workspaceId)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppWorkflows(workflows), http.StatusOK)
}

func (h *Handlers) ListByTeam(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	teamIdParam := web.Params(r, "teamId")

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	teamId, err := uuid.Parse(teamIdParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	workflows, err := h.workflows.ListByTeam(ctx, workspaceId, teamId)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppWorkflows(workflows), http.StatusOK)
}
