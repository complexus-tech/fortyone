package invitations

import (
	invitationsdomain "github.com/complexus-tech/projects-api/internal/modules/invitations/domain"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
)

var (
	ErrInvitationNotFound     = invitationsdomain.ErrInvitationNotFound
	ErrInvitationExpired      = invitationsdomain.ErrInvitationExpired
	ErrInvitationUsed         = invitationsdomain.ErrInvitationUsed
	ErrInvitationRevoked      = invitationsdomain.ErrInvitationRevoked
	ErrInvalidToken           = invitationsdomain.ErrInvalidToken
	ErrDuplicateInvitation    = invitationsdomain.ErrDuplicateInvitation
	ErrInvalidInvitee         = invitationsdomain.ErrInvalidInvitee
	ErrAlreadyWorkspaceMember = invitationsdomain.ErrAlreadyWorkspaceMember
	ErrInvalidInvitationRole  = invitationsdomain.ErrInvalidInvitationRole
	ErrInvalidInvitationEmail = invitationsdomain.ErrInvalidInvitationEmail
	ErrInvalidInvitationTeam  = invitationsdomain.ErrInvalidInvitationTeam
	ErrTooManyInvitations     = invitationsdomain.ErrTooManyInvitations
	ErrOutboxClaimLost        = invitationsdomain.ErrOutboxClaimLost
)

const (
	InvitationRoleGuest  = string(authorization.WorkspaceRoleGuest)
	InvitationRoleMember = string(authorization.WorkspaceRoleMember)
	InvitationRoleAdmin  = string(authorization.WorkspaceRoleAdmin)
)

func ValidateInvitationRole(role string) error {
	switch role {
	case InvitationRoleGuest, InvitationRoleMember, InvitationRoleAdmin:
		return nil
	default:
		return ErrInvalidInvitationRole
	}
}

type InvitationRequest = invitationsdomain.Request
type StoredInvitationToken = invitationsdomain.StoredToken
type InvitationTokenLookup = invitationsdomain.TokenLookup
type NewWorkspaceInvitation = invitationsdomain.NewWorkspaceInvitation
type InvitationEmailOutboxPayload = invitationsdomain.EmailOutboxPayload
type InvitationEmailDelivery = invitationsdomain.EmailDelivery
type AcceptInvitationCommand = invitationsdomain.AcceptCommand
type CoreInvitationOutboxEvent = invitationsdomain.OutboxEvent
type CoreWorkspaceInvitation = invitationsdomain.WorkspaceInvitation
