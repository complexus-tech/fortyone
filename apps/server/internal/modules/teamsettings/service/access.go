package teamsettings

import (
	"context"
	"errors"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
)

var readPrincipalKinds = []platformauth.PrincipalKind{
	platformauth.PrincipalHumanUser,
	platformauth.PrincipalPersonalToken,
	platformauth.PrincipalOAuthUser,
}

var writePrincipalKinds = []platformauth.PrincipalKind{
	platformauth.PrincipalHumanUser,
}

func (s *Service) authorizeRead(ctx context.Context, access Access) error {
	if err := authorization.AuthorizeWorkspace(authorization.WorkspacePolicyInput{
		Actor:                 access.Actor,
		WorkspaceID:           access.WorkspaceID,
		WorkspaceRole:         access.WorkspaceRole,
		MinimumWorkspaceRole:  authorization.WorkspaceRoleMember,
		RequiredScopes:        []platformauth.Scope{platformauth.ScopeTeamsRead},
		TeamID:                access.TeamID,
		AllowedPrincipalKinds: readPrincipalKinds,
	}); err != nil {
		return err
	}

	if access.WorkspaceRole == authorization.WorkspaceRoleAdmin {
		return nil
	}
	actorID, err := access.Actor.UserID()
	if err != nil {
		return err
	}
	isMember, err := s.repo.IsActiveTeamMember(ctx, access.TeamID, access.WorkspaceID, actorID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrTeamMembershipRequired
	}
	return nil
}

func (s *Service) authorizeWrite(access Access) error {
	err := authorization.AuthorizeWorkspace(authorization.WorkspacePolicyInput{
		Actor:                 access.Actor,
		WorkspaceID:           access.WorkspaceID,
		WorkspaceRole:         access.WorkspaceRole,
		MinimumWorkspaceRole:  authorization.WorkspaceRoleAdmin,
		RequiredScopes:        []platformauth.Scope{platformauth.ScopeTeamsRead},
		TeamID:                access.TeamID,
		AllowedPrincipalKinds: writePrincipalKinds,
	})
	if errors.Is(err, authorization.ErrInsufficientWorkspaceRole) {
		return authorization.ErrWorkspaceAdminRequired
	}
	return err
}
