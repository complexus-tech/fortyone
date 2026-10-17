package commentshttp

import (
	"context"
	"net/http"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type commentSeenService interface {
	MarkSeen(context.Context, comments.StoryScope, uuid.UUID) error
}

func (h *collaborationHandlers) MarkSeen(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.scope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := web.UUIDPathParameter(r, "storyId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	commentID, err := web.UUIDPathParameter(r, "commentId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	service, ok := h.service.(commentSeenService)
	if !ok {
		return web.RespondError(ctx, w, comments.ErrInvalidComment, http.StatusInternalServerError)
	}
	if err := service.MarkSeen(ctx, comments.StoryScope{StoryID: storyID, WorkspaceID: scope.WorkspaceID, Actor: scope.Actor}, commentID); err != nil {
		return web.RespondError(ctx, w, err, commentMutationStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
