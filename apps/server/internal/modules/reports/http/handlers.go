package reportshttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var (
	ErrInvalidWorkspaceID = errors.New("invalid workspace id")
	ErrInvalidDate        = errors.New("invalid date parameter")
	ErrInvalidFilterID    = errors.New("invalid report filter identifier")
	avatarAccessURLExpiry = 24 * time.Hour
)

type Handlers struct {
	reports     *reports.Service
	log         *logger.Logger
	attachments *attachments.Service
	clock       platformclock.Clock
}

func New(log *logger.Logger, reports *reports.Service, attachments *attachments.Service) *Handlers {
	return NewWithClock(log, reports, attachments, platformclock.System{})
}

func NewWithClock(
	log *logger.Logger,
	reports *reports.Service,
	attachments *attachments.Service,
	clock platformclock.Clock,
) *Handlers {
	if clock == nil {
		clock = platformclock.System{}
	}
	return &Handlers{
		reports:     reports,
		log:         log,
		attachments: attachments,
		clock:       clock,
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

func respondReportError(ctx context.Context, w http.ResponseWriter, err error) error {
	switch {
	case errors.Is(err, reports.ErrReportsAccessDenied):
		return web.RespondError(ctx, w, err, http.StatusForbidden)
	case errors.Is(err, reports.ErrInvalidReportFilters), errors.Is(err, reports.ErrInvalidWorkspaceAnalyticsEvent):
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	default:
		return err
	}
}
