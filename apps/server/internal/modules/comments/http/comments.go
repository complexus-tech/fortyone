package commentshttp

import (
	"context"
	"net/http"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	comments *comments.Service
	log      *logger.Logger
}

func New(log *logger.Logger, comments *comments.Service) *Handlers {
	return &Handlers{
		comments: comments,
		log:      log,
	}
}

func (h *Handlers) UpdateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	commentIDParam := web.Params(r, "id")
	commentID, err := uuid.Parse(commentIDParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidCommentID, http.StatusBadRequest)
		return nil
	}

	var uc UpdateComment
	if err := web.Decode(r, &uc); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.comments.UpdateComment(ctx, commentID, uc.Content, uc.Mentions); err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) DeleteComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	commentIDParam := web.Params(r, "id")
	commentID, err := uuid.Parse(commentIDParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidCommentID, http.StatusBadRequest)
		return nil
	}

	if err := h.comments.DeleteComment(ctx, commentID); err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
