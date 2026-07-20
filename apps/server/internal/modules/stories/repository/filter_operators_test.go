package storiesrepository

import (
	"strings"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestNegatedFiltersAreIncludedInQueryAndParams(t *testing.T) {
	statusID := uuid.New()
	assigneeID := uuid.New()
	content := "deprecated"
	filters := stories.CoreStoryFilters{
		ExcludedStatusIDs:   []uuid.UUID{statusID},
		ExcludedAssigneeIDs: []uuid.UUID{assigneeID},
		TitleNotContains:    &content,
	}
	repository := &repo{}

	query := repository.buildSimpleWhereClause(filters)
	for _, expected := range []string{
		"excluded_status_ids",
		"excluded_assignee_ids",
		"title_not_contains",
	} {
		if !strings.Contains(query, expected) {
			t.Fatalf("expected query to contain %q, got %q", expected, query)
		}
	}

	params := repository.buildQueryParams(filters)
	if _, ok := params["excluded_status_ids"]; !ok {
		t.Fatal("expected excluded status IDs in query params")
	}
	if _, ok := params["excluded_assignee_ids"]; !ok {
		t.Fatal("expected excluded assignee IDs in query params")
	}
	if got := params["title_not_contains"]; got != content {
		t.Fatalf("expected title_not_contains %q, got %v", content, got)
	}
}
