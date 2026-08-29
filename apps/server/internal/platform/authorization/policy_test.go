package authorization

import (
	"errors"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func TestWorkspacePolicyMatrix(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	scopes := platformauth.MustScopeSet(platformauth.ScopeStoriesRead)
	teamAccess, err := platformauth.RestrictedTeamAccess(teamID)
	if err != nil {
		t.Fatalf("restricted team access: %v", err)
	}
	actor, err := platformauth.NewActor(
		uuid.New(),
		platformauth.PrincipalServiceAccount,
		uuid.New(),
		scopes,
		teamAccess,
	)
	if err != nil {
		t.Fatalf("new actor: %v", err)
	}
	actor, err = actor.WithWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("bind workspace: %v", err)
	}

	base := WorkspacePolicyInput{
		Actor:                 actor,
		WorkspaceID:           workspaceID,
		WorkspaceRole:         WorkspaceRoleMember,
		MinimumWorkspaceRole:  WorkspaceRoleMember,
		RequiredScopes:        []platformauth.Scope{platformauth.ScopeStoriesRead},
		TeamID:                teamID,
		AllowedPrincipalKinds: []platformauth.PrincipalKind{platformauth.PrincipalServiceAccount},
	}
	tests := []struct {
		name  string
		input WorkspacePolicyInput
		code  DecisionCode
		cause error
	}{
		{name: "allowed", input: base, code: DecisionAllowed},
		{name: "kind", input: mutatePolicyInput(base, func(input *WorkspacePolicyInput) {
			input.AllowedPrincipalKinds = []platformauth.PrincipalKind{platformauth.PrincipalHumanUser}
		}), code: DecisionPrincipalKindDenied, cause: ErrPrincipalKindDenied},
		{name: "workspace", input: mutatePolicyInput(base, func(input *WorkspacePolicyInput) {
			input.WorkspaceID = uuid.New()
		}), code: DecisionWorkspaceMismatch, cause: ErrWorkspaceMismatch},
		{name: "scope", input: mutatePolicyInput(base, func(input *WorkspacePolicyInput) {
			input.RequiredScopes = []platformauth.Scope{platformauth.ScopeStoriesWrite}
		}), code: DecisionCredentialScopeDenied, cause: ErrCredentialScopeDenied},
		{name: "team", input: mutatePolicyInput(base, func(input *WorkspacePolicyInput) {
			input.TeamID = uuid.New()
		}), code: DecisionTeamRestrictionDenied, cause: ErrTeamRestrictionDenied},
		{name: "role", input: mutatePolicyInput(base, func(input *WorkspacePolicyInput) {
			input.MinimumWorkspaceRole = WorkspaceRoleAdmin
		}), code: DecisionWorkspaceRoleDenied, cause: ErrInsufficientWorkspaceRole},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			decision := EvaluateWorkspace(test.input)
			if decision.Code != test.code {
				t.Fatalf("decision code = %q, want %q", decision.Code, test.code)
			}
			if test.code == DecisionAllowed {
				if !decision.Allowed || decision.Err() != nil {
					t.Fatalf("allowed decision = %+v", decision)
				}
				return
			}
			if decision.Allowed || !errors.Is(decision.Err(), test.cause) {
				t.Fatalf("denied decision = %+v, want cause %v", decision, test.cause)
			}
		})
	}
}

func mutatePolicyInput(input WorkspacePolicyInput, mutate func(*WorkspacePolicyInput)) WorkspacePolicyInput {
	input.RequiredScopes = append([]platformauth.Scope(nil), input.RequiredScopes...)
	input.AllowedPrincipalKinds = append([]platformauth.PrincipalKind(nil), input.AllowedPrincipalKinds...)
	mutate(&input)
	return input
}
