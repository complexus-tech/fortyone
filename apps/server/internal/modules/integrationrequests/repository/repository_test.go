package integrationrequestsrepository

import (
	"strings"
	"testing"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/google/uuid"
)

func TestTeamAccessPredicateMatchesTeamAuthorizationConvention(t *testing.T) {
	t.Parallel()

	predicate := normalizeIntegrationRequestQuery(teamAccessPredicate("request.team_id", "request.workspace_id", "$9"))
	for _, clause := range []string{
		"FROM team_members request_team_member",
		"request_team_member.team_id = request.team_id",
		"request_team_member.user_id = $9",
		"FROM workspace_members request_workspace_member",
		"request_workspace_member.workspace_id = request.workspace_id",
		"request_workspace_member.user_id = $9",
		"request_workspace_member.role = 'admin'",
	} {
		if !strings.Contains(predicate, clause) {
			t.Fatalf("team access predicate is missing %q: %q", clause, predicate)
		}
	}
}

func TestTeamRequestFilterQueryBindsActorBeforeOptionalFilters(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	userID := uuid.New()
	assigneeID := uuid.New()
	query, args := teamRequestFilterQuery(workspaceID, teamID, userID, integrationrequests.CoreListRequestsFilter{
		Provider:   integrationrequests.ProviderSlack,
		Priority:   "High",
		AssigneeID: &assigneeID,
	})
	normalized := normalizeIntegrationRequestQuery(query)

	if len(args) != 7 {
		t.Fatalf("query arguments = %d, want 7", len(args))
	}
	if args[0] != workspaceID || args[1] != teamID || args[3] != userID {
		t.Fatalf("authorization arguments are not bound to workspace/team/user: %#v", args[:4])
	}
	for _, clause := range []string{
		"request_team_member.user_id = $4",
		"request_workspace_member.user_id = $4",
		"provider = $5",
		"priority = $6",
		"assignee_id = $7",
	} {
		if !strings.Contains(normalized, clause) {
			t.Fatalf("filtered query is missing %q: %q", clause, normalized)
		}
	}
}

func normalizeIntegrationRequestQuery(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
