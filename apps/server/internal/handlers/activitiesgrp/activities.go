package activitiesgrp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var ErrInvalidWorkspaceID = errors.New("invalid workspace id")
var ErrInvalidLimit = errors.New("invalid limit")

type Handlers struct {
	activities *activities.Service
	log        *logger.Logger
}

func New(log *logger.Logger, activities *activities.Service) *Handlers {
	return &Handlers{
		activities: activities,
		log:        log,
	}
}

// GetActivities returns a list of activities for the logged-in user.
func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	limit := 10
	if filters["limit"] != nil {

		limit, err = strconv.Atoi(filters["limit"].(string))
		if err != nil {
			web.RespondError(ctx, w, ErrInvalidLimit, http.StatusBadRequest)
			return nil
		}
	}

	acts, err := h.activities.GetActivities(ctx, userID, limit, workspace.ID)
	if err != nil {
		return err
	}

	return web.Respond(ctx, w, toAppActivities(acts), http.StatusOK)
}
