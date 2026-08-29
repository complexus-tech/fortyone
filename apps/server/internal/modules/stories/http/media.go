package storieshttp

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	storyMediaAccessTTL                        = 5 * time.Minute
	storyMediaMultipartOverheadAllowance int64 = 1 << 20
)

type storyMediaService interface {
	UploadStoryMedia(context.Context, multipart.File, *multipart.FileHeader, uuid.UUID, uuid.UUID, uuid.UUID) (attachments.FileInfo, error)
	ResolveStoryMediaAccessURL(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Duration) (attachments.FileInfo, error)
	DeleteStoryMedia(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

func (h *Handlers) UploadStoryMedia(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, storyID, err := storyMediaContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMediaContextHTTPStatus(err))
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := web.ParseMultipartForm(
		w,
		r,
		validate.MaxAttachmentSize+storyMediaMultipartOverheadAllowance,
	); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid upload request: %w", err), http.StatusBadRequest)
	}
	defer func() {
		if err := web.RemoveMultipartForm(r); err != nil {
			h.log.Warn(ctx, "failed to remove story media upload temporary files", "error", err)
		}
	}()

	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("story media file is required: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	fileInfo, err := h.storyMedia.UploadStoryMedia(ctx, file, header, userID, storyID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMediaHTTPStatus(err))
	}

	stableURL := storyMediaURL(workspace.Slug, storyID, fileInfo.ID)
	return web.Respond(ctx, w, toAppStoryMedia(fileInfo, stableURL), http.StatusCreated)
}

func (h *Handlers) ResolveStoryMedia(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, storyID, err := storyMediaContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMediaContextHTTPStatus(err))
	}
	attachmentID, err := storyMediaAttachmentID(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	fileInfo, err := h.storyMedia.ResolveStoryMediaAccessURL(
		ctx,
		storyID,
		attachmentID,
		workspace.ID,
		storyMediaAccessTTL,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMediaHTTPStatus(err))
	}

	redirectStoryMedia(w, r, fileInfo.URL)
	return nil
}

func (h *Handlers) DeleteStoryMedia(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, storyID, err := storyMediaContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, storyMediaContextHTTPStatus(err))
	}
	attachmentID, err := storyMediaAttachmentID(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if err := h.storyMedia.DeleteStoryMedia(ctx, storyID, attachmentID, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, storyMediaHTTPStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func storyMediaContext(ctx context.Context, r *http.Request) (mid.WorkspaceInfo, uuid.UUID, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return mid.WorkspaceInfo{}, uuid.Nil, err
	}
	storyID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil || storyID == uuid.Nil {
		return mid.WorkspaceInfo{}, uuid.Nil, ErrInvalidStoryID
	}
	return workspace, storyID, nil
}

func storyMediaAttachmentID(r *http.Request) (uuid.UUID, error) {
	attachmentID, err := uuid.Parse(web.Params(r, "attachmentId"))
	if err != nil || attachmentID == uuid.Nil {
		return uuid.Nil, attachments.ErrInvalidFile
	}
	return attachmentID, nil
}

func storyMediaURL(workspaceSlug string, storyID, attachmentID uuid.UUID) string {
	return fmt.Sprintf(
		"/workspaces/%s/stories/%s/media/%s",
		url.PathEscape(workspaceSlug),
		storyID,
		attachmentID,
	)
}

func redirectStoryMedia(w http.ResponseWriter, r *http.Request, accessURL string) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.Redirect(w, r, accessURL, http.StatusTemporaryRedirect)
}

func storyMediaContextHTTPStatus(err error) int {
	if errors.Is(err, ErrInvalidStoryID) {
		return http.StatusBadRequest
	}
	return http.StatusUnauthorized
}

func storyMediaHTTPStatus(err error) int {
	switch {
	case errors.Is(err, attachments.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, attachments.ErrInvalidFileType), errors.Is(err, attachments.ErrInvalidFile):
		return http.StatusBadRequest
	case errors.Is(err, attachments.ErrUnauthorized):
		return http.StatusForbidden
	case errors.Is(err, attachments.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
