package storiesgrp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

	cacheKey := cache.StoryDetailCacheKey(workspaceId, storyId)
	var cachedStory stories.CoreSingleStory

	if err := h.cache.Get(ctx, cacheKey, &cachedStory); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppStory(cachedStory), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	story, err := h.stories.Get(ctx, storyId, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, story, cache.DetailTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppStory(story), http.StatusOK)
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
		StoryID: storyId,
		Parent:  requestData.Parent,
		UserID:  userID,
		Comment: requestData.Comment,
	}

	comment, err := h.stories.CreateComment(ctx, ca)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Invalidate the comments cache for this story
	cacheKey := cache.StoryCommentsCacheKey(workspaceId, storyId)
	if err := h.cache.Delete(ctx, cacheKey); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to delete cache", "key", cacheKey, "error", err)
	}

	web.Respond(ctx, w, toAppComment(comment), http.StatusCreated)

	return nil
}

// GetComments returns the comments for a story.
func (h *Handlers) GetComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.GetComments")
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

	// Try to get from cache first
	cacheKey := cache.StoryCommentsCacheKey(workspaceId, storyId)
	var cachedComments []comments.CoreComment

	if err := h.cache.Get(ctx, cacheKey, &cachedComments); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		web.Respond(ctx, w, toAppComments(cachedComments), http.StatusOK)
		return nil
	}

	// Cache miss, get from database
	commentsList, err := h.stories.GetComments(ctx, storyId)
	if err != nil {
		h.log.Error(ctx, "failed to get comments", "error", err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, commentsList, cache.DetailTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
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
	cacheKey := cache.StoryAttachmentsCacheKey(workspaceId, storyId)
	var cachedAttachments []attachments.FileInfo

	if err := h.cache.Get(ctx, cacheKey, &cachedAttachments); err == nil {
		// Cache hit
		span.AddEvent("cache hit", trace.WithAttributes(
			attribute.String("cache_key", cacheKey),
		))
		return web.Respond(ctx, w, cachedAttachments, http.StatusOK)
	}

	// Cache miss, get from database
	fileInfos, err := h.attachments.GetAttachmentsForStory(ctx, storyId)
	if err != nil {
		return fmt.Errorf("error getting attachments for story: %w", err)
	}

	// Store in cache
	if err := h.cache.Set(ctx, cacheKey, fileInfos, cache.DetailTTL); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to set cache", "key", cacheKey, "error", err)
	}

	return web.Respond(ctx, w, fileInfos, http.StatusOK)
}

// DeleteAttachment deletes an attachment from a story.
func (h *Handlers) DeleteAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storiesgrp.handlers.DeleteAttachment")
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

	// Invalidate the attachments cache for this story
	cacheKey := cache.StoryAttachmentsCacheKey(workspaceId, storyId)
	if err := h.cache.Delete(ctx, cacheKey); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to delete cache", "key", cacheKey, "error", err)
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

	// Invalidate the attachments cache for this story
	cacheKey := cache.StoryAttachmentsCacheKey(workspaceId, storyId)
	if err := h.cache.Delete(ctx, cacheKey); err != nil {
		// Log error but continue
		h.log.Error(ctx, "failed to delete cache", "key", cacheKey, "error", err)
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
