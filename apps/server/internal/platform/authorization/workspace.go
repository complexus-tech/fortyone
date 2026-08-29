// Package authorization contains policy decisions shared across delivery layers.
package authorization

import "errors"

// WorkspaceRole is a user's role within a workspace.
type WorkspaceRole string

const (
	WorkspaceRoleGuest  WorkspaceRole = "guest"
	WorkspaceRoleMember WorkspaceRole = "member"
	WorkspaceRoleAdmin  WorkspaceRole = "admin"
)

var (
	// ErrInvalidWorkspaceRole indicates corrupt or unsupported membership data.
	ErrInvalidWorkspaceRole = errors.New("invalid workspace role")
	// ErrInsufficientWorkspaceRole indicates that a valid role does not meet a
	// policy's minimum level.
	ErrInsufficientWorkspaceRole = errors.New("insufficient workspace role")
	// ErrWorkspaceAdminRequired is returned when a privileged workspace operation
	// is attempted by someone who is not a current workspace administrator.
	ErrWorkspaceAdminRequired = errors.New("workspace administrator access is required")
)

var workspaceRoleLevels = map[WorkspaceRole]int{
	WorkspaceRoleGuest:  1,
	WorkspaceRoleMember: 2,
	WorkspaceRoleAdmin:  3,
}

// RequireMinimumWorkspaceRole applies the workspace role hierarchy and fails
// closed for roles that are not part of the public workspace membership model.
func RequireMinimumWorkspaceRole(actual, minimum WorkspaceRole) error {
	minimumLevel, ok := workspaceRoleLevels[minimum]
	if !ok {
		return ErrInvalidWorkspaceRole
	}

	actualLevel, ok := workspaceRoleLevels[actual]
	if !ok {
		return ErrInvalidWorkspaceRole
	}
	if actualLevel < minimumLevel {
		return ErrInsufficientWorkspaceRole
	}

	return nil
}

// RequireWorkspaceAdmin requires the exact minimum capability used by
// workspace-administration operations.
func RequireWorkspaceAdmin(actual WorkspaceRole) error {
	err := RequireMinimumWorkspaceRole(actual, WorkspaceRoleAdmin)
	if errors.Is(err, ErrInsufficientWorkspaceRole) {
		return ErrWorkspaceAdminRequired
	}
	return err
}

// ValidateWorkspaceRole reports whether a value belongs to the public
// workspace membership role model.
func ValidateWorkspaceRole(role WorkspaceRole) error {
	if _, ok := workspaceRoleLevels[role]; !ok {
		return ErrInvalidWorkspaceRole
	}
	return nil
}
