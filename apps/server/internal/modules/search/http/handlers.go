package searchhttp

import (
	"context"
	"errors"
	"net/http"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var ErrInvalidWorkspaceID = errors.New("invalid workspace ID")

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
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	searchParams, err := parseSearchParams(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	result, err := h.searchService.Search(ctx, workspace.ID, userID, searchParams)
	if err != nil {
		if errors.Is(err, search.ErrInvalidSearchParams) {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	response := toAppSearchResponse(result, searchParams.Page, searchParams.PageSize)

	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) FindSimilarStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	query, err := parseSimilarStoriesQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	stories, err := h.searchService.FindSimilarStories(
		ctx,
		workspace.ID,
		userID,
		query.Title,
		query.TeamID,
		query.Limit,
	)
	if err != nil {
		if errors.Is(err, search.ErrInvalidSearchParams) {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppSimilarStories(stories), http.StatusOK)
}
