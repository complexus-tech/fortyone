package notificationshttp

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/url"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var (
	ErrInvalidNotificationID = errors.New("notification id is not in its proper form")
	ErrInvalidWorkspaceID    = errors.New("workspace id is not in its proper form")
)

const notificationAvatarExpiry = 24 * time.Hour

type profileImageResolver interface {
	ResolveProfileImageURL(context.Context, string, time.Duration) (string, error)
}

type Handlers struct {
	notifications *notifications.Service
	users         *users.Service
	profileImages profileImageResolver
	log           *logger.Logger
}

func New(notificationService *notifications.Service, usersService *users.Service, profileImages profileImageResolver, log *logger.Logger) *Handlers {
	return &Handlers{
		notifications: notificationService,
		users:         usersService,
		profileImages: profileImages,
		log:           log,
	}
}

func respondNotificationError(ctx context.Context, response http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, notifications.ErrInvalid), errors.Is(err, safecast.ErrOutOfRange):
		status = http.StatusBadRequest
	case errors.Is(err, notifications.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, notifications.ErrNotificationNotFound):
		status = http.StatusNotFound
	case errors.Is(err, notifications.ErrConflict):
		status = http.StatusConflict
	}
	return web.RespondError(ctx, response, err, status)
}

type inboxListQuery struct {
	Pagination pagination.OffsetParams
	Offset     int
	Search     string
}

type portalListQuery struct {
	Pagination pagination.OffsetParams
	UnreadOnly bool
}

func parsePortalListQuery(values url.Values) (portalListQuery, error) {
	params, err := pagination.ParseOffsetQuery(values, pagination.OffsetQueryConfig{
		DefaultPageSize: 10,
		MaximumPageSize: notificationsdomain.MaximumPageSize,
		MaximumOffset:   math.MaxInt32,
	})
	if err != nil {
		return portalListQuery{}, err
	}
	unreadOnly, _, err := web.OptionalBooleanQueryParameter(values, "unreadOnly")
	if err != nil {
		return portalListQuery{}, err
	}
	return portalListQuery{Pagination: params, UnreadOnly: unreadOnly}, nil
}

// parseInboxListQuery retains the documented legacy limit/offset precedence
// while making every explicit value bounded and unambiguous.
func parseInboxListQuery(values url.Values) (inboxListQuery, error) {
	params, err := pagination.ParseOffsetQuery(values, pagination.OffsetQueryConfig{
		DefaultPageSize: 25,
		MaximumPageSize: notificationsdomain.MaximumPageSize,
		MaximumOffset:   math.MaxInt32,
	})
	if err != nil {
		return inboxListQuery{}, err
	}

	limit, present, err := web.OptionalIntegerQueryParameter(values, "limit", 20, 1, math.MaxInt)
	if err != nil {
		return inboxListQuery{}, err
	}
	if present {
		params.Page = 1
		params.PageSize = min(limit, notificationsdomain.MaximumPageSize)
	}

	offset, offsetPresent, err := web.OptionalIntegerQueryParameter(values, "offset", 20, 0, math.MaxInt32)
	if err != nil {
		return inboxListQuery{}, err
	}
	if !offsetPresent {
		offset = params.Offset()
	} else {
		params.Page = offset/params.PageSize + 1
	}

	search, _, err := web.OptionalTextQueryParameter(
		values, "search", notificationsdomain.MaximumSearchBytes, notificationsdomain.MaximumSearchBytes,
	)
	if err != nil {
		return inboxListQuery{}, err
	}

	return inboxListQuery{
		Pagination: params,
		Offset:     offset,
		Search:     search,
	}, nil
}
