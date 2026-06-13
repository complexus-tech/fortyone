package reportsrepository

import (
	"strings"
	"testing"

	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	"github.com/google/uuid"
)

func TestBuildWorkloadStoryFilterSupportsAssigneeIDs(t *testing.T) {
	t.Parallel()

	assigneeID := uuid.New()
	namedParams := map[string]any{}

	filter := buildWorkloadStoryFilter(reports.ReportFilters{
		AssigneeIDs: []uuid.UUID{assigneeID},
	}, namedParams)

	if !strings.Contains(filter, "s.assignee_id = ANY(:assignee_ids)") {
		t.Fatalf("expected assignee filter in %q", filter)
	}
	if got := namedParams["assignee_ids"]; got == nil {
		t.Fatal("expected assignee_ids named parameter")
	}
}
