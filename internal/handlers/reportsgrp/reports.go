package reportsgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

	days := 7 // Default to 7 days
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
