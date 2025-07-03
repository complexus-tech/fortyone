package storiesgrp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/handlers/linksgrp"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidStoryID     = errors.New("story id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	stories     *stories.Service
	comments    *comments.Service
	links       *links.Service
	attachments *attachments.Service
	cache       *cache.Service
	log         *logger.Logger
}

// NewStoriesHandlers returns a new storiesHandlers instance.
func New(stories *stories.Service, comments *comments.Service, links *links.Service, attachments *attachments.Service, cacheService *cache.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		stories:     stories,
		comments:    comments,
		links:       links,
		attachments: attachments,
		cache:       cacheService,
		log:         log,
	}
}

// Get returns the story with the specified ID, including sub-stories.
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.Get")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		h.log.Error(ctx, "invalid story id", "error", err)
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Cache miss, get from database
	story, err := h.stories.Get(ctx, storyId, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppStory(story), http.StatusOK)
	return nil
}

// List returns a list of stories.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.List")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	// Try to get from cache first
	filtersStr := ""
	if filters != nil {
		// Convert filters to string representation for cache key
		for k, v := range filters {
			filtersStr += fmt.Sprintf("%s:%v;", k, v)
		}
	}
	cacheKey := cache.StoryListCacheKey(workspaceId, filtersStr)
	var cachedStories []stories.CoreStoryList

	if err := h.cache.Get(ctx, cacheKey, &cachedStories); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppStories(cachedStories), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	storyList, err := h.stories.List(ctx, workspaceId, filters)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, storyList, cache.ListTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppStories(storyList), http.StatusOK)
	return nil
}

// Delete removes the story with the specified ID.
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.Delete")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache before deleting the story
	cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			h.cache.DeleteByPattern(ctx, key)
		} else {
			// Handle exact key deletion
			h.cache.Delete(ctx, key)
		}
	}

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	if err := h.stories.Delete(ctx, storyId, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// BulkDelete removes the stories with the specified IDs.
func (h *Handlers) BulkDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.BulkDelete")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	var req AppBulkDeleteRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache before deleting the stories
	// First invalidate individual story caches
	for _, storyId := range req.StoryIDs {
		cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				h.cache.DeleteByPattern(ctx, key)
			} else {
				h.cache.Delete(ctx, key)
			}
		}
	}

	// Then invalidate the list cache
	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	if err := h.stories.BulkDelete(ctx, req.StoryIDs, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

// Restore restores the story with the specified ID.
func (h *Handlers) Restore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.Restore")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	if err := h.stories.Restore(ctx, storyId, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache after successful restore
	// Invalidate specific story cache
	cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	// Invalidate list cache
	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	data := map[string]uuid.UUID{"id": storyId}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (h *Handlers) BulkRestore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.BulkRestore")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	var req AppBulkRestoreRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkRestore(ctx, req.StoryIDs, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache after successful bulk restore
	// First invalidate individual story caches
	for _, storyId := range req.StoryIDs {
		cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				h.cache.DeleteByPattern(ctx, key)
			} else {
				h.cache.Delete(ctx, key)
			}
		}
	}

	// Then invalidate the list cache
	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

// BulkUpdate updates multiple stories with the same updates.
func (h *Handlers) BulkUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.BulkUpdate")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Decode as raw JSON
	var requestData map[string]json.RawMessage
	if err := web.Decode(r, &requestData); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Extract and validate storyIds
	var storyIDs []uuid.UUID
	if err := json.Unmarshal(requestData["storyIds"], &storyIDs); err != nil {
		web.RespondError(ctx, w, errors.New("invalid storyIds"), http.StatusBadRequest)
		return nil
	}

	if len(storyIDs) == 0 {
		web.RespondError(ctx, w, errors.New("storyIds cannot be empty"), http.StatusBadRequest)
		return nil
	}

	// Extract updates and convert to format getUpdates expects
	var updatesRaw map[string]json.RawMessage
	if err := json.Unmarshal(requestData["updates"], &updatesRaw); err != nil {
		web.RespondError(ctx, w, errors.New("invalid updates"), http.StatusBadRequest)
		return nil
	}

	// Reuse existing getUpdates function (same as Update handler)
	updates, err := getUpdates(updatesRaw)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache for all affected stories
	for _, storyId := range storyIDs {
		cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				h.cache.DeleteByPattern(ctx, key)
			} else {
				h.cache.Delete(ctx, key)
			}
		}
	}

	// Then invalidate the list cache
	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	if err := h.stories.BulkUpdate(ctx, storyIDs, workspaceId, updates); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// Create creates a new story.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.Create")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	var ns AppNewStory
	if err := web.Decode(r, &ns); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	story, err := h.stories.Create(ctx, toCoreNewStory(ns), workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// invalidate parent story cache if it exists parent_id is not nil
	if story.Parent != nil {
		cacheKeys := cache.InvalidateStoryKeys(workspaceId, *story.Parent)
		for _, key := range cacheKeys {
			if strings.Contains(key, "*") {
				// Handle pattern deletion
				h.cache.DeleteByPattern(ctx, key)
			} else {
				// Handle exact key deletion
				h.cache.Delete(ctx, key)
			}
		}
	}

	// invalidate list cache
	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	web.Respond(ctx, w, toAppStory(story), http.StatusCreated)
	return nil
}

