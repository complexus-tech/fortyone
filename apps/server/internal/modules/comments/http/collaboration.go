package commentshttp

import (
	"context"
	"net/http"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type collaborationService interface {
	GetStoryState(context.Context, comments.StoryScope) (comments.StoryState, error)
	SetReaction(context.Context, comments.ReactionCommand) (comments.CommentState, error)
	SetResolution(context.Context, comments.ResolutionCommand) (comments.CommentState, error)
	MarkRead(context.Context, comments.ReadCommand) (comments.StoryState, error)
}

type collaborationHandlers struct {
	service     collaborationService
	workspaceID func(context.Context) (uuid.UUID, error)
	actor       func(context.Context) (platformauth.Actor, error)
}

func newCollaborationHandlers(service collaborationService) *collaborationHandlers {
	return &collaborationHandlers{
		service: service,
		workspaceID: func(ctx context.Context) (uuid.UUID, error) {
			workspace, err := mid.GetWorkspace(ctx)
			return workspace.ID, err
		},
		actor: mid.GetActor,
	}
}

func (h *collaborationHandlers) scope(ctx context.Context) (comments.AuthorScope, error) {
	workspaceID, err := h.workspaceID(ctx)
	if err != nil {
		return comments.AuthorScope{}, err
	}
	actor, err := h.actor(ctx)
	if err != nil {
		return comments.AuthorScope{}, err
	}
	return comments.AuthorScope{WorkspaceID: workspaceID, Actor: actor}, nil
}

func (h *collaborationHandlers) GetState(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.scope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := web.UUIDPathParameter(r, "storyId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	state, err := h.service.GetStoryState(ctx, comments.StoryScope{StoryID: storyID, WorkspaceID: scope.WorkspaceID, Actor: scope.Actor})
	if err != nil {
		return web.RespondError(ctx, w, err, commentMutationStatus(err))
	}
	return web.Respond(ctx, w, state, http.StatusOK)
}

func (h *collaborationHandlers) MarkRead(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.scope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := web.UUIDPathParameter(r, "storyId")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input struct {
		ObservedAt time.Time  `json:"observedAt"`
		ThreadID   *uuid.UUID `json:"threadId"`
	}
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	state, err := h.service.MarkRead(ctx, comments.ReadCommand{
		Scope: comments.StoryScope{StoryID: storyID, WorkspaceID: scope.WorkspaceID, Actor: scope.Actor}, ObservedAt: input.ObservedAt,
		ThreadID: input.ThreadID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, commentMutationStatus(err))
	}
	return web.Respond(ctx, w, state, http.StatusOK)
}

func (h *collaborationHandlers) SetReaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.scope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	scope.CommentID, err = web.UUIDPathParameter(r, "id")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input struct {
		Emoji string `json:"emoji"`
	}
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	state, err := h.service.SetReaction(ctx, comments.ReactionCommand{
		Scope: scope, Emoji: input.Emoji, Active: r.Method == http.MethodPut,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, commentMutationStatus(err))
	}
	return web.Respond(ctx, w, state, http.StatusOK)
}

func (h *collaborationHandlers) SetResolution(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.scope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	scope.CommentID, err = web.UUIDPathParameter(r, "id")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input struct {
		Resolved *bool `json:"resolved"`
	}
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if input.Resolved == nil {
		return web.RespondError(ctx, w, comments.ErrInvalidComment, http.StatusBadRequest)
	}
	state, err := h.service.SetResolution(ctx, comments.ResolutionCommand{Scope: scope, Resolved: *input.Resolved})
	if err != nil {
		return web.RespondError(ctx, w, err, commentMutationStatus(err))
	}
	return web.Respond(ctx, w, state, http.StatusOK)
}
