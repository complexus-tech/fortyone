package labelshttp

import (
	"math"
	"net/url"
	"strings"
	"unicode/utf8"

	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	maximumLabelSearchRunes           = 255
	maximumLabelSearchBytes           = 4 * maximumLabelSearchRunes
	maximumLabelIntegerParameterBytes = 20
	maximumLabelOffset                = math.MaxInt32
)

type labelListQuery struct {
	Filters labels.LabelFilters
	Page    *pagination.OffsetParams
}

func parseLabelListQuery(values url.Values) (labelListQuery, error) {
	teamID, err := web.OptionalUUIDQueryParameter(values, "teamId")
	if err != nil {
		return labelListQuery{}, err
	}
	search, err := parseLabelSearch(values)
	if err != nil {
		return labelListQuery{}, err
	}
	result := labelListQuery{Filters: labels.LabelFilters{TeamID: teamID, Search: search}}

	page, hasPage, err := web.OptionalIntegerQueryParameter(
		values, "page", maximumLabelIntegerParameterBytes, 1, maximumLabelOffset,
	)
	if err != nil {
		return labelListQuery{}, err
	}
	pageSize, hasPageSize, err := web.OptionalIntegerQueryParameter(
		values, "pageSize", maximumLabelIntegerParameterBytes, 1, pagination.MaximumPageSize,
	)
	if err != nil {
		return labelListQuery{}, err
	}
	limit, hasLimit, err := web.OptionalIntegerQueryParameter(
		values, "limit", maximumLabelIntegerParameterBytes, 1, pagination.MaximumPageSize+1,
	)
	if err != nil {
		return labelListQuery{}, err
	}
	offset, hasOffset, err := web.OptionalIntegerQueryParameter(
		values, "offset", maximumLabelIntegerParameterBytes, 0, maximumLabelOffset,
	)
	if err != nil {
		return labelListQuery{}, err
	}

	if hasPage || hasPageSize {
		if hasLimit || hasOffset {
			return labelListQuery{}, invalidLabelQueryParameter("pagination")
		}
		if !hasPage {
			page = 1
		}
		if !hasPageSize {
			pageSize = pagination.DefaultMenuPageSize
		}
		if page-1 > maximumLabelOffset/pageSize {
			return labelListQuery{}, invalidLabelQueryParameter("page")
		}
		resultOffset := (page - 1) * pageSize
		resultLimit := pageSize + 1
		result.Filters.Limit = &resultLimit
		result.Filters.Offset = resultOffset
		result.Page = &pagination.OffsetParams{Page: page, PageSize: pageSize}
		return result, nil
	}

	if hasOffset && !hasLimit {
		return labelListQuery{}, invalidLabelQueryParameter("offset")
	}
	if hasLimit {
		result.Filters.Limit = &limit
	}
	if hasOffset {
		result.Filters.Offset = offset
	}
	return result, nil
}

func parseLabelSearch(values url.Values) (string, error) {
	search, _, err := web.OptionalQueryParameter(values, "search", maximumLabelSearchBytes)
	if err != nil {
		return "", err
	}
	if !utf8.ValidString(search) || strings.ContainsRune(search, '\x00') || utf8.RuneCountInString(search) > maximumLabelSearchRunes {
		return "", invalidLabelQueryParameter("search")
	}
	return strings.TrimSpace(search), nil
}

func invalidLabelQueryParameter(name string) error {
	return &web.QueryParameterError{Name: name, Cause: web.ErrInvalidQueryParameter}
}
