package teamsrepository

import (
	"errors"
	"strings"
	"testing"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
)

func TestPublicTeamJoinIsBoundToAuthenticatedWorkspaceActor(t *testing.T) {
	t.Parallel()

	query := strings.ToLower(joinPublicTeamQuery)
	for _, required := range []string{
		"team.team_id = :team_id",
		"team.workspace_id = :workspace_id",
		"team.is_private = false",
		"membership.workspace_id = team.workspace_id",
		"membership.user_id = :actor_id",
		"actor.user_id = membership.user_id",
		"actor.is_active = true",
		"select team_id, user_id",
		"from eligible_team",
	} {
		if !strings.Contains(query, required) {
			t.Fatalf("public team join must enforce %q: %s", required, query)
		}
	}
}

func TestPublicTeamJoinRejectsIneligibleAndDuplicateActors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		outcome publicTeamJoinOutcome
		wantErr error
	}{
		{name: "eligible actor joined", outcome: publicTeamJoinOutcome{Eligible: true, Joined: true}},
		{name: "private cross-workspace or missing team", outcome: publicTeamJoinOutcome{}, wantErr: teams.ErrTeamNotFound},
		{name: "already a member", outcome: publicTeamJoinOutcome{Eligible: true}, wantErr: teams.ErrTeamMemberExists},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validatePublicTeamJoin(tt.outcome)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("join error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
