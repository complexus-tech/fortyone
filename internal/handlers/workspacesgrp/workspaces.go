package workspacesgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"slices"

	"github.com/complexus-tech/projects-api/internal/core/objectivestatus"
	"github.com/complexus-tech/projects-api/internal/core/states"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/validate"
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
	workspaces    *workspaces.Service
	cache         *cache.Service
	log           *logger.Logger
	secretKey     string
	subscriptions *subscriptions.Service
	attachments   workspaces.AttachmentsService
}

func New(workspaces *workspaces.Service, teams *teams.Service, stories *stories.Service, statuses *states.Service, users *users.Service, objectivestatus *objectivestatus.Service, subscriptions *subscriptions.Service, cacheService *cache.Service, log *logger.Logger, secretKey string, attachments workspaces.AttachmentsService) *Handlers {
	return &Handlers{
		workspaces:    workspaces,
		cache:         cacheService,
		log:           log,
		secretKey:     secretKey,
		subscriptions: subscriptions,
		attachments:   attachments,
	}
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "workspacesgrp.handlers.List")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspacesList, err := h.workspaces.List(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
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

	if slices.Contains(restrictedSlugs, input.Slug) {
		return web.RespondError(ctx, w, errors.New("this workspace slug is restricted"), http.StatusBadRequest)
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

	cacheKey := cache.WorkspacesListCacheKey(userID)
	if err := h.cache.Delete(ctx, cacheKey); err != nil {
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

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	updates := workspaces.CoreWorkspace{
		Name: input.Name,
	}

	result, err := h.workspaces.Update(ctx, workspace.ID, updates)
	if err != nil {
		if err.Error() == "workspace not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	cacheKeys := cache.InvalidateWorkspaceKeys(workspace.ID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			if err := h.cache.DeleteByPattern(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache pattern", "key", key, "error", err)
			}
		} else {
			if err := h.cache.Delete(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache", "key", key, "error", err)
			}
		}
	}

	span.AddEvent("workspace updated.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()),
	))

	return web.Respond(ctx, w, toAppWorkspace(result), http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.Delete")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	cacheKeys := cache.InvalidateWorkspaceKeys(workspace.ID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			if err := h.cache.DeleteByPattern(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache pattern", "key", key, "error", err)
			}
		} else {
			if err := h.cache.Delete(ctx, key); err != nil {
				h.log.Error(ctx, "failed to delete cache", "key", key, "error", err)
			}
		}
	}

	userCacheKey := cache.WorkspacesListCacheKey(userID)
	if err := h.cache.Delete(ctx, userCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", userCacheKey, "error", err)
	}

	err = h.subscriptions.CancelSubscription(ctx, workspace.ID)
	if err != nil {
		span.RecordError(err)
		h.log.Error(ctx, "failed to cancel subscription", "workspaceId", workspace.ID.String(), "error", err)
	}

	if err := h.workspaces.Delete(ctx, workspace.ID); err != nil {
		if err.Error() == "workspace not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("workspace deleted.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()),
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

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	role := input.Role
	if role == "" {
		role = "member"
	}

	if err := h.workspaces.AddMember(ctx, workspace.ID, input.UserID, role); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	membersCacheKey := cache.WorkspaceMembersCacheKey(workspace.ID)
	if err := h.cache.Delete(ctx, membersCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", membersCacheKey, "error", err)
	}

	userCacheKey := cache.WorkspacesListCacheKey(input.UserID)
	if err := h.cache.Delete(ctx, userCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", userCacheKey, "error", err)
	}

	span.AddEvent("workspace member added.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()),
		attribute.String("userId", input.UserID.String()),
		attribute.String("role", role),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "workspacesgrp.handlers.Get")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("not authenticated"), http.StatusUnauthorized)
	}

	ws, err := h.workspaces.Get(ctx, workspace.ID, userID)
	if err != nil {
		if err.Error() == "workspace not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	web.Respond(ctx, w, toAppWorkspace(ws), http.StatusOK)
	return nil
}

