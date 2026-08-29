package workspaceshttp

import (
	"context"
	"errors"
	"net/http"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handlers) GetWorkspaceSettings(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.GetWorkspaceSettings")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	settings, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, writer, toAppWorkspaceSettings(settings), http.StatusOK)
}

func (h *Handlers) UpdateWorkspaceSettings(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.UpdateWorkspaceSettings")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	var input AppUpdateWorkspaceSettings
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	current, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	updated, err := h.workspaces.UpdateWorkspaceSettings(ctx, workspace.ID, toCoreWorkspaceSettings(input, workspace.ID, current))
	if err != nil {
		if errors.Is(err, workschedule.ErrInvalidWorkingDays) || errors.Is(err, workschedule.ErrInvalidWorkingHours) {
			return web.RespondError(ctx, writer, err, http.StatusBadRequest)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	span.AddEvent("workspace settings updated.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()), attribute.String("userId", userID.String()),
	))
	return web.Respond(ctx, writer, toAppWorkspaceSettings(updated), http.StatusOK)
}
