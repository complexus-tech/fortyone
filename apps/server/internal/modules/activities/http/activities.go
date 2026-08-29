package activitieshttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
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
	now         func() time.Time
}

func New(log *logger.Logger, activities *activities.Service, attachments *attachments.Service) *Handlers {
	return &Handlers{
		activities:  activities,
		log:         log,
		attachments: attachments,
		now:         time.Now,
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

	query, err := parseActivityListQuery(r.URL.Query(), h.now())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	acts, err := h.activities.GetActivities(ctx, userID, query.Limit, workspace.ID, query.Filters)
	if err != nil {
		return err
	}

	for i := range acts {
		acts[i].User.AvatarURL = h.resolveUserAvatarURL(ctx, acts[i].User.AvatarURL)
	}

	return web.Respond(ctx, w, toAppActivities(acts), http.StatusOK)
}