func (h *Handlers) RemoveMember(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.RemoveMember")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userIDParam := web.Params(r, "userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid user id"), http.StatusBadRequest)
	}

	if err := h.workspaces.RemoveMember(ctx, workspace.ID, userID); err != nil {
		if errors.Is(err, workspaces.ErrMemberNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	membersCacheKey := cache.WorkspaceMembersCacheKey(workspace.ID)
	if err := h.cache.Delete(ctx, membersCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", membersCacheKey, "error", err)
	}

	userCacheKey := cache.WorkspacesListCacheKey(userID)
	if err := h.cache.Delete(ctx, userCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", userCacheKey, "error", err)
	}

	span.AddEvent("workspace member removed.", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()),
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

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userIDParam := web.Params(r, "userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid user id"), http.StatusBadRequest)
	}

	if err := h.workspaces.UpdateMemberRole(ctx, workspace.ID, userID, input.Role); err != nil {
		if err == workspaces.ErrNotFound {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	membersCacheKey := cache.WorkspaceMembersCacheKey(workspace.ID)
	if err := h.cache.Delete(ctx, membersCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", membersCacheKey, "error", err)
	}

	span.AddEvent("workspace member role updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", userID.String()),
		attribute.String("role", input.Role),
	))

	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) CheckSlugAvailability(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.CheckSlugAvailability")
	defer span.End()

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		return web.RespondError(ctx, w, errors.New("slug is required"), http.StatusBadRequest)
	}

	if len(slug) < 3 || len(slug) > 255 {
		return web.RespondError(ctx, w, errors.New("slug must be between 3 and 255 characters"), http.StatusBadRequest)
	}

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

func (h *Handlers) GetWorkspaceSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.GetWorkspaceSettings")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	cacheKey := cache.WorkspaceSettingsCacheKey(workspace.ID)
	var cachedSettings workspaces.CoreWorkspaceSettings

	if err := h.cache.Get(ctx, cacheKey, &cachedSettings); err == nil {
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppWorkspaceSettings(cachedSettings), http.StatusOK)
		return nil
	}

	settings, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if err := h.cache.Set(ctx, cacheKey, settings, cache.DetailTTL); err != nil {
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	return web.Respond(ctx, w, toAppWorkspaceSettings(settings), http.StatusOK)
}

func (h *Handlers) UpdateWorkspaceSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.workspaces.UpdateWorkspaceSettings")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if workspace.UserRole != "admin" {
		return web.RespondError(ctx, w, errors.New("only workspace admins can update workspace settings"), http.StatusForbidden)
	}

	var input AppUpdateWorkspaceSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	currentSettings, err := h.workspaces.GetOrCreateWorkspaceSettings(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	coreSettings := toCoreWorkspaceSettings(input, workspace.ID, currentSettings)
	updatedSettings, err := h.workspaces.UpdateWorkspaceSettings(ctx, workspace.ID, coreSettings)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	settingsCacheKey := cache.WorkspaceSettingsCacheKey(workspace.ID)
	if err := h.cache.Delete(ctx, settingsCacheKey); err != nil {
		h.log.Error(ctx, "failed to delete cache", "key", settingsCacheKey, "error", err)
	}

	span.AddEvent("workspace settings updated.", trace.WithAttributes(
		attribute.String("workspaceId", workspace.ID.String()),
		attribute.String("userId", userID.String()),
	))

	return web.Respond(ctx, w, toAppWorkspaceSettings(updatedSettings), http.StatusOK)
}

func (h *Handlers) UploadWorkspaceLogo(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := r.ParseMultipartForm(6 << 20); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting image file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	err = h.workspaces.UploadWorkspaceLogo(ctx, workspace.ID, file, header, h.attachments)
	if err != nil {
		switch {
		case errors.Is(err, validate.ErrFileTooLarge), errors.Is(err, validate.ErrInvalidFileType):
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		case errors.Is(err, workspaces.ErrNotFound):
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		default:
			return fmt.Errorf("error uploading workspace logo: %w", err)
		}
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	ws, err := h.workspaces.Get(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppWorkspace(ws), http.StatusOK)
}

func (h *Handlers) DeleteWorkspaceLogo(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	err = h.workspaces.DeleteWorkspaceLogo(ctx, workspace.ID, h.attachments)
	if err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	ws, err := h.workspaces.Get(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppWorkspace(ws), http.StatusOK)
}
