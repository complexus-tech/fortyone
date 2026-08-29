package storieshttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handlers) GetAttachmentsForStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.GetAttachmentsForStory")
	defer span.End()

	storyIdParam := web.Params(r, "id")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	fileInfos, err := h.attachments.GetAttachmentsForStory(ctx, storyId, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, storyAttachmentStatus(err))
	}

	return web.Respond(ctx, w, fileInfos, http.StatusOK)
}

func (h *Handlers) DeleteAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.DeleteAttachment")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	storyID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
	}

	attachmentIDStr := web.Params(r, "attachmentId")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid attachment ID"), http.StatusBadRequest)
	}

	err = h.attachments.DeleteStoryAttachment(
		ctx,
		storyID,
		attachmentID,
		workspace.ID,
		userID,
		workspace.UserRole == string(mid.RoleAdmin),
	)
	if err != nil {
		return web.RespondError(ctx, w, err, storyAttachmentStatus(err))
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) UploadStoryAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, _ := mid.GetUserID(ctx)
	storyIdParam := web.Params(r, "id")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	const multipartOverheadAllowance = 1 << 20
	if err := web.ParseMultipartForm(w, r, validate.MaxAttachmentSize+multipartOverheadAllowance); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid upload request: %w", err), http.StatusBadRequest)
	}
	defer func() {
		if err := web.RemoveMultipartForm(r); err != nil {
			h.log.Warn(ctx, "failed to remove story attachment upload temporary files", "error", err)
		}
	}()

	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	fileInfo, err := h.attachments.UploadAndLinkToStory(ctx, file, header, userID, storyId, workspace.ID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrFileTooLarge):
			return web.RespondError(ctx, w, err, http.StatusRequestEntityTooLarge)
		case errors.Is(err, attachments.ErrInvalidFileType), errors.Is(err, attachments.ErrInvalidFile):
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		default:
			return fmt.Errorf("error uploading attachment: %w", err)
		}
	}

	return web.Respond(ctx, w, fileInfo, http.StatusCreated)
}

func (h *Handlers) CountInWorkspace(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.stories.CountInWorkspace")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	count, err := h.stories.CountInWorkspace(ctx, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	span.AddEvent("stories count retrieved", trace.WithAttributes(
		attribute.Int("stories.count", count),
	))

	return web.Respond(ctx, w, map[string]int{"count": count}, http.StatusOK)
}
