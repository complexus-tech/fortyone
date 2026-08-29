package objectiveshttp

import (
	"context"
	"net/http"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectiveshttp.handlers.List")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filters, err := parseObjectiveListFilters(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	query := objectivesdomain.ListQuery{
		WorkspaceID: workspace.ID, ActorID: userID, Search: filters.Search,
		TeamID: filters.TeamID,
	}

	params, err := parseObjectiveListPagination(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if params != nil {
		query.Limit, query.Offset = params.PageSize+1, params.Offset()
		values, err := h.objectives.ListIntent(ctx, query)
		if err != nil {
			return respondObjectiveError(ctx, w, err)
		}
		hasMore := len(values) > params.PageSize
		if hasMore {
			values = values[:params.PageSize]
		}
		return web.Respond(ctx, w, toAppObjectivesResponse(values, params.Page, params.PageSize, hasMore), http.StatusOK)
	}

	values, err := h.objectives.ListIntent(ctx, query)
	if err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	return web.Respond(ctx, w, toAppObjectives(values), http.StatusOK)
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectiveshttp.handlers.Get")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	objectiveID, ok := parseObjectivePathID(ctx, w, r, "id")
	if !ok {
		return nil
	}
	objective, err := h.objectives.Get(ctx, objectiveID, workspace.ID)
	if err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	return web.Respond(ctx, w, toAppObjective(objective), http.StatusOK)
}

func (h *Handlers) GetKeyResults(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectiveshttp.handlers.GetKeyResults")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	objectiveID, ok := parseObjectivePathID(ctx, w, r, "id")
	if !ok {
		return nil
	}
	if _, err := h.objectives.Get(ctx, objectiveID, workspace.ID); err != nil {
		return respondObjectiveError(ctx, w, err)
	}

	cacheKey := cache.KeyResultsListCacheKey(workspace.ID, objectiveID)
	var cached []keyresults.CoreKeyResult
	if h.cache != nil && h.cache.Get(ctx, cacheKey, &cached) == nil {
		span.AddEvent("cache hit", trace.WithAttributes(attribute.String("cache_key", cacheKey)))
		return web.Respond(ctx, w, toAppKeyResults(cached), http.StatusOK)
	}
	values, err := h.keyResults.List(ctx, objectiveID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	if h.cache != nil {
		if err := h.cache.Set(ctx, cacheKey, values, cache.ListTTL); err != nil && h.log != nil {
			h.log.Error(ctx, "failed to cache objective key results", "key", cacheKey, "error", err)
		}
	}
	return web.Respond(ctx, w, toAppKeyResults(values), http.StatusOK)
}

func (h *Handlers) GetAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "objectiveshttp.handlers.GetAnalytics")
	defer span.End()
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	objectiveID, ok := parseObjectivePathID(ctx, w, r, "id")
	if !ok {
		return nil
	}
	analytics, err := h.objectives.GetAnalytics(ctx, objectiveID, workspace.ID)
	if err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	for index := range analytics.TeamAllocation {
		if analytics.TeamAllocation[index].AvatarURL == nil {
			continue
		}
		resolved := h.resolveUserAvatarURL(ctx, *analytics.TeamAllocation[index].AvatarURL)
		if resolved == "" {
			analytics.TeamAllocation[index].AvatarURL = nil
		} else {
			analytics.TeamAllocation[index].AvatarURL = &resolved
		}
	}
	return web.Respond(ctx, w, toAppObjectiveAnalytics(analytics), http.StatusOK)
}

func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	objectiveID, ok := parseObjectivePathID(ctx, w, r, "id")
	if !ok {
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	params, err := parseObjectiveActivityPagination(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if _, err := h.objectives.Get(ctx, objectiveID, workspace.ID); err != nil {
		return respondObjectiveError(ctx, w, err)
	}
	activities, hasMore, err := h.okrActivities.GetObjectiveActivities(
		ctx, objectiveID, workspace.ID, params.Page, params.PageSize,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	for index := range activities {
		activities[index].User.AvatarURL = h.resolveUserAvatarURL(ctx, activities[index].User.AvatarURL)
	}
	return web.Respond(ctx, w, AppObjectiveActivitiesResponse{
		Activities: toAppObjectiveActivities(activities),
		Pagination: AppActivityPagination{Page: params.Page, PageSize: params.PageSize, HasMore: hasMore},
	}, http.StatusOK)
}
