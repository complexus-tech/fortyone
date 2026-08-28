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

func (h *Handlers) Restore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.Restore")
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

	if err := h.stories.Restore(ctx, storyId, workspace.ID); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	cacheKeys := cache.InvalidateStoryKeys(workspace.ID, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	data := map[string]uuid.UUID{"id": storyId}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

func (h *Handlers) BulkRestore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.BulkRestore")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}
	var req AppBulkRestoreRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkRestore(ctx, req.StoryIDs, workspace.ID); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	for _, storyId := range req.StoryIDs {
		cacheKeys := cache.InvalidateStoryKeys(workspace.ID, storyId)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				h.cache.DeleteByPattern(ctx, key)
			} else {
				h.cache.Delete(ctx, key)
			}
		}
	}

	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

func (h *Handlers) BulkUnarchive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.BulkUnarchive")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppBulkUnarchiveRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	for _, storyId := range req.StoryIDs {
		cacheKeys := cache.InvalidateStoryKeys(workspace.ID, storyId)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				h.cache.DeleteByPattern(ctx, key)
			} else {
				h.cache.Delete(ctx, key)
			}
		}
	}

	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	if err := h.stories.BulkUnarchive(ctx, req.StoryIDs, workspace.ID); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

func (h *Handlers) BulkArchive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.BulkArchive")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppBulkArchiveRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	for _, storyId := range req.StoryIDs {
		cacheKeys := cache.InvalidateStoryKeys(workspace.ID, storyId)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				h.cache.DeleteByPattern(ctx, key)
			} else {
				h.cache.Delete(ctx, key)
			}
		}
	}

	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	if err := h.stories.BulkArchive(ctx, req.StoryIDs, workspace.ID); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

func (h *Handlers) UpdateLabels(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.UpdateLabels")
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

	var req AppNewLabels
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.stories.UpdateLabels(ctx, storyId, workspace.ID, req.Labels); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	cacheKeys := cache.InvalidateStoryKeys(workspace.ID, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}
