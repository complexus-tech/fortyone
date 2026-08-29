package storieshttp

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	maxStoryScalarQueryBytes = 512
	maxStoryListQueryBytes   = 4096
	maxStoryListItemBytes    = 256
	maxStoryListItems        = 100
	maxStoryPage             = 1000
	maxStoryPageSize         = 100
)

var storyListQueryLimits = web.QueryListLimits{
	MaxBytes:     maxStoryListQueryBytes,
	MaxItemBytes: maxStoryListItemBytes,
	MaxItems:     maxStoryListItems,
}

type storyDateFilter struct {
	key   string
	value **time.Time
}

func storyDateFilters(filters *StoryFilters) []storyDateFilter {
	return []storyDateFilter{
		{key: "createdAfter", value: &filters.CreatedAfter},
		{key: "createdBefore", value: &filters.CreatedBefore},
		{key: "updatedAfter", value: &filters.UpdatedAfter},
		{key: "updatedBefore", value: &filters.UpdatedBefore},
		{key: "startDateAfter", value: &filters.StartDateAfter},
		{key: "startDateBefore", value: &filters.StartDateBefore},
		{key: "startDateNot", value: &filters.StartDateNot},
		{key: "deadlineAfter", value: &filters.DeadlineAfter},
		{key: "deadlineBefore", value: &filters.DeadlineBefore},
		{key: "deadlineNot", value: &filters.DeadlineNot},
		{key: "completedAfter", value: &filters.CompletedAfter},
		{key: "completedBefore", value: &filters.CompletedBefore},
	}
}

func parseStringParam(r *http.Request, key, defaultValue string, maxBytes int) (string, error) {
	value, present, err := web.OptionalQueryParameter(r.URL.Query(), key, maxBytes)
	if err != nil {
		return "", err
	}
	if !present {
		return defaultValue, nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s must not be blank", key)
	}
	return value, nil
}

func parseOptionalStringParam(r *http.Request, key string, maxBytes int) (*string, error) {
	value, present, err := web.OptionalQueryParameter(r.URL.Query(), key, maxBytes)
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("%s must not be blank", key)
	}
	return &value, nil
}

func parseStrictIntParam(r *http.Request, key string, defaultValue, minimum, maximum int) (int, error) {
	value, present, err := web.OptionalIntegerQueryParameter(r.URL.Query(), key, 20, minimum, maximum)
	if err != nil {
		return 0, err
	}
	if !present {
		return defaultValue, nil
	}
	return value, nil
}

func parseStoryPagination(r *http.Request, defaultPage, defaultPageSize int) (int, int, error) {
	page, err := parseStrictIntParam(r, "page", defaultPage, 1, maxStoryPage)
	if err != nil {
		return 0, 0, err
	}
	pageSize, err := parseStrictIntParam(r, "pageSize", defaultPageSize, 1, maxStoryPageSize)
	if err != nil {
		return 0, 0, err
	}
	return page, pageSize, nil
}

func parseStrictUUIDArray(r *http.Request, key string) ([]uuid.UUID, error) {
	parts, present, err := web.OptionalListQueryParameter(r.URL.Query(), key, storyListQueryLimits)
	if err != nil || !present {
		return nil, err
	}
	result := make([]uuid.UUID, len(parts))
	for index, part := range parts {
		parsed, parseErr := uuid.Parse(part)
		if parseErr != nil || parsed == uuid.Nil {
			return nil, fmt.Errorf("%s must contain valid UUIDs", key)
		}
		result[index] = parsed
	}
	return result, nil
}

func parseStrictInt16Array(r *http.Request, key string) ([]int16, error) {
	parts, present, err := web.OptionalListQueryParameter(r.URL.Query(), key, storyListQueryLimits)
	if err != nil || !present {
		return nil, err
	}
	result := make([]int16, len(parts))
	for index, part := range parts {
		parsed, parseErr := strconv.ParseInt(part, 10, 16)
		if parseErr != nil {
			return nil, fmt.Errorf("%s must contain 16-bit integers", key)
		}
		result[index] = int16(parsed)
	}
	return result, nil
}

func parseStrictStringArray(r *http.Request, key string) ([]string, error) {
	parts, present, err := web.OptionalListQueryParameter(r.URL.Query(), key, storyListQueryLimits)
	if err != nil || !present {
		return nil, err
	}
	return parts, nil
}

func parseStrictOptionalUUID(r *http.Request, key string) (*uuid.UUID, error) {
	value, present, err := web.OptionalQueryParameter(r.URL.Query(), key, 64)
	if err != nil {
		return nil, err
	}
	value = strings.TrimSpace(value)
	if !present {
		return nil, nil
	}
	if value == "" {
		return nil, fmt.Errorf("%s must be one UUID", key)
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil {
		return nil, fmt.Errorf("%s must be a valid UUID", key)
	}
	return &parsed, nil
}

func parseStrictOptionalBool(r *http.Request, key string) (*bool, error) {
	value, present, err := web.OptionalBooleanQueryParameter(r.URL.Query(), key)
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, nil
	}
	return &value, nil
}

func parseStrictOptionalDate(r *http.Request, key string) (*time.Time, error) {
	value, present, err := web.OptionalQueryParameter(r.URL.Query(), key, 64)
	if err != nil {
		return nil, err
	}
	if !present {
		return nil, nil
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("%s must be one date or timestamp", key)
	}
	for _, format := range []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05"} {
		if parsed, parseErr := time.Parse(format, value); parseErr == nil {
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("%s must be an RFC3339 timestamp or ISO date", key)
}

func isValidOrderBy(orderBy string) bool {
	switch orderBy {
	case "created", "updated", "priority", "deadline", "completed":
		return true
	default:
		return false
	}
}

func isValidOrderDirection(direction string) bool {
	return direction == "asc" || direction == "desc"
}

func isValidGroupBy(groupBy string) bool {
	switch groupBy {
	case "status", "assignee", "priority", "team", "sprint", "none":
		return true
	default:
		return false
	}
}

func isValidStoryCategory(category string) bool {
	switch category {
	case "backlog", "unstarted", "started", "paused", "completed", "cancelled":
		return true
	default:
		return false
	}
}
