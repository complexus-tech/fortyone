package search

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var ErrInvalidSearchParams = errors.New("invalid search parameters")

const (
	defaultSearchPage       = 1
	defaultSearchPageSize   = 20
	maxSearchPage           = 1_000
	maxSearchPageSize       = 100
	maxSearchQueryRunes     = 200
	maxSearchPriorityRunes  = 64
	maxSimilarityTitleRunes = 500
)

func normalizeSearchParams(params SearchParams) (SearchParams, error) {
	params.Query = strings.TrimSpace(params.Query)
	if utf8.RuneCountInString(params.Query) > maxSearchQueryRunes {
		return SearchParams{}, invalidSearchParams("query exceeds %d characters", maxSearchQueryRunes)
	}

	if params.Type == "" {
		params.Type = SearchTypeAll
	}
	switch params.Type {
	case SearchTypeAll, SearchTypeStories, SearchTypeObjectives:
	default:
		return SearchParams{}, invalidSearchParams("unsupported search type")
	}

	if params.SortBy == "" {
		params.SortBy = SortByRelevance
	}
	switch params.SortBy {
	case SortByRelevance, SortByUpdated, SortByCreated:
	default:
		return SearchParams{}, invalidSearchParams("unsupported sort option")
	}

	if params.Page == 0 {
		params.Page = defaultSearchPage
	}
	if params.Page < 1 || params.Page > maxSearchPage {
		return SearchParams{}, invalidSearchParams("page must be between 1 and %d", maxSearchPage)
	}
	if params.PageSize == 0 {
		params.PageSize = defaultSearchPageSize
	}
	if params.PageSize < 1 || params.PageSize > maxSearchPageSize {
		return SearchParams{}, invalidSearchParams("page size must be between 1 and %d", maxSearchPageSize)
	}

	for name, value := range map[string]*uuid.UUID{
		"team":     params.TeamID,
		"assignee": params.AssigneeID,
		"label":    params.LabelID,
		"status":   params.StatusID,
	} {
		if value != nil && *value == uuid.Nil {
			return SearchParams{}, invalidSearchParams("%s ID is required", name)
		}
	}

	if params.Priority != nil {
		priority := strings.TrimSpace(*params.Priority)
		if priority == "" {
			params.Priority = nil
		} else {
			if utf8.RuneCountInString(priority) > maxSearchPriorityRunes {
				return SearchParams{}, invalidSearchParams("priority exceeds %d characters", maxSearchPriorityRunes)
			}
			params.Priority = &priority
		}
	}
	return params, nil
}

func normalizeSimilarityTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if utf8.RuneCountInString(title) > maxSimilarityTitleRunes {
		return "", invalidSearchParams("title exceeds %d characters", maxSimilarityTitleRunes)
	}
	return title, nil
}

func invalidSearchParams(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidSearchParams, fmt.Sprintf(format, args...))
}
