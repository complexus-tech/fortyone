package reportsgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/reports"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("invalid workspace id")
	ErrInvalidDays        = errors.New("invalid days parameter")
	ErrInvalidDate        = errors.New("invalid date parameter")
)

type Handlers struct {
	reports *reports.Service
	log     *logger.Logger
}

func New(log *logger.Logger, reports *reports.Service) *Handlers {
	return &Handlers{
		reports: reports,
		log:     log,
	}
}

func (h *Handlers) GetStoryStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	startDate, endDate, err := date.RangeFromQuery(r.URL.Query(), 30)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}

	stats, err := h.reports.GetStoryStats(ctx, workspace.ID, reports.StoryStatsFilters{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return err
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

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	days := 6 // Default to 7 days
	if filters["days"] != nil {
		days, err = strconv.Atoi(filters["days"].(string))
		if err != nil || days <= 0 {
			return web.RespondError(ctx, w, ErrInvalidDays, http.StatusBadRequest)
		}
	}

	stats, err := h.reports.GetContributionStats(ctx, userID, workspace.ID, days)
	if err != nil {
		return err
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
		return err
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

	var af AppStatsFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	coreFilters := reports.StatsFilters{}
	if filters["teamId"] != nil {
		teamID, err := uuid.Parse(filters["teamId"].(string))
		if err == nil {
			coreFilters.TeamID = &teamID
		}
	}
	if filters["sprintId"] != nil {
		sprintID, err := uuid.Parse(filters["sprintId"].(string))
		if err == nil {
			coreFilters.SprintID = &sprintID
		}
	}
	if filters["objectiveId"] != nil {
		objectiveID, err := uuid.Parse(filters["objectiveId"].(string))
		if err == nil {
			coreFilters.ObjectiveID = &objectiveID
		}
	}

	stats, err := h.reports.GetStatusStats(ctx, workspace.ID, coreFilters)
	if err != nil {
		return fmt.Errorf("getting status stats: %w", err)
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

	var af AppStatsFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	coreFilters := reports.StatsFilters{}
	if filters["teamId"] != nil {
		teamID, err := uuid.Parse(filters["teamId"].(string))
		if err == nil {
			coreFilters.TeamID = &teamID
		}
	}
	if filters["sprintId"] != nil {
		sprintID, err := uuid.Parse(filters["sprintId"].(string))
		if err == nil {
			coreFilters.SprintID = &sprintID
		}
	}
	if filters["objectiveId"] != nil {
		objectiveID, err := uuid.Parse(filters["objectiveId"].(string))
		if err == nil {
			coreFilters.ObjectiveID = &objectiveID
		}
	}

	stats, err := h.reports.GetPriorityStats(ctx, workspace.ID, coreFilters)
	if err != nil {
		return fmt.Errorf("getting priority stats: %w", err)
	}

	return web.Respond(ctx, w, toAppPriorityStats(stats), http.StatusOK)
}

func parseReportFilters(query map[string]interface{}) (reports.ReportFilters, error) {
	filters := reports.ReportFilters{}

	if teamIds, ok := query["teamIds"]; ok {
		if teamIdsStr, ok := teamIds.(string); ok && teamIdsStr != "" {
			teamIDStrings := strings.Split(teamIdsStr, ",")
			for _, idStr := range teamIDStrings {
				idStr = strings.TrimSpace(idStr)
				if idStr != "" {
					if id, err := uuid.Parse(idStr); err == nil {
						filters.TeamIDs = append(filters.TeamIDs, id)
					}
				}
			}
		}
	}

	if sprintIds, ok := query["sprintIds"]; ok {
		if sprintIdsStr, ok := sprintIds.(string); ok && sprintIdsStr != "" {
			sprintIDStrings := strings.Split(sprintIdsStr, ",")
			for _, idStr := range sprintIDStrings {
				idStr = strings.TrimSpace(idStr)
				if idStr != "" {
					if id, err := uuid.Parse(idStr); err == nil {
						filters.SprintIDs = append(filters.SprintIDs, id)
					}
				}
			}
		}
	}

	if objectiveIds, ok := query["objectiveIds"]; ok {
		if objectiveIdsStr, ok := objectiveIds.(string); ok && objectiveIdsStr != "" {
			objectiveIDStrings := strings.Split(objectiveIdsStr, ",")
			for _, idStr := range objectiveIDStrings {
				idStr = strings.TrimSpace(idStr)
				if idStr != "" {
					if id, err := uuid.Parse(idStr); err == nil {
						filters.ObjectiveIDs = append(filters.ObjectiveIDs, id)
					}
				}
			}
		}
	}

	now := time.Now()
	defaultEndDate := now
	defaultStartDate := now.AddDate(0, 0, -60)

	if startDate, ok := query["startDate"]; ok {
		if startDateStr, ok := startDate.(string); ok {
			if parsedDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
				filters.StartDate = &parsedDate
			} else {
				filters.StartDate = &defaultStartDate
			}
		}
	} else {
		filters.StartDate = &defaultStartDate
	}

	if endDate, ok := query["endDate"]; ok {
		if endDateStr, ok := endDate.(string); ok {
			if parsedDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
				filters.EndDate = &parsedDate
			} else {
				filters.EndDate = &defaultEndDate
			}
		}
	} else {
		filters.EndDate = &defaultEndDate
	}

	return filters, nil
}

func (h *Handlers) GetWorkspaceOverview(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetWorkspaceOverview")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	overview, err := h.reports.GetWorkspaceOverview(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting workspace overview: %w", err)
	}

	return web.Respond(ctx, w, toAppWorkspaceOverview(overview), http.StatusOK)
}

func (h *Handlers) GetStoryAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetStoryAnalytics")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analytics, err := h.reports.GetStoryAnalytics(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting story analytics: %w", err)
	}

	return web.Respond(ctx, w, toAppStoryAnalytics(analytics), http.StatusOK)
}

func (h *Handlers) GetObjectiveProgress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetObjectiveProgress")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	progress, err := h.reports.GetObjectiveProgress(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting objective progress: %w", err)
	}

	return web.Respond(ctx, w, toAppObjectiveProgress(progress), http.StatusOK)
}

func (h *Handlers) GetTeamPerformance(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetTeamPerformance")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	performance, err := h.reports.GetTeamPerformance(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting team performance: %w", err)
	}

	return web.Respond(ctx, w, toAppTeamPerformance(performance), http.StatusOK)
}

func (h *Handlers) GetSprintAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetSprintAnalytics")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	analytics, err := h.reports.GetSprintAnalytics(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting sprint analytics: %w", err)
	}

	return web.Respond(ctx, w, toAppSprintAnalyticsWorkspace(analytics), http.StatusOK)
}

func (h *Handlers) GetTimelineTrends(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetTimelineTrends")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppReportFilters
	query, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filters, err := parseReportFilters(query)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	trends, err := h.reports.GetTimelineTrends(ctx, workspace.ID, filters)
	if err != nil {
		return fmt.Errorf("getting timeline trends: %w", err)
	}

	return web.Respond(ctx, w, toAppTimelineTrends(trends), http.StatusOK)
}
