package adminhttp

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	maximumAdminSearchRunes = 512
	maximumAdminFilterRunes = 128
	maximumAdminTimeRunes   = 64
)

type auditLogQuery struct {
	WorkspaceID *uuid.UUID
	TargetType  string
	Query       string
	Action      string
	ActorQuery  string
	From        *time.Time
	To          *time.Time
}

func paginationParams(request *http.Request) (int, int, error) {
	query := request.URL.Query()
	translated := make(url.Values, 2)
	if values, present := query["page"]; present {
		translated["page"] = append([]string(nil), values...)
	}
	if values, present := query["limit"]; present {
		translated["pageSize"] = append([]string(nil), values...)
	}
	params, err := pagination.ParseOffsetQuery(translated, pagination.OffsetQueryConfig{
		DefaultPageSize: defaultPageSize,
		MaximumPageSize: maxPageSize,
		MaximumOffset:   math.MaxInt32,
	})
	if err != nil {
		return 0, 0, err
	}
	return params.Page, params.PageSize, nil
}

func optionalUUIDQuery(request *http.Request, key string) (*uuid.UUID, error) {
	return web.OptionalUUIDQueryParameter(request.URL.Query(), key)
}

func optionalTextQuery(request *http.Request, key string, maximumRunes int) (string, error) {
	value, _, err := web.OptionalTextQueryParameter(request.URL.Query(), key, maximumRunes*4, maximumRunes)
	return value, err
}

func optionalTimeQuery(request *http.Request, key string) (*time.Time, error) {
	raw, present, err := web.OptionalTextQueryParameter(
		request.URL.Query(), key, maximumAdminTimeRunes*4, maximumAdminTimeRunes,
	)
	if err != nil {
		return nil, err
	}
	if !present || raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		parsed, err = time.Parse(time.DateOnly, raw)
	}
	if err != nil {
		return nil, &web.QueryParameterError{Name: key, Cause: web.ErrInvalidQueryParameter}
	}
	return &parsed, nil
}

func parseAuditLogQuery(request *http.Request) (auditLogQuery, error) {
	workspaceID, err := optionalUUIDQuery(request, "workspaceId")
	if err != nil {
		return auditLogQuery{}, err
	}
	targetType, err := optionalTextQuery(request, "targetType", maximumAdminFilterRunes)
	if err != nil {
		return auditLogQuery{}, err
	}
	query, err := optionalTextQuery(request, "q", maximumAdminSearchRunes)
	if err != nil {
		return auditLogQuery{}, err
	}
	action, err := optionalTextQuery(request, "action", maximumAdminFilterRunes)
	if err != nil {
		return auditLogQuery{}, err
	}
	actorQuery, err := optionalTextQuery(request, "actor", maximumAdminSearchRunes)
	if err != nil {
		return auditLogQuery{}, err
	}
	from, err := optionalTimeQuery(request, "from")
	if err != nil {
		return auditLogQuery{}, err
	}
	to, err := optionalTimeQuery(request, "to")
	if err != nil {
		return auditLogQuery{}, err
	}
	if from != nil && to != nil && from.After(*to) {
		return auditLogQuery{}, fmt.Errorf("from must not be after to")
	}
	return auditLogQuery{
		WorkspaceID: workspaceID,
		TargetType:  targetType,
		Query:       query,
		Action:      action,
		ActorQuery:  actorQuery,
		From:        from,
		To:          to,
	}, nil
}
