package storieshttp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.GetActivities")
	defer span.End()

	storyIdParam := web.Params(r, "id")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	page, pageSize, err := parseStoryPagination(r, 1, 20)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Note: Cache is disabled for paginated requests since we need different cache keys per page
	// TODO: Consider implementing page-specific caching if needed

	activitiesList, hasMore, err := h.stories.GetActivitiesWithUser(ctx, storyId, workspace.ID, page, pageSize)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	for i := range activitiesList {
		activitiesList[i].User.AvatarURL = h.resolveUserAvatarURL(ctx, activitiesList[i].User.AvatarURL)
	}

	appActivities := toAppActivitiesWithUser(activitiesList)

	nextPage := page + 1
	if !hasMore {
		nextPage = 0
	}

	response := ActivitiesResponseWithUser{
		Activities: appActivities,
		Pagination: ActivitiesPagination{
			Page:     page,
			PageSize: pageSize,
			HasMore:  hasMore,
			NextPage: nextPage,
		},
	}

	web.Respond(ctx, w, response, http.StatusOK)
	return nil
}

func (h *Handlers) CreateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.CreateComment")
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

	comment, err := h.stories.CreateComment(ctx, workspace.ID, ca)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	return h.respondComment(ctx, w, comment, http.StatusCreated)
}

func (h *Handlers) GetComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.GetComments")
	defer span.End()

	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	page, pageSize, err := parseStoryPagination(r, 1, 20)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	commentsList, hasMore, err := h.stories.GetComments(ctx, storyId, workspace.ID, page, pageSize)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	nextPage := page + 1
	if !hasMore {
		nextPage = 0
	}

	response := CommentsResponse{
		Pagination: CommentsPagination{
			Page:     page,
			PageSize: pageSize,
			HasMore:  hasMore,
			NextPage: nextPage,
		},
	}

	return h.respondComments(ctx, w, commentsList, response, http.StatusOK)
}

func (h *Handlers) DuplicateStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

	userID, _ := mid.GetUserID(ctx)

	duplicatedStory, err := h.stories.DuplicateStory(ctx, storyId, workspace.ID, userID)
	if err != nil {
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

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	h.cache.DeleteByPattern(ctx, myStoriesCachePattern)

	return h.respondStory(ctx, w, duplicatedStory, http.StatusCreated)
}
