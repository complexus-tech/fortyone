package storieshttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) AddAssociation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.AddAssociation")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	fromStoryId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppNewAssociation
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	assoc, err := h.stories.AddAssociation(ctx, fromStoryId, req.ToStoryID, req.AssociationType, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate caches
	h.invalidateCacheForStory(ctx, workspace.ID, fromStoryId)
	h.invalidateCacheForStory(ctx, workspace.ID, req.ToStoryID)

	appAssoc := AppStoryAssociation{
		ID:          assoc.ID,
		FromStoryID: assoc.FromStoryID,
		ToStoryID:   assoc.ToStoryID,
		Type:        assoc.Type,
		Story:       toAppStoryListItem(assoc.Story, nil),
	}

	web.Respond(ctx, w, appAssoc, http.StatusCreated)
	return nil
}

func (h *Handlers) UpdateAssociation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.UpdateAssociation")
	defer span.End()

	associationIdParam := web.Params(r, "associationId")
	associationId, err := uuid.Parse(associationIdParam)
	if err != nil {
		web.RespondError(ctx, w, errors.New("invalid association id"), http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppUpdateAssociation
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	assoc, err := h.stories.UpdateAssociation(ctx, associationId, req.FromStoryID, req.ToStoryID, req.AssociationType, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	h.invalidateCacheForStory(ctx, workspace.ID, req.FromStoryID)
	h.invalidateCacheForStory(ctx, workspace.ID, req.ToStoryID)

	appAssoc := AppStoryAssociation{
		ID:          assoc.ID,
		FromStoryID: assoc.FromStoryID,
		ToStoryID:   assoc.ToStoryID,
		Type:        assoc.Type,
		Story:       toAppStoryListItem(assoc.Story, nil),
	}

	web.Respond(ctx, w, appAssoc, http.StatusOK)
	return nil
}

func (h *Handlers) RemoveAssociation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.RemoveAssociation")
	defer span.End()

	associationIdParam := web.Params(r, "associationId")
	associationId, err := uuid.Parse(associationIdParam)
	if err != nil {
		web.RespondError(ctx, w, errors.New("invalid association id"), http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	if err := h.stories.RemoveAssociation(ctx, associationId, workspace.ID); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate generic lists for now as we don't have the specific IDs handy without a fetch
	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// Helper to reduce code duplication in cache invalidation
func (h *Handlers) invalidateCacheForStory(ctx context.Context, workspaceID, storyID uuid.UUID) {
	cacheKeys := cache.InvalidateStoryKeys(workspaceID, storyID)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)
}
