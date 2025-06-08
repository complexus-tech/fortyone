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

// Running returns a list of running sprints.
func (h *Handlers) Running(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	sprints, err := h.sprints.Running(ctx, workspaceId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppSprints(sprints), http.StatusOK)
	return nil
}

// GetByID returns a single sprint by ID without stats.
func (h *Handlers) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	sprintIdParam := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(sprintIdParam)
	if err != nil {
		return errors.New("sprint id is not in its proper form")
	}

	sprint, err := h.sprints.GetByID(ctx, sprintId, workspaceId)
	if err != nil {
		return err
	}

	web.Respond(ctx, w, toAppSprint(sprint), http.StatusOK)
	return nil
}

// Create creates a new sprint.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	var app AppNewSprint
	if err := web.Decode(r, &app); err != nil {
		return err
	}

	sprint := sprints.CoreNewSprint{
		Name:      app.Name,
		Goal:      app.Goal,
		Objective: app.Objective,
		Team:      app.Team,
		Workspace: workspaceId,
		StartDate: app.StartDate,
		EndDate:   app.EndDate,
	}

	result, err := h.sprints.Create(ctx, sprint)
	if err != nil {
		return err
	}

	web.Respond(ctx, w, toAppSprints([]sprints.CoreSprint{result})[0], http.StatusCreated)
	return nil
}

// Update updates an existing sprint.
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	sprintIdParam := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(sprintIdParam)
	if err != nil {
		return errors.New("sprint id is not in its proper form")
	}

	var app AppUpdateSprint
	if err := web.Decode(r, &app); err != nil {
		return err
	}

	// First get the sprint to verify it belongs to the workspace
	existingSprints, err := h.sprints.List(ctx, workspaceId, map[string]any{"sprint_id": sprintId})
	if err != nil {
		return err
	}
	if len(existingSprints) == 0 {
		return errors.New("sprint not found in workspace")
	}

	sprint := sprints.CoreUpdateSprint{
		Name:      app.Name,
		Goal:      app.Goal,
		Objective: app.Objective,
		StartDate: app.StartDate,
		EndDate:   app.EndDate,
	}

	result, err := h.sprints.Update(ctx, sprintId, workspaceId, sprint)
	if err != nil {
		return err
	}

	web.Respond(ctx, w, toAppSprints([]sprints.CoreSprint{result})[0], http.StatusOK)
	return nil
}

// Delete removes a sprint.
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	sprintIdParam := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(sprintIdParam)
	if err != nil {
		return errors.New("sprint id is not in its proper form")
	}

	if err := h.sprints.Delete(ctx, sprintId, workspaceId); err != nil {
		return err
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}
