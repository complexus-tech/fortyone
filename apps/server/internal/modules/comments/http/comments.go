package commentshttp

import (
	"context"
	"errors"
	"net/http"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type commentService interface {
	UpdateComment(ctx context.Context, command comments.UpdateCommentCommand) error
	DeleteComment(ctx context.Context, scope comments.AuthorScope) error
}

type Handlers struct {
	comments    commentService
	workspaceID func(context.Context) (uuid.UUID, error)
	actor       func(context.Context) (platformauth.Actor, error)
}

func New(commentsService commentService) *Handlers {
	return &Handlers{
		comments: commentsService,
		workspaceID: func(ctx context.Context) (uuid.UUID, error) {
			workspace, err := mid.GetWorkspace(ctx)
			return workspace.ID, err
		},
		actor: mid.GetActor,
	}
}

func (h *Handlers) UpdateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, err := h.workspaceID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actor, err := h.actor(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	commentID, err := web.UUIDPathParameter(r, "id")
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidCommentID, http.StatusBadRequest)
		return nil
	}

	var uc UpdateComment
	if err := web.Decode(r, &uc); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	command := comments.UpdateCommentCommand{
		Scope: comments.AuthorScope{
			CommentID:   commentID,
			WorkspaceID: workspaceID,
			Actor:       actor,
		},
		Content:          uc.Content,
		MentionedUserIDs: uc.Mentions,
	}
	if err := h.comments.UpdateComment(ctx, command); err != nil {
		web.RespondError(ctx, w, err, commentMutationStatus(err))
		return nil
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) DeleteComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, err := h.workspaceID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actor, err := h.actor(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	commentID, err := web.UUIDPathParameter(r, "id")
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidCommentID, http.StatusBadRequest)
		return nil
	}

	scope := comments.AuthorScope{
		CommentID:   commentID,
		WorkspaceID: workspaceID,
		Actor:       actor,
	}
	if err := h.comments.DeleteComment(ctx, scope); err != nil {
		web.RespondError(ctx, w, err, commentMutationStatus(err))
		return nil
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func commentMutationStatus(err error) int {
	switch {
	case errors.Is(err, comments.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, comments.ErrInvalidMention):
		return http.StatusBadRequest
	case errors.Is(err, comments.ErrInvalidComment):
		return http.StatusBadRequest
	case errors.Is(err, comments.ErrForbidden):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
