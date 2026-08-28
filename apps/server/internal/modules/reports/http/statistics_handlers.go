package reportshttp

import (
	"context"
	"fmt"
	"net/http"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) GetStoryStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	startDate, endDate, err := date.RangeFromQueryAt(r.URL.Query(), 30, h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}

	stats, err := h.reports.GetStoryStats(ctx, workspace.ID, reports.StoryStatsFilters{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return respondReportError(ctx, w, err)
	}

	return web.Respond(ctx, w, toAppStoryStats(stats), http.StatusOK)
}

func (h *Handlers) GetContributionStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	startDate, endDate, err := date.RangeFromQueryAt(r.URL.Query(), 30, h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}

	stats, err := h.reports.GetContributionStats(ctx, userID, workspace.ID, startDate, endDate)
	if err != nil {
		return respondReportError(ctx, w, err)
	}

	return web.Respond(ctx, w, toAppContributionsStats(stats), http.StatusOK)
}

func (h *Handlers) GetUserStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	stats, err := h.reports.GetUserStats(ctx, userID, workspace.ID)
	if err != nil {
		return respondReportError(ctx, w, err)
	}

	return web.Respond(ctx, w, toAppUserStats(stats), http.StatusOK)
}

func (h *Handlers) GetStatusStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetStatusStats")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	coreFilters, err := parseStatsFilters(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	startDate, endDate, err := date.RangeFromQueryAt(r.URL.Query(), 30, h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}
	coreFilters.StartDate = startDate
	coreFilters.EndDate = endDate

	stats, err := h.reports.GetStatusStats(ctx, workspace.ID, coreFilters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting status stats: %w", err))
	}

	return web.Respond(ctx, w, toAppStatusStats(stats), http.StatusOK)
}

func (h *Handlers) GetPriorityStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetPriorityStats")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	coreFilters, err := parseStatsFilters(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	startDate, endDate, err := date.RangeFromQueryAt(r.URL.Query(), 30, h.clock.Now())
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}
	coreFilters.StartDate = startDate
	coreFilters.EndDate = endDate

	stats, err := h.reports.GetPriorityStats(ctx, workspace.ID, coreFilters)
	if err != nil {
		return respondReportError(ctx, w, fmt.Errorf("getting priority stats: %w", err))
	}

	return web.Respond(ctx, w, toAppPriorityStats(stats), http.StatusOK)
}
