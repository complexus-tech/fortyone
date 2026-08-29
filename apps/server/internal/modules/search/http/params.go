package searchhttp

import (
	"errors"
	"math"
	"net/url"
	"strings"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidSearchType = errors.New("invalid search type")
	ErrInvalidSortOption = errors.New("invalid search sort option")
)

const (
	defaultSearchPageSize       = 20
	defaultSimilarStoriesLimit  = 3
	maximumSearchPage           = 1_000
	maximumSearchPageSize       = 100
	maximumSimilarStoriesLimit  = 5
	maximumSearchQueryRunes     = 200
	maximumSearchPriorityRunes  = 64
	maximumSimilarityTitleRunes = 500
	maximumSearchEnumRunes      = 32
	maximumSearchOffset         = (maximumSearchPage - 1) * maximumSearchPageSize
	maximumIntegerBytes         = 20
)

type similarStoriesQuery struct {
	TeamID *uuid.UUID
	Title  string
	Limit  int
}

func parseSearchParams(values url.Values) (search.SearchParams, error) {
	teamID, err := web.OptionalUUIDQueryParameter(values, "teamId")
	if err != nil {
		return search.SearchParams{}, err
	}
	assigneeID, err := web.OptionalUUIDQueryParameter(values, "assigneeId")
	if err != nil {
		return search.SearchParams{}, err
	}
	labelID, err := web.OptionalUUIDQueryParameter(values, "labelId")
	if err != nil {
		return search.SearchParams{}, err
	}
	statusID, err := web.OptionalUUIDQueryParameter(values, "statusId")
	if err != nil {
		return search.SearchParams{}, err
	}

	typeValue, err := parseSearchString(values, "type", maximumSearchEnumRunes)
	if err != nil {
		return search.SearchParams{}, err
	}
	searchType, err := parseSearchType(typeValue)
	if err != nil {
		return search.SearchParams{}, err
	}
	sortValue, err := parseSearchString(values, "sortBy", maximumSearchEnumRunes)
	if err != nil {
		return search.SearchParams{}, err
	}
	sortBy, err := parseSortOption(sortValue)
	if err != nil {
		return search.SearchParams{}, err
	}

	params, err := parseSearchPagination(values)
	if err != nil {
		return search.SearchParams{}, err
	}

	query, err := parseSearchString(values, "query", maximumSearchQueryRunes)
	if err != nil {
		return search.SearchParams{}, err
	}
	priorityValue, err := parseSearchString(values, "priority", maximumSearchPriorityRunes)
	if err != nil {
		return search.SearchParams{}, err
	}
	var priority *string
	if priorityValue != "" {
		priority = &priorityValue
	}
	return search.SearchParams{
		Type:       searchType,
		Query:      query,
		TeamID:     teamID,
		AssigneeID: assigneeID,
		LabelID:    labelID,
		StatusID:   statusID,
		Priority:   priority,
		SortBy:     sortBy,
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

func parseSearchPagination(values url.Values) (pagination.OffsetParams, error) {
	if _, _, err := web.OptionalIntegerQueryParameter(
		values, "page", maximumIntegerBytes, 1, maximumSearchPage,
	); err != nil {
		return pagination.OffsetParams{}, err
	}
	if _, _, err := web.OptionalIntegerQueryParameter(
		values, "pageSize", maximumIntegerBytes, 1, maximumSearchPageSize,
	); err != nil {
		return pagination.OffsetParams{}, err
	}
	return pagination.ParseOffsetQuery(values, pagination.OffsetQueryConfig{
		DefaultPageSize: defaultSearchPageSize,
		MaximumPageSize: maximumSearchPageSize,
		MaximumOffset:   maximumSearchOffset,
	})
}

func parseSimilarStoriesQuery(values url.Values) (similarStoriesQuery, error) {
	teamID, err := web.OptionalUUIDQueryParameter(values, "teamId")
	if err != nil {
		return similarStoriesQuery{}, err
	}
	title, err := parseSearchString(values, "title", maximumSimilarityTitleRunes)
	if err != nil {
		return similarStoriesQuery{}, err
	}
	limit, present, err := web.OptionalIntegerQueryParameter(values, "limit", maximumIntegerBytes, 1, math.MaxInt)
	if err != nil {
		return similarStoriesQuery{}, err
	}
	if !present {
		limit = defaultSimilarStoriesLimit
	}
	limit = min(limit, maximumSimilarStoriesLimit)
	return similarStoriesQuery{TeamID: teamID, Title: title, Limit: limit}, nil
}

func parseSearchType(value string) (search.SearchType, error) {
	switch strings.TrimSpace(value) {
	case "", string(search.SearchTypeAll):
		return search.SearchTypeAll, nil
	case string(search.SearchTypeStories):
		return search.SearchTypeStories, nil
	case string(search.SearchTypeObjectives):
		return search.SearchTypeObjectives, nil
	default:
		return "", ErrInvalidSearchType
	}
}

func parseSortOption(value string) (search.SortOption, error) {
	switch strings.TrimSpace(value) {
	case "", string(search.SortByRelevance):
		return search.SortByRelevance, nil
	case string(search.SortByUpdated):
		return search.SortByUpdated, nil
	case string(search.SortByCreated):
		return search.SortByCreated, nil
	default:
		return "", ErrInvalidSortOption
	}
}

func parseSearchString(values url.Values, name string, maximumRunes int) (string, error) {
	value, _, err := web.OptionalTextQueryParameter(values, name, maximumRunes*4, maximumRunes)
	return value, err
}
