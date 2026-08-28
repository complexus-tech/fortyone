package integrationrequestshttp

import (
	"fmt"
	"math"
	"net/http"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	maximumRequestSearchRunes = 512
	maximumRequestFilterRunes = 64
	maximumRequestDateRunes   = 64
)

type requestListQuery struct {
	Filter   integrationrequests.CoreListRequestsFilter
	Page     int
	PageSize int
}

func parseRequestListQuery(r *http.Request) (requestListQuery, error) {
	values := r.URL.Query()
	status, _, err := web.OptionalTextQueryParameter(
		values, "status", maximumRequestFilterRunes*4, maximumRequestFilterRunes,
	)
	if err != nil {
		return requestListQuery{}, err
	}
	if !validRequestStatus(status) {
		return requestListQuery{}, &web.QueryParameterError{Name: "status", Cause: web.ErrInvalidQueryParameter}
	}
	provider, _, err := web.OptionalTextQueryParameter(
		values, "provider", maximumRequestFilterRunes*4, maximumRequestFilterRunes,
	)
	if err != nil {
		return requestListQuery{}, err
	}
	if !validRequestProvider(provider) {
		return requestListQuery{}, &web.QueryParameterError{Name: "provider", Cause: web.ErrInvalidQueryParameter}
	}
	priority, _, err := web.OptionalTextQueryParameter(
		values, "priority", maximumRequestFilterRunes*4, maximumRequestFilterRunes,
	)
	if err != nil {
		return requestListQuery{}, err
	}
	search, _, err := web.OptionalTextQueryParameter(
		values, "search", maximumRequestSearchRunes*4, maximumRequestSearchRunes,
	)
	if err != nil {
		return requestListQuery{}, err
	}
	assigneeID, err := web.OptionalUUIDQueryParameter(values, "assigneeId")
	if err != nil {
		return requestListQuery{}, err
	}
	createdAfter, err := optionalDateQuery(r, "createdAfter")
	if err != nil {
		return requestListQuery{}, err
	}
	createdBefore, err := optionalDateQuery(r, "createdBefore")
	if err != nil {
		return requestListQuery{}, err
	}
	if createdAfter != nil && createdBefore != nil && createdAfter.After(*createdBefore) {
		return requestListQuery{}, fmt.Errorf("createdAfter must not be after createdBefore")
	}

	page, err := pagination.ParseOffsetQuery(values, pagination.OffsetQueryConfig{
		DefaultPageSize: defaultRequestsPageSize,
		MaximumPageSize: maxRequestsPageSize,
		MaximumOffset:   math.MaxInt32,
	})
	if err != nil {
		return requestListQuery{}, err
	}
	return requestListQuery{
		Filter: integrationrequests.CoreListRequestsFilter{
			Search:        search,
			Status:        status,
			Provider:      provider,
			Priority:      priority,
			AssigneeID:    assigneeID,
			CreatedAfter:  createdAfter,
			CreatedBefore: createdBefore,
			Page:          page.Page,
			PageSize:      page.PageSize + 1,
		},
		Page:     page.Page,
		PageSize: page.PageSize,
	}, nil
}

func optionalDateQuery(r *http.Request, key string) (*time.Time, error) {
	value, present, err := web.OptionalTextQueryParameter(
		r.URL.Query(), key, maximumRequestDateRunes*4, maximumRequestDateRunes,
	)
	if err != nil {
		return nil, err
	}
	if !present || value == "" {
		return nil, nil
	}
	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		parsed, parseErr := time.Parse(layout, value)
		if parseErr == nil {
			return &parsed, nil
		}
	}
	return nil, &web.QueryParameterError{Name: key, Cause: web.ErrInvalidQueryParameter}
}

func validRequestStatus(status string) bool {
	switch status {
	case "", integrationrequests.StatusPending, integrationrequests.StatusAccepted, integrationrequests.StatusDeclined:
		return true
	default:
		return false
	}
}

func validRequestProvider(provider string) bool {
	switch provider {
	case "", integrationrequests.ProviderGitHub, integrationrequests.ProviderSlack, integrationrequests.ProviderIntercom:
		return true
	default:
		return false
	}
}
