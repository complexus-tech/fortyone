package teamsettings

import (
	"context"
	"errors"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

func TestAuthorizeReadPolicyMatrix(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	base := Access{
		Actor:         testBoundActor(t, platformauth.PrincipalHumanUser, workspaceID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), platformauth.UnrestrictedTeamAccess()),
		WorkspaceRole: authorization.WorkspaceRoleMember,
		WorkspaceID:   workspaceID,
		TeamID:        teamID,
	}
	restrictedToOtherTeam, err := platformauth.RestrictedTeamAccess(uuid.New())
	if err != nil {
		t.Fatalf("restrict team access: %v", err)
	}

	tests := []struct {
		name         string
		access       Access
		isTeamMember bool
		wantErr      error
	}{
		{name: "active member", access: base, isTeamMember: true},
		{name: "workspace admin does not require team membership", access: withRole(base, authorization.WorkspaceRoleAdmin)},
		{name: "inactive team membership", access: base, wantErr: ErrTeamMembershipRequired},
		{name: "guest role", access: withRole(base, authorization.WorkspaceRoleGuest), isTeamMember: true, wantErr: authorization.ErrInsufficientWorkspaceRole},
		{name: "second tenant", access: withWorkspaceResource(base, uuid.New()), isTeamMember: true, wantErr: authorization.ErrWorkspaceMismatch},
		{name: "missing scope", access: withActor(base, testBoundActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.MustScopeSet(platformauth.ScopeStoriesRead), platformauth.UnrestrictedTeamAccess())), isTeamMember: true, wantErr: authorization.ErrCredentialScopeDenied},
		{name: "restricted to another team", access: withActor(base, testBoundActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), restrictedToOtherTeam)), isTeamMember: true, wantErr: authorization.ErrTeamRestrictionDenied},
		{name: "service principal", access: withActor(base, testBoundActor(t, platformauth.PrincipalServiceAccount, workspaceID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), platformauth.UnrestrictedTeamAccess())), isTeamMember: true, wantErr: authorization.ErrPrincipalKindDenied},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo := &repositoryStub{isTeamMember: test.isTeamMember}
			service := New(testLogger(), repo, nil)
			_, err := service.GetSprintSettings(context.Background(), test.access)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("GetSprintSettings() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestAuthorizeWritePolicyMatrix(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	validUpdate := CoreUpdateTeamEstimationSettings{
		Scheme: PatchField[string]{Value: "points", Present: true},
	}
	base := Access{
		Actor:         testBoundActor(t, platformauth.PrincipalHumanUser, workspaceID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), platformauth.UnrestrictedTeamAccess()),
		WorkspaceRole: authorization.WorkspaceRoleAdmin,
		WorkspaceID:   workspaceID,
		TeamID:        teamID,
	}
	restrictedToOtherTeam, err := platformauth.RestrictedTeamAccess(uuid.New())
	if err != nil {
		t.Fatalf("restrict team access: %v", err)
	}

	tests := []struct {
		name    string
		access  Access
		wantErr error
	}{
		{name: "human administrator", access: base},
		{name: "member", access: withRole(base, authorization.WorkspaceRoleMember), wantErr: authorization.ErrWorkspaceAdminRequired},
		{name: "second tenant", access: withWorkspaceResource(base, uuid.New()), wantErr: authorization.ErrWorkspaceMismatch},
		{name: "personal token", access: withActor(base, testBoundActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), platformauth.UnrestrictedTeamAccess())), wantErr: authorization.ErrPrincipalKindDenied},
		{name: "oauth user", access: withActor(base, testBoundActor(t, platformauth.PrincipalOAuthUser, workspaceID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), platformauth.UnrestrictedTeamAccess())), wantErr: authorization.ErrPrincipalKindDenied},
		{name: "missing scope", access: withActor(base, testBoundActor(t, platformauth.PrincipalHumanUser, workspaceID, platformauth.MustScopeSet(platformauth.ScopeStoriesRead), platformauth.UnrestrictedTeamAccess())), wantErr: authorization.ErrCredentialScopeDenied},
		{name: "restricted to another team", access: withActor(base, testBoundActor(t, platformauth.PrincipalHumanUser, workspaceID, platformauth.MustScopeSet(platformauth.ScopeTeamsRead), restrictedToOtherTeam)), wantErr: authorization.ErrTeamRestrictionDenied},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := New(testLogger(), &repositoryStub{}, nil)
			_, err := service.UpdateEstimationSettings(context.Background(), test.access, validUpdate)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("UpdateEstimationSettings() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func testBoundActor(
	t *testing.T,
	kind platformauth.PrincipalKind,
	workspaceID uuid.UUID,
	scopes platformauth.ScopeSet,
	teamAccess platformauth.TeamAccess,
) platformauth.Actor {
	t.Helper()
	actor, err := platformauth.NewActor(uuid.New(), kind, uuid.New(), scopes, teamAccess)
	if err != nil {
		t.Fatalf("new actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind actor workspace: %v", err)
	}
	return actor
}

func withActor(access Access, actor platformauth.Actor) Access {
	access.Actor = actor
	return access
}

func withRole(access Access, role authorization.WorkspaceRole) Access {
	access.WorkspaceRole = role
	return access
}

func withWorkspaceResource(access Access, workspaceID uuid.UUID) Access {
	access.WorkspaceID = workspaceID
	return access
}
