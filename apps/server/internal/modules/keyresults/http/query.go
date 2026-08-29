package keyresultshttp

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	maximumKeyResultQueryOffset = math.MaxInt32
	maximumKeyResultFilterItems = 100
	maximumKeyResultFilterBytes = 4096
	maximumKeyResultUUIDBytes   = 64
	maximumKeyResultDateBytes   = 64
)

var supportedKeyResultSortFields = map[string]struct{}{
	"created_at":     {},
	"name":           {},
	"objective_name": {},
	"updated_at":     {},
}

var supportedKeyResultMeasurements = map[string]struct{}{
	"boolean":    {},
	"number":     {},
	"percentage": {},
}

func parseKeyResultFilters(r *http.Request, workspaceID, userID uuid.UUID) (keyresults.CoreKeyResultFilters, error) {
	filters, err := parseKeyResultFiltersQuery(r.URL.Query(), workspaceID, userID)
	if err != nil {
		if errors.Is(err, ErrInvalidFilters) {
			return keyresults.CoreKeyResultFilters{}, err
		}
		return keyresults.CoreKeyResultFilters{}, fmt.Errorf("%w: %w", ErrInvalidFilters, err)
	}
	return filters, nil
}

func parseKeyResultFiltersQuery(query url.Values, workspaceID, userID uuid.UUID) (keyresults.CoreKeyResultFilters, error) {
	page, err := pagination.ParseOffsetQuery(query, pagination.OffsetQueryConfig{
		DefaultPageSize: keyresults.DefaultPageSize,
		MaximumPageSize: keyresults.MaximumPageSize,
		MaximumOffset:   maximumKeyResultQueryOffset,
	})
	if err != nil {
		return keyresults.CoreKeyResultFilters{}, err
	}
	orderBy, err := keyResultEnumQuery(query, "orderBy", "created_at", supportedKeyResultSortFields)
	if err != nil {
		return keyresults.CoreKeyResultFilters{}, err
	}
	orderDirection, err := keyResultEnumQuery(query, "orderDirection", "desc", map[string]struct{}{
		"asc": {}, "desc": {},
	})
	if err != nil {
		return keyresults.CoreKeyResultFilters{}, err
	}

	filters := keyresults.CoreKeyResultFilters{
		WorkspaceID: workspaceID, CurrentUserID: userID,
		Page: page.Page, PageSize: page.PageSize,
		OrderBy: orderBy, OrderDirection: orderDirection,
	}
	for _, filter := range []struct {
		name        string
		destination *[]uuid.UUID
	}{
		{name: "teamIds", destination: &filters.TeamIDs},
		{name: "objectiveIds", destination: &filters.ObjectiveIDs},
		{name: "leadIds", destination: &filters.LeadIDs},
		{name: "createdBy", destination: &filters.CreatedBy},
	} {
		values, parseErr := parseKeyResultUUIDList(query, filter.name)
		if parseErr != nil {
			return keyresults.CoreKeyResultFilters{}, parseErr
		}
		*filter.destination = values
	}

	measurements, _, err := web.OptionalListQueryParameter(query, "measurementTypes", web.QueryListLimits{
		MaxBytes: maximumKeyResultFilterBytes, MaxItemBytes: 16, MaxItems: 3,
	})
	if err != nil {
		return keyresults.CoreKeyResultFilters{}, err
	}
	for _, measurement := range measurements {
		if _, supported := supportedKeyResultMeasurements[measurement]; !supported {
			return keyresults.CoreKeyResultFilters{}, fmt.Errorf("%w: unsupported measurement type", ErrInvalidFilters)
		}
	}
	filters.MeasurementTypes = measurements

	for _, dateFilter := range []struct {
		name        string
		destination **time.Time
	}{
		{name: "createdAfter", destination: &filters.CreatedAfter},
		{name: "createdBefore", destination: &filters.CreatedBefore},
		{name: "endDateAfter", destination: &filters.EndDateAfter},
		{name: "endDateBefore", destination: &filters.EndDateBefore},
		{name: "updatedAfter", destination: &filters.UpdatedAfter},
		{name: "updatedBefore", destination: &filters.UpdatedBefore},
	} {
		value, parseErr := parseKeyResultDate(query, dateFilter.name)
		if parseErr != nil {
			return keyresults.CoreKeyResultFilters{}, parseErr
		}
		*dateFilter.destination = value
	}
	if invalidKeyResultDateRange(filters.CreatedAfter, filters.CreatedBefore) ||
		invalidKeyResultDateRange(filters.EndDateAfter, filters.EndDateBefore) ||
		invalidKeyResultDateRange(filters.UpdatedAfter, filters.UpdatedBefore) {
		return keyresults.CoreKeyResultFilters{}, fmt.Errorf("%w: date range is reversed", ErrInvalidFilters)
	}

	return filters, nil
}

func keyResultEnumQuery(query url.Values, name, defaultValue string, supported map[string]struct{}) (string, error) {
	value, present, err := web.OptionalTextQueryParameter(query, name, 32, 32)
	if err != nil {
		return "", err
	}
	if !present {
		return defaultValue, nil
	}
	if value == "" {
		return "", &web.QueryParameterError{Name: name, Cause: web.ErrInvalidQueryParameter}
	}
	if _, ok := supported[value]; !ok {
		return "", fmt.Errorf("%w: unsupported %s", ErrInvalidFilters, name)
	}
	return value, nil
}

func parseKeyResultUUIDList(query url.Values, name string) ([]uuid.UUID, error) {
	values, present, err := web.OptionalListQueryParameter(query, name, web.QueryListLimits{
		MaxBytes: maximumKeyResultFilterBytes, MaxItemBytes: maximumKeyResultUUIDBytes, MaxItems: maximumKeyResultFilterItems,
	})
	if err != nil || !present {
		return nil, err
	}
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		id, parseErr := uuid.Parse(value)
		if parseErr != nil || id == uuid.Nil {
			return nil, fmt.Errorf("%w: %s must contain non-zero UUIDs", ErrInvalidFilters, name)
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func parseKeyResultDate(query url.Values, name string) (*time.Time, error) {
	value, present, err := web.OptionalTextQueryParameter(
		query, name, maximumKeyResultDateBytes, maximumKeyResultDateBytes,
	)
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, nil
	}
	if value == "" {
		return nil, &web.QueryParameterError{Name: name, Cause: web.ErrInvalidQueryParameter}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("%w: %s must be RFC3339", ErrInvalidFilters, name)
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func invalidKeyResultDateRange(after, before *time.Time) bool {
	return after != nil && before != nil && after.After(*before)
}
