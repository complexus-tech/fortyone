package workspacesgrp

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	restrictedSlugs       = []string{"admin", "internal", "qa", "staging", "ops", "team", "complexus", "dev", "test", "prod", "staging", "development", "testing", "production", "staff", "hr", "finance", "legal", "marketing", "sales", "support", "it", "security", "engineering", "design", "product", "marketing", "sales", "support", "it", "security", "engineering", "design", "product", "auth"}
)

type Handlers struct {
	workspaces *workspaces.Service
	secretKey  string
	// audit  *audit.Service
}

// New constructs a new workspaces andlers instance.
func New(workspaces *workspaces.Service, teams *teams.Service, stories *stories.Service, statuses *states.Service, users *users.Service, objectivestatus *objectivestatus.Service, secretKey string) *Handlers {
	return &Handlers{
		workspaces: workspaces,
		secretKey:  secretKey,
	}
}

// List returns a list of workspaces for a user.
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

	// Add validation for restricted slugs
	for _, restrictedSlug := range restrictedSlugs {
		if input.Slug == restrictedSlug {
			return web.RespondError(ctx, w, errors.New("this workspace slug is restricted"), http.StatusBadRequest)
		}
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	cw := workspaces.CoreWorkspace{
		Name:     input.Name,
		Slug:     input.Slug,
		TeamSize: input.TeamSize,
	}

	workspace, err := h.workspaces.Create(ctx, cw, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	span.AddEvent("workspace created.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()),
		attribute.String("userId", userID.String()),
	))
	workspace.UserRole = "admin"
	return web.Respond(ctx, w, toAppWorkspace(workspace), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Update")
	defer span.End()

	var input AppUpdateWorkspace
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspaceIDParam := web.Params(r, "workspaceId")
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
		attribute.String("workspaceId", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppWorkspace(result), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Delete")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
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
		attribute.String("workspaceId", workspaceID.String()),
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

	workspaceIDParam := web.Params(r, "workspaceId")
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
		attribute.String("workspaceId", workspaceID.String()),
		attribute.String("userId", input.UserID.String()),
		attribute.String("role", role),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIDParam := web.Params(r, "workspaceId")
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

func (h *Handlers) CheckSlugAvailability(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.CheckSlugAvailability")
	defer span.End()

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		return web.RespondError(ctx, w, errors.New("slug is required"), http.StatusBadRequest)
	}

	// Validate slug format
	if len(slug) < 3 || len(slug) > 255 {
		return web.RespondError(ctx, w, errors.New("slug must be between 3 and 255 characters"), http.StatusBadRequest)
	}

	// Only allow lowercase letters, numbers, and hyphens
	if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(slug) {
		return web.RespondError(ctx, w, errors.New("slug can only contain lowercase letters, numbers, and hyphens"), http.StatusBadRequest)
	}

	available, err := h.workspaces.CheckSlugAvailability(ctx, slug)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	response := AppSlugAvailability{
		Available: available,
		Slug:      slug,
	}

	span.AddEvent("slug availability checked.", trace.WithAttributes(
		attribute.String("slug", slug),
		attribute.Bool("available", available),
	))

	return web.Respond(ctx, w, response, http.StatusOK)
}
