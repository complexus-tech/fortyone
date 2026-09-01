package storieshttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.Delete")
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
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	authorization := stories.BulkDeleteAuthorization{
		ActorID: userID,
		IsAdmin: workspace.UserRole == string(mid.RoleAdmin),
	}

	cacheKeys := cache.InvalidateStoryKeys(workspace.ID, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	if err := h.stories.Delete(ctx, storyId, workspace.ID, authorization); err != nil {
		web.RespondError(ctx, w, err, storyMutationStatus(err))
		return nil
	}
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) BulkDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.BulkDelete")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}
	var req AppBulkDeleteRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	authorization := stories.BulkDeleteAuthorization{
		ActorID: userID,
		IsAdmin: workspace.UserRole == string(mid.RoleAdmin),
	}

	isHardDelete := req.HardDelete != nil && *req.HardDelete
	var deletedStoryIDs []uuid.UUID

	if isHardDelete {
		result, err := h.stories.HardBulkDelete(ctx, req.StoryIDs, workspace.ID, authorization)
		if err != nil {
			web.RespondError(ctx, w, err, bulkDeleteStatus(err))
			return nil
		}
		deletedStoryIDs = result.StoryIDs
		cleanUpLegacyHardDeleteMedia(ctx, h.attachments, h.log, workspace.ID, result)
	} else {
		deletedStoryIDs, err = h.stories.BulkDelete(ctx, req.StoryIDs, workspace.ID, authorization)
		if err != nil {
			web.RespondError(ctx, w, err, bulkDeleteStatus(err))
			return nil
		}
	}

	for _, storyID := range deletedStoryIDs {
		cacheKeys := cache.InvalidateStoryKeys(workspace.ID, storyID)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				h.cache.DeleteByPattern(ctx, key)
			} else {
				h.cache.Delete(ctx, key)
			}
		}
	}
	h.cache.DeleteByPattern(ctx, fmt.Sprintf(cache.StoryListKey+"*", workspace.ID.String()))
	h.cache.DeleteByPattern(ctx, fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String()))

	data := map[string]any{
		"deletedCount": len(deletedStoryIDs),
		"storyIds":     deletedStoryIDs,
	}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

func (h *Handlers) BulkUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.BulkUpdate")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var requestData map[string]json.RawMessage
	if err := web.DecodeJSON(r, &requestData); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	for field := range requestData {
		if field != "storyIds" && field != "updates" {
			web.RespondError(ctx, w, fmt.Errorf("unknown bulk story update field %q", field), http.StatusBadRequest)
			return nil
		}
	}

	var storyIDs []uuid.UUID
	storyIDsRaw, hasStoryIDs := requestData["storyIds"]
	if !hasStoryIDs || isJSONNull(storyIDsRaw) {
		web.RespondError(ctx, w, errors.New("storyIds is required"), http.StatusBadRequest)
		return nil
	}
	if err := json.Unmarshal(storyIDsRaw, &storyIDs); err != nil {
		web.RespondError(ctx, w, errors.New("invalid storyIds"), http.StatusBadRequest)
		return nil
	}

	if len(storyIDs) == 0 {
		web.RespondError(ctx, w, errors.New("storyIds cannot be empty"), http.StatusBadRequest)
		return nil
	}

	var updatesRaw map[string]json.RawMessage
	updatesBody, hasUpdates := requestData["updates"]
	if !hasUpdates || isJSONNull(updatesBody) {
		web.RespondError(ctx, w, errors.New("updates is required"), http.StatusBadRequest)
		return nil
	}
	if err := json.Unmarshal(updatesBody, &updatesRaw); err != nil {
		web.RespondError(ctx, w, errors.New("invalid updates"), http.StatusBadRequest)
		return nil
	}

	patch, err := parseStoryPatch(updatesRaw)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	for _, storyId := range storyIDs {
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

	result, err := h.stories.BulkUpdatePatch(ctx, storyIDs, workspace.ID, patch)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppBulkUpdateResult(result), http.StatusOK)
	return nil
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.Create")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}
	var ns AppNewStory
	if err := web.Decode(r, &ns); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	story, err := h.stories.Create(ctx, toCoreNewStory(ns, userID), workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, storyMutationStatus(err))
		return nil
	}

	if story.Parent != nil {
		cacheKeys := cache.InvalidateStoryKeys(workspace.ID, *story.Parent)
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

	return h.respondCreatedStory(ctx, w, story)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.Update")
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

	var requestData map[string]json.RawMessage
	if err := web.DecodeJSON(r, &requestData); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	reconcileMedia, referencedMediaIDs, err := storyMediaReconciliationRequest(
		requestData,
		workspace.Slug,
		storyId,
	)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	patch, err := parseStoryPatch(requestData)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if reconcileMedia {
		orphanedMediaIDs, err := h.stories.UpdatePatchWithMediaReconciliation(
			ctx,
			storyId,
			workspace.ID,
			patch,
			referencedMediaIDs,
		)
		if err != nil {
			web.RespondError(ctx, w, err, storyMutationStatus(err))
			return nil
		}
		for _, attachmentID := range orphanedMediaIDs {
			if h.attachments == nil {
				break
			}
			if err := h.attachments.DeleteOrphanedMedia(ctx, attachmentID, workspace.ID); err != nil && h.log != nil {
				h.log.Error(
					ctx,
					"failed to clean up reconciled story media",
					"error",
					err,
					"attachment_id",
					attachmentID,
				)
			}
		}
	} else if err := h.stories.UpdatePatch(ctx, storyId, workspace.ID, patch); err != nil {
		web.RespondError(ctx, w, err, storyMutationStatus(err))
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

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}
