package workspaceshttp

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

func (h *Handlers) List(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	ctx, span := web.AddSpan(ctx, "workspaceshttp.handlers.List")
	defer span.End()
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	workspaceList, err := h.workspaces.List(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	h.resolveWorkspaceLogos(ctx, workspaceList)
	return web.Respond(ctx, writer, toAppWorkspaces(workspaceList), http.StatusOK)
}

func (h *Handlers) GetPortal(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "workspaceshttp.handlers.GetPortal")
	defer span.End()
	slug := web.Params(request, "workspaceSlug")
	if slug == "" {
		slug = web.Params(request, "portalSlug")
	}
	if slug == "" {
		return web.RespondError(ctx, writer, errors.New("workspace slug is required"), http.StatusBadRequest)
	}
	workspace, err := h.workspaces.GetPublicBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	workspace.AvatarURL = h.resolveWorkspaceLogoURL(ctx, workspace.AvatarURL)
	return web.Respond(ctx, writer, toAppPortalWorkspace(workspace), http.StatusOK)
}

func (h *Handlers) Create(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Create")
	defer span.End()
	var input AppNewWorkspace
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	workspace, err := h.workspaces.CreateWithOptions(ctx, workspaces.CoreWorkspace{
		Name: input.Name, Slug: input.Slug, TeamSize: input.TeamSize,
	}, userID, workspaces.CreationOptions{
		IncludeExamples: input.IncludeExamples, WorkType: input.WorkType,
	})
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	span.AddEvent("workspace created.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()), attribute.String("userId", userID.String()),
	))
	workspace.UserRole = "admin"
	workspace.AvatarURL = h.resolveWorkspaceLogoURL(ctx, workspace.AvatarURL)
	return web.Respond(ctx, writer, toAppWorkspace(workspace), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Update")
	defer span.End()
	var input AppUpdateWorkspace
	if err := web.Decode(request, &input); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	updated, err := h.workspaces.Update(ctx, workspace.ID, workspaces.CoreWorkspace{Name: input.Name})
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	span.AddEvent("workspace updated.", trace.WithAttributes(attribute.String("workspaceId", workspace.ID.String())))
	updated.AvatarURL = h.resolveWorkspaceLogoURL(ctx, updated.AvatarURL)
	return web.Respond(ctx, writer, toAppWorkspace(updated), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Delete")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	if err := h.workspaces.Delete(ctx, workspace.ID, userID); err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	span.AddEvent("workspace scheduled for deletion.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()), attribute.String("deletedBy", userID.String()),
	))
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (h *Handlers) Restore(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Restore")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	if err := h.workspaces.Restore(ctx, workspace.ID, userID); err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	span.AddEvent("workspace restored.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()), attribute.String("restoredBy", userID.String()),
	))
	return web.Respond(ctx, writer, nil, http.StatusNoContent)
}

func (h *Handlers) Get(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	ctx, span := web.AddSpan(ctx, "workspaceshttp.handlers.Get")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, errors.New("not authenticated"), http.StatusUnauthorized)
	}
	result, err := h.workspaces.Get(ctx, workspace.ID, userID)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	result.AvatarURL = h.resolveWorkspaceLogoURL(ctx, result.AvatarURL)
	return web.Respond(ctx, writer, toAppWorkspace(result), http.StatusOK)
}

func (h *Handlers) CheckSlugAvailability(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.CheckSlugAvailability")
	defer span.End()
	slug, err := parseSlugAvailabilityQuery(request.URL.Query())
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	available, err := h.workspaces.CheckSlugAvailability(ctx, slug)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	span.AddEvent("slug availability checked.", trace.WithAttributes(
		attribute.String("slug", slug), attribute.Bool("available", available),
	))
	return web.Respond(ctx, writer, AppSlugAvailability{Available: available, Slug: slug}, http.StatusOK)
}
