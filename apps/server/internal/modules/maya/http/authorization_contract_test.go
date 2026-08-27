package mayahttp

import (
	"os"
	"strings"
	"testing"
)

func TestWorkPlanRoutesRequireWorkspaceMemberRole(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read Maya routes: %v", err)
	}
	routes := string(source)

	required := []string{
		`app.Post("/workspaces/{workspaceSlug}/maya/work-plans", h.CreateWorkPlan, auth, workspace, memberAndAdmin)`,
		`app.Post("/workspaces/{workspaceSlug}/maya/work-plans/{runId}/apply", h.ApplyWorkPlan, auth, workspace, memberAndAdmin)`,
	}
	for _, contract := range required {
		if !strings.Contains(routes, contract) {
			t.Fatalf("Maya work-plan route is missing member authorization contract %q", contract)
		}
	}
}
