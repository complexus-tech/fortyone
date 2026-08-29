package okractivitiesrepository

import (
	"errors"
	"math"
	"os"
	"regexp"
	"strings"
	"testing"

	okractivitiesdomain "github.com/complexus-tech/projects-api/internal/modules/okractivities/domain"
)

func TestActivityQueriesEnforceTenantMembershipAndResourceScope(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/activities.sql")
	if err != nil {
		t.Fatalf("read activity queries: %v", err)
	}
	queries := string(contents)
	for _, contract := range []string{
		"objective.workspace_id = sqlc.arg(workspace_id)",
		"activity.workspace_id = sqlc.arg(workspace_id)",
		"public.workspace_members",
		"public.team_members",
		"actor.is_active = TRUE",
		"membership.role IN ('member', 'admin')",
		"key_result.objective_id = objective.objective_id",
		"key_result.team_id = objective.team_id",
	} {
		if !strings.Contains(queries, contract) {
			t.Errorf("activity query contract is missing %q", contract)
		}
	}
	if regexp.MustCompile(`(?i)select\s+(?:[a-z_][a-z0-9_]*\.)?\*`).MatchString(queries) {
		t.Fatal("activity SQL contains a wildcard projection")
	}
}

func TestActivityQueriesUseDeterministicBoundedPagination(t *testing.T) {
	t.Parallel()

	contents, err := os.ReadFile("queries/activities.sql")
	if err != nil {
		t.Fatalf("read activity queries: %v", err)
	}
	queries := string(contents)
	if count := strings.Count(queries, "ORDER BY activity.created_at DESC, activity.activity_id DESC"); count != 2 {
		t.Fatalf("stable activity order appears %d times, want 2", count)
	}
	if count := strings.Count(queries, "LIMIT CAST(sqlc.arg(result_limit) AS integer)"); count != 2 {
		t.Fatalf("bounded activity limit appears %d times, want 2", count)
	}
}

func TestActivityPageBoundsAreCheckedAndFetchOneExtraRow(t *testing.T) {
	t.Parallel()

	offset, limit, err := activityPageBounds(3, 20)
	if err != nil || offset != 40 || limit != 21 {
		t.Fatalf("activityPageBounds(3,20) = %d/%d/%v", offset, limit, err)
	}
	if _, _, err := activityPageBounds(math.MaxInt, 100); !errors.Is(err, okractivitiesdomain.ErrInvalid) {
		t.Fatalf("overflow bounds error = %v, want ErrInvalid", err)
	}
}
