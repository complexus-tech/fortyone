package teamshttp

import (
	"net/http"
	"strconv"
)

const (
	menuPageSize = 15
	maxPageSize  = 100
)

func paginationRequested(r *http.Request) bool {
	query := r.URL.Query()
	return query.Has("page") || query.Has("pageSize")
}

func paginationParams(r *http.Request, defaultPageSize, maximumPageSize int) (int, int) {
	page := 1
	pageSize := defaultPageSize
	query := r.URL.Query()

	if pageValue := query.Get("page"); pageValue != "" {
		if parsed, err := strconv.Atoi(pageValue); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pageSizeValue := query.Get("pageSize"); pageSizeValue != "" {
		if parsed, err := strconv.Atoi(pageSizeValue); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	if pageSize > maximumPageSize {
		pageSize = maximumPageSize
	}

	return page, pageSize
}
