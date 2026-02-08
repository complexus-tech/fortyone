package searchhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("invalid workspace ID")
	ErrInvalidSearchType  = errors.New("invalid search type")
)

type Handlers struct {
	searchService *search.Service
}

func New(searchService *search.Service) *Handlers {
	return &Handlers{
		searchService: searchService,
	}
}

func (h *Handlers) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, _ := mid.GetUserID(ctx)

	var params AppSearchParams

	params.Type = r.URL.Query().Get("type")
	if params.Type == "" {
		params.Type = "all"
	}

	params.Query = r.URL.Query().Get("query")

	if teamIDStr := r.URL.Query().Get("teamId"); teamIDStr != "" {
		teamID, err := uuid.Parse(teamIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid team ID"), http.StatusBadRequest)
		}
		params.TeamID = &teamID
	}

	if assigneeIDStr := r.URL.Query().Get("assigneeId"); assigneeIDStr != "" {
		assigneeID, err := uuid.Parse(assigneeIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid assignee ID"), http.StatusBadRequest)
		}
		params.AssigneeID = &assigneeID
	}

	if labelIDStr := r.URL.Query().Get("labelId"); labelIDStr != "" {
		labelID, err := uuid.Parse(labelIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid label ID"), http.StatusBadRequest)
		}
		params.LabelID = &labelID
	}

	if statusIDStr := r.URL.Query().Get("statusId"); statusIDStr != "" {
		statusID, err := uuid.Parse(statusIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid status ID"), http.StatusBadRequest)
		}
		params.StatusID = &statusID
	}

	if priority := r.URL.Query().Get("priority"); priority != "" {
		params.Priority = &priority
	}

	params.SortBy = r.URL.Query().Get("sortBy")
	if params.SortBy == "" {
		params.SortBy = "relevance"
	}

	params.Page = 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	params.PageSize = 20
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	}

	searchType := search.SearchTypeAll
	switch params.Type {
	case "stories":
		searchType = search.SearchTypeStories
	case "objectives":
		searchType = search.SearchTypeObjectives
	case "all":
		searchType = search.SearchTypeAll
	default:
		return web.RespondError(ctx, w, ErrInvalidSearchType, http.StatusBadRequest)
	}

	sortOption := search.SortByRelevance
	switch params.SortBy {
	case "updated":
		sortOption = search.SortByUpdated
	case "created":
		sortOption = search.SortByCreated
	case "relevance":
		sortOption = search.SortByRelevance
	}

	searchParams := search.SearchParams{
		Type:       searchType,
		Query:      params.Query,
		TeamID:     params.TeamID,
		AssigneeID: params.AssigneeID,
		LabelID:    params.LabelID,
		StatusID:   params.StatusID,
		Priority:   params.Priority,
		SortBy:     sortOption,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}

	result, err := h.searchService.Search(ctx, workspace.ID, userID, searchParams)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	response := toAppSearchResponse(result, params.Page, params.PageSize)

	return web.Respond(ctx, w, response, http.StatusOK)
}
