package activitiesgrp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var ErrInvalidWorkspaceID = errors.New("invalid workspace id")
var ErrInvalidLimit = errors.New("invalid limit")
var ErrInvalidDate = errors.New("invalid date")
var avatarAccessURLExpiry = 24 * time.Hour

type Handlers struct {
	activities  *activities.Service
	log         *logger.Logger
	attachments *attachments.Service
}

func New(log *logger.Logger, activities *activities.Service, attachments *attachments.Service) *Handlers {
	return &Handlers{
		activities:  activities,
		log:         log,
		attachments: attachments,
	}
}

func (h *Handlers) resolveUserAvatarURL(ctx context.Context, avatar string) string {
	if h.attachments == nil {
		return avatar
	}
	resolved, err := h.attachments.ResolveProfileImageURL(ctx, avatar, avatarAccessURLExpiry)
	if err != nil {
		return ""
	}
	return resolved
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

	startDate, endDate, err := date.RangeFromQuery(r.URL.Query(), 30)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidDate, http.StatusBadRequest)
	}

	acts, err := h.activities.GetActivities(ctx, userID, limit, workspace.ID, activities.ActivityFilters{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return err
	}

	for i := range acts {
		acts[i].User.AvatarURL = h.resolveUserAvatarURL(ctx, acts[i].User.AvatarURL)
	}

	return web.Respond(ctx, w, toAppActivities(acts), http.StatusOK)
}