// UpdateLabels replaces the labels for a story.
func (h *Handlers) UpdateLabels(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.UpdateLabels")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var req AppNewLabels
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.stories.UpdateLabels(ctx, storyId, workspaceId, req.Labels); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache after updating labels
	// Invalidate specific story cache
	cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			h.cache.DeleteByPattern(ctx, key)
		} else {
			h.cache.Delete(ctx, key)
		}
	}

	// Invalidate list cache (since labels may be used in filtering)
	listCachePattern := fmt.Sprintf(cache.StoryListKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, listCachePattern)

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// GetStoryLinks returns the links for a story.
func (h *Handlers) GetStoryLinks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyID, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	links, err := h.stories.GetStoryLinks(ctx, storyID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, linksgrp.ToLinks(links), http.StatusOK)
}

// MyStories returns a list of stories.
func (h *Handlers) MyStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.MyStories")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	// Use the predefined MyStoriesKey with the user ID as part of the key
	cacheKey := fmt.Sprintf(cache.MyStoriesKey+":%s", workspaceId.String(), userID.String())
	var cachedStories []stories.CoreStoryList

	if err := h.cache.Get(ctx, cacheKey, &cachedStories); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppStories(cachedStories), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	storiesList, err := h.stories.MyStories(ctx, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, storiesList, cache.ListTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppStories(storiesList), http.StatusOK)
	return nil
}

// Update updates the story with the specified ID.
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.Update")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var requestData map[string]json.RawMessage
	if err := web.Decode(r, &requestData); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	updates, err := getUpdates(requestData)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.Update(ctx, storyId, workspaceId, updates); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache after successful update
	cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			h.cache.DeleteByPattern(ctx, key)
		} else {
			// Handle exact key deletion
			h.cache.Delete(ctx, key)
		}
	}

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// GetActivities returns the activities for a story.
func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.GetActivities")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Try to get from cache first
	cacheKey := cache.StoryActivitiesCacheKey(workspaceId, storyId)
	var cachedActivities []stories.CoreActivity

	if err := h.cache.Get(ctx, cacheKey, &cachedActivities); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppActivities(cachedActivities), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	activitiesList, err := h.stories.GetActivities(ctx, storyId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, activitiesList, cache.ListTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppActivities(activitiesList), http.StatusOK)
	return nil
}

// CreateComment creates a comment for a story.
func (h *Handlers) CreateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.CreateComment")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var requestData AppNewComment
	if err := web.Decode(r, &requestData); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	ca := stories.CoreNewComment{
		StoryID:  storyId,
		Parent:   requestData.Parent,
		UserID:   userID,
		Comment:  requestData.Comment,
		Mentions: requestData.Mentions,
	}

	comment, err := h.stories.CreateComment(ctx, workspaceId, ca)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppComment(comment), http.StatusCreated)

	return nil
}

// GetComments returns the comments for a story.
func (h *Handlers) GetComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.GetComments")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		h.log.Error(ctx, "invalid story id", "error", err)
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	commentsList, err := h.stories.GetComments(ctx, storyId)
	if err != nil {
		h.log.Error(ctx, "failed to get comments", "error", err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppComments(commentsList), http.StatusOK)
	return nil
}

