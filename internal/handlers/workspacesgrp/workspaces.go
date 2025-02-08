package workspacesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	workspaces *workspaces.Service
	secretKey  string
	// audit  *audit.Service
}

// New constructs a new workspaces andlers instance.
func New(workspaces *workspaces.Service, secretKey string) *Handlers {
	return &Handlers{
		workspaces: workspaces,
		secretKey:  secretKey,
	}
}

// List returns a list of users for a workspace.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	workspaces, err := h.workspaces.List(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	web.Respond(ctx, w, toAppWorkspaces(workspaces), http.StatusOK)
	return nil
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Create")
	defer span.End()

	var input AppNewWorkspace
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace := workspaces.CoreWorkspace{
		Name:     input.Name,
		Slug:     input.Slug,
		TeamSize: input.TeamSize,
	}

	result, err := h.workspaces.Create(ctx, workspace)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Add creator as member
	if err := h.workspaces.AddMember(ctx, result.ID, userID, "admin"); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Fetch the workspace again to include the role
	result, err = h.workspaces.Get(ctx, result.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("workspace created.", trace.WithAttributes(
		attribute.String("workspace_id", result.ID.String()),
		attribute.String("user_id", userID.String()),
	))

	return web.Respond(ctx, w, toAppWorkspace(result), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Update")
	defer span.End()

	var input AppUpdateWorkspace
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspaceIDParam := web.Params(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	updates := workspaces.CoreWorkspace{
		Name: input.Name,
	}

	result, err := h.workspaces.Update(ctx, workspaceID, updates)
	if err != nil {
		if err.Error() == "workspace not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("workspace updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppWorkspace(result), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Delete")
	defer span.End()

	workspaceIDParam := web.Params(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	if err := h.workspaces.Delete(ctx, workspaceID); err != nil {
		if err.Error() == "workspace not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("workspace deleted.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) AddMember(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.AddMember")
	defer span.End()

	var input AppNewWorkspaceMember
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspaceIDParam := web.Params(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	// Default to member role if not provided
	role := input.Role
	if role == "" {
		role = "member"
	}

	if err := h.workspaces.AddMember(ctx, workspaceID, input.UserID, role); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("workspace member added.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", input.UserID.String()),
		attribute.String("role", role),
	))

	return web.Respond(ctx, w, nil, http.StatusCreated)
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIDParam := web.Params(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("not authenticated"), http.StatusUnauthorized)
	}

	workspace, err := h.workspaces.Get(ctx, workspaceID, userID)
	if err != nil {
		if err.Error() == "workspace not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppWorkspace(workspace), http.StatusOK)
}

func (h *Handlers) RemoveMember(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.RemoveMember")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	userIDParam := web.Params(r, "userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid user id"), http.StatusBadRequest)
	}

	if err := h.workspaces.RemoveMember(ctx, workspaceID, userID); err != nil {
		if err.Error() == "member not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("workspace member removed.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
