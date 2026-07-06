package teamshttp

import (
	"net/http/httptest"
	"testing"
)

func TestPaginationParamsUsesMenuDefaultPageSize(t *testing.T) {
	request := httptest.NewRequest("GET", "/teams?page=2", nil)

	page, pageSize := paginationParams(request, menuPageSize, maxPageSize)

	if page != 2 {
		t.Fatalf("expected page 2, got %d", page)
	}
	if pageSize != menuPageSize {
		t.Fatalf("expected page size %d, got %d", menuPageSize, pageSize)
	}
}

func TestToAppTeamsResponseClearsNextPageWhenComplete(t *testing.T) {
	response := toAppTeamsResponse(nil, 3, menuPageSize, false)

	if response.Pagination.NextPage != 0 {
		t.Fatalf("expected no next page, got %d", response.Pagination.NextPage)
	}
}