// DuplicateStory creates a copy of an existing story.
func (h *Handlers) DuplicateStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Get the user ID from context
	userID, _ := mid.GetUserID(ctx)

	duplicatedStory, err := h.stories.DuplicateStory(ctx, storyId, workspaceId, userID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate cache after successful copy
	cacheKeys := cache.InvalidateStoryKeys(workspaceId, storyId)
	for _, key := range cacheKeys {
		if strings.Contains(key, "*") {
			// Handle pattern deletion
			h.cache.DeleteByPattern(ctx, key)
		} else {
			// Handle exact key deletion
			h.cache.Delete(ctx, key)
		}
	}

	// Also invalidate my-stories cache pattern
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspaceId.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	web.Respond(ctx, w, toAppStory(duplicatedStory), http.StatusCreated)
	return nil
}

// GetAttachmentsForStory returns the attachments for a story.
func (h *Handlers) GetAttachmentsForStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.GetAttachmentsForStory")
	defer span.End()

	storyIdParam := web.Params(r, "id")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	// Cache miss, get from database
	fileInfos, err := h.attachments.GetAttachmentsForStory(ctx, storyId)
	if err != nil {
		return fmt.Errorf("error getting attachments for story: %w", err)
	}

	return web.Respond(ctx, w, fileInfos, http.StatusOK)
}

// DeleteAttachment deletes an attachment from a story.
func (h *Handlers) DeleteAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.DeleteAttachment")
	defer span.End()

	userID, _ := mid.GetUserID(ctx)

	attachmentIDStr := web.Params(r, "attachmentId")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid attachment ID"), http.StatusBadRequest)
	}

	// delete attachment from Azure
	err = h.attachments.DeleteAttachment(ctx, attachmentID, userID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error deleting attachment: %w", err), http.StatusBadRequest)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// UploadStoryAttachment handles the upload of a new attachment specifically for a story
func (h *Handlers) UploadStoryAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, _ := mid.GetUserID(ctx)
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Parse multipart form, limit to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	// Upload file and automatically link to story
	fileInfo, err := h.attachments.UploadAndLinkToStory(ctx, file, header, userID, storyId, workspaceId)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrFileTooLarge), errors.Is(err, attachments.ErrInvalidFileType):
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		default:
			return fmt.Errorf("error uploading attachment: %w", err)
		}
	}

	return web.Respond(ctx, w, fileInfo, http.StatusCreated)
}

// CountInWorkspace returns the count of stories in a workspace.
func (h *Handlers) CountInWorkspace(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.stories.CountInWorkspace")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	count, err := h.stories.CountInWorkspace(ctx, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	span.AddEvent("stories count retrieved", trace.WithAttributes(
		attribute.Int("stories.count", count),
	))

	return web.Respond(ctx, w, map[string]int{"count": count}, http.StatusOK)
}

