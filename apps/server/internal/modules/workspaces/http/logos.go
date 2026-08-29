package workspaceshttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) UploadWorkspaceLogo(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	const multipartOverheadAllowance int64 = 1 << 20
	if err := web.ParseMultipartForm(writer, request, validate.MaxWorkspaceLogoSize+multipartOverheadAllowance); err != nil {
		return web.RespondError(ctx, writer, err, http.StatusBadRequest)
	}
	defer func() {
		if err := web.RemoveMultipartForm(request); err != nil {
			h.log.Warn(ctx, "failed to remove workspace logo upload temporary files", "error", err)
		}
	}()
	file, header, err := request.FormFile("image")
	if err != nil {
		return web.RespondError(ctx, writer, fmt.Errorf("get workspace logo file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()
	if err := h.workspaces.UploadWorkspaceLogo(ctx, workspace.ID, file, header, h.attachments); err != nil {
		switch {
		case errors.Is(err, validate.ErrFileTooLarge), errors.Is(err, validate.ErrInvalidFileType):
			return web.RespondError(ctx, writer, err, http.StatusBadRequest)
		case errors.Is(err, workspaces.ErrNotFound):
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		default:
			return fmt.Errorf("upload workspace logo: %w", err)
		}
	}
	return h.respondWithCurrentWorkspace(ctx, writer, workspace.ID)
}

func (h *Handlers) DeleteWorkspaceLogo(ctx context.Context, writer http.ResponseWriter, _ *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	if err := h.workspaces.DeleteWorkspaceLogo(ctx, workspace.ID, h.attachments); err != nil {
		if errors.Is(err, workspaces.ErrNotFound) {
			return web.RespondError(ctx, writer, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	return h.respondWithCurrentWorkspace(ctx, writer, workspace.ID)
}

func (h *Handlers) respondWithCurrentWorkspace(ctx context.Context, writer http.ResponseWriter, workspaceID uuid.UUID) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
	}
	workspace, err := h.workspaces.Get(ctx, workspaceID, userID)
	if err != nil {
		return web.RespondError(ctx, writer, err, http.StatusInternalServerError)
	}
	workspace.AvatarURL = h.resolveWorkspaceLogoURL(ctx, workspace.AvatarURL)
	return web.Respond(ctx, writer, toAppWorkspace(workspace), http.StatusOK)
}
