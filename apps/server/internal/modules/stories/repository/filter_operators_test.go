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

func TestCompletionStateFiltersAreIncludedInQuery(t *testing.T) {
	repository := &repo{}
	enabled := true
	builders := map[string]func(stories.CoreStoryFilters) string{
		"grouped SQL": repository.buildSimpleWhereClause,
		"flat group":  repository.buildSimpleStoriesQuery,
		"Go fallback": repository.buildStoriesQuery,
	}

	for name, build := range builders {
		t.Run(name, func(t *testing.T) {
			completedQuery := build(stories.CoreStoryFilters{IsCompleted: &enabled})
			if !strings.Contains(completedQuery, "s.completed_at IS NOT NULL") {
				t.Fatalf("expected completed query to require completed_at, got %q", completedQuery)
			}

			openQuery := build(stories.CoreStoryFilters{IsNotCompleted: &enabled})
			if !strings.Contains(openQuery, "s.completed_at IS NULL") {
				t.Fatalf("expected open query to exclude completed stories, got %q", openQuery)
			}
		})
	}
}

func TestAssigneeGroupsAreRestrictedToFilteredAssignees(t *testing.T) {
	repository := &repo{}
	cte := repository.buildAllGroupsCTE("assignee", stories.CoreStoryFilters{
		TeamIDs:     []uuid.UUID{uuid.New()},
		AssigneeIDs: []uuid.UUID{uuid.New()},
	})

	if !strings.Contains(cte, "u.user_id = ANY(:assignee_ids)") {
		t.Fatalf("expected assignee groups to use the assignee filter, got %q", cte)
	}
	if strings.Contains(cte, "Unassigned") {
		t.Fatalf("filtered assignee groups must not include the unassigned group, got %q", cte)
	}
}

func TestOrderDirectionIsAppliedToSelectedField(t *testing.T) {
	repository := &repo{}

	if got := repository.buildOrderByClause("created", "asc"); got != "s.created_at ASC, s.id ASC" {
		t.Fatalf("expected ascending created order, got %q", got)
	}
	if got := repository.buildOrderByClause("updated", "desc"); got != "s.updated_at DESC, s.id ASC" {
		t.Fatalf("expected descending updated order, got %q", got)
	}
	if got := repository.buildOrderByClause("priority", "desc"); !strings.HasSuffix(got, "s.created_at DESC, s.id ASC") {
		t.Fatalf("expected stable descending priority order, got %q", got)
	}
	if got := repository.buildOrderByClause("deadline", "desc"); got != "s.end_date DESC NULLS LAST, s.created_at DESC, s.id ASC" {
		t.Fatalf("expected descending deadline order, got %q", got)
	}
	if got := repository.buildOrderByClause("completed", "desc"); got != "s.completed_at DESC NULLS LAST, s.created_at DESC, s.id ASC" {
		t.Fatalf("expected descending completion order, got %q", got)
	}
	if got := repository.buildOrderByClauseWithAlias("completed", "asc", "ls"); got != "ls.completed_at ASC NULLS LAST, ls.created_at DESC, ls.id ASC" {
		t.Fatalf("expected ascending aliased completion order, got %q", got)
	}
	if got := repository.buildOrderByClauseWithAlias("created", "desc", "ls"); got != "ls.created_at DESC, ls.id ASC" {
		t.Fatalf("expected stable aliased created order, got %q", got)
	}
}

func TestPaginatedStoryOrderClausesHaveStableIDTieBreaker(t *testing.T) {
	repository := &repo{}

	for _, alias := range []string{"s", "ls"} {
		for _, orderBy := range []string{"created", "updated", "priority", "deadline"} {
			t.Run(alias+"_"+orderBy, func(t *testing.T) {
				got := repository.buildOrderByClauseWithAlias(orderBy, "desc", alias)
				if !strings.HasSuffix(got, alias+".id ASC") {
					t.Fatalf("expected %s order to end with a stable story ID tie-breaker, got %q", orderBy, got)
				}
			})
		}
	}
}

func TestSortStoriesInGroupOrdersByCompletedAt(t *testing.T) {
	repository := &repo{}
	earlier := time.Date(2026, time.August, 13, 9, 0, 0, 0, time.UTC)
	later := earlier.Add(time.Hour)
	createdEarlier := earlier.Add(-2 * time.Hour)
	createdLater := earlier.Add(-time.Hour)
	firstID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	secondID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	nilID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	tiedID := uuid.MustParse("00000000-0000-0000-0000-000000000004")
	tiedByID := uuid.MustParse("00000000-0000-0000-0000-000000000005")

	tests := []struct {
		name      string
		direction string
		want      []uuid.UUID
	}{
		{name: "ascending", direction: "asc", want: []uuid.UUID{tiedID, tiedByID, firstID, secondID, nilID}},
		{name: "descending", direction: "desc", want: []uuid.UUID{secondID, tiedID, tiedByID, firstID, nilID}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := stories.CoreStoryGroup{Stories: []stories.CoreStoryList{
				{ID: nilID, CreatedAt: createdLater},
				{ID: secondID, CompletedAt: &later, CreatedAt: createdEarlier},
				{ID: firstID, CompletedAt: &earlier, CreatedAt: createdLater},
				{ID: tiedID, CompletedAt: &earlier, CreatedAt: later},
				{ID: tiedByID, CompletedAt: &earlier, CreatedAt: later},
			}}

			repository.sortStoriesInGroup(&group, "completed", tt.direction)

			for index, wantID := range tt.want {
				if got := group.Stories[index].ID; got != wantID {
					t.Fatalf("story %d = %s, want %s", index, got, wantID)
				}
			}
		})
	}
}

func TestCollaborationFiltersAreIncludedInQueryAndPersonalScope(t *testing.T) {
	collaboratorID := uuid.New()
	include := true
	filters := stories.CoreStoryFilters{
		CollaboratorIDs:     []uuid.UUID{collaboratorID},
		AssignedToMe:        &include,
		CollaboratingWithMe: &include,
		CreatedByMe:         &include,
	}
	repository := &repo{}

	query := repository.buildSimpleWhereClause(filters)
	for _, expected := range []string{
		"sc_filter.user_id = ANY(:collaborator_ids)",
		"s.assignee_id = :current_user_id",
		"s.reporter_id = :current_user_id",
		"sc_me.user_id = :current_user_id",
		" OR ",
	} {
		if !strings.Contains(query, expected) {
			t.Fatalf("expected collaboration query to contain %q, got %q", expected, query)
		}
	}

	params := repository.buildQueryParams(filters)
	collaboratorIDs, ok := params["collaborator_ids"].([]uuid.UUID)
	if !ok || len(collaboratorIDs) != 1 || collaboratorIDs[0] != collaboratorID {
		t.Fatalf("expected collaborator filter %s, got %v", collaboratorID, params["collaborator_ids"])
	}
	if got := params["current_user_id"]; got == nil {
		t.Fatal("expected personal collaboration scope to include current_user_id")
	}
}