// ListGrouped handles the initial load of grouped stories
func (h *Handlers) ListGrouped(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.ListGrouped")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	// Parse query parameters
	query, err := parseStoryQuery(r, userID, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Set default values
	if query.StoriesPerGroup == 0 {
		query.StoriesPerGroup = 15 // Increased from 5 for better UX
	}

	// Convert to core query
	coreQuery := toCoreStoryQuery(query)
	coreQuery.Filters.CurrentUserID = userID
	coreQuery.Filters.WorkspaceID = workspaceId

	// Handle "none" groupBy case - create a single group with key "none"
	if query.GroupBy == "none" {
		// Use the grouped stories method but force groupBy to be handled as a single group
		coreQuery.GroupBy = "none"

		// Fetch grouped stories (which will handle "none" as a special case)
		groups, err := h.stories.ListGroupedStories(ctx, coreQuery)
		if err != nil {
			web.RespondError(ctx, w, err, http.StatusInternalServerError)
			return nil
		}

		// Convert to app response (same format as other grouped responses)
		response := convertGroupsToResponse(groups, query)
		return web.Respond(ctx, w, response, http.StatusOK)
	}

	// Fetch grouped stories
	groups, err := h.stories.ListGroupedStories(ctx, coreQuery)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Convert to app response
	response := convertGroupsToResponse(groups, query)

	return web.Respond(ctx, w, response, http.StatusOK)
}

// LoadMoreGroup handles loading more stories for a specific group
func (h *Handlers) LoadMoreGroup(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.LoadMoreGroup")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	// Parse query parameters
	query, err := parseStoryQuery(r, userID, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Set default values
	if query.PageSize == 0 {
		query.PageSize = 15 // Increased from 10 for better performance
	}

	// Get group key from query parameters
	groupKey := r.URL.Query().Get("groupKey")
	if groupKey == "" {
		web.RespondError(ctx, w, fmt.Errorf("groupKey is required"), http.StatusBadRequest)
		return nil
	}

	// Convert to core query
	coreQuery := toCoreStoryQuery(query)
	coreQuery.Filters.CurrentUserID = userID
	coreQuery.Filters.WorkspaceID = workspaceId

	// Fetch stories for the specific group
	stories, hasMore, err := h.stories.ListGroupStories(ctx, groupKey, coreQuery)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Convert to app stories
	appStories := toAppStories(stories)

	// Calculate pagination info
	nextPage := query.Page + 1
	if !hasMore {
		nextPage = 0
	}

	response := GroupStoriesResponse{
		GroupKey: groupKey,
		Stories:  appStories,
		Pagination: GroupPagination{
			Page:     query.Page,
			PageSize: query.PageSize,
			HasMore:  hasMore,
			NextPage: nextPage,
		},
		Filters:        query.Filters,
		OrderBy:        query.OrderBy,
		OrderDirection: query.OrderDirection,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

// ListByCategory returns stories filtered by category with pagination
func (h *Handlers) ListByCategory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.ListByCategory")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	// Get required category parameter
	category := r.URL.Query().Get("category")
	if category == "" {
		web.RespondError(ctx, w, fmt.Errorf("category parameter is required"), http.StatusBadRequest)
		return nil
	}

	// Get required teamId parameter
	teamIdParam := r.URL.Query().Get("teamId")
	if teamIdParam == "" {
		web.RespondError(ctx, w, fmt.Errorf("teamId parameter is required"), http.StatusBadRequest)
		return nil
	}

	teamId, err := uuid.Parse(teamIdParam)
	if err != nil {
		web.RespondError(ctx, w, fmt.Errorf("invalid teamId format"), http.StatusBadRequest)
		return nil
	}

	// Parse pagination parameters
	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "pageSize", 20)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Fetch stories by category and team
	stories, hasMore, err := h.stories.ListByCategory(ctx, workspaceId, userID, teamId, category, page, pageSize)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	// Convert to app stories
	appStories := toAppStories(stories)

	// Calculate pagination info
	nextPage := page + 1
	if !hasMore {
		nextPage = 0
	}

	response := CategoryStoriesResponse{
		Stories: appStories,
		Pagination: CategoryPagination{
			Page:     page,
			PageSize: pageSize,
			HasMore:  hasMore,
			NextPage: nextPage,
		},
		Meta: CategoryMeta{
			Category:    category,
			TeamID:      teamId,
			TotalLoaded: len(appStories),
		},
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

// parseStoryQuery parses URL query parameters into a StoryQuery struct
func parseStoryQuery(r *http.Request, userID, workspaceID uuid.UUID) (StoryQuery, error) {
	query := StoryQuery{
		Filters: StoryFilters{},
	}

	// Parse simple string parameters
	query.GroupBy = getStringParam(r, "groupBy", "status")
	query.OrderBy = getStringParam(r, "orderBy", "created")
	query.OrderDirection = getStringParam(r, "orderDirection", "desc")
	query.GroupKey = r.URL.Query().Get("groupKey")

	// Validate groupBy values
	if !isValidGroupBy(query.GroupBy) {
		return query, fmt.Errorf("invalid groupBy value: %s. Must be one of: status, assignee, priority, team, sprint, none", query.GroupBy)
	}

	// Validate orderBy values
	if !isValidOrderBy(query.OrderBy) {
		return query, fmt.Errorf("invalid orderBy value: %s. Must be one of: created, updated, priority, deadline", query.OrderBy)
	}

	// Validate orderDirection values
	if !isValidOrderDirection(query.OrderDirection) {
		return query, fmt.Errorf("invalid orderDirection value: %s. Must be asc or desc", query.OrderDirection)
	}

	// Parse integer parameters
	query.StoriesPerGroup = getIntParam(r, "storiesPerGroup", 0)
	query.Page = getIntParam(r, "page", 1)
	query.PageSize = getIntParam(r, "pageSize", 0)

	// Parse array filters using helper functions
	query.Filters.StatusIDs = parseUUIDArray(r, "statusIds")
	query.Filters.AssigneeIDs = parseUUIDArray(r, "assigneeIds")
	query.Filters.ReporterIDs = parseUUIDArray(r, "reporterIds")
	query.Filters.TeamIDs = parseUUIDArray(r, "teamIds")
	query.Filters.SprintIDs = parseUUIDArray(r, "sprintIds")
	query.Filters.LabelIDs = parseUUIDArray(r, "labelIds")

	// Parse string arrays (comma-separated)
	query.Filters.Priorities = parseStringArray(r, "priorities")
	query.Filters.Categories = parseStringArray(r, "categories")

	// Parse single UUID parameters
	query.Filters.Parent = parseUUIDParam(r, "parentId")
	query.Filters.Objective = parseUUIDParam(r, "objectiveId")
	query.Filters.Epic = parseUUIDParam(r, "epicId")

	// Parse boolean parameters
	query.Filters.HasNoAssignee = parseBoolParam(r, "hasNoAssignee")
	query.Filters.AssignedToMe = parseBoolParam(r, "assignedToMe")
	query.Filters.CreatedByMe = parseBoolParam(r, "createdByMe")

	// Parse date range parameters
	query.Filters.CreatedAfter = parseDateParam(r, "createdAfter")
	query.Filters.CreatedBefore = parseDateParam(r, "createdBefore")
	query.Filters.UpdatedAfter = parseDateParam(r, "updatedAfter")
	query.Filters.UpdatedBefore = parseDateParam(r, "updatedBefore")
	query.Filters.DeadlineAfter = parseDateParam(r, "deadlineAfter")
	query.Filters.DeadlineBefore = parseDateParam(r, "deadlineBefore")

	return query, nil
}

// Helper functions for parameter parsing
func getStringParam(r *http.Request, key, defaultValue string) string {
	if value := r.URL.Query().Get(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntParam(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseUUIDArray(r *http.Request, key string) []uuid.UUID {
	values := r.URL.Query()[key]
	if len(values) == 0 {
		return nil
	}

	var result []uuid.UUID
	for _, value := range values {
		if parsed, err := uuid.Parse(value); err == nil {
			result = append(result, parsed)
		}
	}
	return result
}

func parseStringArray(r *http.Request, key string) []string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}

	// Support comma-separated format: priorities=high,medium,low
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseUUIDParam(r *http.Request, key string) *uuid.UUID {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := uuid.Parse(value); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseBoolParam(r *http.Request, key string) *bool {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseDateParam(r *http.Request, key string) *time.Time {
	if value := r.URL.Query().Get(key); value != "" {
		// Support multiple date formats
		formats := []string{
			time.RFC3339,          // "2006-01-02T15:04:05Z07:00"
			"2006-01-02",          // "2024-06-02"
			"2006-01-02T15:04:05", // "2024-06-02T15:04:05"
		}

		for _, format := range formats {
			if parsed, err := time.Parse(format, value); err == nil {
				return &parsed
			}
		}
	}
	return nil
}

// isValidOrderBy validates the orderBy parameter
func isValidOrderBy(orderBy string) bool {
	validValues := []string{"created", "updated", "priority", "deadline"}
	for _, v := range validValues {
		if v == orderBy {
			return true
		}
	}
	return false
}

// isValidOrderDirection validates the orderDirection parameter
func isValidOrderDirection(direction string) bool {
	return direction == "asc" || direction == "desc"
}

// isValidGroupBy validates the groupBy parameter
func isValidGroupBy(groupBy string) bool {
	validValues := []string{"status", "assignee", "priority", "team", "sprint", "none"}
	for _, v := range validValues {
		if v == groupBy {
			return true
		}
	}
	return false
}

// toCoreStoryQuery converts a handler StoryQuery to a core CoreStoryQuery
func toCoreStoryQuery(query StoryQuery) stories.CoreStoryQuery {
	return stories.CoreStoryQuery{
		Filters: stories.CoreStoryFilters{
			StatusIDs:      query.Filters.StatusIDs,
			AssigneeIDs:    query.Filters.AssigneeIDs,
			ReporterIDs:    query.Filters.ReporterIDs,
			Priorities:     query.Filters.Priorities,
			Categories:     query.Filters.Categories,
			TeamIDs:        query.Filters.TeamIDs,
			SprintIDs:      query.Filters.SprintIDs,
			LabelIDs:       query.Filters.LabelIDs,
			Parent:         query.Filters.Parent,
			Objective:      query.Filters.Objective,
			Epic:           query.Filters.Epic,
			HasNoAssignee:  query.Filters.HasNoAssignee,
			AssignedToMe:   query.Filters.AssignedToMe,
			CreatedByMe:    query.Filters.CreatedByMe,
			CreatedAfter:   query.Filters.CreatedAfter,
			CreatedBefore:  query.Filters.CreatedBefore,
			UpdatedAfter:   query.Filters.UpdatedAfter,
			UpdatedBefore:  query.Filters.UpdatedBefore,
			DeadlineAfter:  query.Filters.DeadlineAfter,
			DeadlineBefore: query.Filters.DeadlineBefore,
			CurrentUserID:  uuid.Nil, // Will be set in handler
			WorkspaceID:    uuid.Nil, // Will be set in handler
		},
		GroupBy:         query.GroupBy,
		OrderBy:         query.OrderBy,
		OrderDirection:  query.OrderDirection,
		StoriesPerGroup: query.StoriesPerGroup,
		GroupKey:        query.GroupKey,
		Page:            query.Page,
		PageSize:        query.PageSize,
	}
}

// coreFiltersToMap converts CoreStoryFilters to map[string]any format for the List method
func coreFiltersToMap(filters stories.CoreStoryFilters) map[string]any {
	result := make(map[string]any)

	// Always include CurrentUserID and WorkspaceID as they're needed for user-specific filters
	result["current_user_id"] = filters.CurrentUserID
	result["workspace_id"] = filters.WorkspaceID

	// Only add non-empty/non-nil filters
	if len(filters.StatusIDs) > 0 {
		result["status_ids"] = filters.StatusIDs
	}
	if len(filters.AssigneeIDs) > 0 {
		result["assignee_ids"] = filters.AssigneeIDs
	}
	if len(filters.ReporterIDs) > 0 {
		result["reporter_ids"] = filters.ReporterIDs
	}
	if len(filters.Priorities) > 0 {
		result["priorities"] = filters.Priorities
	}
	if len(filters.TeamIDs) > 0 {
		result["team_ids"] = filters.TeamIDs
	}
	if len(filters.SprintIDs) > 0 {
		result["sprint_ids"] = filters.SprintIDs
	}
	if len(filters.LabelIDs) > 0 {
		result["label_ids"] = filters.LabelIDs
	}
	if filters.Parent != nil {
		result["parent_id"] = *filters.Parent
	}
	if filters.Objective != nil {
		result["objective_id"] = *filters.Objective
	}
	if filters.Epic != nil {
		result["epic_id"] = *filters.Epic
	}
	if filters.HasNoAssignee != nil {
		result["has_no_assignee"] = *filters.HasNoAssignee
	}
	if filters.AssignedToMe != nil {
		result["assigned_to_me"] = *filters.AssignedToMe
	}
	if filters.CreatedByMe != nil {
		result["created_by_me"] = *filters.CreatedByMe
	}
	if filters.CreatedAfter != nil {
		result["created_after"] = *filters.CreatedAfter
	}
	if filters.CreatedBefore != nil {
		result["created_before"] = *filters.CreatedBefore
	}
	if filters.UpdatedAfter != nil {
		result["updated_after"] = *filters.UpdatedAfter
	}
	if filters.UpdatedBefore != nil {
		result["updated_before"] = *filters.UpdatedBefore
	}
	if filters.DeadlineAfter != nil {
		result["deadline_after"] = *filters.DeadlineAfter
	}
	if filters.DeadlineBefore != nil {
		result["deadline_before"] = *filters.DeadlineBefore
	}

	return result
}

// convertGroupsToResponse converts core story groups to handler response
func convertGroupsToResponse(groups []stories.CoreStoryGroup, query StoryQuery) StoriesResponse {
	appGroups := make([]StoryGroup, len(groups))
	for i, group := range groups {
		appStories := toAppStories(group.Stories)

		appGroups[i] = StoryGroup{
			Key:         group.Key,
			LoadedCount: group.LoadedCount,
			HasMore:     group.HasMore,
			Stories:     appStories,
			NextPage:    group.NextPage,
		}
	}

	return StoriesResponse{
		Groups: appGroups,
		Meta: GroupsMeta{
			TotalGroups:    len(appGroups),
			Filters:        query.Filters,
			GroupBy:        query.GroupBy,
			OrderBy:        query.OrderBy,
			OrderDirection: query.OrderDirection,
		},
	}
}
