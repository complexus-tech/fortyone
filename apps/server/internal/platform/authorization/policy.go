package authorization

import (
	"errors"
	"fmt"
	"slices"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type DecisionCode string

const (
	DecisionAllowed               DecisionCode = "allowed"
	DecisionInvalidActor          DecisionCode = "invalid_actor"
	DecisionPrincipalKindDenied   DecisionCode = "principal_kind_denied"
	DecisionWorkspaceMismatch     DecisionCode = "workspace_mismatch"
	DecisionCredentialScopeDenied DecisionCode = "credential_scope_denied" // #nosec G101 -- Audit decision code, not a credential.
	DecisionTeamRestrictionDenied DecisionCode = "team_restriction_denied"
	DecisionWorkspaceRoleDenied   DecisionCode = "workspace_role_denied"
)

var (
	ErrPrincipalKindDenied   = errors.New("principal kind is not allowed")
	ErrWorkspaceMismatch     = errors.New("actor workspace does not match resource workspace")
	ErrCredentialScopeDenied = errors.New("credential scope is required")
	ErrTeamRestrictionDenied = errors.New("credential does not grant access to the team")
)

// Decision is safe to record in audit metadata. It deliberately contains no
// credentials, provider payloads, or other attacker-controlled detail.
type Decision struct {
	Allowed bool
	Code    DecisionCode
	Cause   error
}

func (d Decision) Err() error {
	if d.Allowed {
		return nil
	}
	return &PolicyError{Code: d.Code, Cause: d.Cause}
}

type PolicyError struct {
	Code  DecisionCode
	Cause error
}

func (e *PolicyError) Error() string {
	return fmt.Sprintf("authorization denied (%s): %v", e.Code, e.Cause)
}

func (e *PolicyError) Unwrap() error {
	return e.Cause
}

// WorkspacePolicyInput combines credential narrowing with current product
// authorization state. AllowedPrincipalKinds must be explicit so a new actor
// type cannot acquire access merely by being added to authentication.
type WorkspacePolicyInput struct {
	Actor                 platformauth.Actor
	WorkspaceID           uuid.UUID
	WorkspaceRole         WorkspaceRole
	MinimumWorkspaceRole  WorkspaceRole
	RequiredScopes        []platformauth.Scope
	TeamID                uuid.UUID
	AllowedPrincipalKinds []platformauth.PrincipalKind
}

func EvaluateWorkspace(input WorkspacePolicyInput) Decision {
	if err := input.Actor.Validate(); err != nil {
		return Decision{Code: DecisionInvalidActor, Cause: err}
	}
	if len(input.AllowedPrincipalKinds) == 0 || !slices.Contains(input.AllowedPrincipalKinds, input.Actor.Kind) {
		return Decision{Code: DecisionPrincipalKindDenied, Cause: ErrPrincipalKindDenied}
	}
	if input.WorkspaceID == uuid.Nil || input.Actor.WorkspaceID == uuid.Nil || input.Actor.WorkspaceID != input.WorkspaceID {
		return Decision{Code: DecisionWorkspaceMismatch, Cause: ErrWorkspaceMismatch}
	}
	if !input.Actor.Scopes.ContainsAll(input.RequiredScopes...) {
		return Decision{Code: DecisionCredentialScopeDenied, Cause: ErrCredentialScopeDenied}
	}
	if input.TeamID != uuid.Nil && !input.Actor.TeamAccess.Allows(input.TeamID) {
		return Decision{Code: DecisionTeamRestrictionDenied, Cause: ErrTeamRestrictionDenied}
	}
	if err := RequireMinimumWorkspaceRole(input.WorkspaceRole, input.MinimumWorkspaceRole); err != nil {
		return Decision{Code: DecisionWorkspaceRoleDenied, Cause: err}
	}
	return Decision{Allowed: true, Code: DecisionAllowed}
}

func AuthorizeWorkspace(input WorkspacePolicyInput) error {
	return EvaluateWorkspace(input).Err()
}
