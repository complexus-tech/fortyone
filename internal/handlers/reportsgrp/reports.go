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
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("invalid workspace id")
	ErrInvalidDays        = errors.New("invalid days parameter")
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

// GetStoryStats returns story statistics for a workspace.
func (h *Handlers) GetStoryStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	stats, err := h.reports.GetStoryStats(ctx, workspaceID)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppStoryStats(stats), http.StatusOK)
}

// GetContributionStats returns contribution statistics for the logged-in user.
func (h *Handlers) GetContributionStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	stats, err := h.reports.GetContributionStats(ctx, userID, workspaceID, days)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppContributionsStats(stats), http.StatusOK)
}

// GetUserStats returns user-specific statistics for the logged-in user.
func (h *Handlers) GetUserStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	stats, err := h.reports.GetUserStats(ctx, userID, workspaceID)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppUserStats(stats), http.StatusOK)
}

// GetStatusStats returns status statistics for stories
func (h *Handlers) GetStatusStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetStatusStats")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	stats, err := h.reports.GetStatusStats(ctx, workspaceID, coreFilters)
	if err != nil {
		return fmt.Errorf("getting status stats: %w", err)
	}

	return web.Respond(ctx, w, toAppStatusStats(stats), http.StatusOK)
}

// GetPriorityStats returns priority statistics for stories
func (h *Handlers) GetPriorityStats(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetPriorityStats")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	stats, err := h.reports.GetPriorityStats(ctx, workspaceID, coreFilters)
	if err != nil {
		return fmt.Errorf("getting priority stats: %w", err)
	}

	return web.Respond(ctx, w, toAppPriorityStats(stats), http.StatusOK)
}

// Workspace Reports Handlers

// parseReportFilters parses query parameters into ReportFilters
// If no teamIds are provided, empty array means "include all teams in workspace"
// If no sprintIds are provided, empty array means "include all sprints in workspace"
// If no objectiveIds are provided, empty array means "include all objectives in workspace"
func parseReportFilters(query map[string]interface{}) (reports.ReportFilters, error) {
	filters := reports.ReportFilters{}

	// Parse multiple team IDs (empty = all teams in workspace)
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
	// Note: Empty TeamIDs slice means "include all teams in workspace"

	// Parse multiple sprint IDs (empty = all sprints in workspace)
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
	// Note: Empty SprintIDs slice means "include all sprints in workspace"

	// Parse multiple objective IDs (empty = all objectives in workspace)
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
	// Note: Empty ObjectiveIDs slice means "include all objectives in workspace"

	// Set default dates: current date as end, 60 days before as start
	now := time.Now()
	defaultEndDate := now
	defaultStartDate := now.AddDate(0, 0, -60) // 60 days ago

	// Parse start date or use default
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

	// Parse end date or use default
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

// GetWorkspaceOverview returns workspace overview with key metrics and trends
func (h *Handlers) GetWorkspaceOverview(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetWorkspaceOverview")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	overview, err := h.reports.GetWorkspaceOverview(ctx, workspaceID, filters)
	if err != nil {
		return fmt.Errorf("getting workspace overview: %w", err)
	}

	return web.Respond(ctx, w, toAppWorkspaceOverview(overview), http.StatusOK)
}

// GetStoryAnalytics returns story analytics including status breakdown and burndown
func (h *Handlers) GetStoryAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetStoryAnalytics")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	analytics, err := h.reports.GetStoryAnalytics(ctx, workspaceID, filters)
	if err != nil {
		return fmt.Errorf("getting story analytics: %w", err)
	}

	return web.Respond(ctx, w, toAppStoryAnalytics(analytics), http.StatusOK)
}

// GetObjectiveProgress returns objective progress including health and key results
func (h *Handlers) GetObjectiveProgress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetObjectiveProgress")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	progress, err := h.reports.GetObjectiveProgress(ctx, workspaceID, filters)
	if err != nil {
		return fmt.Errorf("getting objective progress: %w", err)
	}

	return web.Respond(ctx, w, toAppObjectiveProgress(progress), http.StatusOK)
}

// GetTeamPerformance returns team performance including workload and velocity
func (h *Handlers) GetTeamPerformance(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetTeamPerformance")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	performance, err := h.reports.GetTeamPerformance(ctx, workspaceID, filters)
	if err != nil {
		return fmt.Errorf("getting team performance: %w", err)
	}

	return web.Respond(ctx, w, toAppTeamPerformance(performance), http.StatusOK)
}

// GetSprintAnalytics returns sprint analytics including progress and burndown
func (h *Handlers) GetSprintAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetSprintAnalytics")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	analytics, err := h.reports.GetSprintAnalytics(ctx, workspaceID, filters)
	if err != nil {
		return fmt.Errorf("getting sprint analytics: %w", err)
	}

	return web.Respond(ctx, w, toAppSprintAnalyticsWorkspace(analytics), http.StatusOK)
}

// GetTimelineTrends returns timeline trends for all key metrics
func (h *Handlers) GetTimelineTrends(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.reports.GetTimelineTrends")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	trends, err := h.reports.GetTimelineTrends(ctx, workspaceID, filters)
	if err != nil {
		return fmt.Errorf("getting timeline trends: %w", err)
	}

	return web.Respond(ctx, w, toAppTimelineTrends(trends), http.StatusOK)
}
