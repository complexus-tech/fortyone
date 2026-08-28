package storieshttp

import (
	"context"
	"net/http"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.Get")
	defer span.End()

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
	story, err := h.stories.Get(ctx, storyId, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	return h.respondStory(ctx, w, story, http.StatusOK)
}

func (h *Handlers) QueryByRef(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.QueryByRef")
	defer span.End()

	storyRef := web.Params(r, "ref")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}
	story, err := h.stories.QueryByRef(ctx, workspace.ID, storyRef)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	return h.respondStory(ctx, w, story, http.StatusOK)
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.List")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	filters, err := parseListStoryFilters(r)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	storyList, err := h.stories.List(ctx, workspace.ID, filters)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	return h.respondStories(ctx, w, storyList, http.StatusOK)
}
