package teamsrepository

import (
	"errors"
	"strings"
	"testing"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
)

func TestTeamMembershipQueriesBindWorkspaceAndActor(t *testing.T) {
	t.Parallel()

	addQuery := strings.ToLower(addMemberQuery)
	for _, required := range []string{
		"team.workspace_id = :workspace_id",
		"membership.workspace_id = team.workspace_id",
		"membership.user_id = :user_id",
		"member.is_active = true",
		"select team_id, user_id",
		"from eligible_member",
	} {
		if !strings.Contains(addQuery, required) {
			t.Fatalf("generic member add must enforce %q: %s", required, addQuery)
		}
	}

	removeQuery := strings.ToLower(removeMemberQuery)
	for _, required := range []string{
		"tm.team_id = :team_id",
		"tm.user_id = :user_id",
		"t.workspace_id = :workspace_id",
	} {
		if !strings.Contains(removeQuery, required) {
			t.Fatalf("generic member removal must enforce %q: %s", required, removeQuery)
		}
	}

	leaveQuery := strings.ToLower(leaveTeamQuery)
	for _, required := range []string{
		"membership.team_id = :team_id",
		"membership.user_id = :actor_id",
		"team.workspace_id = :workspace_id",
		"workspace_membership.user_id = :actor_id",
		"actor.is_active = true",
	} {
		if !strings.Contains(leaveQuery, required) {
			t.Fatalf("self-leave must enforce %q: %s", required, leaveQuery)
		}
	}
	if strings.Contains(leaveQuery, ":user_id") {
		t.Fatal("self-leave query must not accept a caller-selected user ID")
	}
}

func TestValidateAddMemberOutcome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		outcome addMemberOutcome
		wantErr error
	}{
		{name: "eligible member added", outcome: addMemberOutcome{Eligible: true, Added: true}},
		{name: "cross-workspace team or member", outcome: addMemberOutcome{}, wantErr: teams.ErrTeamNotFound},
		{name: "duplicate membership", outcome: addMemberOutcome{Eligible: true}, wantErr: teams.ErrTeamMemberExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateAddMember(tt.outcome); !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateAddMember() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
