package searchrepository

import (
	"os"
	"strings"
	"testing"
)

func TestSearchEnrichmentIsPartOfTheBoundedPageQueries(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("queries.go")
	if err != nil {
		t.Fatalf("read queries.go: %v", err)
	}
	source := strings.ToLower(string(data))

	for _, required := range []string{
		"search_status.name as status_name",
		"search_team.name as team_name",
		"search_assignee.full_name as assignee_full_name",
		"search_lead.full_name as lead_full_name",
		"search_assignee.is_active = true",
		"search_lead.is_active = true",
		"search_assignee_membership.workspace_id = s.workspace_id",
		"search_lead_membership.workspace_id = o.workspace_id",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("page-local search enrichment must include %q", required)
		}
	}

	if count := strings.Count(source, "limit :page_size offset :offset"); count != 2 {
		t.Fatalf("story and objective enrichment must remain page-bounded; got %d page limits", count)
	}
}
