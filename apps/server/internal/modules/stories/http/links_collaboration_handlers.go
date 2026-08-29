package storieshttp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) GetStoryLinks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyID, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	links, err := h.stories.GetStoryLinks(ctx, storyID, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	return web.Respond(ctx, w, toAppStoryLinks(links), http.StatusOK)
}

func (h *Handlers) MyStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.MyStories")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	storiesList, err := h.stories.MyStories(ctx, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	return h.respondStories(ctx, w, storiesList, http.StatusOK)
}

func (h *Handlers) UpdateCollaborators(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var request AppUpdateCollaborators
	if err := web.Decode(r, &request); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.UpdateCollaborators(ctx, storyID, workspace.ID, request.CollaboratorIDs); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	h.invalidateStoryCaches(ctx, workspace.ID, storyID)
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) SetWatching(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var request AppSetStoryWatching
	if err := web.Decode(r, &request); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.SetWatching(ctx, storyID, workspace.ID, userID, request.Watching); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	h.invalidateStoryCaches(ctx, workspace.ID, storyID)
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) invalidateStoryCaches(ctx context.Context, workspaceID, storyID uuid.UUID) {
	for _, key := range cache.InvalidateStoryKeys(workspaceID, storyID) {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}
	h.cache.DeleteByPattern(ctx, fmt.Sprintf(cache.MyStoriesKey+"*", workspaceID.String()))
}
