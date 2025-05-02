package workspacesgrp

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
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
	cache      *cache.Service
	log        *logger.Logger
	secretKey  string
}

// New constructs a new workspaces andlers instance.
func New(workspaces *workspaces.Service, teams *teams.Service, stories *stories.Service, statuses *states.Service, users *users.Service, objectivestatus *objectivestatus.Service, cacheService *cache.Service, log *logger.Logger, secretKey string) *Handlers {
	return &Handlers{
		workspaces: workspaces,
		cache:      cacheService,
		log:        log,
		secretKey:  secretKey,
	}
}

// List returns a list of workspaces for a user.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "workspacesgrp.handlers.List")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	// Try to get from cache first
	cacheKey := cache.WorkspacesListCacheKey(userID)
	var cachedWorkspaces []workspaces.CoreWorkspace

	if err := h.cache.Get(ctx, cacheKey, &cachedWorkspaces); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppWorkspaces(cachedWorkspaces), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	workspacesList, err := h.workspaces.List(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, workspacesList, cache.ListTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppWorkspaces(workspacesList), http.StatusOK)
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

	// Invalidate the user's workspaces list cache
	cacheKey := cache.WorkspacesListCacheKey(userID)
	if err := h.cache.Delete(ctx, cacheKey); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to delete cache", "key", cacheKey, "error", err)
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

	// Invalidate the workspace caches
	cacheKeys := cache.InvalidateWorkspaceKeys(workspaceID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			if err := h.cache.DeleteByPattern(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache pattern", "key", key, "error", err)
			}
		} else {
			// Handle exact key deletion
			if err := h.cache.Delete(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache", "key", key, "error", err)
			}
		}
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

	// Get the user ID before deleting the workspace
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	// Invalidate the workspace caches before deleting
	cacheKeys := cache.InvalidateWorkspaceKeys(workspaceID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			if err := h.cache.DeleteByPattern(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache pattern", "key", key, "error", err)
			}
		} else {
			// Handle exact key deletion
			if err := h.cache.Delete(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache", "key", key, "error", err)
			}
		}
	}

	// Also invalidate the user's workspaces list cache
	userCacheKey := cache.WorkspacesListCacheKey(userID)
	if err := h.cache.Delete(ctx, userCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", userCacheKey, "error", err)
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

	// Invalidate workspace members cache
	membersCacheKey := cache.WorkspaceMembersCacheKey(workspaceID)
	if err := h.cache.Delete(ctx, membersCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", membersCacheKey, "error", err)
	}

	// Invalidate user's workspaces list cache
	userCacheKey := cache.WorkspacesListCacheKey(input.UserID)
	if err := h.cache.Delete(ctx, userCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", userCacheKey, "error", err)
	}

	span.AddEvent("workspace member added.", trace.WithAttributes(
		attribute.String("workspaceId", workspaceID.String()),
		attribute.String("userId", input.UserID.String()),
		attribute.String("role", role),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Get returns a workspace by ID
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "workspacesgrp.handlers.Get")
	defer span.End()

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

	web.Respond(ctx, w, toAppWorkspace(workspace), http.StatusOK)
	return nil
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

	// Invalidate workspace members cache
	membersCacheKey := cache.WorkspaceMembersCacheKey(workspaceID)
	if err := h.cache.Delete(ctx, membersCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", membersCacheKey, "error", err)
	}

	// Invalidate user's workspaces list cache
	userCacheKey := cache.WorkspacesListCacheKey(userID)
	if err := h.cache.Delete(ctx, userCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", userCacheKey, "error", err)
	}

	span.AddEvent("workspace member removed.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) UpdateMemberRole(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.UpdateMemberRole")
	defer span.End()

	var input AppUpdateWorkspaceMemberRole
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

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

	if err := h.workspaces.UpdateMemberRole(ctx, workspaceID, userID, input.Role); err != nil {
		if err == workspaces.ErrNotFound {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Invalidate workspace members cache
	membersCacheKey := cache.WorkspaceMembersCacheKey(workspaceID)
	if err := h.cache.Delete(ctx, membersCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", membersCacheKey, "error", err)
	}

	span.AddEvent("workspace member role updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role", input.Role),
	))

	return web.Respond(ctx, w, nil, http.StatusOK)
}

// CheckSlugAvailability checks if a slug is available for use.
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

// GetWorkspaceSettings retrieves the settings for a workspace.
func (h *Handlers) GetWorkspaceSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.GetWorkspaceSettings")
	defer span.End()

	// Get workspace ID from path
	workspaceIDStr := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	// Try to get from cache first
	cacheKey := cache.WorkspaceSettingsCacheKey(workspaceID)
	var cachedSettings workspaces.CoreWorkspaceSettings

	if err := h.cache.Get(ctx, cacheKey, &cachedSettings); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppWorkspaceSettings(cachedSettings), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	settings, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, settings, cache.DetailTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	return web.Respond(ctx, w, toAppWorkspaceSettings(settings), http.StatusOK)
}

// UpdateWorkspaceSettings updates the settings for a workspace.
func (h *Handlers) UpdateWorkspaceSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.UpdateWorkspaceSettings")
	defer span.End()

	// Get workspace ID from path
	workspaceIDStr := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	// Check if user has access to the workspace
	workspace, err := h.workspaces.Get(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Check if user has admin role
	if workspace.UserRole != "admin" {
		return web.RespondError(ctx, w, errors.New("only workspace admins can update workspace settings"), http.StatusForbidden)
	}

	// Decode the request
	var input AppUpdateWorkspaceSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// First, ensure settings exist by getting or creating them
	currentSettings, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspaceID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Update the settings
	coreSettings := toCoreWorkspaceSettings(input, workspaceID, currentSettings)
	updatedSettings, err := h.workspaces.UpdateWorkspaceSettings(ctx, workspaceID, coreSettings)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Invalidate the workspace settings cache
	settingsCacheKey := cache.WorkspaceSettingsCacheKey(workspaceID)
	if err := h.cache.Delete(ctx, settingsCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", settingsCacheKey, "error", err)
	}

	span.AddEvent("workspace settings updated.", trace.WithAttributes(
		attribute.String("workspaceId", workspaceID.String()),
		attribute.String("userId", userID.String()),
	))

	return web.Respond(ctx, w, toAppWorkspaceSettings(updatedSettings), http.StatusOK)
}
