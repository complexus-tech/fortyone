package searchrepository

import (
	"os"
	"strings"
	"testing"
)

func TestSearchEnrichmentIsPartOfTheBoundedPageQueries(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("queries/search.sql")
	if err != nil {
		t.Fatalf("read typed search queries: %v", err)
	}
	source := strings.ToLower(string(data))

	for _, required := range []string{
		"search_status.name as status_name",
		"search_team.name as team_name",
		"search_assignee.full_name as assignee_full_name",
		"search_lead.full_name as lead_full_name",
		"search_assignee.is_active = true",
		"search_lead.is_active = true",
		"search_assignee_membership.workspace_id = story.workspace_id",
		"search_lead_membership.workspace_id = objective.workspace_id",
		"actor.is_active = true",
		"actor.is_system = false",
		"count(*) over () as total_count",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("page-local search enrichment must include %q", required)
		}
	}

	if count := strings.Count(source, "limit sqlc.arg(page_limit)"); count != 2 {
		t.Fatalf("story and objective enrichment must remain page-bounded; got %d page limits", count)
	}
	if count := strings.Count(source, "offset sqlc.arg(page_offset)"); count != 2 {
		t.Fatalf("story and objective queries must use typed page offsets; got %d", count)
	}
}
