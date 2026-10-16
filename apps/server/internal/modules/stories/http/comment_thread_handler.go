package storieshttp

import (
	"context"
	"net/http"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) GetCommentThread(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyID, err := web.UUIDPathParameter(r, "id")
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
	}
	commentID, err := web.UUIDPathParameter(r, "commentId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	thread, err := h.stories.GetCommentThread(ctx, commentID, storyID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, storyReadStatus(err))
	}
	return h.respondComment(ctx, w, thread, http.StatusOK)
}
