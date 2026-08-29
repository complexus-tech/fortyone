package workspaceshttp

import (
	"context"
	"errors"
	"net/http"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handlers) AddMember(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.AddMember")
	defer span.End()
	var input AppNewWorkspaceMember
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	role := input.Role
	if role == "" {
		role = string(authorization.WorkspaceRoleMember)
	}
	if err := h.workspaces.AddMember(ctx, workspace.ID, input.UserID, role); err != nil {
		if errors.Is(err, authorization.ErrInvalidWorkspaceRole) {
			return web.RespondError(ctx, writer, err, http.StatusBadRequest)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	span.AddEvent("workspace member added.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()), attribute.String("userId", input.UserID.String()),
		attribute.String("role", role),
	))
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (h *Handlers) RemoveMember(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.RemoveMember")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	userID, err := uuid.Parse(web.Params(request, "userId"))
	if err != nil {
		return web.RespondError(ctx, writer, errors.New("invalid user id"), http.StatusBadRequest)
	}
	if err := h.workspaces.RemoveMember(ctx, workspace.ID, userID); err != nil {
		if errors.Is(err, workspaces.ErrMemberNotFound) {
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	span.AddEvent("workspace member removed.", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()), attribute.String("user_id", userID.String()),
	))
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (h *Handlers) UpdateMemberRole(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.UpdateMemberRole")
	defer span.End()
	var input AppUpdateWorkspaceMemberRole
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	userID, err := uuid.Parse(web.Params(request, "userId"))
	if err != nil {
		return web.RespondError(ctx, writer, errors.New("invalid user id"), http.StatusBadRequest)
	}
	if err := h.workspaces.UpdateMemberRole(ctx, workspace.ID, userID, input.Role); err != nil {
		switch {
		case errors.Is(err, workspaces.ErrMemberNotFound):
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		case errors.Is(err, authorization.ErrInvalidWorkspaceRole):
			return web.RespondError(ctx, writer, err, http.StatusBadRequest)
		default:
			return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
		}
	}
	span.AddEvent("workspace member role updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()), attribute.String("user_id", userID.String()),
		attribute.String("role", input.Role),
	))
	return web.Respond(ctx, writer, nil, http.StatusOK)
}
