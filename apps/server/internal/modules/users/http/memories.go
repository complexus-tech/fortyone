package usershttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) AddUserMemory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.AddUserMemory")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req AddUserMemoryRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	newItem := users.NewUserMemoryItem{
		UserID:      userID,
		WorkspaceID: workspace.ID,
		Content:     req.Content,
	}

	item, err := h.users.AddUserMemory(ctx, newItem)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("adding user memory: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUserMemoryItem(item), http.StatusCreated)
}

// UpdateUserMemory updates a memory item.
func (h *Handlers) UpdateUserMemory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.UpdateUserMemory")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	idParam := web.Params(r, "id")
	memoryID, err := uuid.Parse(idParam)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req UpdateUserMemoryRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	update := users.UpdateUserMemoryItem{
		Content: &req.Content,
	}
	scope := users.UserMemoryScope{
		UserID:      userID,
		WorkspaceID: workspace.ID,
	}

	if err := h.users.UpdateUserMemory(ctx, memoryID, scope, update); err != nil {
		if errors.Is(err, users.ErrMemoryNotFound) {
			return web.RespondError(ctx, w, users.ErrMemoryNotFound, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, fmt.Errorf("updating user memory: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// DeleteUserMemory deletes a memory item.
func (h *Handlers) DeleteUserMemory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.DeleteUserMemory")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	idParam := web.Params(r, "id")
	memoryID, err := uuid.Parse(idParam)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	scope := users.UserMemoryScope{
		UserID:      userID,
		WorkspaceID: workspace.ID,
	}
	if err := h.users.DeleteUserMemory(ctx, memoryID, scope); err != nil {
		if errors.Is(err, users.ErrMemoryNotFound) {
			return web.RespondError(ctx, w, users.ErrMemoryNotFound, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, fmt.Errorf("deleting user memory: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// ListUserMemories retrieves all memory items for the user in the workspace.
func (h *Handlers) ListUserMemories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.ListUserMemories")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	items, err := h.users.ListUserMemories(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("listing user memories: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUserMemoryItems(items), http.StatusOK)
}
