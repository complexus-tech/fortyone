package storiesrepository

import (
	"strings"
	"testing"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

func TestNegatedFiltersAreIncludedInQueryAndParams(t *testing.T) {
	statusID := uuid.New()
	assigneeID := uuid.New()
	objectiveID := uuid.New()
	content := "deprecated"
	excludedDate := time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC)
	hasAssignee := true
	filters := stories.CoreStoryFilters{
		ExcludedStatusIDs:   []uuid.UUID{statusID},
		ExcludedAssigneeIDs: []uuid.UUID{assigneeID},
		TitleNotContains:    &content,
		ExcludedObjective:   &objectiveID,
		HasAssignee:         &hasAssignee,
		DeadlineNot:         &excludedDate,
	}
	repository := &repo{}

	query := repository.buildSimpleWhereClause(filters)
	for _, expected := range []string{
		"excluded_status_ids",
		"excluded_assignee_ids",
		"title_not_contains",
		"excluded_objective_id",
		"assignee_id IS NOT NULL",
		"deadline_not",
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
	if got := params["excluded_objective_id"]; got != objectiveID {
		t.Fatalf("expected excluded objective %s, got %v", objectiveID, got)
	}
	if got := params["deadline_not"]; got != excludedDate {
		t.Fatalf("expected deadline_not %s, got %v", excludedDate, got)
	}
}

func TestOrderDirectionIsAppliedToSelectedField(t *testing.T) {
	repository := &repo{}

	if got := repository.buildOrderByClause("created", "asc"); got != "s.created_at ASC" {
		t.Fatalf("expected ascending created order, got %q", got)
	}
	if got := repository.buildOrderByClause("deadline", "desc"); !strings.Contains(got, "s.end_date DESC") {
		t.Fatalf("expected descending deadline order, got %q", got)
	}
}
