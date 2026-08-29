package storieshttp

import (
	"context"
	"fmt"
	"net/http"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) ListGrouped(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.ListGrouped")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	query, err := parseStoryQuery(r)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if query.StoriesPerGroup == 0 {
		query.StoriesPerGroup = 15
	}
	coreQuery := toCoreStoryQuery(query)
	coreQuery.Filters.WorkspaceID = workspace.ID

	if query.GroupBy == "none" {
		coreQuery.GroupBy = "none"

		groups, err := h.stories.ListGroupedStories(ctx, coreQuery)
		if err != nil {
			web.RespondError(ctx, w, err, storyReadStatus(err))
			return nil
		}

		response, err := h.convertGroupsToResponse(ctx, groups, query)
		if err != nil {
			web.RespondError(ctx, w, err, http.StatusInternalServerError)
			return nil
		}
		return web.Respond(ctx, w, response, http.StatusOK)
	}

	groups, err := h.stories.ListGroupedStories(ctx, coreQuery)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	response, err := h.convertGroupsToResponse(ctx, groups, query)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) LoadMoreGroup(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.LoadMoreGroup")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	query, err := parseStoryQuery(r)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if query.PageSize == 0 {
		query.PageSize = 15
	}
	groupKey := query.GroupKey
	if groupKey == "" {
		web.RespondError(ctx, w, fmt.Errorf("groupKey is required"), http.StatusBadRequest)
		return nil
	}

	coreQuery := toCoreStoryQuery(query)
	coreQuery.Filters.WorkspaceID = workspace.ID

	stories, hasMore, err := h.stories.ListGroupStories(ctx, groupKey, coreQuery)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	usersByID, err := h.buildStoriesUsersByID(ctx, stories)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	appStories := toAppStories(stories, usersByID)

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

func (h *Handlers) ListByCategory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "storieshttp.handlers.ListByCategory")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	category, err := parseStringParam(r, "category", "", 32)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if category == "" {
		web.RespondError(ctx, w, fmt.Errorf("category parameter is required"), http.StatusBadRequest)
		return nil
	}
	if !isValidStoryCategory(category) {
		web.RespondError(ctx, w, fmt.Errorf("category parameter is invalid"), http.StatusBadRequest)
		return nil
	}

	teamID, err := parseStrictOptionalUUID(r, "teamId")
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if teamID == nil {
		web.RespondError(ctx, w, fmt.Errorf("teamId parameter is required"), http.StatusBadRequest)
		return nil
	}

	page, pageSize, err := parseStoryPagination(r, 1, 20)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	showSubStories, err := parseStrictOptionalBool(r, "showSubStories")
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	includeSubStories := showSubStories != nil && *showSubStories

	stories, hasMore, err := h.stories.ListByCategory(ctx, workspace.ID, *teamID, category, page, pageSize, includeSubStories)
	if err != nil {
		web.RespondError(ctx, w, err, storyReadStatus(err))
		return nil
	}

	usersByID, err := h.buildStoriesUsersByID(ctx, stories)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	appStories := toAppStories(stories, usersByID)

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
			TeamID:      *teamID,
			TotalLoaded: len(appStories),
		},
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) convertGroupsToResponse(ctx context.Context, groups []stories.CoreStoryGroup, query StoryQuery) (StoriesResponse, error) {
	usersByID, err := h.buildStoryGroupsUsersByID(ctx, groups)
	if err != nil {
		return StoriesResponse{}, err
	}

	appGroups := make([]StoryGroup, len(groups))
	for i, group := range groups {
		appStories := toAppStories(group.Stories, usersByID)

		appGroups[i] = StoryGroup{
			Key:         group.Key,
			LoadedCount: group.LoadedCount,
			TotalCount:  group.TotalCount,
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
	}, nil
}
