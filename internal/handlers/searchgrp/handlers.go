package searchgrp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/complexus-tech/projects-api/internal/core/search"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("invalid workspace ID")
	ErrInvalidSearchType  = errors.New("invalid search type")
)

// Handlers manages the handlers for search endpoints.
type Handlers struct {
	searchService *search.Service
}

// New creates a new instance of search handlers.
func New(searchService *search.Service) *Handlers {
	return &Handlers{
		searchService: searchService,
	}
}

// Search handles the search request.
func (h *Handlers) Search(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}
	userID, _ := mid.GetUserID(ctx)

	// Extract query parameters
	var params AppSearchParams

	// Parse type
	params.Type = r.URL.Query().Get("type")
	if params.Type == "" {
		params.Type = "all"
	}

	// Parse query
	params.Query = r.URL.Query().Get("query")

	// Parse optional UUIDs
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

	// Parse optional string
	if priority := r.URL.Query().Get("priority"); priority != "" {
		params.Priority = &priority
	}

	// Parse sort option
	params.SortBy = r.URL.Query().Get("sortBy")
	if params.SortBy == "" {
		params.SortBy = "relevance"
	}

	// Parse pagination
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

	// Convert to core parameters
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

	// Convert sort option
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

	// Call search service
	result, err := h.searchService.Search(ctx, workspaceID, userID, searchParams)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Convert to app response
	response := toAppSearchResponse(result, params.Page, params.PageSize)

	return web.Respond(ctx, w, response, http.StatusOK)
}
