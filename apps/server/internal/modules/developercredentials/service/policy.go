package developercredentials

import (
	"errors"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
)

func authorizePersonalToken(access developercredentialsdomain.Access) error {
	err := authorization.AuthorizeWorkspace(authorization.WorkspacePolicyInput{
		Actor:                 access.Actor,
		WorkspaceID:           access.WorkspaceID,
		WorkspaceRole:         access.WorkspaceRole,
		MinimumWorkspaceRole:  authorization.WorkspaceRoleGuest,
		AllowedPrincipalKinds: []platformauth.PrincipalKind{platformauth.PrincipalHumanUser},
	})
	if err != nil {
		return errors.Join(developercredentialsdomain.ErrAccessDenied, err)
	}
	return nil
}

func authorizeHumanPrincipalResolution(access developercredentialsdomain.Access) error {
	err := authorization.AuthorizeWorkspace(authorization.WorkspacePolicyInput{
		Actor:                 access.Actor,
		WorkspaceID:           access.WorkspaceID,
		WorkspaceRole:         access.WorkspaceRole,
		MinimumWorkspaceRole:  authorization.WorkspaceRoleGuest,
		AllowedPrincipalKinds: []platformauth.PrincipalKind{platformauth.PrincipalHumanUser, platformauth.PrincipalPersonalToken},
	})
	if err != nil {
		return errors.Join(developercredentialsdomain.ErrAccessDenied, err)
	}
	return nil
}

func authorizeServiceAccountManagement(access developercredentialsdomain.Access) error {
	err := authorization.AuthorizeWorkspace(authorization.WorkspacePolicyInput{
		Actor:                 access.Actor,
		WorkspaceID:           access.WorkspaceID,
		WorkspaceRole:         access.WorkspaceRole,
		MinimumWorkspaceRole:  authorization.WorkspaceRoleAdmin,
		RequiredScopes:        []platformauth.Scope{platformauth.ScopeServiceAccountsManage},
		AllowedPrincipalKinds: []platformauth.PrincipalKind{platformauth.PrincipalHumanUser},
	})
	if err != nil {
		return errors.Join(developercredentialsdomain.ErrAccessDenied, err)
	}
	return nil
}
